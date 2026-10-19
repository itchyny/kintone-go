package kintone

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"reflect"
	"testing"
)

func TestRecordUnmarshalJSON(t *testing.T) {
	input := `{
		"$id": {"type": "__ID__", "value": "1"},
		"$revision": {"type": "__REVISION__", "value": "5"},
		"record_number": {"type": "RECORD_NUMBER", "value": "APP-1"},
		"creator": {"type": "CREATOR", "value": {"code": "alice", "name": "Alice"}},
		"created_time": {"type": "CREATED_TIME", "value": "2026-10-16T01:23:00Z"},
		"modifier": {"type": "MODIFIER", "value": {"code": "bob", "name": "Bob"}},
		"updated_time": {"type": "UPDATED_TIME", "value": "2026-10-16T04:56:00Z"},
		"text": {"type": "SINGLE_LINE_TEXT", "value": "Hello"},
		"multi_line_text": {"type": "MULTI_LINE_TEXT", "value": "Hello\nworld"},
		"rich_text": {"type": "RICH_TEXT", "value": "<b>Hello</b>"},
		"number": {"type": "NUMBER", "value": "12.5"},
		"empty_number": {"type": "NUMBER", "value": ""},
		"calc": {"type": "CALC", "value": "25"},
		"check_box": {"type": "CHECK_BOX", "value": ["a", "b"]},
		"empty_check_box": {"type": "CHECK_BOX", "value": []},
		"radio_button": {"type": "RADIO_BUTTON", "value": "a"},
		"multi_select": {"type": "MULTI_SELECT", "value": ["a"]},
		"drop_down": {"type": "DROP_DOWN", "value": "b"},
		"empty_drop_down": {"type": "DROP_DOWN", "value": null},
		"user_select": {"type": "USER_SELECT", "value": [{"code": "alice", "name": "Alice"}]},
		"organization_select": {"type": "ORGANIZATION_SELECT", "value": [{"code": "dev", "name": "Development"}]},
		"group_select": {"type": "GROUP_SELECT", "value": [{"code": "everyone", "name": "Everyone"}]},
		"date": {"type": "DATE", "value": "2026-10-16"},
		"empty_date": {"type": "DATE", "value": null},
		"time": {"type": "TIME", "value": "12:34"},
		"datetime": {"type": "DATETIME", "value": "2026-10-16T03:34:00Z"},
		"link": {"type": "LINK", "value": "https://example.com"},
		"file": {"type": "FILE", "value": [{"contentType": "text/plain", "fileKey": "key1", "name": "a.txt", "size": "123"}]},
		"category": {"type": "CATEGORY", "value": ["c1"]},
		"status": {"type": "STATUS", "value": "In progress"},
		"assignee": {"type": "STATUS_ASSIGNEE", "value": [{"code": "bob", "name": "Bob"}]},
		"table": {"type": "SUBTABLE", "value": [
			{"id": "10", "value": {
				"item": {"type": "SINGLE_LINE_TEXT", "value": "apple"},
				"price": {"type": "NUMBER", "value": "100"},
				"users": {"type": "USER_SELECT", "value": []}
			}}
		]},
		"unknown": {"type": "UNKNOWN_TYPE", "value": {"x": [1, "y"]}}
	}`
	want := Record{
		"$id":                 {Type: FieldTypeID, Value: "1"},
		"$revision":           {Type: FieldTypeRevision, Value: "5"},
		"record_number":       {Type: FieldTypeRecordNumber, Value: "APP-1"},
		"creator":             {Type: FieldTypeCreator, Value: Entity{Code: "alice", Name: "Alice"}},
		"created_time":        {Type: FieldTypeCreatedTime, Value: "2026-10-16T01:23:00Z"},
		"modifier":            {Type: FieldTypeModifier, Value: Entity{Code: "bob", Name: "Bob"}},
		"updated_time":        {Type: FieldTypeUpdatedTime, Value: "2026-10-16T04:56:00Z"},
		"text":                {Type: FieldTypeSingleLineText, Value: "Hello"},
		"multi_line_text":     {Type: FieldTypeMultiLineText, Value: "Hello\nworld"},
		"rich_text":           {Type: FieldTypeRichText, Value: "<b>Hello</b>"},
		"number":              {Type: FieldTypeNumber, Value: "12.5"},
		"empty_number":        {Type: FieldTypeNumber, Value: ""},
		"calc":                {Type: FieldTypeCalc, Value: "25"},
		"check_box":           {Type: FieldTypeCheckBox, Value: []string{"a", "b"}},
		"empty_check_box":     {Type: FieldTypeCheckBox, Value: []string{}},
		"radio_button":        {Type: FieldTypeRadioButton, Value: "a"},
		"multi_select":        {Type: FieldTypeMultiSelect, Value: []string{"a"}},
		"drop_down":           {Type: FieldTypeDropDown, Value: "b"},
		"empty_drop_down":     {Type: FieldTypeDropDown, Value: ""},
		"user_select":         {Type: FieldTypeUserSelect, Value: []Entity{{Code: "alice", Name: "Alice"}}},
		"organization_select": {Type: FieldTypeOrganizationSelect, Value: []Entity{{Code: "dev", Name: "Development"}}},
		"group_select":        {Type: FieldTypeGroupSelect, Value: []Entity{{Code: "everyone", Name: "Everyone"}}},
		"date":                {Type: FieldTypeDate, Value: "2026-10-16"},
		"empty_date":          {Type: FieldTypeDate, Value: ""},
		"time":                {Type: FieldTypeTime, Value: "12:34"},
		"datetime":            {Type: FieldTypeDateTime, Value: "2026-10-16T03:34:00Z"},
		"link":                {Type: FieldTypeLink, Value: "https://example.com"},
		"file": {Type: FieldTypeFile, Value: []FileInfo{
			{ContentType: "text/plain", FileKey: "key1", Name: "a.txt", Size: 123},
		}},
		"category": {Type: FieldTypeCategory, Value: []string{"c1"}},
		"status":   {Type: FieldTypeStatus, Value: "In progress"},
		"assignee": {Type: FieldTypeStatusAssignee, Value: []Entity{{Code: "bob", Name: "Bob"}}},
		"table": {Type: FieldTypeSubtable, Value: []SubtableRow{
			{ID: "10", Value: Record{
				"item":  {Type: FieldTypeSingleLineText, Value: "apple"},
				"price": {Type: FieldTypeNumber, Value: "100"},
				"users": {Type: FieldTypeUserSelect, Value: []Entity{}},
			}},
		}},
		"unknown": {Type: "UNKNOWN_TYPE", Value: map[string]any{"x": []any{1.0, "y"}}},
	}
	var got Record
	if err := json.Unmarshal([]byte(input), &got); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	for code, field := range want {
		if !reflect.DeepEqual(got[code], field) {
			t.Errorf("field %s: got %#v, want %#v", code, got[code], field)
		}
	}
	for code := range got {
		if _, ok := want[code]; !ok {
			t.Errorf("unexpected field %s", code)
		}
	}
}

