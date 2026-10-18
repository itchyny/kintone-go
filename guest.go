package kintone

import "context"

// Guest is a guest user to add. Code is the email address of the guest user,
// and Image is the file key of the image.
type Guest struct {
	Code             string `json:"code"`
	Password         string `json:"password"`
	Timezone         string `json:"timezone"`
	Locale           string `json:"locale,omitempty"`
	Image            string `json:"image,omitempty"`
	Name             string `json:"name"`
	SurNameReading   string `json:"surNameReading,omitempty"`
	GivenNameReading string `json:"givenNameReading,omitempty"`
	Company          string `json:"company,omitempty"`
	Division         string `json:"division,omitempty"`
	Phone            string `json:"phone,omitempty"`
	Callto           string `json:"callto,omitempty"`
}

// AddGuests adds guest users. The guest space of the client is not used.
func (c *Client) AddGuests(ctx context.Context, guests []Guest) error {
	params := struct {
		Guests []Guest `json:"guests"`
	}{guests}
	_, err := c.withoutGuestSpace().post[struct{}](ctx, "guests.json", params)
	return err
}

// DeleteGuests deletes the guest users of the email addresses. The guest space
// of the client is not used.
func (c *Client) DeleteGuests(ctx context.Context, codes []string) error {
	params := struct {
		Guests []string `json:"guests"`
	}{codes}
	_, err := c.withoutGuestSpace().delete[struct{}](ctx, "guests.json", params)
	return err
}

// UpdateGuestMembers updates the guest members of the guest space with the
// email addresses of the guest users. The guest space of the client is not
// used, and id is used instead.
func (c *Client) UpdateGuestMembers(ctx context.Context, id int64, codes []string) error {
	params := struct {
		ID     int64    `json:"id"`
		Guests []string `json:"guests"`
	}{id, codes}
	_, err := c.withGuestSpace(id).put[struct{}](ctx, "space/guests.json", params)
	return err
}
