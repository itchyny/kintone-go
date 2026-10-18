package kintone

import (
	"context"
	"encoding/json"
	"net/http"
)

// FieldProperty is the setting of a field in the form.
//
// When updating fields, the properties not set are not changed.
type FieldProperty struct {
	Type            FieldType                `json:"type"`
	Code            string                   `json:"code,omitempty"`
	Label           string                   `json:"label,omitempty"`
	NoLabel         *bool                    `json:"noLabel,omitempty"`
	Required        *bool                    `json:"required,omitempty"`
	Unique          *bool                    `json:"unique,omitempty"`
	MaxValue        string                   `json:"maxValue,omitempty"`
	MinValue        string                   `json:"minValue,omitempty"`
	MaxLength       string                   `json:"maxLength,omitempty"`
	MinLength       string                   `json:"minLength,omitempty"`
	DefaultValue    any                      `json:"defaultValue,omitempty"`
	DefaultNowValue *bool                    `json:"defaultNowValue,omitempty"`
	Options         map[string]FieldOption   `json:"options,omitempty"`
	Align           string                   `json:"align,omitempty"`
	Expression      string                   `json:"expression,omitempty"`
	HideExpression  *bool                    `json:"hideExpression,omitempty"`
	Digit           *bool                    `json:"digit,omitempty"`
	ThumbnailSize   string                   `json:"thumbnailSize,omitempty"`
	Protocol        string                   `json:"protocol,omitempty"`
	Format          string                   `json:"format,omitempty"`
	DisplayScale    string                   `json:"displayScale,omitempty"`
	Unit            string                   `json:"unit,omitempty"`
	UnitPosition    string                   `json:"unitPosition,omitempty"`
	Entities        []FieldEntity            `json:"entities,omitempty"`
	ReferenceTable  *ReferenceTable          `json:"referenceTable,omitempty"`
	Lookup          *Lookup                  `json:"lookup,omitempty"`
	OpenGroup       *bool                    `json:"openGroup,omitempty"`
	Fields          map[string]FieldProperty `json:"fields,omitempty"`
	Enabled         *bool                    `json:"enabled,omitempty"`
}

// FieldOption is an option of a check box, radio button, multi-choice or
// drop-down field.
type FieldOption struct {
	Label string `json:"label"`
	Index int    `json:"index,string"`
}

// FieldEntity is an entity specified by the type and the code, such as a user,
// department or group of the options of a user selection field.
type FieldEntity struct {
	Code string     `json:"code,omitempty"`
	Type EntityType `json:"type"`
}

// ReferenceTable is the setting of a related records field.
type ReferenceTable struct {
	RelatedApp    RelatedApp              `json:"relatedApp"`
	Condition     ReferenceTableCondition `json:"condition"`
	FilterCond    string                  `json:"filterCond,omitempty"`
	DisplayFields []string                `json:"displayFields,omitempty"`
	Sort          string                  `json:"sort,omitempty"`
	Size          string                  `json:"size,omitempty"`
}

// ReferenceTableCondition is the condition of the records to show in a
// related records field.
type ReferenceTableCondition struct {
	Field        string `json:"field"`
	RelatedField string `json:"relatedField"`
}

// RelatedApp is an app referenced by a related records or lookup field,
// specified by either App or Code.
type RelatedApp struct {
	App  int64  `json:"app,string,omitempty"`
	Code string `json:"code,omitempty"`
}

// Lookup is the setting of a lookup field.
type Lookup struct {
	RelatedApp         RelatedApp     `json:"relatedApp"`
	RelatedKeyField    string         `json:"relatedKeyField"`
	FieldMappings      []FieldMapping `json:"fieldMappings,omitempty"`
	LookupPickerFields []string       `json:"lookupPickerFields,omitempty"`
	FilterCond         string         `json:"filterCond,omitempty"`
	Sort               string         `json:"sort,omitempty"`
}

// FieldMapping is a mapping of a field copied by a lookup field.
type FieldMapping struct {
	Field        string `json:"field"`
	RelatedField string `json:"relatedField"`
}

