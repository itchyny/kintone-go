package kintone

import (
	"net/http"
	"reflect"
	"testing"
)

func TestClientUpdateAssignees(t *testing.T) {
	testCases := []struct {
		name     string
		update   AssigneesUpdate
		wantBody string
	}{
		{
			name:     "assignees with revision",
			update:   AssigneesUpdate{ID: 2, Assignees: []string{"alice", "bob"}, Revision: 4},
			wantBody: `{"app": 1, "id": 2, "assignees": ["alice", "bob"], "revision": 4}`,
		},
		{
			name:     "nil assignees",
			update:   AssigneesUpdate{ID: 2},
			wantBody: `{"app": 1, "id": 2, "assignees": []}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{"revision": "5"}`)
			c := newTestClient(t, ts)
			revision, err := c.UpdateAssignees(t.Context(), 1, tc.update)
			if err != nil {
				t.Fatalf("UpdateAssignees returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPut, "/k/v1/record/assignees.json", tc.wantBody)
			if revision != 5 {
				t.Errorf("revision: got %d, want %d", revision, 5)
			}
		})
	}
}

func TestClientUpdateStatus(t *testing.T) {
	testCases := []struct {
		name     string
		update   StatusUpdate
		wantBody string
	}{
		{
			name:     "action with assignee and revision",
			update:   StatusUpdate{ID: 2, Action: "Submit", Assignee: "alice", Revision: 4},
			wantBody: `{"app": 1, "id": 2, "action": "Submit", "assignee": "alice", "revision": 4}`,
		},
		{
			name:     "only action",
			update:   StatusUpdate{ID: 2, Action: "Approve"},
			wantBody: `{"app": 1, "id": 2, "action": "Approve"}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{"revision": "6"}`)
			c := newTestClient(t, ts)
			revision, err := c.UpdateStatus(t.Context(), 1, tc.update)
			if err != nil {
				t.Fatalf("UpdateStatus returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPut, "/k/v1/record/status.json", tc.wantBody)
			if revision != 6 {
				t.Errorf("revision: got %d, want %d", revision, 6)
			}
		})
	}
}

func TestClientUpdateStatuses(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{
		"records": [
			{"id": "2", "revision": "6"},
			{"id": "3", "revision": "9"}
		]
	}`)
	c := newTestClient(t, ts)
	res, err := c.UpdateStatuses(t.Context(), 1, []StatusUpdate{
		{ID: 2, Action: "Submit", Assignee: "alice", Revision: 4},
		{ID: 3, Action: "Approve"},
	})
	if err != nil {
		t.Fatalf("UpdateStatuses returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPut, "/k/v1/records/status.json", `{
		"app": 1,
		"records": [
			{"id": 2, "action": "Submit", "assignee": "alice", "revision": 4},
			{"id": 3, "action": "Approve"}
		]
	}`)
	want := []RecordRevision{{ID: 2, Revision: 6}, {ID: 3, Revision: 9}}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("response: got %+v, want %+v", res, want)
	}
}
