package kintone

import (
	"errors"
	"net/http"
	"reflect"
	"testing"
)

func TestBulkRequestSend(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{"results": [
		{"id": "10", "revision": "1"},
		{"ids": ["11", "12"], "revisions": ["1", "1"]},
		{"revision": "5"},
		{"records": [{"id": "2", "revision": "6"}]},
		{"records": [{"id": "3", "revision": "2", "operation": "INSERT"}]},
		{},
		{"revision": "7"},
		{"records": [{"id": "4", "revision": "3"}]},
		{"revision": "8"}
	]}`)
	c := newTestClient(t, ts)
	bulk := c.NewBulkRequest()
	addRecord := bulk.AddRecord(1, Record{"title": {Value: "a"}})
	addRecords := bulk.AddRecords(1, []Record{{"title": {Value: "b"}}, {"title": {Value: "c"}}})
	updateRecord := bulk.UpdateRecord(2, RecordUpdate{ID: 2, Record: Record{"title": {Value: "d"}}, Revision: 4})
	updateRecords := bulk.UpdateRecords(2, []RecordUpdate{{ID: 2, Record: Record{"title": {Value: "e"}}}})
	upsertRecords := bulk.UpsertRecords(2, []RecordUpdate{{UpdateKey: &UpdateKey{Field: "code", Value: "A"}, Record: Record{}}})
	deleteRecords := bulk.DeleteRecords(1, []int64{5}, nil)
	updateStatus := bulk.UpdateStatus(1, StatusUpdate{ID: 10, Action: "Submit"})
	updateStatuses := bulk.UpdateStatuses(1, []StatusUpdate{{ID: 4, Action: "Approve"}})
	updateAssignees := bulk.UpdateAssignees(1, AssigneesUpdate{ID: 10})
	if addRecord.Value() != (RecordRevision{}) {
		t.Errorf("value before sending: got %+v", addRecord.Value())
	}
	if err := bulk.Send(t.Context()); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPost, "/k/v1/bulkRequest.json", `{"requests": [
		{"method": "POST", "api": "/k/v1/record.json", "payload": {"app": 1, "record": {"title": {"value": "a"}}}},
		{"method": "POST", "api": "/k/v1/records.json", "payload": {"app": 1, "records": [
			{"title": {"value": "b"}}, {"title": {"value": "c"}}
		]}},
		{"method": "PUT", "api": "/k/v1/record.json", "payload": {"app": 2, "id": 2, "record": {"title": {"value": "d"}}, "revision": 4}},
		{"method": "PUT", "api": "/k/v1/records.json", "payload": {"app": 2, "records": [
			{"id": 2, "record": {"title": {"value": "e"}}}
		]}},
		{"method": "PUT", "api": "/k/v1/records.json", "payload": {"app": 2, "records": [
			{"updateKey": {"field": "code", "value": "A"}, "record": {}}
		], "upsert": true}},
		{"method": "DELETE", "api": "/k/v1/records.json", "payload": {"app": 1, "ids": [5]}},
		{"method": "PUT", "api": "/k/v1/record/status.json", "payload": {"app": 1, "id": 10, "action": "Submit"}},
		{"method": "PUT", "api": "/k/v1/records/status.json", "payload": {"app": 1, "records": [{"id": 4, "action": "Approve"}]}},
		{"method": "PUT", "api": "/k/v1/record/assignees.json", "payload": {"app": 1, "id": 10, "assignees": []}}
	]}`)
	for _, tc := range []struct {
		name      string
		got, want any
	}{
		{"AddRecord", addRecord.Value(), RecordRevision{ID: 10, Revision: 1}},
		{"AddRecords", addRecords.Value(), []RecordRevision{{ID: 11, Revision: 1}, {ID: 12, Revision: 1}}},
		{"UpdateRecord", updateRecord.Value(), int64(5)},
		{"UpdateRecords", updateRecords.Value(), []UpdatedRecord{{ID: 2, Revision: 6}}},
		{"UpsertRecords", upsertRecords.Value(), []UpdatedRecord{{ID: 3, Revision: 2, Operation: OperationInsert}}},
		{"DeleteRecords", deleteRecords.Value(), struct{}{}},
		{"UpdateStatus", updateStatus.Value(), int64(7)},
		{"UpdateStatuses", updateStatuses.Value(), []RecordRevision{{ID: 4, Revision: 3}}},
		{"UpdateAssignees", updateAssignees.Value(), int64(8)},
	} {
		if !reflect.DeepEqual(tc.got, tc.want) {
			t.Errorf("%s: got %+v, want %+v", tc.name, tc.got, tc.want)
		}
	}
}

func TestBulkRequestSendPath(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		opts     []Option
		wantPath string
		wantAPI  string
	}{
		{
			name:     "guest space",
			opts:     []Option{WithGuestSpace(4)},
			wantPath: "/k/guest/4/v1/bulkRequest.json",
			wantAPI:  "/k/guest/4/v1/record.json",
		},
		{
			name:     "base URL with path",
			path:     "/proxy",
			wantPath: "/proxy/k/v1/bulkRequest.json",
			wantAPI:  "/k/v1/record.json",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{"results": [{"id": "10", "revision": "1"}]}`)
			c := newClient(t, ts.URL+tc.path, append([]Option{WithHTTPClient(ts.Client())}, tc.opts...)...)
			bulk := c.NewBulkRequest()
			bulk.AddRecord(1, Record{})
			if err := bulk.Send(t.Context()); err != nil {
				t.Fatalf("Send returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPost, tc.wantPath,
				`{"requests": [{"method": "POST", "api": "`+tc.wantAPI+`", "payload": {"app": 1, "record": {}}}]}`)
		})
	}
}

