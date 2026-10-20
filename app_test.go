package kintone

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"testing/synctest"
	"time"
)

func TestClientGetApp(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{
		"appId": "1",
		"code": "DEAL",
		"name": "Deals",
		"description": "Deals and activities.",
		"spaceId": "2",
		"threadId": "3",
		"createdAt": "2026-10-01T05:14:05.000Z",
		"creator": {"code": "alice", "name": "Alice"},
		"modifiedAt": "2026-10-02T05:22:05.000Z",
		"modifier": {"code": "bob", "name": "Bob"}
	}`)
	c := newTestClient(t, ts)
	app, err := c.GetApp(t.Context(), 1)
	if err != nil {
		t.Fatalf("GetApp returned error: %v", err)
	}
	assertRequest(t, got, http.MethodGet, "/k/v1/app.json", `{"id": 1}`)
	want := &App{
		AppID:       1,
		Code:        "DEAL",
		Name:        "Deals",
		Description: "Deals and activities.",
		SpaceID:     2,
		ThreadID:    3,
		CreatedAt:   time.Date(2026, 10, 1, 5, 14, 5, 0, time.UTC),
		Creator:     Entity{Code: "alice", Name: "Alice"},
		ModifiedAt:  time.Date(2026, 10, 2, 5, 22, 5, 0, time.UTC),
		Modifier:    Entity{Code: "bob", Name: "Bob"},
	}
	if !reflect.DeepEqual(app, want) {
		t.Errorf("app: got %+v, want %+v", app, want)
	}
}

func TestClientGetApps(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetAppsRequest
		wantBody string
	}{
		{
			name: "all parameters",
			req: GetAppsRequest{
				IDs:      []int64{1, 2},
				Codes:    []string{"DEAL"},
				Name:     "Deal",
				SpaceIDs: []int64{3},
				Offset:   10,
				Limit:    20,
			},
			wantBody: `{
				"ids": [1, 2],
				"codes": ["DEAL"],
				"name": "Deal",
				"spaceIds": [3],
				"offset": 10,
				"limit": 20
			}`,
		},
		{
			name:     "no parameters",
			wantBody: `{}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{
				"apps": [
					{
						"appId": "1",
						"code": "DEAL",
						"name": "Deals",
						"description": "",
						"spaceId": null,
						"threadId": null,
						"createdAt": "2026-10-01T05:14:05.000Z",
						"creator": {"code": "", "name": ""},
						"modifiedAt": "2026-10-01T05:14:05.000Z",
						"modifier": {"code": "alice", "name": "Alice"}
					}
				]
			}`)
			c := newTestClient(t, ts)
			apps, err := c.GetApps(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetApps returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, "/k/v1/apps.json", tc.wantBody)
			want := []App{
				{
					AppID:      1,
					Code:       "DEAL",
					Name:       "Deals",
					CreatedAt:  time.Date(2026, 10, 1, 5, 14, 5, 0, time.UTC),
					ModifiedAt: time.Date(2026, 10, 1, 5, 14, 5, 0, time.UTC),
					Modifier:   Entity{Code: "alice", Name: "Alice"},
				},
			}
			if !reflect.DeepEqual(apps, want) {
				t.Errorf("apps: got %+v, want %+v", apps, want)
			}
		})
	}
}

func TestClientAllApps(t *testing.T) {
	setValue(t, &allAppsLimit, 2)
	testCases := []struct {
		name       string
		req        GetAppsRequest
		responses  []response
		wantBodies []string
		wantIDs    []int64
	}{
		{
			name: "default limit",
			req:  GetAppsRequest{Name: "Deals"},
			responses: []response{
				{http.StatusOK, `{"apps": [{"appId": "1"}, {"appId": "2"}]}`},
				{http.StatusOK, `{"apps": [{"appId": "3"}]}`},
			},
			wantBodies: []string{
				`{"name": "Deals", "limit": 2}`,
				`{"name": "Deals", "offset": 2, "limit": 2}`,
			},
			wantIDs: []int64{1, 2, 3},
		},
		{
			name: "offset and limit",
			req:  GetAppsRequest{Offset: 1, Limit: 3},
			responses: []response{
				{http.StatusOK, `{"apps": [{"appId": "2"}, {"appId": "3"}, {"appId": "4"}]}`},
				{http.StatusOK, `{"apps": []}`},
			},
			wantBodies: []string{
				`{"offset": 1, "limit": 3}`,
				`{"offset": 4, "limit": 3}`,
			},
			wantIDs: []int64{2, 3, 4},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, requests := newSequenceServer(t, tc.responses...)
			c := newTestClient(t, ts)
			var ids []int64
			for app, err := range c.AllApps(t.Context(), tc.req) {
				if err != nil {
					t.Fatalf("AllApps returned error: %v", err)
				}
				ids = append(ids, app.AppID)
			}
			reqs := requests()
			if len(reqs) != len(tc.wantBodies) {
				t.Fatalf("number of requests: got %d, want %d", len(reqs), len(tc.wantBodies))
			}
			for i, body := range tc.wantBodies {
				assertRequest(t, &reqs[i], http.MethodGet, "/k/v1/apps.json", body)
			}
			if !reflect.DeepEqual(ids, tc.wantIDs) {
				t.Errorf("app IDs: got %v, want %v", ids, tc.wantIDs)
			}
		})
	}
}

func TestClientAllAppsError(t *testing.T) {
	setValue(t, &allAppsLimit, 2)
	ts, _ := newSequenceServer(t,
		response{http.StatusOK, `{"apps": [{"appId": "1"}, {"appId": "2"}]}`},
		response{http.StatusInternalServerError, `{"code": "TEST_ERROR", "id": "abc123", "message": "Test error."}`},
	)
	c := newTestClient(t, ts)
	var ids []int64
	var err error
	for app, e := range c.AllApps(t.Context(), GetAppsRequest{}) {
		if e != nil {
			err = e
			break
		}
		ids = append(ids, app.AppID)
	}
	if !IsErrorCode(err, "TEST_ERROR") {
		t.Errorf("AllApps returned unexpected error: %v", err)
	}
	if want := []int64{1, 2}; !reflect.DeepEqual(ids, want) {
		t.Errorf("app IDs: got %v, want %v", ids, want)
	}
}

func TestClientAllAppsBreak(t *testing.T) {
	setValue(t, &allAppsLimit, 2)
	ts, requests := newSequenceServer(t, response{http.StatusOK, `{"apps": [{"appId": "1"}, {"appId": "2"}]}`})
	c := newTestClient(t, ts)
	for app, err := range c.AllApps(t.Context(), GetAppsRequest{}) {
		if err != nil {
			t.Fatalf("AllApps returned error: %v", err)
		}
		if app.AppID == 2 {
			break
		}
	}
	if reqs := requests(); len(reqs) != 1 {
		t.Errorf("number of requests: got %d, want %d", len(reqs), 1)
	}
}

func TestClientAddApp(t *testing.T) {
	testCases := []struct {
		name     string
		req      AddAppRequest
		wantPath string
		wantBody string
		opts     []Option
	}{
		{
			name:     "name",
			req:      AddAppRequest{Name: "Deals"},
			wantPath: "/k/v1/preview/app.json",
			wantBody: `{"name": "Deals"}`,
		},
		{
			name:     "space and thread",
			req:      AddAppRequest{Name: "Deals", Space: 2, Thread: 3},
			wantPath: "/k/v1/preview/app.json",
			wantBody: `{"name": "Deals", "space": 2, "thread": 3}`,
		},
		{
			name:     "guest space",
			req:      AddAppRequest{Name: "Deals", Space: 4, Thread: 5},
			opts:     []Option{WithGuestSpace(4)},
			wantPath: "/k/guest/4/v1/preview/app.json",
			wantBody: `{"name": "Deals", "space": 4, "thread": 5}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{"app": "23", "revision": "1"}`)
			c := newTestClient(t, ts, tc.opts...)
			res, err := c.AddApp(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("AddApp returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPost, tc.wantPath, tc.wantBody)
			if want := (AppRevision{App: 23, Revision: 1}); res != want {
				t.Errorf("response: got %+v, want %+v", res, want)
			}
		})
	}
}

func TestClientDeployApps(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{}`)
	c := newTestClient(t, ts)
	if err := c.DeployApps(t.Context(), []AppDeployment{{App: 1, Revision: 4}, {App: 2}}); err != nil {
		t.Fatalf("DeployApps returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPost, "/k/v1/preview/app/deploy.json",
		`{"apps": [{"app": 1, "revision": 4}, {"app": 2}]}`)
}

