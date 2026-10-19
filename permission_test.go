package kintone

import (
	"net/http"
	"reflect"
	"testing"
)

func TestClientEvaluateRecordPermissions(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{
		"rights": [
			{
				"id": "2",
				"record": {"viewable": true, "editable": false, "deletable": false},
				"fields": {
					"title": {"viewable": true, "editable": false},
					"note": {"viewable": false, "editable": false}
				}
			},
			{
				"id": "3",
				"record": {"viewable": true, "editable": true, "deletable": true},
				"fields": {
					"title": {"viewable": true, "editable": true},
					"note": {"viewable": true, "editable": true}
				}
			}
		]
	}`)
	c := newTestClient(t, ts)
	res, err := c.EvaluateRecordPermissions(t.Context(), 1, []int64{2, 3})
	if err != nil {
		t.Fatalf("EvaluateRecordPermissions returned error: %v", err)
	}
	assertRequest(t, got, http.MethodGet, "/k/v1/records/acl/evaluate.json", `{"app": 1, "ids": [2, 3]}`)
	want := []RecordAccess{
		{
			ID:     2,
			Record: RecordAccessRights{Viewable: true},
			Fields: map[string]FieldAccessRights{
				"title": {Viewable: true},
				"note":  {},
			},
		},
		{
			ID:     3,
			Record: RecordAccessRights{Viewable: true, Editable: true, Deletable: true},
			Fields: map[string]FieldAccessRights{
				"title": {Viewable: true, Editable: true},
				"note":  {Viewable: true, Editable: true},
			},
		},
	}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("response: got %+v, want %+v", res, want)
	}
}

func TestClientGetAppPermissions(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetAppSettingsRequest
		wantPath string
	}{
		{
			name:     "production",
			req:      GetAppSettingsRequest{App: 1, Lang: "en"},
			wantPath: "/k/v1/app/acl.json",
		},
		{
			name:     "preview",
			req:      GetAppSettingsRequest{App: 1, Preview: true},
			wantPath: "/k/v1/preview/app/acl.json",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{
				"rights": [
					{
						"entity": {"type": "CREATOR", "code": null},
						"includeSubs": false,
						"appEditable": true,
						"recordViewable": true,
						"recordAddable": true,
						"recordEditable": true,
						"recordDeletable": true,
						"recordImportable": true,
						"recordExportable": true
					},
					{
						"entity": {"type": "GROUP", "code": "everyone"},
						"includeSubs": false,
						"appEditable": false,
						"recordViewable": true,
						"recordAddable": false,
						"recordEditable": false,
						"recordDeletable": false,
						"recordImportable": false,
						"recordExportable": false
					}
				],
				"revision": "3"
			}`)
			c := newTestClient(t, ts)
			res, err := c.GetAppPermissions(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetAppPermissions returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, tc.wantPath, `{"app": 1}`)
			want := &AppPermissions{
				Rights: []AppRight{
					{
						Entity:           FieldEntity{Type: EntityTypeCreator},
						AppEditable:      true,
						RecordViewable:   true,
						RecordAddable:    true,
						RecordEditable:   true,
						RecordDeletable:  true,
						RecordImportable: true,
						RecordExportable: true,
					},
					{
						Entity:         FieldEntity{Type: EntityTypeGroup, Code: "everyone"},
						RecordViewable: true,
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

func TestClientUpdateAppPermissions(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{"revision": "4"}`)
	c := newTestClient(t, ts)
	revision, err := c.UpdateAppPermissions(t.Context(), 1, []AppRight{
		{Entity: FieldEntity{Type: EntityTypeCreator}, AppEditable: true, RecordViewable: true},
		{Entity: FieldEntity{Type: EntityTypeOrganization, Code: "dev"}, IncludeSubs: true, RecordViewable: true, RecordAddable: true},
	}, 3)
	if err != nil {
		t.Fatalf("UpdateAppPermissions returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPut, "/k/v1/preview/app/acl.json", `{
		"app": 1,
		"rights": [
			{"entity": {"type": "CREATOR"}, "appEditable": true, "recordViewable": true},
			{"entity": {"type": "ORGANIZATION", "code": "dev"}, "includeSubs": true, "recordViewable": true, "recordAddable": true}
		],
		"revision": 3
	}`)
	if revision != 4 {
		t.Errorf("revision: got %d, want %d", revision, 4)
	}
}

func TestClientGetRecordPermissions(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetAppSettingsRequest
		wantPath string
		wantBody string
	}{
		{
			name:     "production",
			req:      GetAppSettingsRequest{App: 1, Lang: "en"},
			wantPath: "/k/v1/record/acl.json",
			wantBody: `{"app": 1, "lang": "en"}`,
		},
		{
			name:     "preview",
			req:      GetAppSettingsRequest{App: 1, Preview: true},
			wantPath: "/k/v1/preview/record/acl.json",
			wantBody: `{"app": 1}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{
				"rights": [
					{
						"filterCond": "status = \"Secret\"",
						"entities": [
							{"entity": {"type": "FIELD_ENTITY", "code": "owner"}, "viewable": true, "editable": true, "deletable": false, "includeSubs": false},
							{"entity": {"type": "GROUP", "code": "everyone"}, "viewable": false, "editable": false, "deletable": false, "includeSubs": false}
						]
					}
				],
				"revision": "3"
			}`)
			c := newTestClient(t, ts)
			res, err := c.GetRecordPermissions(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetRecordPermissions returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, tc.wantPath, tc.wantBody)
			want := &RecordPermissions{
				Rights: []RecordRight{
					{
						FilterCond: `status = "Secret"`,
						Entities: []RecordRightEntity{
							{Entity: FieldEntity{Type: EntityTypeFieldEntity, Code: "owner"}, Viewable: true, Editable: true},
							{Entity: FieldEntity{Type: EntityTypeGroup, Code: "everyone"}},
						},
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

func TestClientUpdateRecordPermissions(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{"revision": "4"}`)
	c := newTestClient(t, ts)
	revision, err := c.UpdateRecordPermissions(t.Context(), 1, []RecordRight{
		{
			Entities: []RecordRightEntity{
				{Entity: FieldEntity{Type: EntityTypeUser, Code: "alice"}, Viewable: true, Deletable: true},
			},
		},
	}, 0)
	if err != nil {
		t.Fatalf("UpdateRecordPermissions returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPut, "/k/v1/preview/record/acl.json", `{
		"app": 1,
		"rights": [{"entities": [{"entity": {"type": "USER", "code": "alice"}, "viewable": true, "deletable": true}]}]
	}`)
	if revision != 4 {
		t.Errorf("revision: got %d, want %d", revision, 4)
	}
}

func TestClientGetFieldPermissions(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetAppSettingsRequest
		wantPath string
	}{
		{
			name:     "production",
			req:      GetAppSettingsRequest{App: 1, Lang: "en"},
			wantPath: "/k/v1/field/acl.json",
		},
		{
			name:     "preview",
			req:      GetAppSettingsRequest{App: 1, Preview: true},
			wantPath: "/k/v1/preview/field/acl.json",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{
				"rights": [
					{
						"code": "salary",
						"entities": [
							{"accessibility": "WRITE", "entity": {"type": "ORGANIZATION", "code": "hr"}, "includeSubs": true},
							{"accessibility": "NONE", "entity": {"type": "GROUP", "code": "everyone"}, "includeSubs": false}
						]
					}
				],
				"revision": "3"
			}`)
			c := newTestClient(t, ts)
			res, err := c.GetFieldPermissions(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetFieldPermissions returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, tc.wantPath, `{"app": 1}`)
			want := &FieldPermissions{
				Rights: []FieldRight{
					{
						Code: "salary",
						Entities: []FieldRightEntity{
							{Accessibility: "WRITE", Entity: FieldEntity{Type: EntityTypeOrganization, Code: "hr"}, IncludeSubs: true},
							{Accessibility: "NONE", Entity: FieldEntity{Type: EntityTypeGroup, Code: "everyone"}},
						},
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

func TestClientUpdateFieldPermissions(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{"revision": "4"}`)
	c := newTestClient(t, ts)
	revision, err := c.UpdateFieldPermissions(t.Context(), 1, []FieldRight{
		{
			Code: "salary",
			Entities: []FieldRightEntity{
				{Accessibility: "READ", Entity: FieldEntity{Type: EntityTypeUser, Code: "alice"}},
			},
		},
	}, 3)
	if err != nil {
		t.Fatalf("UpdateFieldPermissions returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPut, "/k/v1/preview/field/acl.json", `{
		"app": 1,
		"rights": [{"code": "salary", "entities": [{"accessibility": "READ", "entity": {"type": "USER", "code": "alice"}}]}],
		"revision": 3
	}`)
	if revision != 4 {
		t.Errorf("revision: got %d, want %d", revision, 4)
	}
}

func TestClientUpdateRecordPermissionsNil(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{"revision": "4"}`)
	c := newTestClient(t, ts)
	if _, err := c.UpdateRecordPermissions(t.Context(), 1, nil, 0); err != nil {
		t.Fatalf("UpdateRecordPermissions returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPut, "/k/v1/preview/record/acl.json", `{"app": 1, "rights": []}`)
}

func TestClientUpdateFieldPermissionsNil(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{"revision": "4"}`)
	c := newTestClient(t, ts)
	if _, err := c.UpdateFieldPermissions(t.Context(), 1, nil, 0); err != nil {
		t.Fatalf("UpdateFieldPermissions returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPut, "/k/v1/preview/field/acl.json", `{"app": 1, "rights": []}`)
}
