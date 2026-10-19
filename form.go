package kintone

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"net/http"
)

// FieldProperty is the setting of a field in the form.
type FieldProperty struct {
	Type            FieldType                `json:"type"`
	Code            string                   `json:"code,omitzero"`
	Label           string                   `json:"label,omitzero"`
	NoLabel         *bool                    `json:"noLabel,omitzero"`
	Required        *bool                    `json:"required,omitzero"`
	Unique          *bool                    `json:"unique,omitzero"`
	MaxValue        *string                  `json:"maxValue,omitzero"`
	MinValue        *string                  `json:"minValue,omitzero"`
	MaxLength       *string                  `json:"maxLength,omitzero"`
	MinLength       *string                  `json:"minLength,omitzero"`
	DefaultValue    any                      `json:"defaultValue,omitzero"`
	DefaultNowValue *bool                    `json:"defaultNowValue,omitzero"`
	Options         map[string]FieldOption   `json:"options,omitzero"`
	Align           Alignment                `json:"align,omitzero"`
	Expression      *string                  `json:"expression,omitzero"`
	HideExpression  *bool                    `json:"hideExpression,omitzero"`
	Digit           *bool                    `json:"digit,omitzero"`
	ThumbnailSize   string                   `json:"thumbnailSize,omitzero"`
	Protocol        LinkProtocol             `json:"protocol,omitzero"`
	Format          CalcFormat               `json:"format,omitzero"`
	DisplayScale    *string                  `json:"displayScale,omitzero"`
	Unit            *string                  `json:"unit,omitzero"`
	UnitPosition    UnitPosition             `json:"unitPosition,omitzero"`
	Entities        []FieldEntity            `json:"entities,omitzero"`
	ReferenceTable  *ReferenceTable          `json:"referenceTable,omitzero"`
	Lookup          *Lookup                  `json:"lookup,omitzero"`
	OpenGroup       *bool                    `json:"openGroup,omitzero"`
	Fields          map[string]FieldProperty `json:"fields,omitzero"`
	Enabled         *bool                    `json:"enabled,omitzero"`
}

// Alignment is the alignment of the options of radio button and check box fields.
type Alignment string

// Alignments.
const (
	AlignmentHorizontal Alignment = "HORIZONTAL"
	AlignmentVertical   Alignment = "VERTICAL"
)

// LinkProtocol is the type of the values of a link field.
type LinkProtocol string

// Link protocols.
const (
	LinkProtocolWeb  LinkProtocol = "WEB"
	LinkProtocolCall LinkProtocol = "CALL"
	LinkProtocolMail LinkProtocol = "MAIL"
)

// CalcFormat is the display format of a calculated field.
type CalcFormat string

// Calculation formats.
const (
	CalcFormatNumber        CalcFormat = "NUMBER"
	CalcFormatNumberDigit   CalcFormat = "NUMBER_DIGIT"
	CalcFormatDateTime      CalcFormat = "DATETIME"
	CalcFormatDate          CalcFormat = "DATE"
	CalcFormatTime          CalcFormat = "TIME"
	CalcFormatHourMinute    CalcFormat = "HOUR_MINUTE"
	CalcFormatDayHourMinute CalcFormat = "DAY_HOUR_MINUTE"
)

// UnitPosition is the position of the unit of number and calculated fields.
type UnitPosition string

// Unit positions.
const (
	UnitPositionBefore UnitPosition = "BEFORE"
	UnitPositionAfter  UnitPosition = "AFTER"
)

// FieldOption is an option of a check box, radio button, multi-choice or
// drop-down field.
type FieldOption struct {
	Label string `json:"label"`
	Index int    `json:"index,string"`
}

// FieldEntity is an entity specified by the type and the code, such as a user,
// department or group of the options of a user selection field.
type FieldEntity struct {
	Code string     `json:"code,omitzero"`
	Type EntityType `json:"type"`
}