func TestFieldUnmarshalJSONError(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"not an object", `"x"`},
		{"string field with array", `{"type": "SINGLE_LINE_TEXT", "value": ["x"]}`},
		{"check box field with string", `{"type": "CHECK_BOX", "value": "x"}`},
		{"user field with object", `{"type": "USER_SELECT", "value": {"code": "x"}}`},
		{"file size not string", `{"type": "FILE", "value": [{"fileKey": "x", "size": 1}]}`},
		{"subtable field with invalid row", `{"type": "SUBTABLE", "value": [{"value": {"x": 1}}]}`},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var f Field
			if err := json.Unmarshal([]byte(tc.input), &f); err == nil {
				t.Errorf("json.Unmarshal(%s) should return error but got %#v", tc.input, f)
			}
		})
	}
}

func TestRecordMarshalJSON(t *testing.T) {
	record := Record{
		"text":   {Value: "Hello"},
		"number": {Type: FieldTypeNumber, Value: "10"},
		"users":  {Value: []Entity{{Code: "alice"}}},
		"files":  {Value: []FileInfo{{FileKey: "key1"}}},
		"table": {Value: []SubtableRow{
			{ID: "10", Value: Record{"item": {Value: "apple"}}},
			{Value: Record{"item": {Value: "banana"}}},
		}},
	}
	want := `{
	"files": {
		"value": [
			{
				"fileKey": "key1"
			}
		]
	},
	"number": {
		"type": "NUMBER",
		"value": "10"
	},
	"table": {
		"value": [
			{
				"id": "10",
				"value": {
					"item": {
						"value": "apple"
					}
				}
			},
			{
				"value": {
					"item": {
						"value": "banana"
					}
				}
			}
		]
	},
	"text": {
		"value": "Hello"
	},
	"users": {
		"value": [
			{
				"code": "alice"
			}
		]
	}
}`
	got, err := json.Marshal(record, json.Deterministic(true), jsontext.WithIndent("\t"))
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	if string(got) != want {
		t.Errorf("json.Marshal:\n got: %s\nwant: %s", got, want)
	}
}
