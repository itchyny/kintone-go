package kintone

import (
	"context"
	"encoding/json/v2"
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
	CoverType           CoverType        `json:"coverType"`
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

// CoverType is the type of the cover image of a space.
type CoverType string

// Cover types.
const (
	CoverTypeBlob   CoverType = "BLOB"
	CoverTypePreset CoverType = "PRESET"
)

// SpacePermissions is the permissions of a space.
type SpacePermissions struct {
	CreateApp CreateAppPermission `json:"createApp"`
}

// CreateAppPermission is who can create apps in a space.
type CreateAppPermission string

// Create app permissions.
const (
	CreateAppPermissionEveryone CreateAppPermission = "EVERYONE"
	CreateAppPermissionAdmin    CreateAppPermission = "ADMIN"
)

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
type UpdateSpaceRequest struct {
	ID                  int64             `json:"id"`
	Name                string            `json:"name,omitzero"`
	IsPrivate           *bool             `json:"isPrivate,omitzero"`
	FixedMember         *bool             `json:"fixedMember,omitzero"`
	UseMultiThread      *bool             `json:"useMultiThread,omitzero"`
	ShowAnnouncement    *bool             `json:"showAnnouncement,omitzero"`
	ShowThreadList      *bool             `json:"showThreadList,omitzero"`
	ShowAppList         *bool             `json:"showAppList,omitzero"`
	ShowMemberList      *bool             `json:"showMemberList,omitzero"`
	ShowRelatedLinkList *bool             `json:"showRelatedLinkList,omitzero"`
	Permissions         *SpacePermissions `json:"permissions,omitzero"`
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
		ID number `json:"id"`
	}](ctx, "space.json", params)
	if err != nil {
		return 0, err
	}
	return int64(result.ID), nil
}

// SpaceMember is a member of a space. IsImplicit is true for the users who
// belong to the space through a group or a department, and is not sent.
type SpaceMember struct {
	Entity      FieldEntity `json:"entity"`
	IsAdmin     bool        `json:"isAdmin,omitzero"`
	IsImplicit  bool        `json:"isImplicit,omitzero"`
	IncludeSubs bool        `json:"includeSubs,omitzero"`
}

type spaceMemberParam struct {
	Entity      FieldEntity `json:"entity"`
	IsAdmin     bool        `json:"isAdmin,omitzero"`
	IncludeSubs bool        `json:"includeSubs,omitzero"`
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
		IsPrivate   bool               `json:"isPrivate,omitzero"`
		IsGuest     bool               `json:"isGuest,omitzero"`
		FixedMember bool               `json:"fixedMember,omitzero"`
	}{req.ID, req.Name, spaceMemberParams(req.Members), req.IsPrivate, req.IsGuest, req.FixedMember}
	result, err := c.withoutGuestSpace().post[struct {
		ID number `json:"id"`
	}](ctx, "template/space.json", params)
	if err != nil {
		return 0, err
	}
	return int64(result.ID), nil
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
		ID                 number `json:"id"`
		AdministratorCount number `json:"administratorCount"`
		MemberCount        number `json:"memberCount"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*s = SpaceStatistics(v.statistics)
	s.ID, s.AdministratorCount, s.MemberCount = int64(v.ID), int64(v.AdministratorCount), int64(v.MemberCount)
	return nil
}

// GetSpacesStatisticsRequest is the request of [Client.GetSpacesStatistics].
type GetSpacesStatisticsRequest struct {
	Offset int64 `json:"offset,omitzero"`
	Limit  int64 `json:"limit,omitzero"`
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

// GetSpaceMembers gets the members of a space.
func (c *Client) GetSpaceMembers(ctx context.Context, id int64) ([]SpaceMember, error) {
	params := struct {
		ID int64 `json:"id"`
	}{id}
	result, err := c.get[struct {
		Members []SpaceMember `json:"members"`
	}](ctx, "space/members.json", params)
	if err != nil {
		return nil, err
	}
	return result.Members, nil
}

// UpdateSpaceMembers updates the members of a space. The members with
// IsImplicit are sent as explicit members.
func (c *Client) UpdateSpaceMembers(ctx context.Context, id int64, members []SpaceMember) error {
	params := struct {
		ID      int64              `json:"id"`
		Members []spaceMemberParam `json:"members"`
	}{id, spaceMemberParams(members)}
	_, err := c.put[struct{}](ctx, "space/members.json", params)
	return err
}
