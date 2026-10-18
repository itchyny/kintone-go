package kintone

import (
	"net/http"
	"reflect"
	"testing"
)

func TestClientGetCategories(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetAppSettingsRequest
		wantPath string
		wantBody string
	}{
		{
			name:     "production",
			req:      GetAppSettingsRequest{App: 1, Lang: "en"},
			wantPath: "/k/v1/app/categories.json",
			wantBody: `{"app": 1, "lang": "en"}`,
		},
		{
			name:     "preview",
			req:      GetAppSettingsRequest{App: 1, Preview: true},
			wantPath: "/k/v1/preview/app/categories.json",
			wantBody: `{"app": 1}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{
				"enabled": true,
				"categories": {
					"Products": {
						"name": "Products",
						"index": "0",
						"children": {
							"Hardware": {"name": "Hardware", "index": "0", "children": {}},
							"Software": {"name": "Software", "index": "1", "children": {}}
						}
					}
				},
				"revision": "2"
			}`)
			c := newTestClient(t, ts)
			res, err := c.GetCategories(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetCategories returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, tc.wantPath, tc.wantBody)
			want := &Categories{
				Enabled: new(true),
				Categories: map[string]Category{
					"Products": {
						Name:  "Products",
						Index: 0,
						Children: map[string]Category{
							"Hardware": {Name: "Hardware", Index: 0, Children: map[string]Category{}},
							"Software": {Name: "Software", Index: 1, Children: map[string]Category{}},
						},
					},
				},
				Revision: 2,
			}
			if !reflect.DeepEqual(res, want) {
				t.Errorf("response: got %+v, want %+v", res, want)
			}
		})
	}
}

func TestClientUpdateCategories(t *testing.T) {
	testCases := []struct {
		name     string
		settings Categories
		revision int64
		wantBody string
	}{
		{
			name: "categories",
			settings: Categories{
				Enabled: new(true),
				Categories: map[string]Category{
					"Products": {
						Name:  "Goods",
						Index: 0,
						Children: map[string]Category{
							"Hardware": {Name: "Hardware", Index: 0},
						},
					},
					"Services": {Name: "Services", Index: 1},
				},
				Revision: 9,
			},
			revision: 2,
			wantBody: `{
				"app": 1,
				"enabled": true,
				"categories": {
					"Products": {
						"name": "Goods",
						"index": "0",
						"children": {"Hardware": {"name": "Hardware", "index": "0", "children": {}}}
					},
					"Services": {"name": "Services", "index": "1", "children": {}}
				},
				"revision": 2
			}`,
		},
		{
			name:     "disable",
			settings: Categories{Enabled: new(false)},
			wantBody: `{"app": 1, "enabled": false}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{"revision": "3"}`)
			c := newTestClient(t, ts)
			revision, err := c.UpdateCategories(t.Context(), 1, tc.settings, tc.revision)
			if err != nil {
				t.Fatalf("UpdateCategories returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPut, "/k/v1/preview/app/categories.json", tc.wantBody)
			if revision != 3 {
				t.Errorf("revision: got %d, want %d", revision, 3)
			}
		})
	}
}
