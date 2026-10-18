package kintone

import (
	"net/http"
	"reflect"
	"testing"
)

func TestClientGetAppPlugins(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetAppSettingsRequest
		wantPath string
		wantBody string
	}{
		{
			name:     "production",
			req:      GetAppSettingsRequest{App: 1, Lang: "es"},
			wantPath: "/k/v1/app/plugins.json",
			wantBody: `{"app": 1, "lang": "es"}`,
		},
		{
			name:     "preview",
			req:      GetAppSettingsRequest{App: 1, Preview: true},
			wantPath: "/k/v1/preview/app/plugins.json",
			wantBody: `{"app": 1}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{
				"plugins": [
					{"id": "abcdefghijklmnopabcdefghijklmnop", "name": "Calendar", "enabled": true},
					{"id": "ponmlkjihgfedcbaponmlkjihgfedcba", "name": "Chart", "enabled": false}
				],
				"revision": "3"
			}`)
			c := newTestClient(t, ts)
			res, err := c.GetAppPlugins(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetAppPlugins returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, tc.wantPath, tc.wantBody)
			want := &AppPlugins{
				Plugins: []AppPlugin{
					{ID: "abcdefghijklmnopabcdefghijklmnop", Name: "Calendar", Enabled: true},
					{ID: "ponmlkjihgfedcbaponmlkjihgfedcba", Name: "Chart"},
				},
				Revision: 3,
			}
			if !reflect.DeepEqual(res, want) {
				t.Errorf("response: got %+v, want %+v", res, want)
			}
		})
	}
}

func TestClientAddAppPlugins(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{"revision": "4"}`)
	c := newTestClient(t, ts)
	revision, err := c.AddAppPlugins(t.Context(), 1, []string{"abcdefghijklmnopabcdefghijklmnop"}, 3)
	if err != nil {
		t.Fatalf("AddAppPlugins returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPost, "/k/v1/preview/app/plugins.json",
		`{"app": 1, "ids": ["abcdefghijklmnopabcdefghijklmnop"], "revision": 3}`)
	if revision != 4 {
		t.Errorf("revision: got %d, want %d", revision, 4)
	}
}
