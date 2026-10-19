package kintone

import (
	"context"
	"encoding/json/jsontext"
	"fmt"
	"slices"
	"strings"
)

// APIs is the list of the kintone REST APIs, keyed by the API ID.
type APIs struct {
	BaseURL string             `json:"baseUrl"`
	APIs    map[string]APILink `json:"apis"`
}

// APILink is the link to get the schema of an API.
type APILink struct {
	Link string `json:"link"`
}

// GetAPIs gets the list of the kintone REST APIs. The guest space of the
// client is not used.
func (c *Client) GetAPIs(ctx context.Context) (*APIs, error) {
	result, err := c.withoutGuestSpace().get[APIs](ctx, "apis.json", nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// APISchema is the schema of an API. Request, Response and Schemas are JSON
// Schemas.
type APISchema struct {
	ID         string                    `json:"id"`
	BaseURL    string                    `json:"baseUrl"`
	Path       string                    `json:"path"`
	HTTPMethod string                    `json:"httpMethod"`
	Request    jsontext.Value            `json:"request"`
	Response   jsontext.Value            `json:"response"`
	Schemas    map[string]jsontext.Value `json:"schemas"`
}

// GetAPISchema gets the schema of the API of the ID, for example
// "records/get". The guest space of the client is not used.
func (c *Client) GetAPISchema(ctx context.Context, id string) (*APISchema, error) {
	if id == "" || slices.Contains(strings.Split(id, "/"), "..") {
		return nil, fmt.Errorf("kintone: invalid API ID: %q", id)
	}
	result, err := c.withoutGuestSpace().get[APISchema](ctx, "apis/"+id+".json", nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