// ReferenceTable is the setting of a related records field. When RelatedApp
// is set in an update, the server resets the omitted settings.
type ReferenceTable struct {
	RelatedApp    RelatedApp              `json:"relatedApp,omitzero"`
	Condition     ReferenceTableCondition `json:"condition,omitzero"`
	FilterCond    *string                 `json:"filterCond,omitzero"`
	DisplayFields []string                `json:"displayFields,omitzero"`
	Sort          *string                 `json:"sort,omitzero"`
	Size          string                  `json:"size,omitzero"`
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
	App  int64  `json:"app,string,omitzero"`
	Code string `json:"code,omitzero"`
}

// Lookup is the setting of a lookup field.
type Lookup struct {
	RelatedApp         RelatedApp     `json:"relatedApp,omitzero"`
	RelatedKeyField    string         `json:"relatedKeyField,omitzero"`
	FieldMappings      []FieldMapping `json:"fieldMappings,omitzero"`
	LookupPickerFields []string       `json:"lookupPickerFields,omitzero"`
	FilterCond         *string        `json:"filterCond,omitzero"`
	Sort               *string        `json:"sort,omitzero"`
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
		Entities jsontext.Value `json:"entities"`
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
		FieldMappings      jsontext.Value `json:"fieldMappings"`
		LookupPickerFields jsontext.Value `json:"lookupPickerFields"`
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

func unmarshalList[T any](value jsontext.Value, v *[]T) error {
	if len(value) == 0 || string(value) == `""` {
		return nil
	}
	return json.Unmarshal(value, v)
}

// FormFields is the fields of an app.
type FormFields struct {
	Properties map[string]FieldProperty `json:"properties"`
	Revision   int64                    `json:"revision,string"`
}

// GetFormFields gets the fields of an app, keyed by the field code.
func (c *Client) GetFormFields(ctx context.Context, req GetAppSettingsRequest) (*FormFields, error) {
	return c.getAppSettings[FormFields](ctx, "app/form/fields.json", req.Preview, req)
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
		Revision   int64                    `json:"revision,omitzero"`
	}{app, properties, revision}
	return c.updateAppSettings(ctx, method, "app/form/fields.json", params)
}

// DeleteFormFields deletes fields of an app in the preview environment and
// returns the new revision. The revision is not checked if revision is zero.
func (c *Client) DeleteFormFields(ctx context.Context, app int64, fields []string, revision int64) (int64, error) {
	params := struct {
		App      int64    `json:"app"`
		Fields   []string `json:"fields"`
		Revision int64    `json:"revision,omitzero"`
	}{app, fields, revision}
	return c.updateAppSettings(ctx, http.MethodDelete, "app/form/fields.json", params)
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
	Code   string        `json:"code,omitzero"`
	Fields []LayoutField `json:"fields,omitzero"`
	Layout []LayoutRow   `json:"layout,omitzero"`
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
	Code      string     `json:"code,omitzero"`
	Label     string     `json:"label,omitzero"`
	ElementID string     `json:"elementId,omitzero"`
	Size      *FieldSize `json:"size,omitzero"`
}

// FieldSize is the size of a field in pixels. Height includes the field name,
// and InnerHeight does not.
type FieldSize struct {
	Width       string `json:"width,omitzero"`
	Height      string `json:"height,omitzero"`
	InnerHeight string `json:"innerHeight,omitzero"`
}

// GetFormLayout gets the layout of the form of an app. Lang of the request is
// not used.
func (c *Client) GetFormLayout(ctx context.Context, req GetAppSettingsRequest) (*FormLayout, error) {
	params := struct {
		App int64 `json:"app"`
	}{req.App}
	return c.getAppSettings[FormLayout](ctx, "app/form/layout.json", req.Preview, params)
}

// UpdateFormLayout updates the layout of the form of an app in the preview
// environment and returns the new revision. The revision is not checked if
// revision is zero.
func (c *Client) UpdateFormLayout(ctx context.Context, app int64, layout []LayoutRow, revision int64) (int64, error) {
	params := struct {
		App      int64       `json:"app"`
		Layout   []LayoutRow `json:"layout"`
		Revision int64       `json:"revision,omitzero"`
	}{app, layout, revision}
	return c.updateAppSettings(ctx, http.MethodPut, "app/form/layout.json", params)
}
