package kintone

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Comment is a comment on a record.
type Comment struct {
	ID        int64     `json:"id,string"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
	Creator   Entity    `json:"creator"`
	Mentions  []Mention `json:"mentions"`
}

// Mention is a mention in a comment.
type Mention struct {
	Code string     `json:"code"`
	Type EntityType `json:"type"`
}

// GetCommentsRequest is the request of [Client.GetComments].
// Order is "asc" or "desc".
type GetCommentsRequest struct {
	App    int64  `json:"app"`
	Record int64  `json:"record"`
	Order  string `json:"order,omitempty"`
	Offset int64  `json:"offset,omitempty"`
	Limit  int64  `json:"limit,omitempty"`
}

// GetCommentsResponse is the response of [Client.GetComments].
// Older and Newer report whether there are older or newer comments.
type GetCommentsResponse struct {
	Comments []Comment `json:"comments"`
	Older    bool      `json:"older"`
	Newer    bool      `json:"newer"`
}

// GetComments gets comments on a record.
func (c *Client) GetComments(ctx context.Context, req GetCommentsRequest) (*GetCommentsResponse, error) {
	result, err := c.get[GetCommentsResponse](ctx, "record/comments.json", req)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// AddComment adds a comment on a record and returns the comment ID.
func (c *Client) AddComment(ctx context.Context, app, record int64, text string, mentions ...Mention) (int64, error) {
	type comment struct {
		Text     string    `json:"text"`
		Mentions []Mention `json:"mentions,omitempty"`
	}
	params := struct {
		App     int64   `json:"app"`
		Record  int64   `json:"record"`
		Comment comment `json:"comment"`
	}{app, record, comment{text, mentions}}
	result, err := c.post[struct {
		ID json.Number `json:"id"`
	}](ctx, "record/comment.json", params)
	if err != nil {
		return 0, err
	}
	id, err := result.ID.Int64()
	if err != nil {
		return 0, fmt.Errorf("kintone: decode response: %w", err)
	}
	return id, nil
}

// DeleteComment deletes a comment on a record.
func (c *Client) DeleteComment(ctx context.Context, app, record, comment int64) error {
	params := struct {
		App     int64 `json:"app"`
		Record  int64 `json:"record"`
		Comment int64 `json:"comment"`
	}{app, record, comment}
	_, err := c.delete[struct{}](ctx, "record/comment.json", params)
	return err
}
