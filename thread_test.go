package kintone

import (
	"net/http"
	"testing"
)

func TestClientAddThread(t *testing.T) {
	testCases := []struct {
		name     string
		response string
	}{
		{"string ID", `{"id": "8"}`},
		{"numeric ID", `{"id": 8}`},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, tc.response)
			c := newTestClient(t, ts, WithGuestSpace(4))
			id, err := c.AddThread(t.Context(), 4, "Weekly")
			if err != nil {
				t.Fatalf("AddThread returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPost, "/k/guest/4/v1/space/thread.json", `{"space": 4, "name": "Weekly"}`)
			if id != 8 {
				t.Errorf("thread ID: got %d, want %d", id, 8)
			}
		})
	}
}

func TestClientAddThreadDecodeError(t *testing.T) {
	ts, _ := newTestServer(t, http.StatusOK, `{"id": "x"}`)
	c := newTestClient(t, ts)
	if _, err := c.AddThread(t.Context(), 4, "Weekly"); err == nil {
		t.Error("AddThread should return error")
	}
}

func TestClientUpdateThread(t *testing.T) {
	testCases := []struct {
		name     string
		req      UpdateThreadRequest
		wantBody string
	}{
		{
			name:     "name and body",
			req:      UpdateThreadRequest{ID: 8, Name: "Weekly", Body: new("<p>Agenda</p>")},
			wantBody: `{"id": 8, "name": "Weekly", "body": "<p>Agenda</p>"}`,
		},
		{
			name:     "empty body",
			req:      UpdateThreadRequest{ID: 8, Body: new("")},
			wantBody: `{"id": 8, "body": ""}`,
		},
		{
			name:     "only name",
			req:      UpdateThreadRequest{ID: 8, Name: "Daily"},
			wantBody: `{"id": 8, "name": "Daily"}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{}`)
			c := newTestClient(t, ts)
			if err := c.UpdateThread(t.Context(), tc.req); err != nil {
				t.Fatalf("UpdateThread returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPut, "/k/v1/space/thread.json", tc.wantBody)
		})
	}
}

func TestClientAddThreadComment(t *testing.T) {
	testCases := []struct {
		name     string
		response string
		comment  ThreadComment
		wantBody string
	}{
		{
			name:     "text with numeric ID",
			response: `{"id": 3}`,
			comment: ThreadComment{
				Text:     "Hello",
				Mentions: []Mention{{Code: "alice", Type: EntityTypeUser}},
			},
			wantBody: `{"space": 2, "thread": 8, "comment": {
				"text": "Hello",
				"mentions": [{"code": "alice", "type": "USER"}]
			}}`,
		},
		{
			name:     "files with string ID",
			response: `{"id": "3"}`,
			comment: ThreadComment{
				Files: []ThreadCommentFile{{FileKey: "key1", Width: 250}, {FileKey: "key2"}},
			},
			wantBody: `{"space": 2, "thread": 8, "comment": {
				"files": [{"fileKey": "key1", "width": 250}, {"fileKey": "key2"}]
			}}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, tc.response)
			c := newTestClient(t, ts)
			id, err := c.AddThreadComment(t.Context(), 2, 8, tc.comment)
			if err != nil {
				t.Fatalf("AddThreadComment returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPost, "/k/v1/space/thread/comment.json", tc.wantBody)
			if id != 3 {
				t.Errorf("comment ID: got %d, want %d", id, 3)
			}
		})
	}
}

func TestClientAddThreadCommentDecodeError(t *testing.T) {
	ts, _ := newTestServer(t, http.StatusOK, `{"id": "x"}`)
	c := newTestClient(t, ts)
	if _, err := c.AddThreadComment(t.Context(), 2, 8, ThreadComment{Text: "Hello"}); err == nil {
		t.Error("AddThreadComment should return error")
	}
}
