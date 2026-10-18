package kintone

import (
	"context"
	"encoding/json"
	"fmt"
)

// AddThread adds a thread to a space and returns the thread ID.
func (c *Client) AddThread(ctx context.Context, space int64, name string) (int64, error) {
	params := struct {
		Space int64  `json:"space"`
		Name  string `json:"name"`
	}{space, name}
	result, err := c.post[struct {
		ID json.Number `json:"id"`
	}](ctx, "space/thread.json", params)
	if err != nil {
		return 0, err
	}
	id, err := result.ID.Int64()
	if err != nil {
		return 0, fmt.Errorf("kintone: decode response: %w", err)
	}
	return id, nil
}

// UpdateThreadRequest is the request of [Client.UpdateThread].
// The name or the body is not changed if not set.
type UpdateThreadRequest struct {
	ID   int64   `json:"id"`
	Name string  `json:"name,omitempty"`
	Body *string `json:"body,omitempty"`
}

// UpdateThread updates the name or the body of a thread.
func (c *Client) UpdateThread(ctx context.Context, req UpdateThreadRequest) error {
	_, err := c.put[struct{}](ctx, "space/thread.json", req)
	return err
}

// ThreadComment is a comment on a thread.
type ThreadComment struct {
	Text     string              `json:"text,omitempty"`
	Mentions []Mention           `json:"mentions,omitempty"`
	Files    []ThreadCommentFile `json:"files,omitempty"`
}

// ThreadCommentFile is a file attached to a thread comment. Width is the
// width of the image in pixels.
type ThreadCommentFile struct {
	FileKey string `json:"fileKey"`
	Width   int    `json:"width,omitempty"`
}

// AddThreadComment adds a comment on a thread and returns the comment ID.
func (c *Client) AddThreadComment(ctx context.Context, space, thread int64, comment ThreadComment) (int64, error) {
	params := struct {
		Space   int64         `json:"space"`
		Thread  int64         `json:"thread"`
		Comment ThreadComment `json:"comment"`
	}{space, thread, comment}
	result, err := c.post[struct {
		ID json.Number `json:"id"`
	}](ctx, "space/thread/comment.json", params)
	if err != nil {
		return 0, err
	}
	id, err := result.ID.Int64()
	if err != nil {
		return 0, fmt.Errorf("kintone: decode response: %w", err)
	}
	return id, nil
}