// UnmarshalJSON decodes the field property, accepting an empty string for
// the entities.
func (p *FieldProperty) UnmarshalJSON(b []byte) error {
	type property FieldProperty
	var v struct {
		property
		Entities json.RawMessage `json:"entities"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*p = FieldProperty(v.property)
	return unmarshalList(v.Entities, &p.Entities)
}

// UnmarshalJSON decodes the lookup setting, accepting an empty string for
// the field mappings and the lookup picker fields.
func (l *Lookup) UnmarshalJSON(b []byte) error {
	type lookup Lookup
	var v struct {
		lookup
		FieldMappings      json.RawMessage `json:"fieldMappings"`
		LookupPickerFields json.RawMessage `json:"lookupPickerFields"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*l = Lookup(v.lookup)
	if err := unmarshalList(v.FieldMappings, &l.FieldMappings); err != nil {
		return err
	}
	return unmarshalList(v.LookupPickerFields, &l.LookupPickerFields)
}

func unmarshalList[T any](b json.RawMessage, v *[]T) error {
	if len(b) == 0 || string(b) == `""` {
		return nil
	}
	return json.Unmarshal(b, v)
}

// FormFields is the fields of an app.
type FormFields struct {
	Properties map[string]FieldProperty `json:"properties"`
	Revision   int64                    `json:"revision,string"`
}

// GetFormFields gets the fields of an app, keyed by the field code.
func (c *Client) GetFormFields(ctx context.Context, req GetAppSettingsRequest) (*FormFields, error) {
	return c.getAppSettings[FormFields](ctx, "form/fields.json", req.Preview, req)
}

// AddFormFields adds fields to an app in the preview environment and returns
// the new revision. The revision is not checked if revision is zero.
func (c *Client) AddFormFields(ctx context.Context, app int64, properties map[string]FieldProperty, revision int64) (int64, error) {
	return c.updateFormFields(ctx, http.MethodPost, app, properties, revision)
}

// UpdateFormFields updates fields of an app in the preview environment and
// returns the new revision. The keys of properties are the field codes before
// the update. The revision is not checked if revision is zero.
func (c *Client) UpdateFormFields(ctx context.Context, app int64, properties map[string]FieldProperty, revision int64) (int64, error) {
	return c.updateFormFields(ctx, http.MethodPut, app, properties, revision)
}

func (c *Client) updateFormFields(ctx context.Context, method string, app int64, properties map[string]FieldProperty, revision int64) (int64, error) {
	params := struct {
		App        int64                    `json:"app"`
		Properties map[string]FieldProperty `json:"properties"`
		Revision   int64                    `json:"revision,omitempty"`
	}{app, properties, revision}
	return c.updateAppSettings(ctx, method, "form/fields.json", params)
}

// DeleteFormFields deletes fields of an app in the preview environment and
// returns the new revision. The revision is not checked if revision is zero.
func (c *Client) DeleteFormFields(ctx context.Context, app int64, fields []string, revision int64) (int64, error) {
	params := struct {
		App      int64    `json:"app"`
		Fields   []string `json:"fields"`
		Revision int64    `json:"revision,omitempty"`
	}{app, fields, revision}
	return c.updateAppSettings(ctx, http.MethodDelete, "form/fields.json", params)
}

// FormLayout is the layout of the form of an app.
type FormLayout struct {
	Layout   []LayoutRow `json:"layout"`
	Revision int64       `json:"revision,string"`
}

// LayoutRow is a row of the form layout. Code is the field code of the table
// or the group, and Layout is the rows in the group.
type LayoutRow struct {
	Type   LayoutType    `json:"type"`
	Code   string        `json:"code,omitempty"`
	Fields []LayoutField `json:"fields,omitempty"`
	Layout []LayoutRow   `json:"layout,omitempty"`
}

// LayoutType is the type of a layout row.
type LayoutType string

// Layout types.
const (
	LayoutTypeRow      LayoutType = "ROW"
	LayoutTypeSubtable LayoutType = "SUBTABLE"
	LayoutTypeGroup    LayoutType = "GROUP"
)

// LayoutField is a field in a layout row. Label is the text of a label field,
// and ElementID is the element ID of a label, spacer or border field.
type LayoutField struct {
	Type      FieldType  `json:"type"`
	Code      string     `json:"code,omitempty"`
	Label     string     `json:"label,omitempty"`
	ElementID string     `json:"elementId,omitempty"`
	Size      *FieldSize `json:"size,omitempty"`
}

// FieldSize is the size of a field in pixels. Height includes the field name,
// and InnerHeight does not.
type FieldSize struct {
	Width       string `json:"width,omitempty"`
	Height      string `json:"height,omitempty"`
	InnerHeight string `json:"innerHeight,omitempty"`
}

// GetFormLayout gets the layout of the form of an app. Lang of the request is
// not used.
func (c *Client) GetFormLayout(ctx context.Context, req GetAppSettingsRequest) (*FormLayout, error) {
	params := struct {
		App int64 `json:"app"`
	}{req.App}
	return c.getAppSettings[FormLayout](ctx, "form/layout.json", req.Preview, params)
}

// UpdateFormLayout updates the layout of the form of an app in the preview
// environment and returns the new revision. The revision is not checked if
// revision is zero.
func (c *Client) UpdateFormLayout(ctx context.Context, app int64, layout []LayoutRow, revision int64) (int64, error) {
	params := struct {
		App      int64       `json:"app"`
		Layout   []LayoutRow `json:"layout"`
		Revision int64       `json:"revision,omitempty"`
	}{app, layout, revision}
	return c.updateAppSettings(ctx, http.MethodPut, "form/layout.json", params)
}