func TestClientRevertApps(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{}`)
	c := newTestClient(t, ts, WithGuestSpace(3))
	if err := c.RevertApps(t.Context(), []AppDeployment{{App: 1}}); err != nil {
		t.Fatalf("RevertApps returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPost, "/k/guest/3/v1/preview/app/deploy.json",
		`{"apps": [{"app": 1}], "revert": true}`)
}

func TestClientGetDeployStatus(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{
		"apps": [
			{"app": "1", "status": "PROCESSING"},
			{"app": "2", "status": "SUCCESS"},
			{"app": "3", "status": "FAIL"},
			{"app": "4", "status": "CANCEL"}
		]
	}`)
	c := newTestClient(t, ts)
	statuses, err := c.GetDeployStatus(t.Context(), []int64{1, 2, 3, 4})
	if err != nil {
		t.Fatalf("GetDeployStatus returned error: %v", err)
	}
	assertRequest(t, got, http.MethodGet, "/k/v1/preview/app/deploy.json", `{"apps": [1, 2, 3, 4]}`)
	want := []AppDeployStatus{
		{App: 1, Status: DeployStatusProcessing},
		{App: 2, Status: DeployStatusSuccess},
		{App: 3, Status: DeployStatusFail},
		{App: 4, Status: DeployStatusCancel},
	}
	if !reflect.DeepEqual(statuses, want) {
		t.Errorf("statuses: got %+v, want %+v", statuses, want)
	}
}