func TestBulkRequestSendEmpty(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{"results": []}`)
	c := newTestClient(t, ts)
	if err := c.NewBulkRequest().Send(t.Context()); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPost, "/k/v1/bulkRequest.json", `{"requests": []}`)
}

func TestBulkRequestSendError(t *testing.T) {
	ts, _ := newTestServer(t, http.StatusBadRequest,
		`{"results": [{}, {"code": "TEST_ERROR", "id": "abc123", "message": "Test error."}]}`)
	c := newTestClient(t, ts)
	bulk := c.NewBulkRequest()
	added := bulk.AddRecord(1, Record{})
	updated := bulk.UpdateRecord(1, RecordUpdate{ID: 2})
	err := bulk.Send(t.Context())
	if bulkErr, ok := errors.AsType[*BulkError](err); !ok || bulkErr.Index != 1 || bulkErr.Err.Code != "TEST_ERROR" {
		t.Fatalf("Send should return *BulkError of index 1 but got: %v", err)
	}
	if !IsErrorCode(err, "TEST_ERROR") {
		t.Error("IsErrorCode should find the code")
	}
	if added.Value() != (RecordRevision{}) || updated.Value() != 0 {
		t.Errorf("values should be zero on error: %+v, %d", added.Value(), updated.Value())
	}
}

func TestBulkRequestSendDecodeError(t *testing.T) {
	addAndUpdate := func(b *BulkRequest) {
		b.AddRecord(1, Record{})
		b.UpdateRecord(1, RecordUpdate{ID: 2})
	}
	testCases := []struct {
		name     string
		add      func(*BulkRequest)
		response string
	}{
		{"fewer results", addAndUpdate, `{"results": [{"id": "10", "revision": "1"}]}`},
		{"more results", addAndUpdate, `{"results": [{"id": "10", "revision": "1"}, {"revision": "5"}, {}]}`},
		{"invalid result", addAndUpdate, `{"results": [{"id": "x", "revision": "1"}, {"revision": "5"}]}`},
		{
			"invalid IDs",
			func(b *BulkRequest) { b.AddRecords(1, []Record{{}}) },
			`{"results": [{"ids": ["x"], "revisions": ["1"]}]}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, _ := newTestServer(t, http.StatusOK, tc.response)
			c := newTestClient(t, ts)
			bulk := c.NewBulkRequest()
			tc.add(bulk)
			if err := bulk.Send(t.Context()); err == nil {
				t.Error("Send should return error")
			}
		})
	}
}

