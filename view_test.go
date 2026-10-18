package kintone

import (
	"net/http"
	"reflect"
	"testing"
)

func TestClientGetViews(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetAppSettingsRequest
		wantPath string
		wantBody string
	}{
		{
			name:     "production",
			req:      GetAppSettingsRequest{App: 1, Lang: "ja"},
			wantPath: "/k/v1/app/views.json",
			wantBody: `{"app": 1, "lang": "ja"}`,
		},
		{
			name:     "preview",
			req:      GetAppSettingsRequest{App: 1, Preview: true},
			wantPath: "/k/v1/preview/app/views.json",
			wantBody: `{"app": 1}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{
				"views": {
					"All": {
						"type": "LIST", "name": "All", "id": "10", "index": "0",
						"fields": ["$id", "title"], "filterCond": "", "sort": "$id desc"
					},
					"Calendar": {
						"type": "CALENDAR", "name": "Calendar", "id": "11", "index": "1",
						"date": "due", "title": "title", "filterCond": "", "sort": ""
					},
					"Custom": {
						"type": "CUSTOM", "name": "Custom", "id": "12", "index": "2",
						"html": "<div></div>", "pager": false, "device": "ANY", "filterCond": "", "sort": ""
					},
					"(Assigned to me)": {
						"type": "LIST", "builtinType": "ASSIGNEE", "name": "(Assigned to me)", "id": "13", "index": "3",
						"fields": ["title"], "filterCond": "", "sort": ""
					}
				},
				"revision": "4"
			}`)
			c := newTestClient(t, ts)
			res, err := c.GetViews(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetViews returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, tc.wantPath, tc.wantBody)
			want := &Views{
				Views: map[string]View{
					"All": {
						ID: 10, Name: "All", Type: ViewTypeList, Index: 0,
						Fields: []string{"$id", "title"}, Sort: "$id desc",
					},
					"Calendar": {
						ID: 11, Name: "Calendar", Type: ViewTypeCalendar, Index: 1,
						Date: "due", Title: "title",
					},
					"Custom": {
						ID: 12, Name: "Custom", Type: ViewTypeCustom, Index: 2,
						HTML: "<div></div>", Pager: new(false), Device: "ANY",
					},
					"(Assigned to me)": {
						ID: 13, Name: "(Assigned to me)", Type: ViewTypeList, BuiltinType: "ASSIGNEE", Index: 3,
						Fields: []string{"title"},
					},
				},
				Revision: 4,
			}
			if !reflect.DeepEqual(res, want) {
				t.Errorf("response: got %+v, want %+v", res, want)
			}
		})
	}
}

func TestClientUpdateViews(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{
		"views": {"All": {"id": "10"}, "Open": {"id": "14"}},
		"revision": "5"
	}`)
	c := newTestClient(t, ts)
	res, err := c.UpdateViews(t.Context(), 1, map[string]View{
		"All":  {Name: "All", Type: ViewTypeList, Index: 0, Fields: []string{"$id", "title"}},
		"Open": {Name: "Open", Type: ViewTypeCustom, Index: 1, HTML: "<p></p>", Pager: new(false), FilterCond: `status in ("Open")`},
	}, 4)
	if err != nil {
		t.Fatalf("UpdateViews returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPut, "/k/v1/preview/app/views.json", `{
		"app": 1,
		"views": {
			"All": {"name": "All", "type": "LIST", "index": "0", "fields": ["$id", "title"]},
			"Open": {
				"name": "Open", "type": "CUSTOM", "index": "1",
				"html": "<p></p>", "pager": false, "filterCond": "status in (\"Open\")"
			}
		},
		"revision": 4
	}`)
	want := &UpdatedViews{
		Views:    map[string]UpdatedView{"All": {ID: 10}, "Open": {ID: 14}},
		Revision: 5,
	}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("response: got %+v, want %+v", res, want)
	}
}