func TestClientGetAppsStatistics(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetAppsStatisticsRequest
		opts     []Option
		wantBody string
	}{
		{
			name:     "offset and limit",
			req:      GetAppsStatisticsRequest{Offset: 100, Limit: 50},
			wantBody: `{"offset": 100, "limit": 50}`,
		},
		{
			name:     "guest space is not used",
			opts:     []Option{WithGuestSpace(3)},
			wantBody: `{}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{
				"apps": [
					{
						"id": "1",
						"name": "Deals",
						"space": {"id": "2", "name": "Sales"},
						"appGroup": "Public",
						"status": "ACTIVATED",
						"recordUpdatedAt": "2026-10-04T05:45:36.000Z",
						"recordCount": "100",
						"fieldCount": "10",
						"dailyRequestCount": "1",
						"apiTokenCount": "2",
						"webhookCount": "3",
						"storageUsage": "47000000",
						"customized": true,
						"creator": {"code": "alice", "name": "Alice"},
						"createdAt": "2026-10-01T05:45:36.000Z",
						"modifier": {"code": "bob", "name": "Bob"},
						"modifiedAt": "2026-10-02T05:45:36.000Z"
					},
					{
						"id": "4",
						"name": "Notes",
						"space": null,
						"appGroup": "",
						"status": "NOT_ACTIVATED",
						"recordUpdatedAt": "2026-10-01T00:00:00.000Z",
						"recordCount": "0",
						"fieldCount": "0",
						"dailyRequestCount": "0",
						"apiTokenCount": "0",
						"webhookCount": "0",
						"storageUsage": "0",
						"customized": false,
						"creator": {"code": "alice", "name": "Alice"},
						"createdAt": "2026-10-01T00:00:00.000Z",
						"modifier": {"code": "alice", "name": "Alice"},
						"modifiedAt": "2026-10-01T00:00:00.000Z"
					}
				]
			}`)
			c := newTestClient(t, ts, tc.opts...)
			apps, err := c.GetAppsStatistics(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetAppsStatistics returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, "/k/v1/apps/statistics.json", tc.wantBody)
			want := []AppStatistics{
				{
					ID:                1,
					Name:              "Deals",
					Space:             &AppSpace{ID: 2, Name: "Sales"},
					AppGroup:          "Public",
					Status:            AppStatusActivated,
					RecordUpdatedAt:   time.Date(2026, 10, 4, 5, 45, 36, 0, time.UTC),
					RecordCount:       100,
					FieldCount:        10,
					DailyRequestCount: 1,
					APITokenCount:     2,
					WebhookCount:      3,
					StorageUsage:      47000000,
					Customized:        true,
					Creator:           Entity{Code: "alice", Name: "Alice"},
					CreatedAt:         time.Date(2026, 10, 1, 5, 45, 36, 0, time.UTC),
					Modifier:          Entity{Code: "bob", Name: "Bob"},
					ModifiedAt:        time.Date(2026, 10, 2, 5, 45, 36, 0, time.UTC),
				},
				{
					ID:              4,
					Name:            "Notes",
					Status:          AppStatusNotActivated,
					RecordUpdatedAt: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
					Creator:         Entity{Code: "alice", Name: "Alice"},
					CreatedAt:       time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
					Modifier:        Entity{Code: "alice", Name: "Alice"},
					ModifiedAt:      time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
				},
			}
			if !reflect.DeepEqual(apps, want) {
				t.Errorf("apps: got %+v, want %+v", apps, want)
			}
		})
	}
}

func TestClientMoveApp(t *testing.T) {
	testCases := []struct {
		name     string
		space    int64
		opts     []Option
		wantBody string
	}{
		{
			name:     "to a space",
			space:    2,
			wantBody: `{"app": 1, "space": 2}`,
		},
		{
			name:     "out of the space",
			wantBody: `{"app": 1, "space": null}`,
		},
		{
			name:     "guest space is not used",
			space:    2,
			opts:     []Option{WithGuestSpace(3)},
			wantBody: `{"app": 1, "space": 2}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{}`)
			c := newTestClient(t, ts, tc.opts...)
			if err := c.MoveApp(t.Context(), 1, tc.space); err != nil {
				t.Fatalf("MoveApp returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPost, "/k/v1/app/move.json", tc.wantBody)
		})
	}
}

