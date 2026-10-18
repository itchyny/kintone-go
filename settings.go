package kintone

import (
	"context"
	"net/http"
)

// GeneralSettings is the general settings of an app.
//
// When updating the settings, the settings not set are not changed.
type GeneralSettings struct {
	Name                      string           `json:"name,omitempty"`
	Description               string           `json:"description,omitempty"`
	Icon                      *AppIcon         `json:"icon,omitempty"`
	Theme                     string           `json:"theme,omitempty"`
	TitleField                *TitleField      `json:"titleField,omitempty"`
	EnableThumbnails          *bool            `json:"enableThumbnails,omitempty"`
	EnableBulkDeletion        *bool            `json:"enableBulkDeletion,omitempty"`
	EnableComments            *bool            `json:"enableComments,omitempty"`
	EnableDuplicateRecord     *bool            `json:"enableDuplicateRecord,omitempty"`
	EnableInlineRecordEditing *bool            `json:"enableInlineRecordEditing,omitempty"`
	NumberPrecision           *NumberPrecision `json:"numberPrecision,omitempty"`
	FirstMonthOfFiscalYear    string           `json:"firstMonthOfFiscalYear,omitempty"`
	Revision                  int64            `json:"revision,string,omitempty"`
}

// AppIcon is the icon of an app. Key is set for preset icons, and File is set
// for uploaded images.
type AppIcon struct {
	Type string    `json:"type"`
	Key  string    `json:"key,omitempty"`
	File *FileInfo `json:"file,omitempty"`
}

// TitleField is the field used as the title of records.
type TitleField struct {
	SelectionMode string `json:"selectionMode"`
	Code          string `json:"code,omitempty"`
}

// NumberPrecision is the precision of number and calculated fields.
type NumberPrecision struct {
	Digits        string `json:"digits,omitempty"`
	DecimalPlaces string `json:"decimalPlaces,omitempty"`
	RoundingMode  string `json:"roundingMode,omitempty"`
}

// GetGeneralSettings gets the general settings of an app.
func (c *Client) GetGeneralSettings(ctx context.Context, req GetAppSettingsRequest) (*GeneralSettings, error) {
	return c.getAppSettings[GeneralSettings](ctx, "settings.json", req.Preview, req)
}

// UpdateGeneralSettings updates the general settings of an app in the preview
// environment and returns the new revision. Revision of the settings is not
// used. The revision is not checked if revision is zero.
func (c *Client) UpdateGeneralSettings(ctx context.Context, app int64, settings GeneralSettings, revision int64) (int64, error) {
	params := struct {
		App int64 `json:"app"`
		GeneralSettings
		Revision int64 `json:"revision,omitempty"`
	}{app, settings, revision}
	return c.updateAppSettings(ctx, http.MethodPut, "settings.json", params)
}

// AdminNotes is the notes for the app administrators.
type AdminNotes struct {
	Content                        string `json:"content"`
	IncludeInTemplateAndDuplicates *bool  `json:"includeInTemplateAndDuplicates,omitempty"`
	Revision                       int64  `json:"revision,string,omitempty"`
}

// GetAdminNotes gets the notes for the administrators of an app. Lang of the
// request is not used.
func (c *Client) GetAdminNotes(ctx context.Context, req GetAppSettingsRequest) (*AdminNotes, error) {
	params := struct {
		App int64 `json:"app"`
	}{req.App}
	return c.getAppSettings[AdminNotes](ctx, "adminNotes.json", req.Preview, params)
}

// UpdateAdminNotes updates the notes for the administrators of an app in the
// preview environment and returns the new revision. Content is always sent,
// and Revision of the notes is not used. The revision is not checked if
// revision is zero.
func (c *Client) UpdateAdminNotes(ctx context.Context, app int64, notes AdminNotes, revision int64) (int64, error) {
	params := struct {
		App int64 `json:"app"`
		AdminNotes
		Revision int64 `json:"revision,omitempty"`
	}{app, notes, revision}
	return c.updateAppSettings(ctx, http.MethodPut, "adminNotes.json", params)
}