func TestClientAddAllRecords(t *testing.T) {
	setValue(t, &allRecordsChunkSize, 2)
	setValue(t, &allRecordsBulkSize, 2)
	ts, requests := newSequenceServer(t,
		response{http.StatusOK, `{"results": [
			{"ids": ["1", "2"], "revisions": ["1", "1"]},
			{"ids": ["3", "4"], "revisions": ["1", "1"]}
		]}`},
		response{http.StatusOK, `{"results": [{"ids": ["5"], "revisions": ["1"]}]}`},
	)
	c := newTestClient(t, ts)
	records := make([]Record, 5)
	for i := range records {
		records[i] = Record{"title": {Value: string(rune('a' + i))}}
	}
	res, err := c.AddAllRecords(t.Context(), 1, records)
	if err != nil {
		t.Fatalf("AddAllRecords returned error: %v", err)
	}
	reqs := requests()
	if len(reqs) != 2 {
		t.Fatalf("number of requests: got %d, want %d", len(reqs), 2)
	}
	assertRequest(t, &reqs[0], http.MethodPost, "/k/v1/bulkRequest.json", `{"requests": [
		{"method": "POST", "api": "/k/v1/records.json", "payload": {"app": 1, "records": [
			{"title": {"value": "a"}}, {"title": {"value": "b"}}
		]}},
		{"method": "POST", "api": "/k/v1/records.json", "payload": {"app": 1, "records": [
			{"title": {"value": "c"}}, {"title": {"value": "d"}}
		]}}
	]}`)
	assertRequest(t, &reqs[1], http.MethodPost, "/k/v1/bulkRequest.json", `{"requests": [
		{"method": "POST", "api": "/k/v1/records.json", "payload": {"app": 1, "records": [{"title": {"value": "e"}}]}}
	]}`)
	want := []RecordRevision{{1, 1}, {2, 1}, {3, 1}, {4, 1}, {5, 1}}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("response: got %+v, want %+v", res, want)
	}
}

