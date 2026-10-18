package kintone

import "context"

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
	if update.Assignees == nil {
		update.Assignees = []string{}
	}
	params := struct {
		App int64 `json:"app"`
		AssigneesUpdate
	}{app, update}
	result, err := c.put[struct {
		Revision int64 `json:"revision,string"`
	}](ctx, "record/assignees.json", params)
	if err != nil {
		return 0, err
	}
	return result.Revision, nil
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
	params := struct {
		App int64 `json:"app"`
		StatusUpdate
	}{app, update}
	result, err := c.put[struct {
		Revision int64 `json:"revision,string"`
	}](ctx, "record/status.json", params)
	if err != nil {
		return 0, err
	}
	return result.Revision, nil
}

// UpdateStatuses updates the statuses of records.
func (c *Client) UpdateStatuses(ctx context.Context, app int64, updates []StatusUpdate) ([]RecordRevision, error) {
	params := struct {
		App     int64          `json:"app"`
		Records []StatusUpdate `json:"records"`
	}{app, updates}
	result, err := c.put[struct {
		Records []RecordRevision `json:"records"`
	}](ctx, "records/status.json", params)
	if err != nil {
		return nil, err
	}
	return result.Records, nil
}
