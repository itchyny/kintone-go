package kintone

import (
	"encoding/json/jsontext"
	"net/http"
	"reflect"
	"testing"
)

func TestClientGetAPIs(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{
		"baseUrl": "https://example.cybozu.com/k/v1/",
		"apis": {
			"records/get": {"link": "apis/records/get.json"},
			"app/form/fields/get": {"link": "apis/app/form/fields/get.json"}
		}
	}`)
	c := newTestClient(t, ts, WithGuestSpace(4))
	apis, err := c.GetAPIs(t.Context())
	if err != nil {
		t.Fatalf("GetAPIs returned error: %v", err)
	}
	if got.Method != http.MethodGet || got.Path != "/k/v1/apis.json" || got.Body != "" {
		t.Errorf("request: got %s %s %q", got.Method, got.Path, got.Body)
	}
	want := &APIs{
		BaseURL: "https://example.cybozu.com/k/v1/",
		APIs: map[string]APILink{
			"records/get":         {Link: "apis/records/get.json"},
			"app/form/fields/get": {Link: "apis/app/form/fields/get.json"},
		},
	}
	if !reflect.DeepEqual(apis, want) {
		t.Errorf("apis: got %+v, want %+v", apis, want)
	}
}

func TestClientGetAPISchema(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{
		"id": "records/get",
		"baseUrl": "https://example.cybozu.com/k/v1/",
		"path": "records.json",
		"httpMethod": "GET",
		"request": {"type": "object", "required": ["app"]},
		"response": {"type": "object"},
		"schemas": {"SingleLineTextSimpleValue": {"type": "object"}}
	}`)
	c := newTestClient(t, ts, WithGuestSpace(4))
	schema, err := c.GetAPISchema(t.Context(), "records/get")
	if err != nil {
		t.Fatalf("GetAPISchema returned error: %v", err)
	}
	if got.Method != http.MethodGet || got.Path != "/k/v1/apis/records/get.json" || got.Body != "" {
		t.Errorf("request: got %s %s %q", got.Method, got.Path, got.Body)
	}
	want := &APISchema{
		ID:         "records/get",
		BaseURL:    "https://example.cybozu.com/k/v1/",
		Path:       "records.json",
		HTTPMethod: "GET",
		Request:    jsontext.Value(`{"type": "object", "required": ["app"]}`),
		Response:   jsontext.Value(`{"type": "object"}`),
		Schemas:    map[string]jsontext.Value{"SingleLineTextSimpleValue": jsontext.Value(`{"type": "object"}`)},
	}
	if !reflect.DeepEqual(schema, want) {
		t.Errorf("schema: got %+v, want %+v", schema, want)
	}
}

func TestClientGetAPISchemaInvalidID(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{}`)
	c := newTestClient(t, ts)
	for _, id := range []string{"", "..", "../record", "records/../../record", "records/.."} {
		if _, err := c.GetAPISchema(t.Context(), id); err == nil {
			t.Errorf("GetAPISchema(%q) should return error", id)
		}
	}
	if got.Method != "" {
		t.Errorf("request should not be sent but got %s %s", got.Method, got.Path)
	}
}