func TestClientUpdateAllRecords(t *testing.T) {
	setValue(t, &allRecordsChunkSize, 2)
	setValue(t, &allRecordsBulkSize, 2)
	testCases := []struct {
		name     string
		update   func(*Client, []RecordUpdate) ([]UpdatedRecord, error)
		response string
		wantBody string
		want     []UpdatedRecord
	}{
		{
			name: "update",
			update: func(c *Client, updates []RecordUpdate) ([]UpdatedRecord, error) {
				return c.UpdateAllRecords(t.Context(), 1, updates)
			},
			response: `{"results": [
				{"records": [{"id": "1", "revision": "2"}, {"id": "2", "revision": "3"}]},
				{"records": [{"id": "3", "revision": "4"}]}
			]}`,
			wantBody: `{"requests": [
				{"method": "PUT", "api": "/k/v1/records.json", "payload": {"app": 1, "records": [{"id": 1}, {"id": 2}]}},
				{"method": "PUT", "api": "/k/v1/records.json", "payload": {"app": 1, "records": [{"id": 3}]}}
			]}`,
			want: []UpdatedRecord{{ID: 1, Revision: 2}, {ID: 2, Revision: 3}, {ID: 3, Revision: 4}},
		},
		{
			name: "upsert",
			update: func(c *Client, updates []RecordUpdate) ([]UpdatedRecord, error) {
				return c.UpsertAllRecords(t.Context(), 1, updates)
			},
			response: `{"results": [
				{"records": [{"id": "1", "revision": "2", "operation": "UPDATE"}, {"id": "2", "revision": "3", "operation": "UPDATE"}]},
				{"records": [{"id": "4", "revision": "1", "operation": "INSERT"}]}
			]}`,
			wantBody: `{"requests": [
				{"method": "PUT", "api": "/k/v1/records.json", "payload": {"app": 1, "records": [{"id": 1}, {"id": 2}], "upsert": true}},
				{"method": "PUT", "api": "/k/v1/records.json", "payload": {"app": 1, "records": [{"id": 3}], "upsert": true}}
			]}`,
			want: []UpdatedRecord{
				{ID: 1, Revision: 2, Operation: OperationUpdate},
				{ID: 2, Revision: 3, Operation: OperationUpdate},
				{ID: 4, Revision: 1, Operation: OperationInsert},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, tc.response)
			c := newTestClient(t, ts)
			res, err := tc.update(c, []RecordUpdate{{ID: 1}, {ID: 2}, {ID: 3}})
			if err != nil {
				t.Fatalf("returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPost, "/k/v1/bulkRequest.json", tc.wantBody)
			if !reflect.DeepEqual(res, tc.want) {
				t.Errorf("response: got %+v, want %+v", res, tc.want)
			}
		})
	}
}

func TestClientDeleteAllRecords(t *testing.T) {
	setValue(t, &allRecordsChunkSize, 2)
	setValue(t, &allRecordsBulkSize, 2)
	testCases := []struct {
		name      string
		revisions []int64
		wantBody  string
	}{
		{
			name:      "revisions",
			revisions: []int64{4, 5, 6},
			wantBody: `{"requests": [
				{"method": "DELETE", "api": "/k/v1/records.json", "payload": {"app": 1, "ids": [1, 2], "revisions": [4, 5]}},
				{"method": "DELETE", "api": "/k/v1/records.json", "payload": {"app": 1, "ids": [3], "revisions": [6]}}
			]}`,
		},
		{
			name: "nil revisions",
			wantBody: `{"requests": [
				{"method": "DELETE", "api": "/k/v1/records.json", "payload": {"app": 1, "ids": [1, 2]}},
				{"method": "DELETE", "api": "/k/v1/records.json", "payload": {"app": 1, "ids": [3]}}
			]}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{"results": [{}, {}]}`)
			c := newTestClient(t, ts)
			if err := c.DeleteAllRecords(t.Context(), 1, []int64{1, 2, 3}, tc.revisions); err != nil {
				t.Fatalf("DeleteAllRecords returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPost, "/k/v1/bulkRequest.json", tc.wantBody)
		})
	}
}

func TestClientDeleteAllRecordsRevisionsLength(t *testing.T) {
	ts, requests := newSequenceServer(t)
	c := newTestClient(t, ts)
	err := c.DeleteAllRecords(t.Context(), 1, []int64{1, 2}, []int64{1})
	if err == nil || err.Error() != "kintone: got 1 revisions for 2 ids" {
		t.Errorf("DeleteAllRecords returned unexpected error: %v", err)
	}
	if reqs := requests(); len(reqs) != 0 {
		t.Errorf("number of requests: got %d, want %d", len(reqs), 0)
	}
}

func TestClientAddAllRecordsEmpty(t *testing.T) {
	ts, requests := newSequenceServer(t)
	c := newTestClient(t, ts)
	res, err := c.AddAllRecords(t.Context(), 1, nil)
	if err != nil {
		t.Fatalf("AddAllRecords returned error: %v", err)
	}
	if res != nil {
		t.Errorf("response: got %+v, want nil", res)
	}
	if reqs := requests(); len(reqs) != 0 {
		t.Errorf("number of requests: got %d, want %d", len(reqs), 0)
	}
}

func TestClientAddAllRecordsError(t *testing.T) {
	setValue(t, &allRecordsChunkSize, 2)
	setValue(t, &allRecordsBulkSize, 2)
	ts, requests := newSequenceServer(t,
		response{http.StatusOK, `{"results": [
			{"ids": ["1", "2"], "revisions": ["1", "1"]},
			{"ids": ["3", "4"], "revisions": ["1", "1"]}
		]}`},
		response{http.StatusBadRequest, `{"results": [
			{},
			{"code": "TEST_ERROR", "id": "abc123", "message": "Test error."}
		]}`},
	)
	c := newTestClient(t, ts)
	res, err := c.AddAllRecords(t.Context(), 1, make([]Record, 7))
	want := []RecordRevision{{1, 1}, {2, 1}, {3, 1}, {4, 1}}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("response: got %+v, want %+v", res, want)
	}
	if allErr, ok := errors.AsType[*AllRecordsError](err); !ok || allErr.Processed != 4 {
		t.Fatalf("AddAllRecords returned unexpected error: %v", err)
	}
	if bulkErr, ok := errors.AsType[*BulkError](err); !ok || bulkErr.Index != 1 || !IsErrorCode(err, "TEST_ERROR") {
		t.Errorf("AddAllRecords returned unexpected error: %v", err)
	}
	if want := "kintone: processed 4 records: bulk request 1: Test error. (status: 400, code: TEST_ERROR, id: abc123)"; err.Error() != want {
		t.Errorf("error message: got %q, want %q", err.Error(), want)
	}
	if reqs := requests(); len(reqs) != 2 {
		t.Errorf("number of requests: got %d, want %d", len(reqs), 2)
	}
}
