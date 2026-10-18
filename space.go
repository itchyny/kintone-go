package kintone

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Space is the information of a space. The portal settings ShowAnnouncement,
// ShowThreadList, ShowAppList, ShowMemberList and ShowRelatedLinkList are nil
// for single thread spaces and guest spaces.
type Space struct {
	ID                  int64            `json:"id,string"`
	Name                string           `json:"name"`
	DefaultThread       int64            `json:"defaultThread,string"`
	IsPrivate           bool             `json:"isPrivate"`
	Creator             Entity           `json:"creator"`
	Modifier            Entity           `json:"modifier"`
	MemberCount         int64            `json:"memberCount,string"`
	CoverType           string           `json:"coverType"`
	CoverKey            string           `json:"coverKey"`
	CoverURL            string           `json:"coverUrl"`
	Body                string           `json:"body"`
	UseMultiThread      bool             `json:"useMultiThread"`
	IsGuest             bool             `json:"isGuest"`
	AttachedApps        []App            `json:"attachedApps"`
	FixedMember         bool             `json:"fixedMember"`
	ShowAnnouncement    *bool            `json:"showAnnouncement"`
	ShowThreadList      *bool            `json:"showThreadList"`
	ShowAppList         *bool            `json:"showAppList"`
	ShowMemberList      *bool            `json:"showMemberList"`
	ShowRelatedLinkList *bool            `json:"showRelatedLinkList"`
	Permissions         SpacePermissions `json:"permissions"`
}

// SpacePermissions is the permissions of a space.
type SpacePermissions struct {
	CreateApp string `json:"createApp"`
}

