package kintone

import (
	"context"
	"net/http"
)

// ProcessManagement is the process management settings of an app.
// States and Actions are nil if the process management has never been set.
type ProcessManagement struct {
	Enable   *bool                   `json:"enable,omitempty"`
	States   map[string]ProcessState `json:"states,omitempty"`
	Actions  []ProcessAction         `json:"actions,omitempty"`
	Revision int64                   `json:"revision,string,omitempty"`
}

// ProcessState is a status of the process management.
type ProcessState struct {
	Name     string           `json:"name,omitempty"`
	Index    int              `json:"index,string"`
	Assignee *ProcessAssignee `json:"assignee,omitempty"`
}

// ProcessAssignee is the assignees of a status.
type ProcessAssignee struct {
	Type     string         `json:"type"`
	Entities []TargetEntity `json:"entities"`
}

// TargetEntity is an entity with whether to include the sub organizations,
// used for the assignees, the executable users and the notification targets.
type TargetEntity struct {
	Entity      FieldEntity `json:"entity"`
	IncludeSubs bool        `json:"includeSubs,omitempty"`
}

// ProcessAction is an action of the process management. ExecutableUser is the
// users who can execute the action of type SECONDARY.
type ProcessAction struct {
	Name           string          `json:"name"`
	From           string          `json:"from"`
	To             string          `json:"to"`
	FilterCond     string          `json:"filterCond,omitempty"`
	Type           string          `json:"type,omitempty"`
	ExecutableUser *ExecutableUser `json:"executableUser,omitempty"`
}

// ExecutableUser is the users who can execute an action.
type ExecutableUser struct {
	Entities []TargetEntity `json:"entities"`
}

// GetProcessManagement gets the process management settings of an app.
func (c *Client) GetProcessManagement(ctx context.Context, req GetAppSettingsRequest) (*ProcessManagement, error) {
	return c.getAppSettings[ProcessManagement](ctx, "app/status.json", req.Preview, req)
}

// UpdateProcessManagement updates the process management settings of an app
// in the preview environment and returns the new revision. Revision of the
// settings is not used. The revision is not checked if revision is zero.
func (c *Client) UpdateProcessManagement(ctx context.Context, app int64, settings ProcessManagement, revision int64) (int64, error) {
	params := struct {
		App int64 `json:"app"`
		ProcessManagement
		Revision int64 `json:"revision,omitempty"`
	}{app, settings, revision}
	return c.updateAppSettings(ctx, http.MethodPut, "app/status.json", params)
}
