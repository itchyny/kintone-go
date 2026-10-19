package kintone

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
)

type response struct {
	status int
	body   string
}

func newSequenceServer(t *testing.T, responses ...response) (*httptest.Server, func() []request) {
	t.Helper()
	var mu sync.Mutex
	var requests []request
	ts := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		requests = append(requests, request{r.Method, r.URL.Path, r.Header, string(b)})
		i := len(requests) - 1
		mu.Unlock()
		if i >= len(responses) {
			t.Errorf("unexpected request: %s %s %s", r.Method, r.URL.Path, b)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responses[i].status)
		_, _ = io.WriteString(w, responses[i].body)
	}))
	return ts, func() []request {
		mu.Lock()
		defer mu.Unlock()
		return requests
	}
}

func TestClientCreateCursor(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{"id": "cursor1", "totalCount": "1234"}`)
	c := newTestClient(t, ts)
	cursor, err := c.CreateCursor(t.Context(), CreateCursorRequest{
		App:    1,
		Fields: []string{"$id", "title"},
		Query:  `title = "Hello" order by $id asc`,
		Size:   200,
	})
	if err != nil {
		t.Fatalf("CreateCursor returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPost, "/k/v1/records/cursor.json", `{
		"app": 1,
		"fields": ["$id", "title"],
		"query": "title = \"Hello\" order by $id asc",
		"size": 200
	}`)
	if want := (&Cursor{ID: "cursor1", TotalCount: 1234}); !reflect.DeepEqual(cursor, want) {
		t.Errorf("cursor: got %+v, want %+v", cursor, want)
	}
}

func TestClientGetRecordsByCursor(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{
		"records": [{"$id": {"type": "__ID__", "value": "1"}}],
		"next": true
	}`)
	c := newTestClient(t, ts)
	records, next, err := c.GetRecordsByCursor(t.Context(), "cursor1")
	if err != nil {
		t.Fatalf("GetRecordsByCursor returned error: %v", err)
	}
	assertRequest(t, got, http.MethodGet, "/k/v1/records/cursor.json", `{"id": "cursor1"}`)
	if want := []Record{{"$id": {Type: FieldTypeID, Value: "1"}}}; !reflect.DeepEqual(records, want) {
		t.Errorf("records: got %#v, want %#v", records, want)
	}
	if !next {
		t.Error("next: got false, want true")
	}
}