// GetSpace gets the information of a space.
func (c *Client) GetSpace(ctx context.Context, id int64) (*Space, error) {
	params := struct {
		ID int64 `json:"id"`
	}{id}
	result, err := c.get[Space](ctx, "space.json", params)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateSpaceRequest is the request of [Client.UpdateSpace].
// The settings not set are not changed.
type UpdateSpaceRequest struct {
	ID                  int64             `json:"id"`
	Name                string            `json:"name,omitempty"`
	IsPrivate           *bool             `json:"isPrivate,omitempty"`
	FixedMember         *bool             `json:"fixedMember,omitempty"`
	UseMultiThread      *bool             `json:"useMultiThread,omitempty"`
	ShowAnnouncement    *bool             `json:"showAnnouncement,omitempty"`
	ShowThreadList      *bool             `json:"showThreadList,omitempty"`
	ShowAppList         *bool             `json:"showAppList,omitempty"`
	ShowMemberList      *bool             `json:"showMemberList,omitempty"`
	ShowRelatedLinkList *bool             `json:"showRelatedLinkList,omitempty"`
	Permissions         *SpacePermissions `json:"permissions,omitempty"`
}

// UpdateSpace updates the settings of a space.
func (c *Client) UpdateSpace(ctx context.Context, req UpdateSpaceRequest) error {
	_, err := c.put[struct{}](ctx, "space.json", req)
	return err
}

// DeleteSpace deletes a space.
func (c *Client) DeleteSpace(ctx context.Context, id int64) error {
	params := struct {
		ID int64 `json:"id"`
	}{id}
	_, err := c.delete[struct{}](ctx, "space.json", params)
	return err
}

// UpdateSpaceBody updates the body of a space in HTML.
func (c *Client) UpdateSpaceBody(ctx context.Context, id int64, body string) error {
	params := struct {
		ID   int64  `json:"id"`
		Body string `json:"body"`
	}{id, body}
	_, err := c.put[struct{}](ctx, "space/body.json", params)
	return err
}

// AddSpace adds a public single thread space and returns the space ID. This
// uses an API in the API lab of kintone. The guest space of the client is not
// used.
func (c *Client) AddSpace(ctx context.Context, name string) (int64, error) {
	params := struct {
		Name string `json:"name"`
	}{name}
	result, err := c.withoutGuestSpace().post[struct {
		ID json.Number `json:"id"`
	}](ctx, "space.json", params)
	if err != nil {
		return 0, err
	}
	id, err := result.ID.Int64()
	if err != nil {
		return 0, fmt.Errorf("kintone: decode response: %w", err)
	}
	return id, nil
}

// SpaceMember is a member of a space. IsImplicit is true for the users who
// belong to the space through a group or a department, and is not sent.
type SpaceMember struct {
	Entity      FieldEntity `json:"entity"`
	IsAdmin     bool        `json:"isAdmin,omitempty"`
	IsImplicit  bool        `json:"isImplicit,omitempty"`
	IncludeSubs bool        `json:"includeSubs,omitempty"`
}

type spaceMemberParam struct {
	Entity      FieldEntity `json:"entity"`
	IsAdmin     bool        `json:"isAdmin,omitempty"`
	IncludeSubs bool        `json:"includeSubs,omitempty"`
}

func spaceMemberParams(members []SpaceMember) []spaceMemberParam {
	params := make([]spaceMemberParam, len(members))
	for i, m := range members {
		params[i] = spaceMemberParam{m.Entity, m.IsAdmin, m.IncludeSubs}
	}
	return params
}

// AddSpaceFromTemplateRequest is the request of [Client.AddSpaceFromTemplate].
// ID is the ID of the space template.
type AddSpaceFromTemplateRequest struct {
	ID          int64
	Name        string
	Members     []SpaceMember
	IsPrivate   bool
	IsGuest     bool
	FixedMember bool
}

// AddSpaceFromTemplate adds a space from a space template and returns the
// space ID. The guest space of the client is not used.
func (c *Client) AddSpaceFromTemplate(ctx context.Context, req AddSpaceFromTemplateRequest) (int64, error) {
	params := struct {
		ID          int64              `json:"id"`
		Name        string             `json:"name"`
		Members     []spaceMemberParam `json:"members"`
		IsPrivate   bool               `json:"isPrivate,omitempty"`
		IsGuest     bool               `json:"isGuest,omitempty"`
		FixedMember bool               `json:"fixedMember,omitempty"`
	}{req.ID, req.Name, spaceMemberParams(req.Members), req.IsPrivate, req.IsGuest, req.FixedMember}
	result, err := c.withoutGuestSpace().post[struct {
		ID json.Number `json:"id"`
	}](ctx, "template/space.json", params)
	if err != nil {
		return 0, err
	}
	id, err := result.ID.Int64()
	if err != nil {
		return 0, fmt.Errorf("kintone: decode response: %w", err)
	}
	return id, nil
}

// SpaceStatistics is the usage of a space.
type SpaceStatistics struct {
	ID                 int64     `json:"id,string"`
	Name               string    `json:"name"`
	AdministratorCount int64     `json:"administratorCount,string"`
	MemberCount        int64     `json:"memberCount,string"`
	IsPrivate          bool      `json:"isPrivate"`
	IsGuest            bool      `json:"isGuest"`
	Creator            Entity    `json:"creator"`
	CreatedAt          time.Time `json:"createdAt"`
	Modifier           Entity    `json:"modifier"`
	ModifiedAt         time.Time `json:"modifiedAt"`
}

// UnmarshalJSON decodes the space statistics, accepting both numbers and
// strings for the ID and the counts.
func (s *SpaceStatistics) UnmarshalJSON(b []byte) error {
	type statistics SpaceStatistics
	var v struct {
		statistics
		ID                 json.Number `json:"id"`
		AdministratorCount json.Number `json:"administratorCount"`
		MemberCount        json.Number `json:"memberCount"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*s = SpaceStatistics(v.statistics)
	for _, f := range []struct {
		n json.Number
		p *int64
	}{{v.ID, &s.ID}, {v.AdministratorCount, &s.AdministratorCount}, {v.MemberCount, &s.MemberCount}} {
		if f.n == "" {
			continue
		}
		var err error
		if *f.p, err = f.n.Int64(); err != nil {
			return err
		}
	}
	return nil
}

// GetSpacesStatisticsRequest is the request of [Client.GetSpacesStatistics].
type GetSpacesStatisticsRequest struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

// GetSpacesStatistics gets the usage of spaces. The guest space of the client
// is not used.
func (c *Client) GetSpacesStatistics(ctx context.Context, req GetSpacesStatisticsRequest) ([]SpaceStatistics, error) {
	result, err := c.withoutGuestSpace().get[struct {
		Spaces []SpaceStatistics `json:"spaces"`
	}](ctx, "spaces/statistics.json", req)
	if err != nil {
		return nil, err
	}
	return result.Spaces, nil
}
