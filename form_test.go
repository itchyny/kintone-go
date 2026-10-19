package kintone

import (
	"encoding/json/v2"
	"net/http"
	"reflect"
	"testing"
)

func TestClientGetFormFields(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetAppSettingsRequest
		wantPath string
		wantBody string
	}{
		{
			name:     "production",
			req:      GetAppSettingsRequest{App: 1, Lang: "en"},
			wantPath: "/k/v1/app/form/fields.json",
			wantBody: `{"app": 1, "lang": "en"}`,
		},
		{
			name:     "preview",
			req:      GetAppSettingsRequest{App: 1, Preview: true},
			wantPath: "/k/v1/preview/app/form/fields.json",
			wantBody: `{"app": 1}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{
				"properties": {
					"title": {
						"type": "SINGLE_LINE_TEXT",
						"code": "title",
						"label": "Title",
						"noLabel": false,
						"required": true,
						"unique": false,
						"maxLength": "64",
						"minLength": "",
						"defaultValue": "",
						"expression": "",
						"hideExpression": false
					}
				},
				"revision": "3"
			}`)
			c := newTestClient(t, ts)
			res, err := c.GetFormFields(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetFormFields returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, tc.wantPath, tc.wantBody)
			want := &FormFields{
				Properties: map[string]FieldProperty{
					"title": {
						Type:           FieldTypeSingleLineText,
						Code:           "title",
						Label:          "Title",
						NoLabel:        new(false),
						Required:       new(true),
						Unique:         new(false),
						MaxLength:      new("64"),
						MinLength:      new(""),
						Expression:     new(""),
						DefaultValue:   "",
						HideExpression: new(false),
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

func TestFieldPropertyUnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  FieldProperty
	}{
		{
			name: "options",
			input: `{
				"type": "DROP_DOWN",
				"code": "priority",
				"label": "Priority",
				"defaultValue": "Low",
				"options": {
					"Low": {"label": "Low", "index": "0"},
					"High": {"label": "High", "index": "1"}
				}
			}`,
			want: FieldProperty{
				Type:         FieldTypeDropDown,
				Code:         "priority",
				Label:        "Priority",
				DefaultValue: "Low",
				Options: map[string]FieldOption{
					"Low":  {Label: "Low", Index: 0},
					"High": {Label: "High", Index: 1},
				},
			},
		},
		{
			name: "entities",
			input: `{
				"type": "USER_SELECT",
				"code": "users",
				"entities": [{"code": "alice", "type": "USER"}, {"code": "dev", "type": "ORGANIZATION"}],
				"defaultValue": [{"code": "LOGINUSER()", "type": "FUNCTION"}]
			}`,
			want: FieldProperty{
				Type: FieldTypeUserSelect,
				Code: "users",
				Entities: []FieldEntity{
					{Code: "alice", Type: EntityTypeUser},
					{Code: "dev", Type: EntityTypeOrganization},
				},
				DefaultValue: []any{map[string]any{"code": "LOGINUSER()", "type": "FUNCTION"}},
			},
		},
		{
			name:  "empty entities",
			input: `{"type": "GROUP_SELECT", "code": "groups", "entities": ""}`,
			want:  FieldProperty{Type: FieldTypeGroupSelect, Code: "groups"},
		},
		{
			name: "lookup",
			input: `{
				"type": "SINGLE_LINE_TEXT",
				"code": "customer",
				"lookup": {
					"relatedApp": {"app": "3", "code": ""},
					"relatedKeyField": "name",
					"fieldMappings": [{"field": "phone", "relatedField": "tel"}],
					"lookupPickerFields": ["name", "tel"],
					"filterCond": "",
					"sort": "$id desc"
				}
			}`,
			want: FieldProperty{
				Type: FieldTypeSingleLineText,
				Code: "customer",
				Lookup: &Lookup{
					RelatedApp:         RelatedApp{App: 3},
					RelatedKeyField:    "name",
					FieldMappings:      []FieldMapping{{Field: "phone", RelatedField: "tel"}},
					LookupPickerFields: []string{"name", "tel"},
					FilterCond:         new(""),
					Sort:               new("$id desc"),
				},
			},
		},
		{
			name: "lookup with empty lists",
			input: `{
				"type": "NUMBER",
				"code": "price",
				"lookup": {
					"relatedApp": {"app": "3", "code": "ITEM"},
					"relatedKeyField": "code",
					"fieldMappings": "",
					"lookupPickerFields": ""
				}
			}`,
			want: FieldProperty{
				Type: FieldTypeNumber,
				Code: "price",
				Lookup: &Lookup{
					RelatedApp:      RelatedApp{App: 3, Code: "ITEM"},
					RelatedKeyField: "code",
				},
			},
		},
		{
			name: "reference table",
			input: `{
				"type": "REFERENCE_TABLE",
				"code": "orders",
				"referenceTable": {
					"relatedApp": {"app": "4", "code": ""},
					"condition": {"field": "id", "relatedField": "customer_id"},
					"filterCond": "",
					"displayFields": ["date", "amount"],
					"sort": "date desc",
					"size": "5"
				}
			}`,
			want: FieldProperty{
				Type: FieldTypeReferenceTable,
				Code: "orders",
				ReferenceTable: &ReferenceTable{
					RelatedApp:    RelatedApp{App: 4},
					Condition:     ReferenceTableCondition{Field: "id", RelatedField: "customer_id"},
					DisplayFields: []string{"date", "amount"},
					Sort:          "date desc",
					Size:          "5",
				},
			},
		},
		{
			name: "subtable",
			input: `{
				"type": "SUBTABLE",
				"code": "items",
				"fields": {
					"item": {"type": "SINGLE_LINE_TEXT", "code": "item", "entities": ""}
				}
			}`,
			want: FieldProperty{
				Type: FieldTypeSubtable,
				Code: "items",
				Fields: map[string]FieldProperty{
					"item": {Type: FieldTypeSingleLineText, Code: "item"},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var got FieldProperty
			if err := json.Unmarshal([]byte(tc.input), &got); err != nil {
				t.Fatalf("json.Unmarshal returned error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("json.Unmarshal:\n got: %+v\nwant: %+v", got, tc.want)
			}
		})
	}
}

func TestFieldPropertyUnmarshalJSONError(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"not an object", `"x"`},
		{"invalid entities", `{"type": "USER_SELECT", "entities": "x"}`},
		{"invalid field mappings", `{"type": "NUMBER", "lookup": {"fieldMappings": "x"}}`},
		{"invalid lookup picker fields", `{"type": "NUMBER", "lookup": {"lookupPickerFields": [1]}}`},
		{"invalid lookup", `{"type": "NUMBER", "lookup": "x"}`},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var p FieldProperty
			if err := json.Unmarshal([]byte(tc.input), &p); err == nil {
				t.Errorf("json.Unmarshal(%s) should return error but got %+v", tc.input, p)
			}
		})
	}
}

