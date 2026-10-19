package kintone

import (
	"context"
	"net/http"
)

// GeneralSettings is the general settings of an app.
type GeneralSettings struct {
	Name                      string           `json:"name,omitzero"`
	Description               *string          `json:"description,omitzero"`
	Icon                      *AppIcon         `json:"icon,omitzero"`
	Theme                     string           `json:"theme,omitzero"`
	TitleField                *TitleField      `json:"titleField,omitzero"`
	EnableThumbnails          *bool            `json:"enableThumbnails,omitzero"`
	EnableBulkDeletion        *bool            `json:"enableBulkDeletion,omitzero"`
	EnableComments            *bool            `json:"enableComments,omitzero"`
	EnableDuplicateRecord     *bool            `json:"enableDuplicateRecord,omitzero"`
	EnableInlineRecordEditing *bool            `json:"enableInlineRecordEditing,omitzero"`
	NumberPrecision           *NumberPrecision `json:"numberPrecision,omitzero"`
	FirstMonthOfFiscalYear    string           `json:"firstMonthOfFiscalYear,omitzero"`
	Revision                  int64            `json:"revision,string,omitzero"`
}

// AppIcon is the icon of an app. Key is set for preset icons, and File is set
// for uploaded images.
type AppIcon struct {
	Type string    `json:"type"`
	Key  string    `json:"key,omitzero"`
	File *FileInfo `json:"file,omitzero"`
}

// TitleField is the field used as the title of records.
type TitleField struct {
	SelectionMode string `json:"selectionMode"`
	Code          string `json:"code,omitzero"`
}

// NumberPrecision is the precision of number and calculated fields.
type NumberPrecision struct {
	Digits        string `json:"digits,omitzero"`
	DecimalPlaces string `json:"decimalPlaces,omitzero"`
	RoundingMode  string `json:"roundingMode,omitzero"`
}

// GetGeneralSettings gets the general settings of an app.
func (c *Client) GetGeneralSettings(ctx context.Context, req GetAppSettingsRequest) (*GeneralSettings, error) {
	return c.getAppSettings[GeneralSettings](ctx, "app/settings.json", req.Preview, req)
}

// UpdateGeneralSettings updates the general settings of an app in the preview
// environment and returns the new revision. Revision of the settings is not
// used. The revision is not checked if revision is zero.
func (c *Client) UpdateGeneralSettings(ctx context.Context, app int64, settings GeneralSettings, revision int64) (int64, error) {
	params := struct {
		App int64 `json:"app"`
		GeneralSettings
		Revision int64 `json:"revision,omitzero"`
	}{app, settings, revision}
	return c.updateAppSettings(ctx, http.MethodPut, "app/settings.json", params)
}

// AdminNotes is the notes for the app administrators.
type AdminNotes struct {
	Content                        string `json:"content"`
	IncludeInTemplateAndDuplicates *bool  `json:"includeInTemplateAndDuplicates,omitzero"`
	Revision                       int64  `json:"revision,string,omitzero"`
}

// GetAdminNotes gets the notes for the administrators of an app. Lang of the
// request is not used.
func (c *Client) GetAdminNotes(ctx context.Context, req GetAppSettingsRequest) (*AdminNotes, error) {
	params := struct {
		App int64 `json:"app"`
	}{req.App}
	return c.getAppSettings[AdminNotes](ctx, "app/adminNotes.json", req.Preview, params)
}

// UpdateAdminNotes updates the notes for the administrators of an app in the
// preview environment and returns the new revision. Content is always sent,
// and Revision of the notes is not used. The revision is not checked if
// revision is zero.
func (c *Client) UpdateAdminNotes(ctx context.Context, app int64, notes AdminNotes, revision int64) (int64, error) {
	params := struct {
		App int64 `json:"app"`
		AdminNotes
		Revision int64 `json:"revision,omitzero"`
	}{app, notes, revision}
	return c.updateAppSettings(ctx, http.MethodPut, "app/adminNotes.json", params)
}
