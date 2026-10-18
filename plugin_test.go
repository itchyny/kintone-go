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

func TestClientGetPlugins(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetPluginsRequest
		opts     []Option
		wantBody string
	}{
		{
			name:     "all parameters",
			req:      GetPluginsRequest{Offset: 10, Limit: 20, IDs: []string{"abcdefghijklmnopabcdefghijklmnop"}},
			wantBody: `{"offset": 10, "limit": 20, "ids": ["abcdefghijklmnopabcdefghijklmnop"]}`,
		},
		{
			name:     "guest space is not used",
			opts:     []Option{WithGuestSpace(4)},
			wantBody: `{}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{
				"plugins": [
					{
						"id": "abcdefghijklmnopabcdefghijklmnop",
						"name": "Calendar",
						"description": "Shows records in a calendar.",
						"isMarketPlugin": true,
						"version": "1.2.0"
					}
				]
			}`)
			c := newTestClient(t, ts, tc.opts...)
			plugins, err := c.GetPlugins(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetPlugins returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, "/k/v1/plugins.json", tc.wantBody)
			want := []Plugin{
				{
					ID:             "abcdefghijklmnopabcdefghijklmnop",
					Name:           "Calendar",
					Description:    "Shows records in a calendar.",
					IsMarketPlugin: true,
					Version:        "1.2.0",
				},
			}
			if !reflect.DeepEqual(plugins, want) {
				t.Errorf("plugins: got %+v, want %+v", plugins, want)
			}
		})
	}
}

func TestClientGetRequiredPlugins(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{
		"plugins": [
			{"id": "abcdefghijklmnopabcdefghijklmnop", "name": "Calendar", "isMarketPlugin": false},
			{"id": "ponmlkjihgfedcbaponmlkjihgfedcba", "name": null, "isMarketPlugin": true}
		]
	}`)
	c := newTestClient(t, ts, WithGuestSpace(4))
	plugins, err := c.GetRequiredPlugins(t.Context(), GetRequiredPluginsRequest{Limit: 50})
	if err != nil {
		t.Fatalf("GetRequiredPlugins returned error: %v", err)
	}
	assertRequest(t, got, http.MethodGet, "/k/v1/plugins/required.json", `{"limit": 50}`)
	want := []Plugin{
		{ID: "abcdefghijklmnopabcdefghijklmnop", Name: "Calendar"},
		{ID: "ponmlkjihgfedcbaponmlkjihgfedcba", IsMarketPlugin: true},
	}
	if !reflect.DeepEqual(plugins, want) {
		t.Errorf("plugins: got %+v, want %+v", plugins, want)
	}
}

func TestClientGetPluginApps(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{
		"apps": [{"id": "1", "name": "Deals"}, {"id": "5", "name": "Invoices"}]
	}`)
	c := newTestClient(t, ts, WithGuestSpace(4))
	apps, err := c.GetPluginApps(t.Context(), GetPluginAppsRequest{ID: "abcdefghijklmnopabcdefghijklmnop", Offset: 100, Limit: 500})
	if err != nil {
		t.Fatalf("GetPluginApps returned error: %v", err)
	}
	assertRequest(t, got, http.MethodGet, "/k/v1/plugin/apps.json",
		`{"id": "abcdefghijklmnopabcdefghijklmnop", "offset": 100, "limit": 500}`)
	want := []PluginApp{{ID: 1, Name: "Deals"}, {ID: 5, Name: "Invoices"}}
	if !reflect.DeepEqual(apps, want) {
		t.Errorf("apps: got %+v, want %+v", apps, want)
	}
}

func TestClientAddPlugin(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{"id": "abcdefghijklmnopabcdefghijklmnop", "version": "1.0.0"}`)
	c := newTestClient(t, ts, WithGuestSpace(4))
	res, err := c.AddPlugin(t.Context(), "c15b3870-7505-4ab6-9d8d-b9bdbc74f5d6")
	if err != nil {
		t.Fatalf("AddPlugin returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPost, "/k/v1/plugin.json", `{"fileKey": "c15b3870-7505-4ab6-9d8d-b9bdbc74f5d6"}`)
	if want := (PluginVersion{ID: "abcdefghijklmnopabcdefghijklmnop", Version: "1.0.0"}); res != want {
		t.Errorf("response: got %+v, want %+v", res, want)
	}
}

func TestClientUpdatePlugin(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{"id": "abcdefghijklmnopabcdefghijklmnop", "version": "1.1.0"}`)
	c := newTestClient(t, ts)
	res, err := c.UpdatePlugin(t.Context(), "abcdefghijklmnopabcdefghijklmnop", "c15b3870-7505-4ab6-9d8d-b9bdbc74f5d6")
	if err != nil {
		t.Fatalf("UpdatePlugin returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPut, "/k/v1/plugin.json",
		`{"id": "abcdefghijklmnopabcdefghijklmnop", "fileKey": "c15b3870-7505-4ab6-9d8d-b9bdbc74f5d6"}`)
	if want := (PluginVersion{ID: "abcdefghijklmnopabcdefghijklmnop", Version: "1.1.0"}); res != want {
		t.Errorf("response: got %+v, want %+v", res, want)
	}
}

func TestClientDeletePlugin(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{}`)
	c := newTestClient(t, ts, WithGuestSpace(4))
	if err := c.DeletePlugin(t.Context(), "abcdefghijklmnopabcdefghijklmnop"); err != nil {
		t.Fatalf("DeletePlugin returned error: %v", err)
	}
	assertRequest(t, got, http.MethodDelete, "/k/v1/plugin.json", `{"id": "abcdefghijklmnopabcdefghijklmnop"}`)
}