func TestClientAddFormFields(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{"revision": "4"}`)
	c := newTestClient(t, ts)
	revision, err := c.AddFormFields(t.Context(), 1, map[string]FieldProperty{
		"priority": {
			Type:     FieldTypeRadioButton,
			Code:     "priority",
			Label:    "Priority",
			Required: new(true),
			Options: map[string]FieldOption{
				"Low":  {Label: "Low", Index: 0},
				"High": {Label: "High", Index: 1},
			},
			DefaultValue: "Low",
			Align:        "HORIZONTAL",
		},
	}, 3)
	if err != nil {
		t.Fatalf("AddFormFields returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPost, "/k/v1/preview/app/form/fields.json", `{
		"app": 1,
		"properties": {
			"priority": {
				"type": "RADIO_BUTTON",
				"code": "priority",
				"label": "Priority",
				"required": true,
				"options": {
					"Low": {"label": "Low", "index": "0"},
					"High": {"label": "High", "index": "1"}
				},
				"defaultValue": "Low",
				"align": "HORIZONTAL"
			}
		},
		"revision": 3
	}`)
	if revision != 4 {
		t.Errorf("revision: got %d, want %d", revision, 4)
	}
}

func TestClientUpdateFormFields(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{"revision": "5"}`)
	c := newTestClient(t, ts)
	revision, err := c.UpdateFormFields(t.Context(), 1, map[string]FieldProperty{
		"title":  {Type: FieldTypeSingleLineText, Code: "subject", Required: new(false)},
		"amount": {Type: FieldTypeNumber, MaxValue: new(""), Unit: new("")},
		"owner":  {Type: FieldTypeUserSelect, Entities: []FieldEntity{}},
		"customer": {Type: FieldTypeSingleLineText, Lookup: &Lookup{
			RelatedApp: RelatedApp{App: 3}, RelatedKeyField: "name",
			LookupPickerFields: []string{}, FilterCond: new(""),
		}},
	}, 0)
	if err != nil {
		t.Fatalf("UpdateFormFields returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPut, "/k/v1/preview/app/form/fields.json", `{
		"app": 1,
		"properties": {
			"title": {"type": "SINGLE_LINE_TEXT", "code": "subject", "required": false},
			"amount": {"type": "NUMBER", "maxValue": "", "unit": ""},
			"owner": {"type": "USER_SELECT", "entities": []},
			"customer": {
				"type": "SINGLE_LINE_TEXT",
				"lookup": {
					"relatedApp": {"app": "3"}, "relatedKeyField": "name",
					"lookupPickerFields": [], "filterCond": ""
				}
			}
		}
	}`)
	if revision != 5 {
		t.Errorf("revision: got %d, want %d", revision, 5)
	}
}

func TestClientDeleteFormFields(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{"revision": "6"}`)
	c := newTestClient(t, ts)
	revision, err := c.DeleteFormFields(t.Context(), 1, []string{"title", "priority"}, 5)
	if err != nil {
		t.Fatalf("DeleteFormFields returned error: %v", err)
	}
	assertRequest(t, got, http.MethodDelete, "/k/v1/preview/app/form/fields.json",
		`{"app": 1, "fields": ["title", "priority"], "revision": 5}`)
	if revision != 6 {
		t.Errorf("revision: got %d, want %d", revision, 6)
	}
}

