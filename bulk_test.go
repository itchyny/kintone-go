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
			{"updateKey": {"field": "code", "value": "A"}}
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
				`{"requests": [{"method": "POST", "api": "`+tc.wantAPI+`", "payload": {"app": 1}}]}`)
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
	var bulkErr *BulkError
	if !errors.As(err, &bulkErr) || bulkErr.Index != 1 || bulkErr.Err.Code != "TEST_ERROR" {
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
