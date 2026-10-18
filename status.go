package kintone

import (
	"context"
	"net/http"
)

// AssigneesUpdate is an update of the assignees of a record.
// Assignees are the login names of the users.
// The revision is not checked if Revision is zero.
type AssigneesUpdate struct {
	ID        int64    `json:"id"`
	Assignees []string `json:"assignees"`
	Revision  int64    `json:"revision,omitempty"`
}

// UpdateAssignees updates the assignees of a record and returns the new
// revision. Nil Assignees is sent as an empty list.
func (c *Client) UpdateAssignees(ctx context.Context, app int64, update AssigneesUpdate) (int64, error) {
	return c.call(ctx, updateAssigneesRequest(app, update))
}

func updateAssigneesRequest(app int64, update AssigneesUpdate) apiRequest[int64] {
	if update.Assignees == nil {
		update.Assignees = []string{}
	}
	params := struct {
		App int64 `json:"app"`
		AssigneesUpdate
	}{app, update}
	return apiRequest[int64]{http.MethodPut, "record/assignees.json", params, decodeWith(decodeRevision)}
}

// StatusUpdate is an update of the status of a record by an action.
// Assignee is the login name of the user.
// The revision is not checked if Revision is zero.
type StatusUpdate struct {
	ID       int64  `json:"id"`
	Action   string `json:"action"`
	Assignee string `json:"assignee,omitempty"`
	Revision int64  `json:"revision,omitempty"`
}

// UpdateStatus updates the status of a record and returns the new revision.
func (c *Client) UpdateStatus(ctx context.Context, app int64, update StatusUpdate) (int64, error) {
	return c.call(ctx, updateStatusRequest(app, update))
}

func updateStatusRequest(app int64, update StatusUpdate) apiRequest[int64] {
	params := struct {
		App int64 `json:"app"`
		StatusUpdate
	}{app, update}
	return apiRequest[int64]{http.MethodPut, "record/status.json", params, decodeWith(decodeRevision)}
}

// UpdateStatuses updates the statuses of records.
func (c *Client) UpdateStatuses(ctx context.Context, app int64, updates []StatusUpdate) ([]RecordRevision, error) {
	return c.call(ctx, updateStatusesRequest(app, updates))
}

func updateStatusesRequest(app int64, updates []StatusUpdate) apiRequest[[]RecordRevision] {
	params := struct {
		App     int64          `json:"app"`
		Records []StatusUpdate `json:"records"`
	}{app, updates}
	return apiRequest[[]RecordRevision]{http.MethodPut, "records/status.json", params,
		decodeWith(func(r struct {
			Records []RecordRevision `json:"records"`
		}) ([]RecordRevision, error) {
			return r.Records, nil
		})}
}
