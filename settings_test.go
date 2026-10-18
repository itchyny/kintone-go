package kintone

import (
	"net/http"
	"reflect"
	"testing"
)

func TestClientGetGeneralSettings(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetAppSettingsRequest
		wantPath string
		wantBody string
	}{
		{
			name:     "production",
			req:      GetAppSettingsRequest{App: 1, Lang: "en"},
			wantPath: "/k/v1/app/settings.json",
			wantBody: `{"app": 1, "lang": "en"}`,
		},
		{
			name:     "preview",
			req:      GetAppSettingsRequest{App: 1, Preview: true},
			wantPath: "/k/v1/preview/app/settings.json",
			wantBody: `{"app": 1}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{
				"name": "Deals",
				"description": "<b>Deals</b>",
				"icon": {"type": "FILE", "file": {"contentType": "image/png", "fileKey": "key1", "name": "icon.png", "size": "1024"}},
				"theme": "WHITE",
				"titleField": {"selectionMode": "MANUAL", "code": "title"},
				"enableThumbnails": true,
				"enableBulkDeletion": false,
				"enableComments": true,
				"enableDuplicateRecord": true,
				"enableInlineRecordEditing": false,
				"numberPrecision": {"digits": "16", "decimalPlaces": "4", "roundingMode": "HALF_EVEN"},
				"firstMonthOfFiscalYear": "4",
				"revision": "3"
			}`)
			c := newTestClient(t, ts)
			res, err := c.GetGeneralSettings(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetGeneralSettings returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, tc.wantPath, tc.wantBody)
			want := &GeneralSettings{
				Name:        "Deals",
				Description: "<b>Deals</b>",
				Icon: &AppIcon{
					Type: "FILE",
					File: &FileInfo{ContentType: "image/png", FileKey: "key1", Name: "icon.png", Size: 1024},
				},
				Theme:                     "WHITE",
				TitleField:                &TitleField{SelectionMode: "MANUAL", Code: "title"},
				EnableThumbnails:          new(true),
				EnableBulkDeletion:        new(false),
				EnableComments:            new(true),
				EnableDuplicateRecord:     new(true),
				EnableInlineRecordEditing: new(false),
				NumberPrecision:           &NumberPrecision{Digits: "16", DecimalPlaces: "4", RoundingMode: "HALF_EVEN"},
				FirstMonthOfFiscalYear:    "4",
				Revision:                  3,
			}
			if !reflect.DeepEqual(res, want) {
				t.Errorf("response: got %+v, want %+v", res, want)
			}
		})
	}
}

func TestClientUpdateGeneralSettings(t *testing.T) {
	testCases := []struct {
		name     string
		settings GeneralSettings
		revision int64
		wantBody string
	}{
		{
			name: "some settings with revision",
			settings: GeneralSettings{
				Name:           "Deals",
				Icon:           &AppIcon{Type: "PRESET", Key: "APP72"},
				EnableComments: new(false),
				Revision:       9,
			},
			revision: 3,
			wantBody: `{
				"app": 1,
				"name": "Deals",
				"icon": {"type": "PRESET", "key": "APP72"},
				"enableComments": false,
				"revision": 3
			}`,
		},
		{
			name:     "revision of the settings is not used",
			settings: GeneralSettings{Theme: "BLUE", Revision: 9},
			wantBody: `{"app": 1, "theme": "BLUE"}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{"revision": "4"}`)
			c := newTestClient(t, ts)
			revision, err := c.UpdateGeneralSettings(t.Context(), 1, tc.settings, tc.revision)
			if err != nil {
				t.Fatalf("UpdateGeneralSettings returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPut, "/k/v1/preview/app/settings.json", tc.wantBody)
			if revision != 4 {
				t.Errorf("revision: got %d, want %d", revision, 4)
			}
		})
	}
}

func TestClientGetAdminNotes(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetAppSettingsRequest
		wantPath string
	}{
		{
			name:     "production",
			req:      GetAppSettingsRequest{App: 1, Lang: "en"},
			wantPath: "/k/v1/app/adminNotes.json",
		},
		{
			name:     "preview",
			req:      GetAppSettingsRequest{App: 1, Preview: true},
			wantPath: "/k/v1/preview/app/adminNotes.json",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK,
				`{"content": "<p>Notes</p>", "includeInTemplateAndDuplicates": true, "revision": "3"}`)
			c := newTestClient(t, ts)
			res, err := c.GetAdminNotes(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetAdminNotes returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, tc.wantPath, `{"app": 1}`)
			want := &AdminNotes{Content: "<p>Notes</p>", IncludeInTemplateAndDuplicates: new(true), Revision: 3}
			if !reflect.DeepEqual(res, want) {
				t.Errorf("response: got %+v, want %+v", res, want)
			}
		})
	}
}

func TestClientUpdateAdminNotes(t *testing.T) {
	testCases := []struct {
		name     string
		notes    AdminNotes
		revision int64
		wantBody string
	}{
		{
			name:     "content and flag with revision",
			notes:    AdminNotes{Content: "<p>Notes</p>", IncludeInTemplateAndDuplicates: new(false)},
			revision: 3,
			wantBody: `{"app": 1, "content": "<p>Notes</p>", "includeInTemplateAndDuplicates": false, "revision": 3}`,
		},
		{
			name:     "empty content",
			notes:    AdminNotes{Revision: 9},
			wantBody: `{"app": 1, "content": ""}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{"revision": "4"}`)
			c := newTestClient(t, ts)
			revision, err := c.UpdateAdminNotes(t.Context(), 1, tc.notes, tc.revision)
			if err != nil {
				t.Fatalf("UpdateAdminNotes returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPut, "/k/v1/preview/app/adminNotes.json", tc.wantBody)
			if revision != 4 {
				t.Errorf("revision: got %d, want %d", revision, 4)
			}
		})
	}
}
