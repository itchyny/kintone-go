package kintone

import (
	"net/http"
	"testing"
)

func TestClientAddGuests(t *testing.T) {
	testCases := []struct {
		name string
		opts []Option
	}{
		{name: "guests"},
		{name: "guest space is not used", opts: []Option{WithGuestSpace(4)}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{}`)
			c := newTestClient(t, ts, tc.opts...)
			err := c.AddGuests(t.Context(), []Guest{
				{
					Code:           "guest@example.com",
					Password:       "password",
					Timezone:       "Asia/Tokyo",
					Locale:         LocaleJa,
					Name:           "Guest",
					SurNameReading: "げすと",
					Company:        "Example",
				},
				{Code: "guest2@example.com", Password: "password", Timezone: "UTC", Name: "Guest 2"},
			})
			if err != nil {
				t.Fatalf("AddGuests returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPost, "/k/v1/guests.json", `{"guests": [
				{
					"code": "guest@example.com",
					"password": "password",
					"timezone": "Asia/Tokyo",
					"locale": "ja",
					"name": "Guest",
					"surNameReading": "げすと",
					"company": "Example"
				},
				{"code": "guest2@example.com", "password": "password", "timezone": "UTC", "name": "Guest 2"}
			]}`)
		})
	}
}

func TestClientDeleteGuests(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{}`)
	c := newTestClient(t, ts, WithGuestSpace(4))
	if err := c.DeleteGuests(t.Context(), []string{"guest@example.com"}); err != nil {
		t.Fatalf("DeleteGuests returned error: %v", err)
	}
	assertRequest(t, got, http.MethodDelete, "/k/v1/guests.json", `{"guests": ["guest@example.com"]}`)
}

func TestClientUpdateGuestMembers(t *testing.T) {
	testCases := []struct {
		name string
		opts []Option
	}{
		{name: "client without guest space"},
		{name: "client with another guest space", opts: []Option{WithGuestSpace(5)}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{}`)
			c := newTestClient(t, ts, tc.opts...)
			if err := c.UpdateGuestMembers(t.Context(), 4, []string{"guest@example.com", "guest2@example.com"}); err != nil {
				t.Fatalf("UpdateGuestMembers returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPut, "/k/guest/4/v1/space/guests.json",
				`{"id": 4, "guests": ["guest@example.com", "guest2@example.com"]}`)
		})
	}
}
