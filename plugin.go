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
	return c.getAppSettings[AppPlugins](ctx, "app/plugins.json", req.Preview, req)
}

// AddAppPlugins adds plugins to an app in the preview environment and returns
// the new revision. The revision is not checked if revision is zero.
func (c *Client) AddAppPlugins(ctx context.Context, app int64, ids []string, revision int64) (int64, error) {
	params := struct {
		App      int64    `json:"app"`
		IDs      []string `json:"ids"`
		Revision int64    `json:"revision,omitempty"`
	}{app, ids, revision}
	return c.updateAppSettings(ctx, http.MethodPost, "app/plugins.json", params)
}

// Plugin is a plugin installed in kintone. Description and Version are empty
// for the plugins returned by [Client.GetRequiredPlugins].
type Plugin struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	IsMarketPlugin bool   `json:"isMarketPlugin"`
	Version        string `json:"version"`
}

// GetPluginsRequest is the request of [Client.GetPlugins].
type GetPluginsRequest struct {
	Offset int64    `json:"offset,omitempty"`
	Limit  int64    `json:"limit,omitempty"`
	IDs    []string `json:"ids,omitempty"`
}

// GetPlugins gets the installed plugins. The guest space of the client is not
// used.
func (c *Client) GetPlugins(ctx context.Context, req GetPluginsRequest) ([]Plugin, error) {
	return c.getPlugins(ctx, "plugins.json", req)
}

// GetRequiredPluginsRequest is the request of [Client.GetRequiredPlugins].
type GetRequiredPluginsRequest struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

// GetRequiredPlugins gets the plugins that are required to be installed.
// The guest space of the client is not used.
func (c *Client) GetRequiredPlugins(ctx context.Context, req GetRequiredPluginsRequest) ([]Plugin, error) {
	return c.getPlugins(ctx, "plugins/required.json", req)
}

func (c *Client) getPlugins(ctx context.Context, path string, params any) ([]Plugin, error) {
	result, err := c.withoutGuestSpace().get[struct {
		Plugins []Plugin `json:"plugins"`
	}](ctx, path, params)
	if err != nil {
		return nil, err
	}
	return result.Plugins, nil
}

// GetPluginAppsRequest is the request of [Client.GetPluginApps].
// ID is the plugin ID.
type GetPluginAppsRequest struct {
	ID     string `json:"id"`
	Offset int64  `json:"offset,omitempty"`
	Limit  int64  `json:"limit,omitempty"`
}

// PluginApp is an app which the plugin is added to.
type PluginApp struct {
	ID   int64  `json:"id,string"`
	Name string `json:"name"`
}

// GetPluginApps gets the apps which the plugin is added to. The guest space of
// the client is not used.
func (c *Client) GetPluginApps(ctx context.Context, req GetPluginAppsRequest) ([]PluginApp, error) {
	result, err := c.withoutGuestSpace().get[struct {
		Apps []PluginApp `json:"apps"`
	}](ctx, "plugin/apps.json", req)
	if err != nil {
		return nil, err
	}
	return result.Apps, nil
}

// PluginVersion is the ID and the version of a plugin.
type PluginVersion struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

// AddPlugin installs the plugin of the file key, which is obtained by
// [Client.UploadFile]. The guest space of the client is not used.
func (c *Client) AddPlugin(ctx context.Context, fileKey string) (PluginVersion, error) {
	params := struct {
		FileKey string `json:"fileKey"`
	}{fileKey}
	return c.withoutGuestSpace().post[PluginVersion](ctx, "plugin.json", params)
}

// UpdatePlugin updates the plugin with the file key, which is obtained by
// [Client.UploadFile]. The guest space of the client is not used.
func (c *Client) UpdatePlugin(ctx context.Context, id, fileKey string) (PluginVersion, error) {
	params := struct {
		ID      string `json:"id"`
		FileKey string `json:"fileKey"`
	}{id, fileKey}
	return c.withoutGuestSpace().put[PluginVersion](ctx, "plugin.json", params)
}

// DeletePlugin uninstalls the plugin. The guest space of the client is not
// used.
func (c *Client) DeletePlugin(ctx context.Context, id string) error {
	params := struct {
		ID string `json:"id"`
	}{id}
	_, err := c.withoutGuestSpace().delete[struct{}](ctx, "plugin.json", params)
	return err
}
