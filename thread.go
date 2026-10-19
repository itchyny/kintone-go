package kintone

import (
	"context"
)

// AddThread adds a thread to a space and returns the thread ID.
func (c *Client) AddThread(ctx context.Context, space int64, name string) (int64, error) {
	params := struct {
		Space int64  `json:"space"`
		Name  string `json:"name"`
	}{space, name}
	result, err := c.post[struct {
		ID number `json:"id"`
	}](ctx, "space/thread.json", params)
	if err != nil {
		return 0, err
	}
	return int64(result.ID), nil
}

// UpdateThreadRequest is the request of [Client.UpdateThread].
type UpdateThreadRequest struct {
	ID   int64   `json:"id"`
	Name string  `json:"name,omitzero"`
	Body *string `json:"body,omitzero"`
}

// UpdateThread updates the name or the body of a thread.
func (c *Client) UpdateThread(ctx context.Context, req UpdateThreadRequest) error {
	_, err := c.put[struct{}](ctx, "space/thread.json", req)
	return err
}

// ThreadComment is a comment on a thread.
type ThreadComment struct {
	Text     string              `json:"text,omitzero"`
	Mentions []Mention           `json:"mentions,omitzero"`
	Files    []ThreadCommentFile `json:"files,omitzero"`
}

// ThreadCommentFile is a file attached to a thread comment. Width is the
// width of the image in pixels.
type ThreadCommentFile struct {
	FileKey string `json:"fileKey"`
	Width   int    `json:"width,omitzero"`
}

// AddThreadComment adds a comment on a thread and returns the comment ID.
func (c *Client) AddThreadComment(ctx context.Context, space, thread int64, comment ThreadComment) (int64, error) {
	params := struct {
		Space   int64         `json:"space"`
		Thread  int64         `json:"thread"`
		Comment ThreadComment `json:"comment"`
	}{space, thread, comment}
	result, err := c.post[struct {
		ID number `json:"id"`
	}](ctx, "space/thread/comment.json", params)
	if err != nil {
		return 0, err
	}
	return int64(result.ID), nil
}
