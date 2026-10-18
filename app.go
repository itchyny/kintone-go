package kintone

import (
	"context"
	"time"
)

// App is the information of an app.
// SpaceID and ThreadID are zero if the app does not belong to a space.
type App struct {
	AppID       int64     `json:"appId,string"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SpaceID     int64     `json:"spaceId,string"`
	ThreadID    int64     `json:"threadId,string"`
	CreatedAt   time.Time `json:"createdAt"`
	Creator     Entity    `json:"creator"`
	ModifiedAt  time.Time `json:"modifiedAt"`
	Modifier    Entity    `json:"modifier"`
}

// GetApp gets the information of an app.
func (c *Client) GetApp(ctx context.Context, id int64) (*App, error) {
	params := struct {
		ID int64 `json:"id"`
	}{id}
	result, err := c.get[App](ctx, "app.json", params)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAppsRequest is the request of [Client.GetApps].
type GetAppsRequest struct {
	IDs      []int64  `json:"ids,omitempty"`
	Codes    []string `json:"codes,omitempty"`
	Name     string   `json:"name,omitempty"`
	SpaceIDs []int64  `json:"spaceIds,omitempty"`
	Offset   int64    `json:"offset,omitempty"`
	Limit    int64    `json:"limit,omitempty"`
}

// GetApps gets the information of apps.
func (c *Client) GetApps(ctx context.Context, req GetAppsRequest) ([]App, error) {
	result, err := c.get[struct {
		Apps []App `json:"apps"`
	}](ctx, "apps.json", req)
	if err != nil {
		return nil, err
	}
	return result.Apps, nil
}

// AddAppRequest is the request of [Client.AddApp].
type AddAppRequest struct {
	Name   string `json:"name"`
	Space  int64  `json:"space,omitempty"`
	Thread int64  `json:"thread,omitempty"`
}

// AppRevision is the ID and the revision of an app.
type AppRevision struct {
	App      int64 `json:"app,string"`
	Revision int64 `json:"revision,string"`
}

// AddApp adds an app to the preview environment.
func (c *Client) AddApp(ctx context.Context, req AddAppRequest) (AppRevision, error) {
	return c.post[AppRevision](ctx, "preview/app.json", req)
}

// AppDeployment is an app to deploy or revert.
// The revision is not checked if Revision is zero.
type AppDeployment struct {
	App      int64 `json:"app"`
	Revision int64 `json:"revision,omitempty"`
}

// DeployApps deploys the settings of the apps in the preview environment to
// the production environment. Use [Client.GetDeployStatus] to check whether
// the deployment is finished.
func (c *Client) DeployApps(ctx context.Context, apps []AppDeployment) error {
	return c.deployApps(ctx, apps, false)
}

// RevertApps discards the settings of the apps in the preview environment.
func (c *Client) RevertApps(ctx context.Context, apps []AppDeployment) error {
	return c.deployApps(ctx, apps, true)
}

func (c *Client) deployApps(ctx context.Context, apps []AppDeployment, revert bool) error {
	params := struct {
		Apps   []AppDeployment `json:"apps"`
		Revert bool            `json:"revert,omitempty"`
	}{apps, revert}
	_, err := c.post[struct{}](ctx, "preview/app/deploy.json", params)
	return err
}

// AppDeployStatus is the deploy status of an app.
type AppDeployStatus struct {
	App    int64        `json:"app,string"`
	Status DeployStatus `json:"status"`
}

// DeployStatus is the status of a deployment.
type DeployStatus string

// Deploy statuses.
const (
	DeployStatusProcessing DeployStatus = "PROCESSING"
	DeployStatusSuccess    DeployStatus = "SUCCESS"
	DeployStatusFail       DeployStatus = "FAIL"
	DeployStatusCancel     DeployStatus = "CANCEL"
)

// GetDeployStatus gets the deploy statuses of the apps.
func (c *Client) GetDeployStatus(ctx context.Context, apps []int64) ([]AppDeployStatus, error) {
	params := struct {
		Apps []int64 `json:"apps"`
	}{apps}
	result, err := c.get[struct {
		Apps []AppDeployStatus `json:"apps"`
	}](ctx, "preview/app/deploy.json", params)
	if err != nil {
		return nil, err
	}
	return result.Apps, nil
}
