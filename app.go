package kintone

import (
	"context"
	"fmt"
	"iter"
	"strings"
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
	IDs      []int64  `json:"ids,omitzero"`
	Codes    []string `json:"codes,omitzero"`
	Name     string   `json:"name,omitzero"`
	SpaceIDs []int64  `json:"spaceIds,omitzero"`
	Offset   int64    `json:"offset,omitzero"`
	Limit    int64    `json:"limit,omitzero"`
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

var allAppsLimit int64 = 100

// AllApps iterates over the apps, using [Client.GetApps]. Offset of the
// request is the number of apps to skip, and Limit is the number of apps to
// get in each request.
func (c *Client) AllApps(ctx context.Context, req GetAppsRequest) iter.Seq2[App, error] {
	if req.Limit == 0 {
		req.Limit = allAppsLimit
	}
	return paginate(req.Offset, func(offset int64) ([]App, bool, error) {
		r := req
		r.Offset = offset
		apps, err := c.GetApps(ctx, r)
		return apps, int64(len(apps)) == r.Limit, err
	})
}

// AddAppRequest is the request of [Client.AddApp].
type AddAppRequest struct {
	Name   string `json:"name"`
	Space  int64  `json:"space,omitzero"`
	Thread int64  `json:"thread,omitzero"`
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
	Revision int64 `json:"revision,omitzero"`
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
		Revert bool            `json:"revert,omitzero"`
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

const deployPollInterval = time.Second

// WaitForDeploy waits until the deployments of the apps finish, by polling
// [Client.GetDeployStatus]. It returns an error if any of the deployments
// fails or is canceled.
func (c *Client) WaitForDeploy(ctx context.Context, apps []int64) error {
	for {
		statuses, err := c.GetDeployStatus(ctx, apps)
		if err != nil {
			return err
		}
		var processing bool
		var failed []string
		for _, s := range statuses {
			switch s.Status {
			case DeployStatusProcessing:
				processing = true
			case DeployStatusSuccess:
			default:
				failed = append(failed, fmt.Sprintf("app %d: %s", s.App, s.Status))
			}
		}
		if !processing {
			if len(failed) > 0 {
				return fmt.Errorf("kintone: deployment failed: %s", strings.Join(failed, ", "))
			}
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(deployPollInterval):
		}
	}
}

// GetAppSettingsRequest is the request to get settings of an app.
// If Preview is true, the settings in the preview environment are returned.
type GetAppSettingsRequest struct {
	App     int64  `json:"app"`
	Lang    string `json:"lang,omitzero"`
	Preview bool   `json:"-"`
}

func (c *Client) getAppSettings[T any](ctx context.Context, path string, preview bool, params any) (*T, error) {
	if preview {
		path = "preview/" + path
	}
	result, err := c.get[T](ctx, path, params)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) updateAppSettings(ctx context.Context, method, path string, params any) (int64, error) {
	result, err := c.do[struct {
		Revision int64 `json:"revision,string"`
	}](ctx, method, "preview/"+path, params)
	if err != nil {
		return 0, err
	}
	return result.Revision, nil
}

// AppStatistics is the usage of an app.
// Space is nil if the app does not belong to a space.
type AppStatistics struct {
	ID                int64     `json:"id,string"`
	Name              string    `json:"name"`
	Space             *AppSpace `json:"space"`
	AppGroup          string    `json:"appGroup"`
	Status            string    `json:"status"`
	RecordUpdatedAt   time.Time `json:"recordUpdatedAt"`
	RecordCount       int64     `json:"recordCount,string"`
	FieldCount        int64     `json:"fieldCount,string"`
	DailyRequestCount int64     `json:"dailyRequestCount,string"`
	APITokenCount     int64     `json:"apiTokenCount,string"`
	WebhookCount      int64     `json:"webhookCount,string"`
	StorageUsage      int64     `json:"storageUsage,string"`
	Customized        bool      `json:"customized"`
	Creator           Entity    `json:"creator"`
	CreatedAt         time.Time `json:"createdAt"`
	Modifier          Entity    `json:"modifier"`
	ModifiedAt        time.Time `json:"modifiedAt"`
}

// AppSpace is the space of an app.
type AppSpace struct {
	ID   int64  `json:"id,string"`
	Name string `json:"name"`
}

// GetAppsStatisticsRequest is the request of [Client.GetAppsStatistics].
type GetAppsStatisticsRequest struct {
	Offset int64 `json:"offset,omitzero"`
	Limit  int64 `json:"limit,omitzero"`
}

// GetAppsStatistics gets the usage of apps. The guest space of the client is
// not used.
func (c *Client) GetAppsStatistics(ctx context.Context, req GetAppsStatisticsRequest) ([]AppStatistics, error) {
	result, err := c.withoutGuestSpace().get[struct {
		Apps []AppStatistics `json:"apps"`
	}](ctx, "apps/statistics.json", req)
	if err != nil {
		return nil, err
	}
	return result.Apps, nil
}

// MoveApp moves an app to the space. If space is zero, the app is moved out of
// the space. The guest space of the client is not used.
func (c *Client) MoveApp(ctx context.Context, app, space int64) error {
	params := struct {
		App   int64  `json:"app"`
		Space *int64 `json:"space"`
	}{App: app}
	if space != 0 {
		params.Space = &space
	}
	_, err := c.withoutGuestSpace().post[struct{}](ctx, "app/move.json", params)
	return err
}
