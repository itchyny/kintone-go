package kintone

import (
	"context"
	"net/http"
)

// Categories is the category settings of an app, keyed by the category name.
type Categories struct {
	Enabled    *bool               `json:"enabled,omitzero"`
	Categories map[string]Category `json:"categories,omitzero"`
	Revision   int64               `json:"revision,string,omitzero"`
}

// Category is a category with the child categories keyed by the name.
type Category struct {
	Name     string              `json:"name"`
	Index    int                 `json:"index,string"`
	Children map[string]Category `json:"children"`
}

// GetCategories gets the category settings of an app. This uses an API in the
// API lab of kintone.
func (c *Client) GetCategories(ctx context.Context, req GetAppSettingsRequest) (*Categories, error) {
	return c.getAppSettings[Categories](ctx, "app/categories.json", req.Preview, req)
}

// UpdateCategories updates the category settings of an app in the preview
// environment and returns the new revision. The keys of the categories are the
// category names before the update. Revision of the settings is not used. The
// revision is not checked if revision is zero. This uses an API in the API lab
// of kintone.
func (c *Client) UpdateCategories(ctx context.Context, app int64, settings Categories, revision int64) (int64, error) {
	params := struct {
		App int64 `json:"app"`
		Categories
		Revision int64 `json:"revision,omitzero"`
	}{app, settings, revision}
	return c.updateAppSettings(ctx, http.MethodPut, "app/categories.json", params)
}
