package kintone

import (
	"net/http"
	"reflect"
	"testing"
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
