package kintone

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
)

// Field is a field value of a record.
//
// When decoded from a response, Value has the Go type corresponding to Type:
//
//   - string for text, number, calculated, date and time, radio button,
//     drop-down, link, record number, status, $id and $revision fields
//   - []string for check box, multi-choice and category fields
//   - Entity for created by and updated by fields
//   - []Entity for user, department, group selection and assignee fields
//   - []FileInfo for attachment fields
//   - []SubtableRow for table fields
//
// When writing records, Type can be omitted. For example,
//
//	kintone.Record{
//		"title": {Value: "Hello"},
//		"users": {Value: []kintone.Entity{{Code: "alice"}}},
//		"files": {Value: []kintone.FileInfo{{FileKey: fileKey}}},
//	}
type Field struct {
	Type  FieldType `json:"type,omitzero"`
	Value any       `json:"value"`
}

// FieldType is the type of a field.
type FieldType string

// Field types.
const (
	FieldTypeRecordNumber       FieldType = "RECORD_NUMBER"
	FieldTypeID                 FieldType = "__ID__"
	FieldTypeRevision           FieldType = "__REVISION__"
	FieldTypeCreator            FieldType = "CREATOR"
	FieldTypeCreatedTime        FieldType = "CREATED_TIME"
	FieldTypeModifier           FieldType = "MODIFIER"
	FieldTypeUpdatedTime        FieldType = "UPDATED_TIME"
	FieldTypeSingleLineText     FieldType = "SINGLE_LINE_TEXT"
	FieldTypeMultiLineText      FieldType = "MULTI_LINE_TEXT"
	FieldTypeRichText           FieldType = "RICH_TEXT"
	FieldTypeNumber             FieldType = "NUMBER"
	FieldTypeCalc               FieldType = "CALC"
	FieldTypeCheckBox           FieldType = "CHECK_BOX"
	FieldTypeRadioButton        FieldType = "RADIO_BUTTON"
	FieldTypeMultiSelect        FieldType = "MULTI_SELECT"
	FieldTypeDropDown           FieldType = "DROP_DOWN"
	FieldTypeUserSelect         FieldType = "USER_SELECT"
	FieldTypeOrganizationSelect FieldType = "ORGANIZATION_SELECT"
	FieldTypeGroupSelect        FieldType = "GROUP_SELECT"
	FieldTypeDate               FieldType = "DATE"
	FieldTypeTime               FieldType = "TIME"
	FieldTypeDateTime           FieldType = "DATETIME"
	FieldTypeLink               FieldType = "LINK"
	FieldTypeFile               FieldType = "FILE"
	FieldTypeSubtable           FieldType = "SUBTABLE"
	FieldTypeCategory           FieldType = "CATEGORY"
	FieldTypeStatus             FieldType = "STATUS"
	FieldTypeStatusAssignee     FieldType = "STATUS_ASSIGNEE"
	FieldTypeReferenceTable     FieldType = "REFERENCE_TABLE"
	FieldTypeGroup              FieldType = "GROUP"
	FieldTypeLabel              FieldType = "LABEL"
	FieldTypeSpacer             FieldType = "SPACER"
	FieldTypeHR                 FieldType = "HR"
)

// Entity is a user, department or group.
type Entity struct {
	Code string `json:"code"`
	Name string `json:"name,omitzero"`
}

// EntityType is the type of an entity.
type EntityType string

// Entity types.
const (
	EntityTypeUser         EntityType = "USER"
	EntityTypeGroup        EntityType = "GROUP"
	EntityTypeOrganization EntityType = "ORGANIZATION"
	EntityTypeFieldEntity  EntityType = "FIELD_ENTITY"
	EntityTypeCreator      EntityType = "CREATOR"
	EntityTypeCustomField  EntityType = "CUSTOM_FIELD"
)

// FileInfo is a file of an attachment field.
// Only FileKey is required when writing records.
type FileInfo struct {
	ContentType string `json:"contentType,omitzero"`
	FileKey     string `json:"fileKey"`
	Name        string `json:"name,omitzero"`
	Size        int64  `json:"size,omitzero,string"`
}

// SubtableRow is a row of a table field.
type SubtableRow struct {
	ID    string `json:"id,omitzero"`
	Value Record `json:"value"`
}

// UnmarshalJSON decodes the field value into the Go type of the field type.
func (f *Field) UnmarshalJSON(b []byte) error {
	var v struct {
		Type  FieldType      `json:"type"`
		Value jsontext.Value `json:"value"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var err error
	f.Type = v.Type
	switch v.Type {
	case FieldTypeRecordNumber, FieldTypeID, FieldTypeRevision,
		FieldTypeCreatedTime, FieldTypeUpdatedTime,
		FieldTypeSingleLineText, FieldTypeMultiLineText, FieldTypeRichText,
		FieldTypeNumber, FieldTypeCalc, FieldTypeRadioButton, FieldTypeDropDown,
		FieldTypeDate, FieldTypeTime, FieldTypeDateTime, FieldTypeLink,
		FieldTypeStatus:
		f.Value, err = decodeValue[string](v.Value)
	case FieldTypeCheckBox, FieldTypeMultiSelect, FieldTypeCategory:
		f.Value, err = decodeValue[[]string](v.Value)
	case FieldTypeCreator, FieldTypeModifier:
		f.Value, err = decodeValue[Entity](v.Value)
	case FieldTypeUserSelect, FieldTypeOrganizationSelect,
		FieldTypeGroupSelect, FieldTypeStatusAssignee:
		f.Value, err = decodeValue[[]Entity](v.Value)
	case FieldTypeFile:
		f.Value, err = decodeValue[[]FileInfo](v.Value)
	case FieldTypeSubtable:
		f.Value, err = decodeValue[[]SubtableRow](v.Value)
	default:
		f.Value, err = decodeValue[any](v.Value)
	}
	return err
}

func decodeValue[T any](value jsontext.Value) (any, error) {
	var v T
	if len(value) > 0 {
		if err := json.Unmarshal(value, &v); err != nil {
			return nil, err
		}
	}
	return v, nil
}