func TestClientWaitForDeploy(t *testing.T) {
	testCases := []struct {
		name      string
		responses []response
		wantErr   string
	}{
		{
			name: "success",
			responses: []response{
				{http.StatusOK, `{"apps": [{"app": "1", "status": "PROCESSING"}, {"app": "2", "status": "SUCCESS"}]}`},
				{http.StatusOK, `{"apps": [{"app": "1", "status": "PROCESSING"}, {"app": "2", "status": "SUCCESS"}]}`},
				{http.StatusOK, `{"apps": [{"app": "1", "status": "SUCCESS"}, {"app": "2", "status": "SUCCESS"}]}`},
			},
		},
		{
			name: "failure",
			responses: []response{
				{http.StatusOK, `{"apps": [{"app": "1", "status": "PROCESSING"}, {"app": "2", "status": "PROCESSING"}]}`},
				{http.StatusOK, `{"apps": [{"app": "1", "status": "FAIL"}, {"app": "2", "status": "CANCEL"}]}`},
			},
			wantErr: "kintone: deployment failed: app 1: FAIL, app 2: CANCEL",
		},
		{
			name: "error response",
			responses: []response{
				{http.StatusOK, `{"apps": [{"app": "1", "status": "PROCESSING"}]}`},
				{http.StatusInternalServerError, `{"code": "TEST_ERROR", "id": "abc123", "message": "Test error."}`},
			},
			wantErr: "kintone: Test error. (status: 500, code: TEST_ERROR, id: abc123)",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ts, requests := newSequenceServer(t, tc.responses...)
				c := newTestClient(t, ts)
				start := time.Now()
				err := c.WaitForDeploy(t.Context(), []int64{1, 2})
				if tc.wantErr == "" {
					if err != nil {
						t.Fatalf("WaitForDeploy returned error: %v", err)
					}
				} else if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("WaitForDeploy should return error %q but got: %v", tc.wantErr, err)
				}
				if got, want := time.Since(start), time.Duration(len(tc.responses)-1)*time.Second; got != want {
					t.Errorf("elapsed time: got %v, want %v", got, want)
				}
				reqs := requests()
				if len(reqs) != len(tc.responses) {
					t.Fatalf("number of requests: got %d, want %d", len(reqs), len(tc.responses))
				}
				for _, req := range reqs {
					assertRequest(t, &req, http.MethodGet, "/k/v1/preview/app/deploy.json", `{"apps": [1, 2]}`)
				}
			})
		})
	}
}

func TestClientWaitForDeployContextCanceled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ts, requests := newSequenceServer(t,
			response{http.StatusOK, `{"apps": [{"app": "1", "status": "PROCESSING"}]}`},
		)
		c := newTestClient(t, ts)
		ctx, cancel := context.WithCancel(t.Context())
		time.AfterFunc(500*time.Millisecond, cancel)
		start := time.Now()
		if err := c.WaitForDeploy(ctx, []int64{1}); !errors.Is(err, context.Canceled) {
			t.Errorf("WaitForDeploy should return context.Canceled but got: %v", err)
		}
		if got, want := time.Since(start), 500*time.Millisecond; got != want {
			t.Errorf("elapsed time: got %v, want %v", got, want)
		}
		if reqs := requests(); len(reqs) != 1 {
			t.Errorf("number of requests: got %d, want %d", len(reqs), 1)
		}
	})
}

func TestClientAppSettingsError(t *testing.T) {
	ts, _ := newTestServer(t, http.StatusBadRequest, `{"code": "TEST_ERROR", "id": "abc123", "message": "Test error."}`)
	c := newTestClient(t, ts)
	views, err := c.GetViews(t.Context(), GetAppSettingsRequest{App: 1})
	if views != nil || !IsErrorCode(err, "TEST_ERROR") {
		t.Errorf("GetViews should return nil and TEST_ERROR but got: %v, %v", views, err)
	}
	revision, err := c.UpdateAppPermissions(t.Context(), 1, []AppRight{}, 0)
	if revision != 0 || !IsErrorCode(err, "TEST_ERROR") {
		t.Errorf("UpdateAppPermissions should return 0 and TEST_ERROR but got: %d, %v", revision, err)
	}
	plugins, err := c.GetPlugins(t.Context(), GetPluginsRequest{})
	if plugins != nil || !IsErrorCode(err, "TEST_ERROR") {
		t.Errorf("GetPlugins should return nil and TEST_ERROR but got: %v, %v", plugins, err)
	}
}
