package kintone

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func assertRequest(t *testing.T, got *request, method, path, body string) {
	t.Helper()
	wantMethod, wantOverride := method, ""
	if method == http.MethodGet {
		wantMethod, wantOverride = http.MethodPost, http.MethodGet
	}
	if got.Method != wantMethod {
		t.Errorf("method: got %q, want %q", got.Method, wantMethod)
	}
	if override := got.Header.Get("X-HTTP-Method-Override"); override != wantOverride {
		t.Errorf("X-HTTP-Method-Override: got %q, want %q", override, wantOverride)
	}
	if got.Path != path {
		t.Errorf("path: got %q, want %q", got.Path, path)
	}
	var gotBody, wantBody any
	if err := json.Unmarshal([]byte(got.Body), &gotBody); err != nil {
		t.Fatalf("request body is not JSON: %q", got.Body)
	}
	if err := json.Unmarshal([]byte(body), &wantBody); err != nil {
		t.Fatalf("expected body is not JSON: %q", body)
	}
	if !reflect.DeepEqual(gotBody, wantBody) {
		t.Errorf("body:\n got: %s\nwant: %s", got.Body, body)
	}
}

func TestClientGetRecord(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{
		"record": {
			"$id": {"type": "__ID__", "value": "2"},
			"title": {"type": "SINGLE_LINE_TEXT", "value": "Hello"}
		}
	}`)
	c := newTestClient(t, ts)
	record, err := c.GetRecord(t.Context(), 1, 2)
	if err != nil {
		t.Fatalf("GetRecord returned error: %v", err)
	}
	assertRequest(t, got, http.MethodGet, "/k/v1/record.json", `{"app": 1, "id": 2}`)
	want := Record{
		"$id":   {Type: FieldTypeID, Value: "2"},
		"title": {Type: FieldTypeSingleLineText, Value: "Hello"},
	}
	if !reflect.DeepEqual(record, want) {
		t.Errorf("record: got %#v, want %#v", record, want)
	}
}

func TestClientGetRecordGuestSpace(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{"record": {}}`)
	c := newTestClient(t, ts, WithGuestSpace(3))
	if _, err := c.GetRecord(t.Context(), 1, 2); err != nil {
		t.Fatalf("GetRecord returned error: %v", err)
	}
	assertRequest(t, got, http.MethodGet, "/k/guest/3/v1/record.json", `{"app": 1, "id": 2}`)
}

func TestClientGetRecordError(t *testing.T) {
	ts, _ := newTestServer(t, http.StatusNotFound,
		`{"code": "GAIA_RE01", "id": "abc123", "message": "The specified record is not found."}`)
	c := newTestClient(t, ts)
	record, err := c.GetRecord(t.Context(), 1, 2)
	if !IsErrorCode(err, "GAIA_RE01") {
		t.Errorf("GetRecord should return GAIA_RE01 error but got: %v", err)
	}
	if record != nil {
		t.Errorf("record: got %#v, want nil", record)
	}
}

func TestClientGetRecords(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetRecordsRequest
		response string
		wantBody string
		want     *GetRecordsResponse
	}{
		{
			name: "all parameters",
			req: GetRecordsRequest{
				App:        1,
				Fields:     []string{"$id", "title"},
				Query:      `title = "Hello" order by $id asc limit 10`,
				TotalCount: true,
			},
			response: `{
				"records": [
					{"$id": {"type": "__ID__", "value": "2"}},
					{"$id": {"type": "__ID__", "value": "1"}}
				],
				"totalCount": "25"
			}`,
			wantBody: `{
				"app": 1,
				"fields": ["$id", "title"],
				"query": "title = \"Hello\" order by $id asc limit 10",
				"totalCount": true
			}`,
			want: &GetRecordsResponse{
				Records: []Record{
					{"$id": {Type: FieldTypeID, Value: "2"}},
					{"$id": {Type: FieldTypeID, Value: "1"}},
				},
				TotalCount: 25,
			},
		},
		{
			name:     "only app",
			req:      GetRecordsRequest{App: 1},
			response: `{"records": [], "totalCount": null}`,
			wantBody: `{"app": 1}`,
			want:     &GetRecordsResponse{Records: []Record{}},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, tc.response)
			c := newTestClient(t, ts)
			res, err := c.GetRecords(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetRecords returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, "/k/v1/records.json", tc.wantBody)
			if !reflect.DeepEqual(res, tc.want) {
				t.Errorf("response: got %#v, want %#v", res, tc.want)
			}
		})
	}
}

