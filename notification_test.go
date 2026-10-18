package kintone

import (
	"net/http"
	"reflect"
	"testing"
)

func TestClientGetGeneralNotifications(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetAppSettingsRequest
		wantPath string
	}{
		{
			name:     "production",
			req:      GetAppSettingsRequest{App: 1, Lang: "en"},
			wantPath: "/k/v1/app/notifications/general.json",
		},
		{
			name:     "preview",
			req:      GetAppSettingsRequest{App: 1, Preview: true},
			wantPath: "/k/v1/preview/app/notifications/general.json",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{
				"notifications": [
					{
						"entity": {"type": "USER", "code": "alice"},
						"includeSubs": false,
						"recordAdded": true,
						"recordEdited": true,
						"commentAdded": false,
						"statusChanged": false,
						"fileImported": true
					},
					{
						"entity": {"type": "ORGANIZATION", "code": "dev"},
						"includeSubs": true,
						"recordAdded": false,
						"recordEdited": false,
						"commentAdded": true,
						"statusChanged": true,
						"fileImported": false
					}
				],
				"notifyToCommenter": true,
				"revision": "3"
			}`)
			c := newTestClient(t, ts)
			res, err := c.GetGeneralNotifications(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetGeneralNotifications returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, tc.wantPath, `{"app": 1}`)
			want := &GeneralNotifications{
				Notifications: []GeneralNotification{
					{
						Entity:       FieldEntity{Type: EntityTypeUser, Code: "alice"},
						RecordAdded:  true,
						RecordEdited: true,
						FileImported: true,
					},
					{
						Entity:        FieldEntity{Type: EntityTypeOrganization, Code: "dev"},
						IncludeSubs:   true,
						CommentAdded:  true,
						StatusChanged: true,
					},
				},
				NotifyToCommenter: new(true),
				Revision:          3,
			}
			if !reflect.DeepEqual(res, want) {
				t.Errorf("response: got %+v, want %+v", res, want)
			}
		})
	}
}

func TestClientUpdateGeneralNotifications(t *testing.T) {
	testCases := []struct {
		name     string
		settings GeneralNotifications
		revision int64
		wantBody string
	}{
		{
			name: "notifications",
			settings: GeneralNotifications{
				Notifications: []GeneralNotification{
					{Entity: FieldEntity{Type: EntityTypeFieldEntity, Code: "owner"}, RecordEdited: true},
				},
				NotifyToCommenter: new(false),
				Revision:          9,
			},
			revision: 3,
			wantBody: `{
				"app": 1,
				"notifications": [{"entity": {"type": "FIELD_ENTITY", "code": "owner"}, "recordEdited": true}],
				"notifyToCommenter": false,
				"revision": 3
			}`,
		},
		{
			name:     "remove notifications",
			settings: GeneralNotifications{Notifications: []GeneralNotification{}},
			wantBody: `{"app": 1, "notifications": []}`,
		},
		{
			name:     "only notify to commenter",
			settings: GeneralNotifications{NotifyToCommenter: new(true)},
			wantBody: `{"app": 1, "notifyToCommenter": true}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{"revision": "4"}`)
			c := newTestClient(t, ts)
			revision, err := c.UpdateGeneralNotifications(t.Context(), 1, tc.settings, tc.revision)
			if err != nil {
				t.Fatalf("UpdateGeneralNotifications returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPut, "/k/v1/preview/app/notifications/general.json", tc.wantBody)
			if revision != 4 {
				t.Errorf("revision: got %d, want %d", revision, 4)
			}
		})
	}
}

func TestClientGetPerRecordNotifications(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetAppSettingsRequest
		wantPath string
		wantBody string
	}{
		{
			name:     "production",
			req:      GetAppSettingsRequest{App: 1, Lang: "en"},
			wantPath: "/k/v1/app/notifications/perRecord.json",
			wantBody: `{"app": 1, "lang": "en"}`,
		},
		{
			name:     "preview",
			req:      GetAppSettingsRequest{App: 1, Preview: true},
			wantPath: "/k/v1/preview/app/notifications/perRecord.json",
			wantBody: `{"app": 1}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{
				"notifications": [
					{
						"filterCond": "amount > 100",
						"title": "Large deal",
						"targets": [
							{"entity": {"type": "GROUP", "code": "sales"}, "includeSubs": false},
							{"entity": {"type": "FIELD_ENTITY", "code": "department"}, "includeSubs": true}
						]
					}
				],
				"revision": "3"
			}`)
			c := newTestClient(t, ts)
			res, err := c.GetPerRecordNotifications(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetPerRecordNotifications returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, tc.wantPath, tc.wantBody)
			want := &PerRecordNotifications{
				Notifications: []PerRecordNotification{
					{
						FilterCond: "amount > 100",
						Title:      "Large deal",
						Targets: []TargetEntity{
							{Entity: FieldEntity{Type: EntityTypeGroup, Code: "sales"}},
							{Entity: FieldEntity{Type: EntityTypeFieldEntity, Code: "department"}, IncludeSubs: true},
						},
					},
				},
				Revision: 3,
			}
			if !reflect.DeepEqual(res, want) {
				t.Errorf("response: got %+v, want %+v", res, want)
			}
		})
	}
}

func TestClientUpdatePerRecordNotifications(t *testing.T) {
	testCases := []struct {
		name          string
		notifications []PerRecordNotification
		revision      int64
		wantBody      string
	}{
		{
			name: "notifications",
			notifications: []PerRecordNotification{
				{
					FilterCond: "amount > 100",
					Title:      "Large deal",
					Targets:    []TargetEntity{{Entity: FieldEntity{Type: EntityTypeUser, Code: "alice"}}},
				},
			},
			revision: 3,
			wantBody: `{
				"app": 1,
				"notifications": [{
					"filterCond": "amount > 100",
					"title": "Large deal",
					"targets": [{"entity": {"type": "USER", "code": "alice"}}]
				}],
				"revision": 3
			}`,
		},
		{
			name:     "nil notifications",
			wantBody: `{"app": 1, "notifications": []}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{"revision": "4"}`)
			c := newTestClient(t, ts)
			revision, err := c.UpdatePerRecordNotifications(t.Context(), 1, tc.notifications, tc.revision)
			if err != nil {
				t.Fatalf("UpdatePerRecordNotifications returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPut, "/k/v1/preview/app/notifications/perRecord.json", tc.wantBody)
			if revision != 4 {
				t.Errorf("revision: got %d, want %d", revision, 4)
			}
		})
	}
}

func TestClientGetReminderNotifications(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetAppSettingsRequest
		response string
		wantPath string
		wantBody string
		want     *ReminderNotifications
	}{
		{
			name:     "notifications",
			req:      GetAppSettingsRequest{App: 1, Lang: "en"},
			wantPath: "/k/v1/app/notifications/reminder.json",
			wantBody: `{"app": 1, "lang": "en"}`,
			response: `{
				"notifications": [
					{
						"timing": {"code": "due", "daysLater": "-1", "hoursLater": "0"},
						"filterCond": "",
						"title": "Due soon",
						"targets": [{"entity": {"type": "FIELD_ENTITY", "code": "owner"}, "includeSubs": false}]
					},
					{
						"timing": {"code": "start", "daysLater": "2", "time": "09:30"},
						"filterCond": "status = \"Open\"",
						"title": "",
						"targets": [{"entity": {"type": "USER", "code": "alice"}, "includeSubs": false}]
					}
				],
				"timezone": "Asia/Tokyo",
				"revision": "3"
			}`,
			want: &ReminderNotifications{
				Notifications: []ReminderNotification{
					{
						Timing:  ReminderTiming{Code: "due", DaysLater: -1, HoursLater: new(0)},
						Title:   "Due soon",
						Targets: []TargetEntity{{Entity: FieldEntity{Type: EntityTypeFieldEntity, Code: "owner"}}},
					},
					{
						Timing:     ReminderTiming{Code: "start", DaysLater: 2, Time: "09:30"},
						FilterCond: `status = "Open"`,
						Targets:    []TargetEntity{{Entity: FieldEntity{Type: EntityTypeUser, Code: "alice"}}},
					},
				},
				Timezone: "Asia/Tokyo",
				Revision: 3,
			},
		},
		{
			name:     "never set",
			req:      GetAppSettingsRequest{App: 1, Preview: true},
			wantPath: "/k/v1/preview/app/notifications/reminder.json",
			wantBody: `{"app": 1}`,
			response: `{"notifications": [], "timezone": null, "revision": "1"}`,
			want:     &ReminderNotifications{Notifications: []ReminderNotification{}, Revision: 1},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, tc.response)
			c := newTestClient(t, ts)
			res, err := c.GetReminderNotifications(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetReminderNotifications returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, tc.wantPath, tc.wantBody)
			if !reflect.DeepEqual(res, tc.want) {
				t.Errorf("response: got %+v, want %+v", res, tc.want)
			}
		})
	}
}

func TestClientUpdateReminderNotifications(t *testing.T) {
	testCases := []struct {
		name     string
		settings ReminderNotifications
		revision int64
		wantBody string
	}{
		{
			name: "notifications",
			settings: ReminderNotifications{
				Notifications: []ReminderNotification{
					{
						Timing:  ReminderTiming{Code: "due", DaysLater: 0, HoursLater: new(-2)},
						Targets: []TargetEntity{{Entity: FieldEntity{Type: EntityTypeOrganization, Code: "dev"}, IncludeSubs: true}},
					},
					{
						Timing:  ReminderTiming{Code: "start", DaysLater: 1, Time: "08:00"},
						Title:   "Starts tomorrow",
						Targets: []TargetEntity{{Entity: FieldEntity{Type: EntityTypeUser, Code: "alice"}}},
					},
				},
				Timezone: "Asia/Tokyo",
				Revision: 9,
			},
			revision: 3,
			wantBody: `{
				"app": 1,
				"notifications": [
					{
						"timing": {"code": "due", "daysLater": "0", "hoursLater": "-2"},
						"targets": [{"entity": {"type": "ORGANIZATION", "code": "dev"}, "includeSubs": true}]
					},
					{
						"timing": {"code": "start", "daysLater": "1", "time": "08:00"},
						"title": "Starts tomorrow",
						"targets": [{"entity": {"type": "USER", "code": "alice"}}]
					}
				],
				"timezone": "Asia/Tokyo",
				"revision": 3
			}`,
		},
		{
			name:     "only timezone",
			settings: ReminderNotifications{Timezone: "UTC"},
			wantBody: `{"app": 1, "timezone": "UTC"}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{"revision": "4"}`)
			c := newTestClient(t, ts)
			revision, err := c.UpdateReminderNotifications(t.Context(), 1, tc.settings, tc.revision)
			if err != nil {
				t.Fatalf("UpdateReminderNotifications returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPut, "/k/v1/preview/app/notifications/reminder.json", tc.wantBody)
			if revision != 4 {
				t.Errorf("revision: got %d, want %d", revision, 4)
			}
		})
	}
}
