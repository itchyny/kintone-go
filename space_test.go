package kintone

import (
	"encoding/json/v2"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestClientGetSpace(t *testing.T) {
	testCases := []struct {
		name     string
		opts     []Option
		response string
		wantPath string
		want     *Space
	}{
		{
			name:     "multi thread space",
			wantPath: "/k/v1/space.json",
			response: `{
				"id": "2",
				"name": "Sales",
				"defaultThread": "3",
				"isPrivate": true,
				"creator": {"code": "alice", "name": "Alice"},
				"modifier": {"code": "bob", "name": "Bob"},
				"memberCount": "10",
				"coverType": "PRESET",
				"coverKey": "GREEN",
				"coverUrl": "https://example.cybozu.com/cover.png",
				"body": "<p>Welcome</p>",
				"useMultiThread": true,
				"isGuest": false,
				"attachedApps": [
					{
						"threadId": "3",
						"appId": "1",
						"code": "DEAL",
						"name": "Deals",
						"description": "",
						"createdAt": "2026-10-01T05:14:05.000Z",
						"creator": {"code": "alice", "name": "Alice"},
						"modifiedAt": "2026-10-02T05:14:05.000Z",
						"modifier": {"code": "bob", "name": "Bob"}
					}
				],
				"fixedMember": false,
				"showAnnouncement": true,
				"showThreadList": true,
				"showAppList": false,
				"showMemberList": true,
				"showRelatedLinkList": false,
				"permissions": {"createApp": "ADMIN"}
			}`,
			want: &Space{
				ID:             2,
				Name:           "Sales",
				DefaultThread:  3,
				IsPrivate:      true,
				Creator:        Entity{Code: "alice", Name: "Alice"},
				Modifier:       Entity{Code: "bob", Name: "Bob"},
				MemberCount:    10,
				CoverType:      CoverTypePreset,
				CoverKey:       "GREEN",
				CoverURL:       "https://example.cybozu.com/cover.png",
				Body:           "<p>Welcome</p>",
				UseMultiThread: true,
				AttachedApps: []App{
					{
						AppID:      1,
						Code:       "DEAL",
						Name:       "Deals",
						ThreadID:   3,
						CreatedAt:  time.Date(2026, 10, 1, 5, 14, 5, 0, time.UTC),
						Creator:    Entity{Code: "alice", Name: "Alice"},
						ModifiedAt: time.Date(2026, 10, 2, 5, 14, 5, 0, time.UTC),
						Modifier:   Entity{Code: "bob", Name: "Bob"},
					},
				},
				ShowAnnouncement:    new(true),
				ShowThreadList:      new(true),
				ShowAppList:         new(false),
				ShowMemberList:      new(true),
				ShowRelatedLinkList: new(false),
				Permissions:         SpacePermissions{CreateApp: CreateAppPermissionAdmin},
			},
		},
		{
			name:     "guest space",
			opts:     []Option{WithGuestSpace(4)},
			wantPath: "/k/guest/4/v1/space.json",
			response: `{
				"id": "4",
				"name": "Partners",
				"defaultThread": "5",
				"isPrivate": true,
				"creator": {"code": "", "name": ""},
				"modifier": {"code": "", "name": ""},
				"memberCount": "3",
				"coverType": "BLOB",
				"coverKey": "key1",
				"coverUrl": "",
				"body": null,
				"useMultiThread": false,
				"isGuest": true,
				"attachedApps": [],
				"fixedMember": true,
				"showAnnouncement": null,
				"showThreadList": null,
				"showAppList": null,
				"showMemberList": null,
				"showRelatedLinkList": null,
				"permissions": {"createApp": "EVERYONE"}
			}`,
			want: &Space{
				ID:            4,
				Name:          "Partners",
				DefaultThread: 5,
				IsPrivate:     true,
				MemberCount:   3,
				CoverType:     CoverTypeBlob,
				CoverKey:      "key1",
				IsGuest:       true,
				AttachedApps:  []App{},
				FixedMember:   true,
				Permissions:   SpacePermissions{CreateApp: CreateAppPermissionEveryone},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, tc.response)
			c := newTestClient(t, ts, tc.opts...)
			space, err := c.GetSpace(t.Context(), tc.want.ID)
			if err != nil {
				t.Fatalf("GetSpace returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, tc.wantPath, fmt.Sprintf(`{"id": %d}`, tc.want.ID))
			if !reflect.DeepEqual(space, tc.want) {
				t.Errorf("space: got %+v, want %+v", space, tc.want)
			}
		})
	}
}

func TestClientUpdateSpace(t *testing.T) {
	testCases := []struct {
		name     string
		req      UpdateSpaceRequest
		wantBody string
	}{
		{
			name: "settings",
			req: UpdateSpaceRequest{
				ID:               2,
				Name:             "Sales",
				IsPrivate:        new(false),
				UseMultiThread:   new(true),
				ShowAnnouncement: new(false),
				Permissions:      &SpacePermissions{CreateApp: CreateAppPermissionEveryone},
			},
			wantBody: `{
				"id": 2,
				"name": "Sales",
				"isPrivate": false,
				"useMultiThread": true,
				"showAnnouncement": false,
				"permissions": {"createApp": "EVERYONE"}
			}`,
		},
		{
			name:     "only ID",
			req:      UpdateSpaceRequest{ID: 2},
			wantBody: `{"id": 2}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{}`)
			c := newTestClient(t, ts)
			if err := c.UpdateSpace(t.Context(), tc.req); err != nil {
				t.Fatalf("UpdateSpace returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPut, "/k/v1/space.json", tc.wantBody)
		})
	}
}

func TestClientDeleteSpace(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{}`)
	c := newTestClient(t, ts, WithGuestSpace(4))
	if err := c.DeleteSpace(t.Context(), 4); err != nil {
		t.Fatalf("DeleteSpace returned error: %v", err)
	}
	assertRequest(t, got, http.MethodDelete, "/k/guest/4/v1/space.json", `{"id": 4}`)
}

func TestClientUpdateSpaceBody(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{}`)
	c := newTestClient(t, ts)
	if err := c.UpdateSpaceBody(t.Context(), 2, "<p>Welcome</p>"); err != nil {
		t.Fatalf("UpdateSpaceBody returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPut, "/k/v1/space/body.json", `{"id": 2, "body": "<p>Welcome</p>"}`)
}

func TestClientAddSpace(t *testing.T) {
	testCases := []struct {
		name     string
		response string
		opts     []Option
		want     int64
		wantErr  bool
	}{
		{name: "string ID", response: `{"id": "7"}`, want: 7},
		{name: "numeric ID", response: `{"id": 7}`, want: 7},
		{name: "guest space is not used", response: `{"id": "7"}`, opts: []Option{WithGuestSpace(4)}, want: 7},
		{name: "invalid ID", response: `{"id": "x"}`, wantErr: true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, tc.response)
			c := newTestClient(t, ts, tc.opts...)
			id, err := c.AddSpace(t.Context(), "Sales")
			if tc.wantErr {
				if err == nil {
					t.Error("AddSpace should return error")
				}
				return
			}
			if err != nil {
				t.Fatalf("AddSpace returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPost, "/k/v1/space.json", `{"name": "Sales"}`)
			if id != tc.want {
				t.Errorf("space ID: got %d, want %d", id, tc.want)
			}
		})
	}
}

func TestClientAddSpaceFromTemplate(t *testing.T) {
	testCases := []struct {
		name     string
		req      AddSpaceFromTemplateRequest
		opts     []Option
		wantBody string
	}{
		{
			name: "space",
			req: AddSpaceFromTemplateRequest{
				ID:   1,
				Name: "Sales",
				Members: []SpaceMember{
					{Entity: FieldEntity{Type: EntityTypeUser, Code: "alice"}, IsAdmin: true, IsImplicit: true},
					{Entity: FieldEntity{Type: EntityTypeOrganization, Code: "sales"}, IncludeSubs: true},
				},
				IsPrivate:   true,
				FixedMember: true,
			},
			wantBody: `{
				"id": 1,
				"name": "Sales",
				"members": [
					{"entity": {"type": "USER", "code": "alice"}, "isAdmin": true},
					{"entity": {"type": "ORGANIZATION", "code": "sales"}, "includeSubs": true}
				],
				"isPrivate": true,
				"fixedMember": true
			}`,
		},
		{
			name: "guest space is not used",
			req: AddSpaceFromTemplateRequest{
				ID:      1,
				Name:    "Partners",
				Members: []SpaceMember{{Entity: FieldEntity{Type: EntityTypeUser, Code: "alice"}, IsAdmin: true}},
				IsGuest: true,
			},
			opts: []Option{WithGuestSpace(4)},
			wantBody: `{
				"id": 1,
				"name": "Partners",
				"members": [{"entity": {"type": "USER", "code": "alice"}, "isAdmin": true}],
				"isGuest": true
			}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{"id": "7"}`)
			c := newTestClient(t, ts, tc.opts...)
			id, err := c.AddSpaceFromTemplate(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("AddSpaceFromTemplate returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPost, "/k/v1/template/space.json", tc.wantBody)
			if id != 7 {
				t.Errorf("space ID: got %d, want %d", id, 7)
			}
		})
	}
}

func TestClientGetSpacesStatistics(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{
		"spaces": [
			{
				"id": "1",
				"name": "Sales",
				"administratorCount": "2",
				"memberCount": "10",
				"isPrivate": true,
				"isGuest": false,
				"creator": {"code": "alice", "name": "Alice"},
				"createdAt": "2026-10-04T05:45:36.000Z",
				"modifier": {"code": "bob", "name": "Bob"},
				"modifiedAt": "2026-10-05T05:45:36.000Z"
			},
			{
				"id": 3,
				"name": "Dev",
				"administratorCount": 1,
				"memberCount": 5,
				"isPrivate": false,
				"isGuest": true,
				"creator": {"code": "alice", "name": "Alice"},
				"createdAt": "2026-10-04T05:45:36.000Z",
				"modifier": {"code": "alice", "name": "Alice"},
				"modifiedAt": "2026-10-04T05:45:36.000Z"
			}
		]
	}`)
	c := newTestClient(t, ts, WithGuestSpace(4))
	spaces, err := c.GetSpacesStatistics(t.Context(), GetSpacesStatisticsRequest{Offset: 100, Limit: 50})
	if err != nil {
		t.Fatalf("GetSpacesStatistics returned error: %v", err)
	}
	assertRequest(t, got, http.MethodGet, "/k/v1/spaces/statistics.json", `{"offset": 100, "limit": 50}`)
	want := []SpaceStatistics{
		{
			ID:                 1,
			Name:               "Sales",
			AdministratorCount: 2,
			MemberCount:        10,
			IsPrivate:          true,
			Creator:            Entity{Code: "alice", Name: "Alice"},
			CreatedAt:          time.Date(2026, 10, 4, 5, 45, 36, 0, time.UTC),
			Modifier:           Entity{Code: "bob", Name: "Bob"},
			ModifiedAt:         time.Date(2026, 10, 5, 5, 45, 36, 0, time.UTC),
		},
		{
			ID:                 3,
			Name:               "Dev",
			AdministratorCount: 1,
			MemberCount:        5,
			IsGuest:            true,
			Creator:            Entity{Code: "alice", Name: "Alice"},
			CreatedAt:          time.Date(2026, 10, 4, 5, 45, 36, 0, time.UTC),
			Modifier:           Entity{Code: "alice", Name: "Alice"},
			ModifiedAt:         time.Date(2026, 10, 4, 5, 45, 36, 0, time.UTC),
		},
	}
	if !reflect.DeepEqual(spaces, want) {
		t.Errorf("spaces: got %+v, want %+v", spaces, want)
	}
}

func TestSpaceStatisticsUnmarshalJSONError(t *testing.T) {
	for _, input := range []string{`"x"`, `{"id": "x"}`, `{"administratorCount": true}`, `{"memberCount": "1.5"}`} {
		var s SpaceStatistics
		if err := json.Unmarshal([]byte(input), &s); err == nil {
			t.Errorf("json.Unmarshal(%s) should return error but got %+v", input, s)
		}
	}
}

func TestClientAddSpaceFromTemplateID(t *testing.T) {
	for _, tc := range []struct {
		response string
		want     int64
		wantErr  bool
	}{
		{response: `{"id": "7"}`, want: 7},
		{response: `{"id": 7}`, want: 7},
		{response: `{"id": "x"}`, wantErr: true},
	} {
		ts, _ := newTestServer(t, http.StatusOK, tc.response)
		c := newTestClient(t, ts)
		id, err := c.AddSpaceFromTemplate(t.Context(), AddSpaceFromTemplateRequest{ID: 1, Name: "Sales"})
		if tc.wantErr {
			if err == nil {
				t.Errorf("AddSpaceFromTemplate with %s should return error", tc.response)
			}
			continue
		}
		if err != nil || id != tc.want {
			t.Errorf("AddSpaceFromTemplate with %s: got %d, %v, want %d", tc.response, id, err, tc.want)
		}
	}
}

func TestClientGetSpaceMembers(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{
		"members": [
			{"entity": {"type": "USER", "code": "alice"}, "isAdmin": true, "isImplicit": false},
			{"entity": {"type": "USER", "code": "bob"}, "isAdmin": false, "isImplicit": true},
			{"entity": {"type": "ORGANIZATION", "code": "sales"}, "isAdmin": false, "includeSubs": true}
		]
	}`)
	c := newTestClient(t, ts)
	members, err := c.GetSpaceMembers(t.Context(), 2)
	if err != nil {
		t.Fatalf("GetSpaceMembers returned error: %v", err)
	}
	assertRequest(t, got, http.MethodGet, "/k/v1/space/members.json", `{"id": 2}`)
	want := []SpaceMember{
		{Entity: FieldEntity{Type: EntityTypeUser, Code: "alice"}, IsAdmin: true},
		{Entity: FieldEntity{Type: EntityTypeUser, Code: "bob"}, IsImplicit: true},
		{Entity: FieldEntity{Type: EntityTypeOrganization, Code: "sales"}, IncludeSubs: true},
	}
	if !reflect.DeepEqual(members, want) {
		t.Errorf("members: got %+v, want %+v", members, want)
	}
}

func TestClientUpdateSpaceMembers(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{}`)
	c := newTestClient(t, ts)
	err := c.UpdateSpaceMembers(t.Context(), 2, []SpaceMember{
		{Entity: FieldEntity{Type: EntityTypeUser, Code: "alice"}, IsAdmin: true},
		{Entity: FieldEntity{Type: EntityTypeGroup, Code: "everyone"}, IsImplicit: true},
	})
	if err != nil {
		t.Fatalf("UpdateSpaceMembers returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPut, "/k/v1/space/members.json", `{
		"id": 2,
		"members": [
			{"entity": {"type": "USER", "code": "alice"}, "isAdmin": true},
			{"entity": {"type": "GROUP", "code": "everyone"}}
		]
	}`)
}