func TestClientDeleteCursor(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{}`)
	c := newTestClient(t, ts)
	if err := c.DeleteCursor(t.Context(), "cursor1"); err != nil {
		t.Fatalf("DeleteCursor returned error: %v", err)
	}
	assertRequest(t, got, http.MethodDelete, "/k/v1/records/cursor.json", `{"id": "cursor1"}`)
}

func recordIDs(t *testing.T, seq func(func(Record, error) bool)) ([]string, error) {
	t.Helper()
	var ids []string
	for record, err := range seq {
		if err != nil {
			return ids, err
		}
		ids = append(ids, record["$id"].Value.(string))
	}
	return ids, nil
}

func TestClientAllRecords(t *testing.T) {
	ts, requests := newSequenceServer(t,
		response{http.StatusOK, `{"id": "cursor1", "totalCount": "3"}`},
		response{http.StatusOK, `{"records": [
			{"$id": {"type": "__ID__", "value": "1"}},
			{"$id": {"type": "__ID__", "value": "2"}}
		], "next": true}`},
		response{http.StatusOK, `{"records": [], "next": true}`},
		response{http.StatusOK, `{"records": [
			{"$id": {"type": "__ID__", "value": "3"}}
		], "next": false}`},
	)
	c := newTestClient(t, ts)
	ids, err := recordIDs(t, c.AllRecords(t.Context(), CreateCursorRequest{App: 1, Query: "order by $id asc"}))
	if err != nil {
		t.Fatalf("AllRecords returned error: %v", err)
	}
	if want := []string{"1", "2", "3"}; !reflect.DeepEqual(ids, want) {
		t.Errorf("record IDs: got %v, want %v", ids, want)
	}
	reqs := requests()
	if len(reqs) != 4 {
		t.Fatalf("number of requests: got %d, want %d", len(reqs), 4)
	}
	assertRequest(t, &reqs[0], http.MethodPost, "/k/v1/records/cursor.json",
		`{"app": 1, "query": "order by $id asc", "size": 500}`)
	for _, req := range reqs[1:] {
		assertRequest(t, &req, http.MethodGet, "/k/v1/records/cursor.json", `{"id": "cursor1"}`)
	}
}

func TestClientAllRecordsSize(t *testing.T) {
	ts, requests := newSequenceServer(t,
		response{http.StatusOK, `{"id": "cursor1", "totalCount": "0"}`},
		response{http.StatusOK, `{"records": [], "next": false}`},
	)
	c := newTestClient(t, ts)
	if _, err := recordIDs(t, c.AllRecords(t.Context(), CreateCursorRequest{App: 1, Size: 100})); err != nil {
		t.Fatalf("AllRecords returned error: %v", err)
	}
	reqs := requests()
	if len(reqs) != 2 {
		t.Fatalf("number of requests: got %d, want %d", len(reqs), 2)
	}
	assertRequest(t, &reqs[0], http.MethodPost, "/k/v1/records/cursor.json", `{"app": 1, "size": 100}`)
}

func TestClientAllRecordsBreak(t *testing.T) {
	ts, requests := newSequenceServer(t,
		response{http.StatusOK, `{"id": "cursor1", "totalCount": "3"}`},
		response{http.StatusOK, `{"records": [
			{"$id": {"type": "__ID__", "value": "1"}},
			{"$id": {"type": "__ID__", "value": "2"}}
		], "next": true}`},
		response{http.StatusOK, `{}`},
	)
	c := newTestClient(t, ts)
	for record, err := range c.AllRecords(t.Context(), CreateCursorRequest{App: 1}) {
		if err != nil {
			t.Fatalf("AllRecords returned error: %v", err)
		}
		if record["$id"].Value == "1" {
			break
		}
	}
	reqs := requests()
	if len(reqs) != 3 {
		t.Fatalf("number of requests: got %d, want %d", len(reqs), 3)
	}
	assertRequest(t, &reqs[2], http.MethodDelete, "/k/v1/records/cursor.json", `{"id": "cursor1"}`)
}

func TestClientAllRecordsBreakOnLastRecords(t *testing.T) {
	ts, requests := newSequenceServer(t,
		response{http.StatusOK, `{"id": "cursor1", "totalCount": "2"}`},
		response{http.StatusOK, `{"records": [
			{"$id": {"type": "__ID__", "value": "1"}},
			{"$id": {"type": "__ID__", "value": "2"}}
		], "next": false}`},
	)
	c := newTestClient(t, ts)
	for _, err := range c.AllRecords(t.Context(), CreateCursorRequest{App: 1}) {
		if err != nil {
			t.Fatalf("AllRecords returned error: %v", err)
		}
		break
	}
	if reqs := requests(); len(reqs) != 2 {
		t.Errorf("number of requests: got %d, want %d", len(reqs), 2)
	}
}

func TestClientAllRecordsCreateCursorError(t *testing.T) {
	ts, requests := newSequenceServer(t,
		response{http.StatusBadRequest, `{"code": "TEST_ERROR", "id": "abc123", "message": "Test error."}`},
	)
	c := newTestClient(t, ts)
	ids, err := recordIDs(t, c.AllRecords(t.Context(), CreateCursorRequest{App: 1}))
	if !IsErrorCode(err, "TEST_ERROR") {
		t.Errorf("AllRecords should return TEST_ERROR error but got: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("record IDs: got %v, want none", ids)
	}
	if reqs := requests(); len(reqs) != 1 {
		t.Errorf("number of requests: got %d, want %d", len(reqs), 1)
	}
}

func TestClientAllRecordsGetRecordsError(t *testing.T) {
	ts, requests := newSequenceServer(t,
		response{http.StatusOK, `{"id": "cursor1", "totalCount": "3"}`},
		response{http.StatusOK, `{"records": [
			{"$id": {"type": "__ID__", "value": "1"}}
		], "next": true}`},
		response{http.StatusInternalServerError, `{"code": "TEST_ERROR", "id": "abc123", "message": "Test error."}`},
		response{http.StatusOK, `{}`},
	)
	c := newTestClient(t, ts)
	ids, err := recordIDs(t, c.AllRecords(t.Context(), CreateCursorRequest{App: 1}))
	if !IsErrorCode(err, "TEST_ERROR") {
		t.Errorf("AllRecords should return TEST_ERROR error but got: %v", err)
	}
	if want := []string{"1"}; !reflect.DeepEqual(ids, want) {
		t.Errorf("record IDs: got %v, want %v", ids, want)
	}
	reqs := requests()
	if len(reqs) != 4 {
		t.Fatalf("number of requests: got %d, want %d", len(reqs), 4)
	}
	assertRequest(t, &reqs[3], http.MethodDelete, "/k/v1/records/cursor.json", `{"id": "cursor1"}`)
}

func TestClientAllRecordsContextCanceled(t *testing.T) {
	ts, requests := newSequenceServer(t,
		response{http.StatusOK, `{"id": "cursor1", "totalCount": "3"}`},
		response{http.StatusOK, `{"records": [
			{"$id": {"type": "__ID__", "value": "1"}}
		], "next": true}`},
		response{http.StatusOK, `{}`},
	)
	c := newTestClient(t, ts)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	var err error
	for _, err = range c.AllRecords(ctx, CreateCursorRequest{App: 1}) {
		if err != nil {
			break
		}
		cancel()
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("AllRecords should return context.Canceled but got: %v", err)
	}
	reqs := requests()
	if len(reqs) != 3 {
		t.Fatalf("number of requests: got %d, want %d", len(reqs), 3)
	}
	assertRequest(t, &reqs[2], http.MethodDelete, "/k/v1/records/cursor.json", `{"id": "cursor1"}`)
}

func newCursorWarningServer(t *testing.T, warning string) *httptest.Server {
	t.Helper()
	var calls int
	ts := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		calls++
		switch calls {
		case 1:
			if warning != "" {
				w.Header().Set("X-Cybozu-Warning", warning)
			}
			_, _ = io.WriteString(w, `{"id": "cursor1", "totalCount": "2"}`)
		case 2:
			_, _ = io.WriteString(w, `{"records": [
				{"$id": {"type": "__ID__", "value": "1"}},
				{"$id": {"type": "__ID__", "value": "2"}}
			], "next": false}`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	return ts
}

func TestClientCreateCursorWarning(t *testing.T) {
	ts := newCursorWarningServer(t, "Filter aborted because of too many search results.")
	c := newTestClient(t, ts)
	cursor, err := c.CreateCursor(t.Context(), CreateCursorRequest{App: 1, Query: `title like "a"`})
	if err != nil {
		t.Fatalf("CreateCursor returned error: %v", err)
	}
	if want := "Filter aborted because of too many search results."; cursor.Warning != want {
		t.Errorf("warning: got %q, want %q", cursor.Warning, want)
	}
}

func TestClientAllRecordsWarning(t *testing.T) {
	ts := newCursorWarningServer(t, "Filter aborted because of too many search results.")
	c := newTestClient(t, ts)
	var ids []string
	var errs []error
	for record, err := range c.AllRecords(t.Context(), CreateCursorRequest{App: 1, Query: `title like "a"`}) {
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ids = append(ids, record["$id"].Value.(string))
	}
	if want := []string{"1", "2"}; !reflect.DeepEqual(ids, want) {
		t.Errorf("record IDs: got %v, want %v", ids, want)
	}
	if len(errs) != 1 {
		t.Fatalf("number of errors: got %d, want %d", len(errs), 1)
	}
	if warning, ok := errors.AsType[*WarningError](errs[0]); !ok || warning.Message != "Filter aborted because of too many search results." {
		t.Errorf("AllRecords should yield *WarningError but got: %v", errs[0])
	}
	if want := "kintone: Filter aborted because of too many search results."; errs[0].Error() != want {
		t.Errorf("error message: got %q, want %q", errs[0].Error(), want)
	}
}

func TestClientAllRecordsNoWarning(t *testing.T) {
	ts := newCursorWarningServer(t, "")
	c := newTestClient(t, ts)
	ids, err := recordIDs(t, c.AllRecords(t.Context(), CreateCursorRequest{App: 1}))
	if err != nil {
		t.Fatalf("AllRecords returned error: %v", err)
	}
	if want := []string{"1", "2"}; !reflect.DeepEqual(ids, want) {
		t.Errorf("record IDs: got %v, want %v", ids, want)
	}
}

func TestClientAllRecordsWarningBreak(t *testing.T) {
	ts := newCursorWarningServer(t, "Filter aborted because of too many search results.")
	c := newTestClient(t, ts)
	for record, err := range c.AllRecords(t.Context(), CreateCursorRequest{App: 1}) {
		if err != nil {
			t.Fatalf("AllRecords should not yield the warning after break but got: %v", err)
		}
		if record["$id"].Value == "2" {
			break
		}
	}
}

func idRecords(ids ...string) string {
	var records []string
	for _, id := range ids {
		records = append(records, `{"$id": {"type": "__ID__", "value": "`+id+`"}}`)
	}
	return `{"records": [` + strings.Join(records, ", ") + `], "totalCount": null}`
}

func TestClientAllRecordsByID(t *testing.T) {
	setValue(t, &allRecordsByIDLimit, 2)
	testCases := []struct {
		name      string
		req       AllRecordsByIDRequest
		responses []response
		wantIDs   []string
		wantBody  []string
	}{
		{
			name: "condition and fields",
			req:  AllRecordsByIDRequest{App: 1, Fields: []string{"title"}, Condition: `status = "Open"`},
			responses: []response{
				{http.StatusOK, idRecords("3", "5")},
				{http.StatusOK, idRecords("8", "9")},
				{http.StatusOK, idRecords("12")},
			},
			wantIDs: []string{"3", "5", "8", "9", "12"},
			wantBody: []string{
				`{"app": 1, "fields": ["title", "$id"], "query": "(status = \"Open\") and $id > 0 order by $id asc limit 2"}`,
				`{"app": 1, "fields": ["title", "$id"], "query": "(status = \"Open\") and $id > 5 order by $id asc limit 2"}`,
				`{"app": 1, "fields": ["title", "$id"], "query": "(status = \"Open\") and $id > 9 order by $id asc limit 2"}`,
			},
		},
		{
			name: "no condition and fields with $id",
			req:  AllRecordsByIDRequest{App: 1, Fields: []string{"$id", "title"}},
			responses: []response{
				{http.StatusOK, idRecords("1", "2")},
				{http.StatusOK, idRecords()},
			},
			wantIDs: []string{"1", "2"},
			wantBody: []string{
				`{"app": 1, "fields": ["$id", "title"], "query": "$id > 0 order by $id asc limit 2"}`,
				`{"app": 1, "fields": ["$id", "title"], "query": "$id > 2 order by $id asc limit 2"}`,
			},
		},
		{
			name:      "no records",
			req:       AllRecordsByIDRequest{App: 1},
			responses: []response{{http.StatusOK, idRecords()}},
			wantBody:  []string{`{"app": 1, "query": "$id > 0 order by $id asc limit 2"}`},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, requests := newSequenceServer(t, tc.responses...)
			c := newTestClient(t, ts)
			ids, err := recordIDs(t, c.AllRecordsByID(t.Context(), tc.req))
			if err != nil {
				t.Fatalf("AllRecordsByID returned error: %v", err)
			}
			if !reflect.DeepEqual(ids, tc.wantIDs) {
				t.Errorf("record IDs: got %v, want %v", ids, tc.wantIDs)
			}
			reqs := requests()
			if len(reqs) != len(tc.wantBody) {
				t.Fatalf("number of requests: got %d, want %d", len(reqs), len(tc.wantBody))
			}
			for i, req := range reqs {
				assertRequest(t, &req, http.MethodGet, "/k/v1/records.json", tc.wantBody[i])
			}
		})
	}
}

func TestClientAllRecordsByIDFieldsNotModified(t *testing.T) {
	ts, _ := newSequenceServer(t, response{http.StatusOK, idRecords()})
	c := newTestClient(t, ts)
	fields := make([]string, 1, 2)
	fields[0] = "title"
	if _, err := recordIDs(t, c.AllRecordsByID(t.Context(), AllRecordsByIDRequest{App: 1, Fields: fields})); err != nil {
		t.Fatalf("AllRecordsByID returned error: %v", err)
	}
	if got := fields[:cap(fields)]; got[1] != "" {
		t.Errorf("fields of the request should not be modified but got %q", got)
	}
}

func TestClientAllRecordsByIDError(t *testing.T) {
	setValue(t, &allRecordsByIDLimit, 2)
	testCases := []struct {
		name      string
		responses []response
		wantIDs   []string
		check     func(error) bool
	}{
		{
			name: "error response",
			responses: []response{
				{http.StatusOK, idRecords("1", "2")},
				{http.StatusInternalServerError, `{"code": "TEST_ERROR", "id": "abc123", "message": "Test error."}`},
			},
			wantIDs: []string{"1", "2"},
			check:   func(err error) bool { return IsErrorCode(err, "TEST_ERROR") },
		},
		{
			name:      "record without $id",
			responses: []response{{http.StatusOK, `{"records": [{"title": {"type": "SINGLE_LINE_TEXT", "value": "x"}}]}`}},
			check:     func(err error) bool { return err != nil && err.Error() == "kintone: record without $id" },
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, _ := newSequenceServer(t, tc.responses...)
			c := newTestClient(t, ts)
			ids, err := recordIDs(t, c.AllRecordsByID(t.Context(), AllRecordsByIDRequest{App: 1}))
			if !tc.check(err) {
				t.Errorf("AllRecordsByID returned unexpected error: %v", err)
			}
			if !reflect.DeepEqual(ids, tc.wantIDs) {
				t.Errorf("record IDs: got %v, want %v", ids, tc.wantIDs)
			}
		})
	}
}

func TestClientAllRecordsByIDContextCanceled(t *testing.T) {
	setValue(t, &allRecordsByIDLimit, 2)
	ts, requests := newSequenceServer(t,
		response{http.StatusOK, idRecords("1", "2")},
	)
	c := newTestClient(t, ts)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	var ids []string
	var err error
	for r, e := range c.AllRecordsByID(ctx, AllRecordsByIDRequest{App: 1}) {
		if err = e; err != nil {
			break
		}
		ids = append(ids, r["$id"].Value.(string))
		cancel()
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("AllRecordsByID should return context.Canceled but got: %v", err)
	}
	if want := []string{"1", "2"}; !reflect.DeepEqual(ids, want) {
		t.Errorf("record IDs: got %v, want %v", ids, want)
	}
	if reqs := requests(); len(reqs) != 1 {
		t.Errorf("number of requests: got %d, want %d", len(reqs), 1)
	}
}

func TestClientAllRecordsByIDBreak(t *testing.T) {
	setValue(t, &allRecordsByIDLimit, 2)
	ts, requests := newSequenceServer(t, response{http.StatusOK, idRecords("1", "2")})
	c := newTestClient(t, ts)
	for record, err := range c.AllRecordsByID(t.Context(), AllRecordsByIDRequest{App: 1}) {
		if err != nil {
			t.Fatalf("AllRecordsByID returned error: %v", err)
		}
		if record["$id"].Value == "2" {
			break
		}
	}
	if reqs := requests(); len(reqs) != 1 {
		t.Errorf("number of requests: got %d, want %d", len(reqs), 1)
	}
}

func TestClientAllRecordsByIDWarning(t *testing.T) {
	setValue(t, &allRecordsByIDLimit, 2)
	var calls int
	ts := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		calls++
		if calls == 1 {
			w.Header().Set("X-Cybozu-Warning", "Filter aborted because of too many search results")
			_, _ = io.WriteString(w, idRecords("1", "2"))
			return
		}
		_, _ = io.WriteString(w, idRecords("3"))
	}))
	c := newTestClient(t, ts)
	var ids []string
	var errs []error
	for record, err := range c.AllRecordsByID(t.Context(), AllRecordsByIDRequest{App: 1, Condition: `title like "a"`}) {
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ids = append(ids, record["$id"].Value.(string))
	}
	if want := []string{"1", "2", "3"}; !reflect.DeepEqual(ids, want) {
		t.Errorf("record IDs: got %v, want %v", ids, want)
	}
	var warning *WarningError
	if len(errs) == 1 {
		warning, _ = errors.AsType[*WarningError](errs[0])
	}
	if warning == nil || warning.Message != "Filter aborted because of too many search results" {
		t.Errorf("AllRecordsByID should yield one *WarningError after the records but got: %v", errs)
	}
}
