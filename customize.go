package kintone

import (
	"context"
	"net/http"
)

// Customization is the JavaScript and CSS customization of an app.
type Customization struct {
	Scope    string              `json:"scope,omitzero"`
	Desktop  *CustomizationFiles `json:"desktop,omitzero"`
	Mobile   *CustomizationFiles `json:"mobile,omitzero"`
	Revision int64               `json:"revision,string,omitzero"`
}

// CustomizationFiles is the JavaScript and CSS files of the customization.
type CustomizationFiles struct {
	JS  []CustomizationFile `json:"js,omitzero"`
	CSS []CustomizationFile `json:"css,omitzero"`
}

// CustomizationFile is a file of the customization. URL is set for files
// specified by URL, and File is set for uploaded files.
type CustomizationFile struct {
	Type string    `json:"type"`
	URL  string    `json:"url,omitzero"`
	File *FileInfo `json:"file,omitzero"`
}

// GetCustomization gets the JavaScript and CSS customization of an app. Lang
// of the request is not used.
func (c *Client) GetCustomization(ctx context.Context, req GetAppSettingsRequest) (*Customization, error) {
	params := struct {
		App int64 `json:"app"`
	}{req.App}
	return c.getAppSettings[Customization](ctx, "app/customize.json", req.Preview, params)
}

// UpdateCustomization updates the JavaScript and CSS customization of an app
// in the preview environment and returns the new revision. Revision of the
// customization is not used. The revision is not checked if revision is zero.
func (c *Client) UpdateCustomization(ctx context.Context, app int64, customization Customization, revision int64) (int64, error) {
	params := struct {
		App int64 `json:"app"`
		Customization
		Revision int64 `json:"revision,omitzero"`
	}{app, customization, revision}
	return c.updateAppSettings(ctx, http.MethodPut, "app/customize.json", params)
}
