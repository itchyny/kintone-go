package kintone

import "encoding/json"

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
	Type  FieldType `json:"type,omitempty"`
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
	Name string `json:"name,omitempty"`
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
	ContentType string `json:"contentType,omitempty"`
	FileKey     string `json:"fileKey"`
	Name        string `json:"name,omitempty"`
	Size        int64  `json:"size,omitempty,string"`
}

// SubtableRow is a row of a table field.
type SubtableRow struct {
	ID    string `json:"id,omitempty"`
	Value Record `json:"value"`
}

// UnmarshalJSON decodes the field value into the Go type of the field type.
func (f *Field) UnmarshalJSON(b []byte) error {
	var raw struct {
		Type  FieldType       `json:"type"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	var err error
	f.Type = raw.Type
	switch raw.Type {
	case FieldTypeRecordNumber, FieldTypeID, FieldTypeRevision,
		FieldTypeCreatedTime, FieldTypeUpdatedTime,
		FieldTypeSingleLineText, FieldTypeMultiLineText, FieldTypeRichText,
		FieldTypeNumber, FieldTypeCalc, FieldTypeRadioButton, FieldTypeDropDown,
		FieldTypeDate, FieldTypeTime, FieldTypeDateTime, FieldTypeLink,
		FieldTypeStatus:
		f.Value, err = decodeValue[string](raw.Value)
	case FieldTypeCheckBox, FieldTypeMultiSelect, FieldTypeCategory:
		f.Value, err = decodeValue[[]string](raw.Value)
	case FieldTypeCreator, FieldTypeModifier:
		f.Value, err = decodeValue[Entity](raw.Value)
	case FieldTypeUserSelect, FieldTypeOrganizationSelect,
		FieldTypeGroupSelect, FieldTypeStatusAssignee:
		f.Value, err = decodeValue[[]Entity](raw.Value)
	case FieldTypeFile:
		f.Value, err = decodeValue[[]FileInfo](raw.Value)
	case FieldTypeSubtable:
		f.Value, err = decodeValue[[]SubtableRow](raw.Value)
	default:
		f.Value, err = decodeValue[any](raw.Value)
	}
	return err
}

func decodeValue[T any](b json.RawMessage) (any, error) {
	var v T
	if len(b) > 0 {
		if err := json.Unmarshal(b, &v); err != nil {
			return nil, err
		}
	}
	return v, nil
}
