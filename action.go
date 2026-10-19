package kintone

import "context"

// AppAction is the setting of an action of an app, which copies a record to
// DestApp.
type AppAction struct {
	ID         int64           `json:"id,string,omitzero"`
	Name       string          `json:"name,omitzero"`
	Index      int             `json:"index,string"`
	DestApp    *RelatedApp     `json:"destApp,omitzero"`
	Mappings   []ActionMapping `json:"mappings,omitzero"`
	Entities   []FieldEntity   `json:"entities,omitzero"`
	FilterCond *string         `json:"filterCond,omitzero"`
}

// ActionMapping is a mapping from the source to the field of the destination
// app. SrcField is set if SrcType is FIELD.
type ActionMapping struct {
	SrcType   MappingSourceType `json:"srcType"`
	SrcField  string            `json:"srcField,omitzero"`
	DestField string            `json:"destField"`
}

// MappingSourceType is the source of a mapping of an app action.
type MappingSourceType string

// Mapping source types.
const (
	MappingSourceTypeField     MappingSourceType = "FIELD"
	MappingSourceTypeRecordURL MappingSourceType = "RECORD_URL"
)

// AppActions is the actions of an app, keyed by the action name.
type AppActions struct {
	Actions  map[string]AppAction `json:"actions"`
	Revision int64                `json:"revision,string"`
}

// GetAppActions gets the actions of an app.
func (c *Client) GetAppActions(ctx context.Context, req GetAppSettingsRequest) (*AppActions, error) {
	return c.getAppSettings[AppActions](ctx, "app/actions.json", req.Preview, req)
}

// UpdatedAppActions is the response of [Client.UpdateAppActions], which has
// the action IDs keyed by the action name.
type UpdatedAppActions struct {
	Actions  map[string]UpdatedAppAction `json:"actions"`
	Revision int64                       `json:"revision,string"`
}

// UpdatedAppAction is an action updated by [Client.UpdateAppActions].
type UpdatedAppAction struct {
	ID int64 `json:"id,string"`
}

// UpdateAppActions updates the actions of an app in the preview environment.
// The keys of actions are the action names before the update. The revision is
// not checked if revision is zero.
func (c *Client) UpdateAppActions(ctx context.Context, app int64, actions map[string]AppAction, revision int64) (*UpdatedAppActions, error) {
	params := struct {
		App      int64                `json:"app"`
		Actions  map[string]AppAction `json:"actions"`
		Revision int64                `json:"revision,omitzero"`
	}{app, actions, revision}
	result, err := c.put[UpdatedAppActions](ctx, "preview/app/actions.json", params)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
