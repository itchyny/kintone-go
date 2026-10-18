package kintone

import (
	"net/http"
	"reflect"
	"testing"
)

func TestClientGetAppActions(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetAppSettingsRequest
		wantPath string
		wantBody string
	}{
		{
			name:     "production",
			req:      GetAppSettingsRequest{App: 1, Lang: "en"},
			wantPath: "/k/v1/app/actions.json",
			wantBody: `{"app": 1, "lang": "en"}`,
		},
		{
			name:     "preview",
			req:      GetAppSettingsRequest{App: 1, Preview: true},
			wantPath: "/k/v1/preview/app/actions.json",
			wantBody: `{"app": 1}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{
				"actions": {
					"Create invoice": {
						"name": "Create invoice",
						"id": "30",
						"index": "0",
						"destApp": {"app": "5", "code": "INVOICE"},
						"mappings": [
							{"srcType": "FIELD", "srcField": "customer", "destField": "billTo"},
							{"srcType": "RECORD_URL", "destField": "source"}
						],
						"entities": [{"type": "GROUP", "code": "sales"}],
						"filterCond": "status = \"Won\""
					},
					"Copy": {
						"name": "Copy",
						"id": "31",
						"index": "1",
						"destApp": null,
						"mappings": null,
						"entities": [],
						"filterCond": ""
					}
				},
				"revision": "3"
			}`)
			c := newTestClient(t, ts)
			res, err := c.GetAppActions(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetAppActions returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, tc.wantPath, tc.wantBody)
			want := &AppActions{
				Actions: map[string]AppAction{
					"Create invoice": {
						ID:      30,
						Name:    "Create invoice",
						Index:   0,
						DestApp: &RelatedApp{App: 5, Code: "INVOICE"},
						Mappings: []ActionMapping{
							{SrcType: "FIELD", SrcField: "customer", DestField: "billTo"},
							{SrcType: "RECORD_URL", DestField: "source"},
						},
						Entities:   []FieldEntity{{Type: EntityTypeGroup, Code: "sales"}},
						FilterCond: `status = "Won"`,
					},
					"Copy": {
						ID:       31,
						Name:     "Copy",
						Index:    1,
						Entities: []FieldEntity{},
					},
				},
				Revision: 3,
			}
			if !reflect.DeepEqual(res, want) {
				t.Errorf("response: got %+v, want %+v", res, want)
			}
		})
	}
}

func TestClientUpdateAppActions(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{
		"revision": "4",
		"actions": {"Create invoice": {"id": "30"}, "Archive": {"id": "32"}}
	}`)
	c := newTestClient(t, ts)
	res, err := c.UpdateAppActions(t.Context(), 1, map[string]AppAction{
		"Create invoice": {Name: "Create invoice", Index: 0, Entities: []FieldEntity{}},
		"Archive": {
			Name:     "Archive",
			Index:    1,
			DestApp:  &RelatedApp{Code: "ARCHIVE"},
			Mappings: []ActionMapping{},
			Entities: []FieldEntity{{Type: EntityTypeUser, Code: "alice"}},
		},
	}, 3)
	if err != nil {
		t.Fatalf("UpdateAppActions returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPut, "/k/v1/preview/app/actions.json", `{
		"app": 1,
		"actions": {
			"Create invoice": {"name": "Create invoice", "index": "0", "entities": []},
			"Archive": {
				"name": "Archive",
				"index": "1",
				"destApp": {"code": "ARCHIVE"},
				"mappings": [],
				"entities": [{"type": "USER", "code": "alice"}]
			}
		},
		"revision": 3
	}`)
	want := &UpdatedAppActions{
		Actions:  map[string]UpdatedAppAction{"Create invoice": {ID: 30}, "Archive": {ID: 32}},
		Revision: 4,
	}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("response: got %+v, want %+v", res, want)
	}
}
