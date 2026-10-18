package kintone

import (
	"context"
	"net/http"
)

// AppPlugin is a plugin added to an app.
type AppPlugin struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// AppPlugins is the plugins added to an app.
type AppPlugins struct {
	Plugins  []AppPlugin `json:"plugins"`
	Revision int64       `json:"revision,string"`
}

// GetAppPlugins gets the plugins added to an app.
func (c *Client) GetAppPlugins(ctx context.Context, req GetAppSettingsRequest) (*AppPlugins, error) {
	return c.getAppSettings[AppPlugins](ctx, "plugins.json", req.Preview, req)
}

// AddAppPlugins adds plugins to an app in the preview environment and returns
// the new revision. The revision is not checked if revision is zero.
func (c *Client) AddAppPlugins(ctx context.Context, app int64, ids []string, revision int64) (int64, error) {
	params := struct {
		App      int64    `json:"app"`
		IDs      []string `json:"ids"`
		Revision int64    `json:"revision,omitempty"`
	}{app, ids, revision}
	return c.updateAppSettings(ctx, http.MethodPost, "plugins.json", params)
}
