package kintone

import (
	"context"
	"net/http"
)

// GeneralNotifications is the general notification settings of an app.
type GeneralNotifications struct {
	Notifications     []GeneralNotification `json:"notifications,omitzero"`
	NotifyToCommenter *bool                 `json:"notifyToCommenter,omitzero"`
	Revision          int64                 `json:"revision,string,omitzero"`
}

// GeneralNotification is a general notification setting.
type GeneralNotification struct {
	Entity        FieldEntity `json:"entity"`
	IncludeSubs   bool        `json:"includeSubs,omitzero"`
	RecordAdded   bool        `json:"recordAdded,omitzero"`
	RecordEdited  bool        `json:"recordEdited,omitzero"`
	CommentAdded  bool        `json:"commentAdded,omitzero"`
	StatusChanged bool        `json:"statusChanged,omitzero"`
	FileImported  bool        `json:"fileImported,omitzero"`
}

// GetGeneralNotifications gets the general notification settings of an app.
// Lang of the request is not used.
func (c *Client) GetGeneralNotifications(ctx context.Context, req GetAppSettingsRequest) (*GeneralNotifications, error) {
	params := struct {
		App int64 `json:"app"`
	}{req.App}
	return c.getAppSettings[GeneralNotifications](ctx, "app/notifications/general.json", req.Preview, params)
}

// UpdateGeneralNotifications updates the general notification settings of an
// app in the preview environment and returns the new revision. Revision of the
// settings is not used. The revision is not checked if revision is zero.
func (c *Client) UpdateGeneralNotifications(ctx context.Context, app int64, settings GeneralNotifications, revision int64) (int64, error) {
	params := struct {
		App int64 `json:"app"`
		GeneralNotifications
		Revision int64 `json:"revision,omitzero"`
	}{app, settings, revision}
	return c.updateAppSettings(ctx, http.MethodPut, "app/notifications/general.json", params)
}

// PerRecordNotifications is the per record notification settings of an app.
type PerRecordNotifications struct {
	Notifications []PerRecordNotification `json:"notifications"`
	Revision      int64                   `json:"revision,string,omitzero"`
}

// PerRecordNotification is a per record notification setting.
type PerRecordNotification struct {
	FilterCond string         `json:"filterCond,omitzero"`
	Title      string         `json:"title,omitzero"`
	Targets    []TargetEntity `json:"targets"`
}

// GetPerRecordNotifications gets the per record notification settings of an
// app.
func (c *Client) GetPerRecordNotifications(ctx context.Context, req GetAppSettingsRequest) (*PerRecordNotifications, error) {
	return c.getAppSettings[PerRecordNotifications](ctx, "app/notifications/perRecord.json", req.Preview, req)
}

// UpdatePerRecordNotifications updates the per record notification settings
// of an app in the preview environment and returns the new revision. The
// revision is not checked if revision is zero.
func (c *Client) UpdatePerRecordNotifications(ctx context.Context, app int64, notifications []PerRecordNotification, revision int64) (int64, error) {
	params := struct {
		App           int64                   `json:"app"`
		Notifications []PerRecordNotification `json:"notifications"`
		Revision      int64                   `json:"revision,omitzero"`
	}{app, notifications, revision}
	return c.updateAppSettings(ctx, http.MethodPut, "app/notifications/perRecord.json", params)
}

// ReminderNotifications is the reminder notification settings of an app.
type ReminderNotifications struct {
	Notifications []ReminderNotification `json:"notifications,omitzero"`
	Timezone      string                 `json:"timezone,omitzero"`
	Revision      int64                  `json:"revision,string,omitzero"`
}

// ReminderNotification is a reminder notification setting.
type ReminderNotification struct {
	Timing     ReminderTiming `json:"timing"`
	FilterCond string         `json:"filterCond,omitzero"`
	Title      string         `json:"title,omitzero"`
	Targets    []TargetEntity `json:"targets"`
}

// ReminderTiming is the timing of a reminder notification, relative to the
// field of Code. Either HoursLater or Time is set.
type ReminderTiming struct {
	Code       string `json:"code"`
	DaysLater  int    `json:"daysLater,string"`
	HoursLater *int   `json:"hoursLater,string,omitzero"`
	Time       string `json:"time,omitzero"`
}

// GetReminderNotifications gets the reminder notification settings of an app.
func (c *Client) GetReminderNotifications(ctx context.Context, req GetAppSettingsRequest) (*ReminderNotifications, error) {
	return c.getAppSettings[ReminderNotifications](ctx, "app/notifications/reminder.json", req.Preview, req)
}

// UpdateReminderNotifications updates the reminder notification settings of
// an app in the preview environment and returns the new revision. Revision of
// the settings is not used. The revision is not checked if revision is zero.
func (c *Client) UpdateReminderNotifications(ctx context.Context, app int64, settings ReminderNotifications, revision int64) (int64, error) {
	params := struct {
		App int64 `json:"app"`
		ReminderNotifications
		Revision int64 `json:"revision,omitzero"`
	}{app, settings, revision}
	return c.updateAppSettings(ctx, http.MethodPut, "app/notifications/reminder.json", params)
}
