package kintone

import "context"

// View is the setting of a view of an app.
//
// Fields is used by list views, Date and Title by calendar views, and HTML,
// Pager and Device by custom views.
type View struct {
	ID          int64    `json:"id,string,omitempty"`
	Name        string   `json:"name,omitempty"`
	Type        ViewType `json:"type"`
	BuiltinType string   `json:"builtinType,omitempty"`
	Index       int      `json:"index,string"`
	Fields      []string `json:"fields,omitempty"`
	Date        string   `json:"date,omitempty"`
	Title       string   `json:"title,omitempty"`
	HTML        string   `json:"html,omitempty"`
	Pager       *bool    `json:"pager,omitempty"`
	Device      string   `json:"device,omitempty"`
	FilterCond  string   `json:"filterCond,omitempty"`
	Sort        string   `json:"sort,omitempty"`
}

// ViewType is the type of a view.
type ViewType string

// View types.
const (
	ViewTypeList     ViewType = "LIST"
	ViewTypeCalendar ViewType = "CALENDAR"
	ViewTypeCustom   ViewType = "CUSTOM"
)

// Views is the views of an app, keyed by the view name.
type Views struct {
	Views    map[string]View `json:"views"`
	Revision int64           `json:"revision,string"`
}

// GetViews gets the views of an app.
func (c *Client) GetViews(ctx context.Context, req GetAppSettingsRequest) (*Views, error) {
	return c.getAppSettings[Views](ctx, "app/views.json", req.Preview, req)
}

// UpdatedViews is the response of [Client.UpdateViews], which has the view
// IDs keyed by the view name.
type UpdatedViews struct {
	Views    map[string]UpdatedView `json:"views"`
	Revision int64                  `json:"revision,string"`
}

// UpdatedView is a view updated by [Client.UpdateViews].
type UpdatedView struct {
	ID int64 `json:"id,string"`
}

// UpdateViews updates the views of an app in the preview environment. The keys
// of views are the view names before the update. The revision is not checked
// if revision is zero.
func (c *Client) UpdateViews(ctx context.Context, app int64, views map[string]View, revision int64) (*UpdatedViews, error) {
	params := struct {
		App      int64           `json:"app"`
		Views    map[string]View `json:"views"`
		Revision int64           `json:"revision,omitempty"`
	}{app, views, revision}
	result, err := c.put[UpdatedViews](ctx, "preview/app/views.json", params)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
