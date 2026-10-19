package kintone

import (
	"context"
	"encoding/json"
	"net/http"
)

// Categories is the category settings of an app, keyed by the category name.
// Nil Categories is not sent and the categories are not changed, but an empty
// one is sent. Clearing the categories requires Enabled to be false.
type Categories struct {
	Enabled    *bool               `json:"enabled,omitempty"`
	Categories map[string]Category `json:"categories,omitzero"`
	Revision   int64               `json:"revision,string,omitempty"`
}

// Category is a category with the child categories keyed by the name.
// Nil Children is sent as an empty object.
type Category struct {
	Name     string              `json:"name"`
	Index    int                 `json:"index,string"`
	Children map[string]Category `json:"children"`
}

// MarshalJSON encodes the category, sending nil Children as an empty object.
func (c Category) MarshalJSON() ([]byte, error) {
	type category Category
	if c.Children == nil {
		c.Children = map[string]Category{}
	}
	return json.Marshal(category(c))
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
		Revision int64 `json:"revision,omitempty"`
	}{app, settings, revision}
	return c.updateAppSettings(ctx, http.MethodPut, "app/categories.json", params)
}
