package kintone

import (
	"net/http"
	"reflect"
	"testing"
)

func TestClientGetCustomization(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetAppSettingsRequest
		wantPath string
	}{
		{
			name:     "production",
			req:      GetAppSettingsRequest{App: 1, Lang: "en"},
			wantPath: "/k/v1/app/customize.json",
		},
		{
			name:     "preview",
			req:      GetAppSettingsRequest{App: 1, Preview: true},
			wantPath: "/k/v1/preview/app/customize.json",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{
				"scope": "ALL",
				"desktop": {
					"js": [
						{"type": "URL", "url": "https://example.com/app.js"},
						{"type": "FILE", "file": {"contentType": "text/javascript", "fileKey": "key1", "name": "app.js", "size": "100"}}
					],
					"css": []
				},
				"mobile": {"js": [], "css": []},
				"revision": "3"
			}`)
			c := newTestClient(t, ts)
			res, err := c.GetCustomization(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetCustomization returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, tc.wantPath, `{"app": 1}`)
			want := &Customization{
				Scope: CustomizationScopeAll,
				Desktop: &CustomizationFiles{
					JS: []CustomizationFile{
						{Type: CustomizationFileTypeURL, URL: "https://example.com/app.js"},
						{Type: CustomizationFileTypeFile, File: &FileInfo{ContentType: "text/javascript", FileKey: "key1", Name: "app.js", Size: 100}},
					},
					CSS: []CustomizationFile{},
				},
				Mobile:   &CustomizationFiles{JS: []CustomizationFile{}, CSS: []CustomizationFile{}},
				Revision: 3,
			}
			if !reflect.DeepEqual(res, want) {
				t.Errorf("response: got %+v, want %+v", res, want)
			}
		})
	}
}

func TestClientUpdateCustomization(t *testing.T) {
	testCases := []struct {
		name          string
		customization Customization
		revision      int64
		wantBody      string
	}{
		{
			name: "files",
			customization: Customization{
				Scope: CustomizationScopeAdmin,
				Desktop: &CustomizationFiles{
					JS: []CustomizationFile{
						{Type: CustomizationFileTypeURL, URL: "https://example.com/app.js"},
						{Type: CustomizationFileTypeFile, File: &FileInfo{FileKey: "key2"}},
					},
				},
				Revision: 9,
			},
			revision: 3,
			wantBody: `{
				"app": 1,
				"scope": "ADMIN",
				"desktop": {"js": [
					{"type": "URL", "url": "https://example.com/app.js"},
					{"type": "FILE", "file": {"fileKey": "key2"}}
				]},
				"revision": 3
			}`,
		},
		{
			name: "remove files",
			customization: Customization{
				Mobile: &CustomizationFiles{JS: []CustomizationFile{}, CSS: []CustomizationFile{}},
			},
			wantBody: `{"app": 1, "mobile": {"js": [], "css": []}}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{"revision": "4"}`)
			c := newTestClient(t, ts)
			revision, err := c.UpdateCustomization(t.Context(), 1, tc.customization, tc.revision)
			if err != nil {
				t.Fatalf("UpdateCustomization returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPut, "/k/v1/preview/app/customize.json", tc.wantBody)
			if revision != 4 {
				t.Errorf("revision: got %d, want %d", revision, 4)
			}
		})
	}
}
