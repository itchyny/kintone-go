package kintone

import (
	"net/http"
	"reflect"
	"testing"
)

func TestClientGetProcessManagement(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetAppSettingsRequest
		response string
		wantPath string
		wantBody string
		want     *ProcessManagement
	}{
		{
			name:     "settings",
			req:      GetAppSettingsRequest{App: 1, Lang: "en"},
			wantPath: "/k/v1/app/status.json",
			wantBody: `{"app": 1, "lang": "en"}`,
			response: `{
				"enable": true,
				"states": {
					"Open": {
						"name": "Open",
						"index": "0",
						"assignee": {"type": "ONE", "entities": []}
					},
					"Review": {
						"name": "Review",
						"index": "1",
						"assignee": {
							"type": "ANY",
							"entities": [
								{"entity": {"type": "USER", "code": "alice"}, "includeSubs": false},
								{"entity": {"type": "ORGANIZATION", "code": "dev"}, "includeSubs": true},
								{"entity": {"type": "CREATOR", "code": null}, "includeSubs": false}
							]
						}
					}
				},
				"actions": [
					{"name": "Submit", "from": "Open", "to": "Review", "filterCond": "", "type": "PRIMARY"},
					{
						"name": "Cancel", "from": "Review", "to": "Open", "filterCond": "amount > 100", "type": "SECONDARY",
						"executableUser": {"entities": [{"entity": {"type": "FIELD_ENTITY", "code": "owner"}, "includeSubs": false}]}
					}
				],
				"revision": "3"
			}`,
			want: &ProcessManagement{
				Enable: new(true),
				States: map[string]ProcessState{
					"Open": {Name: "Open", Index: 0, Assignee: &ProcessAssignee{Type: "ONE", Entities: []ProcessEntity{}}},
					"Review": {Name: "Review", Index: 1, Assignee: &ProcessAssignee{
						Type: "ANY",
						Entities: []ProcessEntity{
							{Entity: FieldEntity{Type: EntityTypeUser, Code: "alice"}},
							{Entity: FieldEntity{Type: EntityTypeOrganization, Code: "dev"}, IncludeSubs: true},
							{Entity: FieldEntity{Type: EntityTypeCreator}},
						},
					}},
				},
				Actions: []ProcessAction{
					{Name: "Submit", From: "Open", To: "Review", Type: "PRIMARY"},
					{
						Name: "Cancel", From: "Review", To: "Open", FilterCond: "amount > 100", Type: "SECONDARY",
						ExecutableUser: &ExecutableUser{Entities: []ProcessEntity{
							{Entity: FieldEntity{Type: EntityTypeFieldEntity, Code: "owner"}},
						}},
					},
				},
				Revision: 3,
			},
		},
		{
			name:     "never set",
			req:      GetAppSettingsRequest{App: 1, Preview: true},
			wantPath: "/k/v1/preview/app/status.json",
			wantBody: `{"app": 1}`,
			response: `{"enable": false, "states": null, "actions": null, "revision": "1"}`,
			want:     &ProcessManagement{Enable: new(false), Revision: 1},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, tc.response)
			c := newTestClient(t, ts)
			res, err := c.GetProcessManagement(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetProcessManagement returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, tc.wantPath, tc.wantBody)
			if !reflect.DeepEqual(res, tc.want) {
				t.Errorf("response: got %+v, want %+v", res, tc.want)
			}
		})
	}
}

func TestClientUpdateProcessManagement(t *testing.T) {
	testCases := []struct {
		name     string
		settings ProcessManagement
		revision int64
		wantBody string
	}{
		{
			name: "states and actions",
			settings: ProcessManagement{
				Enable: new(true),
				States: map[string]ProcessState{
					"Open": {Name: "Open", Index: 0},
					"Done": {Name: "Done", Index: 1, Assignee: &ProcessAssignee{
						Type: "ALL",
						Entities: []ProcessEntity{
							{Entity: FieldEntity{Type: EntityTypeCreator}},
							{Entity: FieldEntity{Type: EntityTypeCustomField, Code: "manager"}},
						},
					}},
				},
				Actions: []ProcessAction{
					{Name: "Finish", From: "Open", To: "Done"},
				},
				Revision: 9,
			},
			revision: 3,
			wantBody: `{
				"app": 1,
				"enable": true,
				"states": {
					"Open": {"name": "Open", "index": "0"},
					"Done": {"name": "Done", "index": "1", "assignee": {"type": "ALL", "entities": [
						{"entity": {"type": "CREATOR"}},
						{"entity": {"type": "CUSTOM_FIELD", "code": "manager"}}
					]}}
				},
				"actions": [{"name": "Finish", "from": "Open", "to": "Done"}],
				"revision": 3
			}`,
		},
		{
			name:     "disable",
			settings: ProcessManagement{Enable: new(false)},
			wantBody: `{"app": 1, "enable": false}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{"revision": "4"}`)
			c := newTestClient(t, ts)
			revision, err := c.UpdateProcessManagement(t.Context(), 1, tc.settings, tc.revision)
			if err != nil {
				t.Fatalf("UpdateProcessManagement returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPut, "/k/v1/preview/app/status.json", tc.wantBody)
			if revision != 4 {
				t.Errorf("revision: got %d, want %d", revision, 4)
			}
		})
	}
}
