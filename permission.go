package kintone

import (
	"context"
	"net/http"
)

// RecordAccess is the access rights of the user on a record and its fields.
type RecordAccess struct {
	ID     int64                        `json:"id,string"`
	Record RecordAccessRights           `json:"record"`
	Fields map[string]FieldAccessRights `json:"fields"`
}

// RecordAccessRights is the access rights on a record.
type RecordAccessRights struct {
	Viewable  bool `json:"viewable"`
	Editable  bool `json:"editable"`
	Deletable bool `json:"deletable"`
}

// FieldAccessRights is the access rights on a field.
type FieldAccessRights struct {
	Viewable bool `json:"viewable"`
	Editable bool `json:"editable"`
}

// EvaluateRecordPermissions gets the permissions of the user on the records.
func (c *Client) EvaluateRecordPermissions(ctx context.Context, app int64, ids []int64) ([]RecordAccess, error) {
	params := struct {
		App int64   `json:"app"`
		IDs []int64 `json:"ids"`
	}{app, ids}
	result, err := c.get[struct {
		Rights []RecordAccess `json:"rights"`
	}](ctx, "records/acl/evaluate.json", params)
	if err != nil {
		return nil, err
	}
	return result.Rights, nil
}

// AppPermissions is the app permissions of an app, in order of priority.
type AppPermissions struct {
	Rights   []AppRight `json:"rights"`
	Revision int64      `json:"revision,string,omitzero"`
}

// AppRight is an app permission of an entity.
type AppRight struct {
	Entity           FieldEntity `json:"entity"`
	IncludeSubs      bool        `json:"includeSubs,omitzero"`
	AppEditable      bool        `json:"appEditable,omitzero"`
	RecordViewable   bool        `json:"recordViewable,omitzero"`
	RecordAddable    bool        `json:"recordAddable,omitzero"`
	RecordEditable   bool        `json:"recordEditable,omitzero"`
	RecordDeletable  bool        `json:"recordDeletable,omitzero"`
	RecordImportable bool        `json:"recordImportable,omitzero"`
	RecordExportable bool        `json:"recordExportable,omitzero"`
}

// GetAppPermissions gets the app permissions of an app. Lang of the request is
// not used.
func (c *Client) GetAppPermissions(ctx context.Context, req GetAppSettingsRequest) (*AppPermissions, error) {
	params := struct {
		App int64 `json:"app"`
	}{req.App}
	return c.getAppSettings[AppPermissions](ctx, "app/acl.json", req.Preview, params)
}

// UpdateAppPermissions updates the app permissions of an app in the preview
// environment and returns the new revision. The revision is not checked if
// revision is zero.
func (c *Client) UpdateAppPermissions(ctx context.Context, app int64, rights []AppRight, revision int64) (int64, error) {
	params := struct {
		App      int64      `json:"app"`
		Rights   []AppRight `json:"rights"`
		Revision int64      `json:"revision,omitzero"`
	}{app, rights, revision}
	return c.updateAppSettings(ctx, http.MethodPut, "app/acl.json", params)
}

// RecordPermissions is the record permissions of an app, in order of priority.
type RecordPermissions struct {
	Rights   []RecordRight `json:"rights"`
	Revision int64         `json:"revision,string,omitzero"`
}

// RecordRight is a record permission of the records matching FilterCond.
type RecordRight struct {
	FilterCond string              `json:"filterCond,omitzero"`
	Entities   []RecordRightEntity `json:"entities"`
}

// RecordRightEntity is a record permission of an entity.
type RecordRightEntity struct {
	Entity      FieldEntity `json:"entity"`
	Viewable    bool        `json:"viewable,omitzero"`
	Editable    bool        `json:"editable,omitzero"`
	Deletable   bool        `json:"deletable,omitzero"`
	IncludeSubs bool        `json:"includeSubs,omitzero"`
}

// GetRecordPermissions gets the record permissions of an app.
func (c *Client) GetRecordPermissions(ctx context.Context, req GetAppSettingsRequest) (*RecordPermissions, error) {
	return c.getAppSettings[RecordPermissions](ctx, "record/acl.json", req.Preview, req)
}

// UpdateRecordPermissions updates the record permissions of an app in the
// preview environment and returns the new revision. The revision is not
// checked if revision is zero.
func (c *Client) UpdateRecordPermissions(ctx context.Context, app int64, rights []RecordRight, revision int64) (int64, error) {
	params := struct {
		App      int64         `json:"app"`
		Rights   []RecordRight `json:"rights"`
		Revision int64         `json:"revision,omitzero"`
	}{app, rights, revision}
	return c.updateAppSettings(ctx, http.MethodPut, "record/acl.json", params)
}

// FieldPermissions is the field permissions of an app.
type FieldPermissions struct {
	Rights   []FieldRight `json:"rights"`
	Revision int64        `json:"revision,string,omitzero"`
}

// FieldRight is the permissions of the field of Code, in order of priority.
type FieldRight struct {
	Code     string             `json:"code"`
	Entities []FieldRightEntity `json:"entities"`
}

// FieldRightEntity is a field permission of an entity.
type FieldRightEntity struct {
	Accessibility string      `json:"accessibility"`
	Entity        FieldEntity `json:"entity"`
	IncludeSubs   bool        `json:"includeSubs,omitzero"`
}

// GetFieldPermissions gets the field permissions of an app. Lang of the
// request is not used.
func (c *Client) GetFieldPermissions(ctx context.Context, req GetAppSettingsRequest) (*FieldPermissions, error) {
	params := struct {
		App int64 `json:"app"`
	}{req.App}
	return c.getAppSettings[FieldPermissions](ctx, "field/acl.json", req.Preview, params)
}

// UpdateFieldPermissions updates the field permissions of an app in the
// preview environment and returns the new revision. The revision is not
// checked if revision is zero.
func (c *Client) UpdateFieldPermissions(ctx context.Context, app int64, rights []FieldRight, revision int64) (int64, error) {
	params := struct {
		App      int64        `json:"app"`
		Rights   []FieldRight `json:"rights"`
		Revision int64        `json:"revision,omitzero"`
	}{app, rights, revision}
	return c.updateAppSettings(ctx, http.MethodPut, "field/acl.json", params)
}