func TestClientGetFormLayout(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetAppSettingsRequest
		wantPath string
	}{
		{
			name:     "production",
			req:      GetAppSettingsRequest{App: 1, Lang: "en"},
			wantPath: "/k/v1/app/form/layout.json",
		},
		{
			name:     "preview",
			req:      GetAppSettingsRequest{App: 1, Preview: true},
			wantPath: "/k/v1/preview/app/form/layout.json",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{
				"layout": [
					{
						"type": "ROW",
						"fields": [
							{"type": "SINGLE_LINE_TEXT", "code": "title", "size": {"width": "200"}},
							{"type": "LABEL", "label": "Note", "elementId": "note", "size": {"width": "100"}}
						]
					},
					{
						"type": "SUBTABLE",
						"code": "items",
						"fields": [
							{"type": "MULTI_LINE_TEXT", "code": "detail", "size": {"width": "300", "innerHeight": "80"}}
						]
					},
					{
						"type": "GROUP",
						"code": "group",
						"layout": [
							{"type": "ROW", "fields": [{"type": "SPACER", "elementId": "space", "size": {"width": "50", "height": "30"}}]}
						]
					}
				],
				"revision": "3"
			}`)
			c := newTestClient(t, ts)
			res, err := c.GetFormLayout(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetFormLayout returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, tc.wantPath, `{"app": 1}`)
			want := &FormLayout{
				Layout: []LayoutRow{
					{
						Type: LayoutTypeRow,
						Fields: []LayoutField{
							{Type: FieldTypeSingleLineText, Code: "title", Size: &FieldSize{Width: "200"}},
							{Type: FieldTypeLabel, Label: "Note", ElementID: "note", Size: &FieldSize{Width: "100"}},
						},
					},
					{
						Type: LayoutTypeSubtable,
						Code: "items",
						Fields: []LayoutField{
							{Type: FieldTypeMultiLineText, Code: "detail", Size: &FieldSize{Width: "300", InnerHeight: "80"}},
						},
					},
					{
						Type: LayoutTypeGroup,
						Code: "group",
						Layout: []LayoutRow{
							{
								Type: LayoutTypeRow,
								Fields: []LayoutField{
									{Type: FieldTypeSpacer, ElementID: "space", Size: &FieldSize{Width: "50", Height: "30"}},
								},
							},
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

func TestClientUpdateFormLayout(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{"revision": "4"}`)
	c := newTestClient(t, ts)
	revision, err := c.UpdateFormLayout(t.Context(), 1, []LayoutRow{
		{
			Type: LayoutTypeRow,
			Fields: []LayoutField{
				{Type: FieldTypeSingleLineText, Code: "title", Size: &FieldSize{Width: "200"}},
				{Type: FieldTypeHR, ElementID: "line"},
			},
		},
		{
			Type: LayoutTypeGroup,
			Code: "group",
			Layout: []LayoutRow{
				{Type: LayoutTypeRow, Fields: []LayoutField{{Type: FieldTypeNumber, Code: "price"}}},
			},
		},
	}, 3)
	if err != nil {
		t.Fatalf("UpdateFormLayout returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPut, "/k/v1/preview/app/form/layout.json", `{
		"app": 1,
		"layout": [
			{
				"type": "ROW",
				"fields": [
					{"type": "SINGLE_LINE_TEXT", "code": "title", "size": {"width": "200"}},
					{"type": "HR", "elementId": "line"}
				]
			},
			{
				"type": "GROUP",
				"code": "group",
				"layout": [
					{"type": "ROW", "fields": [{"type": "NUMBER", "code": "price"}]}
				]
			}
		],
		"revision": 3
	}`)
	if revision != 4 {
		t.Errorf("revision: got %d, want %d", revision, 4)
	}
}
