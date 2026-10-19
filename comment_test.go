package kintone

import (
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestClientGetComments(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetCommentsRequest
		wantBody string
	}{
		{
			name:     "all parameters",
			req:      GetCommentsRequest{App: 1, Record: 2, Order: "asc", Offset: 10, Limit: 5},
			wantBody: `{"app": 1, "record": 2, "order": "asc", "offset": 10, "limit": 5}`,
		},
		{
			name:     "only required parameters",
			req:      GetCommentsRequest{App: 1, Record: 2},
			wantBody: `{"app": 1, "record": 2}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{
				"comments": [
					{
						"id": "2",
						"text": "alice \nThanks.",
						"createdAt": "2026-10-16T01:23:00Z",
						"creator": {"code": "bob", "name": "Bob"},
						"mentions": [{"code": "alice", "type": "USER"}]
					}
				],
				"older": true,
				"newer": false
			}`)
			c := newTestClient(t, ts)
			res, err := c.GetComments(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetComments returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, "/k/v1/record/comments.json", tc.wantBody)
			want := &GetCommentsResponse{
				Comments: []Comment{
					{
						ID:        2,
						Text:      "alice \nThanks.",
						CreatedAt: time.Date(2026, 10, 16, 1, 23, 0, 0, time.UTC),
						Creator:   Entity{Code: "bob", Name: "Bob"},
						Mentions:  []Mention{{Code: "alice", Type: EntityTypeUser}},
					},
				},
				Older: true,
			}
			if !reflect.DeepEqual(res, want) {
				t.Errorf("response: got %#v, want %#v", res, want)
			}
		})
	}
}

func TestClientAllComments(t *testing.T) {
	testCases := []struct {
		name       string
		req        GetCommentsRequest
		responses  []response
		wantBodies []string
		wantIDs    []int64
	}{
		{
			name: "default order",
			req:  GetCommentsRequest{App: 1, Record: 2, Limit: 2},
			responses: []response{
				{http.StatusOK, `{"comments": [{"id": "5"}, {"id": "4"}], "older": true, "newer": false}`},
				{http.StatusOK, `{"comments": [{"id": "3"}], "older": false, "newer": true}`},
			},
			wantBodies: []string{
				`{"app": 1, "record": 2, "limit": 2}`,
				`{"app": 1, "record": 2, "offset": 2, "limit": 2}`,
			},
			wantIDs: []int64{5, 4, 3},
		},
		{
			name: "ascending order",
			req:  GetCommentsRequest{App: 1, Record: 2, Order: "asc", Offset: 1},
			responses: []response{
				{http.StatusOK, `{"comments": [{"id": "2"}, {"id": "3"}], "older": true, "newer": true}`},
				{http.StatusOK, `{"comments": [{"id": "4"}], "older": true, "newer": false}`},
			},
			wantBodies: []string{
				`{"app": 1, "record": 2, "order": "asc", "offset": 1}`,
				`{"app": 1, "record": 2, "order": "asc", "offset": 3}`,
			},
			wantIDs: []int64{2, 3, 4},
		},
		{
			name: "ascending order in upper case",
			req:  GetCommentsRequest{App: 1, Record: 2, Order: "ASC"},
			responses: []response{
				{http.StatusOK, `{"comments": [{"id": "1"}], "older": false, "newer": true}`},
				{http.StatusOK, `{"comments": [{"id": "2"}], "older": true, "newer": false}`},
			},
			wantBodies: []string{
				`{"app": 1, "record": 2, "order": "ASC"}`,
				`{"app": 1, "record": 2, "order": "ASC", "offset": 1}`,
			},
			wantIDs: []int64{1, 2},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, requests := newSequenceServer(t, tc.responses...)
			c := newTestClient(t, ts)
			var ids []int64
			for comment, err := range c.AllComments(t.Context(), tc.req) {
				if err != nil {
					t.Fatalf("AllComments returned error: %v", err)
				}
				ids = append(ids, comment.ID)
			}
			reqs := requests()
			if len(reqs) != len(tc.wantBodies) {
				t.Fatalf("number of requests: got %d, want %d", len(reqs), len(tc.wantBodies))
			}
			for i, body := range tc.wantBodies {
				assertRequest(t, &reqs[i], http.MethodGet, "/k/v1/record/comments.json", body)
			}
			if !reflect.DeepEqual(ids, tc.wantIDs) {
				t.Errorf("comment IDs: got %v, want %v", ids, tc.wantIDs)
			}
		})
	}
}

func TestClientAllCommentsError(t *testing.T) {
	ts, _ := newSequenceServer(t,
		response{http.StatusInternalServerError, `{"code": "TEST_ERROR", "id": "abc123", "message": "Test error."}`},
	)
	c := newTestClient(t, ts)
	var err error
	for _, e := range c.AllComments(t.Context(), GetCommentsRequest{App: 1, Record: 2}) {
		err = e
	}
	if !IsErrorCode(err, "TEST_ERROR") {
		t.Errorf("AllComments returned unexpected error: %v", err)
	}
}

func TestClientAddComment(t *testing.T) {
	testCases := []struct {
		name     string
		response string
		mentions []Mention
		wantBody string
	}{
		{
			name:     "numeric ID",
			response: `{"id": 3}`,
			wantBody: `{"app": 1, "record": 2, "comment": {"text": "Hello"}}`,
		},
		{
			name:     "string ID",
			response: `{"id": "3"}`,
			mentions: []Mention{
				{Code: "alice", Type: EntityTypeUser},
				{Code: "dev", Type: EntityTypeOrganization},
				{Code: "everyone", Type: EntityTypeGroup},
			},
			wantBody: `{"app": 1, "record": 2, "comment": {"text": "Hello", "mentions": [
				{"code": "alice", "type": "USER"},
				{"code": "dev", "type": "ORGANIZATION"},
				{"code": "everyone", "type": "GROUP"}
			]}}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, tc.response)
			c := newTestClient(t, ts)
			id, err := c.AddComment(t.Context(), 1, 2, "Hello", tc.mentions...)
			if err != nil {
				t.Fatalf("AddComment returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPost, "/k/v1/record/comment.json", tc.wantBody)
			if id != 3 {
				t.Errorf("comment ID: got %d, want %d", id, 3)
			}
		})
	}
}

func TestClientAddCommentDecodeError(t *testing.T) {
	ts, _ := newTestServer(t, http.StatusOK, `{"id": "x"}`)
	c := newTestClient(t, ts)
	if _, err := c.AddComment(t.Context(), 1, 2, "Hello"); err == nil {
		t.Error("AddComment should return error")
	}
}

func TestClientDeleteComment(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{}`)
	c := newTestClient(t, ts)
	if err := c.DeleteComment(t.Context(), 1, 2, 3); err != nil {
		t.Fatalf("DeleteComment returned error: %v", err)
	}
	assertRequest(t, got, http.MethodDelete, "/k/v1/record/comment.json", `{"app": 1, "record": 2, "comment": 3}`)
}