func TestClientAddRecord(t *testing.T) {
	testCases := []struct {
		name     string
		record   Record
		wantBody string
	}{
		{
			name:     "record",
			record:   Record{"title": {Value: "Hello"}},
			wantBody: `{"app": 1, "record": {"title": {"value": "Hello"}}}`,
		},
		{
			name:     "default values",
			wantBody: `{"app": 1}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{"id": "100", "revision": "1"}`)
			c := newTestClient(t, ts)
			res, err := c.AddRecord(t.Context(), 1, tc.record)
			if err != nil {
				t.Fatalf("AddRecord returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPost, "/k/v1/record.json", tc.wantBody)
			if want := (RecordRevision{ID: 100, Revision: 1}); res != want {
				t.Errorf("response: got %+v, want %+v", res, want)
			}
		})
	}
}

func TestClientAddRecords(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{"ids": ["100", "101"], "revisions": ["1", "1"]}`)
	c := newTestClient(t, ts)
	res, err := c.AddRecords(t.Context(), 1, []Record{
		{"title": {Value: "Hello"}},
		{"title": {Value: "World"}},
	})
	if err != nil {
		t.Fatalf("AddRecords returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPost, "/k/v1/records.json", `{
		"app": 1,
		"records": [
			{"title": {"value": "Hello"}},
			{"title": {"value": "World"}}
		]
	}`)
	want := []RecordRevision{{ID: 100, Revision: 1}, {ID: 101, Revision: 1}}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("response: got %+v, want %+v", res, want)
	}
}

func TestClientAddRecordsDecodeError(t *testing.T) {
	testCases := []struct {
		name     string
		response string
	}{
		{"invalid ID", `{"ids": ["x"], "revisions": ["1"]}`},
		{"invalid revision", `{"ids": ["100"], "revisions": ["x"]}`},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, _ := newTestServer(t, http.StatusOK, tc.response)
			c := newTestClient(t, ts)
			if _, err := c.AddRecords(t.Context(), 1, []Record{{}}); err == nil {
				t.Error("AddRecords should return error")
			}
		})
	}
}

func TestClientUpdateRecord(t *testing.T) {
	testCases := []struct {
		name     string
		update   RecordUpdate
		wantBody string
	}{
		{
			name: "by ID with revision",
			update: RecordUpdate{
				ID:       2,
				Record:   Record{"title": {Value: "Hello"}},
				Revision: 4,
			},
			wantBody: `{"app": 1, "id": 2, "record": {"title": {"value": "Hello"}}, "revision": 4}`,
		},
		{
			name: "by update key",
			update: RecordUpdate{
				UpdateKey: &UpdateKey{Field: "code", Value: "A001"},
				Record:    Record{"title": {Value: "Hello"}},
			},
			wantBody: `{
				"app": 1,
				"updateKey": {"field": "code", "value": "A001"},
				"record": {"title": {"value": "Hello"}}
			}`,
		},
		{
			name: "table rows",
			update: RecordUpdate{
				ID: 2,
				Record: Record{"table": {Value: []SubtableRow{
					{ID: "10", Value: Record{"item": {Value: "apple"}}},
					{Value: Record{"item": {Value: "banana"}}},
				}}},
			},
			wantBody: `{
				"app": 1,
				"id": 2,
				"record": {"table": {"value": [
					{"id": "10", "value": {"item": {"value": "apple"}}},
					{"value": {"item": {"value": "banana"}}}
				]}}
			}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{"revision": "5"}`)
			c := newTestClient(t, ts)
			revision, err := c.UpdateRecord(t.Context(), 1, tc.update)
			if err != nil {
				t.Fatalf("UpdateRecord returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPut, "/k/v1/record.json", tc.wantBody)
			if revision != 5 {
				t.Errorf("revision: got %d, want %d", revision, 5)
			}
		})
	}
}

func TestClientUpdateRecordError(t *testing.T) {
	ts, _ := newTestServer(t, http.StatusConflict,
		`{"code": "TEST_ERROR", "id": "abc123", "message": "Test error."}`)
	c := newTestClient(t, ts)
	revision, err := c.UpdateRecord(t.Context(), 1, RecordUpdate{ID: 2, Revision: 3})
	if !IsErrorCode(err, "TEST_ERROR") {
		t.Errorf("UpdateRecord should return TEST_ERROR error but got: %v", err)
	}
	if revision != 0 {
		t.Errorf("revision: got %d, want %d", revision, 0)
	}
}

func TestClientUpdateRecords(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{
		"records": [
			{"id": "2", "revision": "5"},
			{"id": "3", "revision": "2"}
		]
	}`)
	c := newTestClient(t, ts)
	res, err := c.UpdateRecords(t.Context(), 1, []RecordUpdate{
		{ID: 2, Record: Record{"title": {Value: "Hello"}}, Revision: 4},
		{UpdateKey: &UpdateKey{Field: "code", Value: "A001"}, Record: Record{"title": {Value: "World"}}},
	})
	if err != nil {
		t.Fatalf("UpdateRecords returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPut, "/k/v1/records.json", `{
		"app": 1,
		"records": [
			{"id": 2, "record": {"title": {"value": "Hello"}}, "revision": 4},
			{"updateKey": {"field": "code", "value": "A001"}, "record": {"title": {"value": "World"}}}
		]
	}`)
	want := []UpdatedRecord{
		{ID: 2, Revision: 5},
		{ID: 3, Revision: 2},
	}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("response: got %+v, want %+v", res, want)
	}
}

func TestClientUpsertRecords(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{
		"records": [
			{"id": "2", "revision": "5", "operation": "UPDATE"},
			{"id": "100", "revision": "1", "operation": "INSERT"}
		]
	}`)
	c := newTestClient(t, ts)
	res, err := c.UpsertRecords(t.Context(), 1, []RecordUpdate{
		{UpdateKey: &UpdateKey{Field: "code", Value: "A001"}, Record: Record{"title": {Value: "Hello"}}},
		{UpdateKey: &UpdateKey{Field: "code", Value: "A002"}, Record: Record{"title": {Value: "World"}}},
	})
	if err != nil {
		t.Fatalf("UpsertRecords returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPut, "/k/v1/records.json", `{
		"app": 1,
		"upsert": true,
		"records": [
			{"updateKey": {"field": "code", "value": "A001"}, "record": {"title": {"value": "Hello"}}},
			{"updateKey": {"field": "code", "value": "A002"}, "record": {"title": {"value": "World"}}}
		]
	}`)
	want := []UpdatedRecord{
		{ID: 2, Revision: 5, Operation: OperationUpdate},
		{ID: 100, Revision: 1, Operation: OperationInsert},
	}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("response: got %+v, want %+v", res, want)
	}
}

func TestClientDeleteRecords(t *testing.T) {
	testCases := []struct {
		name      string
		ids       []int64
		revisions []int64
		wantBody  string
	}{
		{
			name:     "without revisions",
			ids:      []int64{2, 3},
			wantBody: `{"app": 1, "ids": [2, 3]}`,
		},
		{
			name:      "with revisions",
			ids:       []int64{2, 3},
			revisions: []int64{4, -1},
			wantBody:  `{"app": 1, "ids": [2, 3], "revisions": [4, -1]}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{}`)
			c := newTestClient(t, ts)
			if err := c.DeleteRecords(t.Context(), 1, tc.ids, tc.revisions); err != nil {
				t.Fatalf("DeleteRecords returned error: %v", err)
			}
			assertRequest(t, got, http.MethodDelete, "/k/v1/records.json", tc.wantBody)
		})
	}
}

func TestClientDeleteRecordsError(t *testing.T) {
	ts, _ := newTestServer(t, http.StatusForbidden,
		`{"code": "TEST_ERROR", "id": "abc123", "message": "Test error."}`)
	c := newTestClient(t, ts)
	err := c.DeleteRecords(t.Context(), 1, []int64{2}, nil)
	if !IsErrorCode(err, "TEST_ERROR") {
		t.Errorf("DeleteRecords should return TEST_ERROR error but got: %v", err)
	}
}

func TestClientGetRecordsWarning(t *testing.T) {
	testCases := []struct {
		name    string
		warning string
	}{
		{"warning", "Filter aborted because of too many search results"},
		{"no warning", ""},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.warning != "" {
					w.Header().Set("X-Cybozu-Warning", tc.warning)
				}
				_, _ = io.WriteString(w, `{"records": [{"$id": {"type": "__ID__", "value": "1"}}], "totalCount": null}`)
			}))
			c := newTestClient(t, ts)
			res, err := c.GetRecords(t.Context(), GetRecordsRequest{App: 1, Query: `title like "a"`})
			if err != nil {
				t.Fatalf("GetRecords returned error: %v", err)
			}
			if res.Warning != tc.warning {
				t.Errorf("warning: got %q, want %q", res.Warning, tc.warning)
			}
			if len(res.Records) != 1 {
				t.Errorf("records should be returned even with the warning but got %d", len(res.Records))
			}
		})
	}
}
