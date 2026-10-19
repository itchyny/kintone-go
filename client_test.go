package kintone

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type request struct {
	Method string
	Path   string
	Header http.Header
	Body   string
}

func newTestServer(t *testing.T, status int, body string) (*httptest.Server, *request) {
	t.Helper()
	var got request
	ts := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got = request{r.Method, r.URL.Path, r.Header, string(b)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	return ts, &got
}

// newTestClient returns a client sending requests to the test server.
func newTestClient(t *testing.T, ts *httptest.Server, opts ...Option) *Client {
	t.Helper()
	return newClient(t, ts.URL, append([]Option{WithHTTPClient(ts.Client())}, opts...)...)
}

func newClient(t *testing.T, baseURL string, opts ...Option) *Client {
	t.Helper()
	c, err := NewClient(baseURL, opts...)
	if err != nil {
		t.Fatalf("NewClient(%q) returned error: %v", baseURL, err)
	}
	return c
}

func TestNewClient(t *testing.T) {
	testCases := []struct {
		name    string
		baseURL string
		want    string
		wantErr bool
	}{
		{
			name:    "https",
			baseURL: "https://example.cybozu.com",
			want:    "https://example.cybozu.com/k/v1/record.json",
		},
		{
			name:    "trailing slash",
			baseURL: "https://example.cybozu.com/",
			want:    "https://example.cybozu.com/k/v1/record.json",
		},
		{
			name:    "path, query and fragment",
			baseURL: "http://localhost:8080/proxy/?x=1#y",
			want:    "http://localhost:8080/proxy/k/v1/record.json",
		},
		{
			name:    "no scheme",
			baseURL: "example.cybozu.com",
			wantErr: true,
		},
		{
			name:    "unsupported scheme",
			baseURL: "ftp://example.cybozu.com",
			wantErr: true,
		},
		{
			name:    "no host",
			baseURL: "https://",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			baseURL: "https://example.cybozu.com/%zz",
			wantErr: true,
		},
		{
			name:    "user information",
			baseURL: "https://user:secret@example.cybozu.com",
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewClient(tc.baseURL)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("NewClient(%q) should return error", tc.baseURL)
				}
				if !strings.HasPrefix(err.Error(), "kintone: invalid base URL: ") {
					t.Errorf("NewClient(%q) returned unexpected error: %v", tc.baseURL, err)
				}
				if strings.Contains(err.Error(), "secret") {
					t.Errorf("NewClient(%q) should not show the password: %v", tc.baseURL, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewClient(%q) returned error: %v", tc.baseURL, err)
			}
			if got := c.endpoint("record.json").String(); got != tc.want {
				t.Errorf("endpoint: got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestClientEndpointGuestSpace(t *testing.T) {
	c := newClient(t, "https://example.cybozu.com", WithGuestSpace(3))
	want := "https://example.cybozu.com/k/guest/3/v1/records/cursor.json"
	if got := c.endpoint("records/cursor.json").String(); got != want {
		t.Errorf("endpoint: got %q, want %q", got, want)
	}
}

func TestClientAuthentication(t *testing.T) {
	testCases := []struct {
		name string
		opts []Option
		want http.Header
	}{
		{
			name: "API token",
			opts: []Option{WithAPIToken("token1")},
			want: http.Header{"X-Cybozu-Api-Token": {"token1"}},
		},
		{
			name: "multiple API tokens",
			opts: []Option{WithAPIToken("token1", "token2")},
			want: http.Header{"X-Cybozu-Api-Token": {"token1,token2"}},
		},
		{
			name: "password",
			opts: []Option{WithPassword("alice", "secret")},
			want: http.Header{"X-Cybozu-Authorization": {"YWxpY2U6c2VjcmV0"}},
		},
		{
			name: "OAuth token",
			opts: []Option{WithOAuthToken("token1")},
			want: http.Header{"Authorization": {"Bearer token1"}},
		},
		{
			name: "password with Basic authentication",
			opts: []Option{WithPassword("alice", "secret"), WithBasicAuth("bob", "pass")},
			want: http.Header{
				"X-Cybozu-Authorization": {"YWxpY2U6c2VjcmV0"},
				"Authorization":          {"Basic Ym9iOnBhc3M="},
			},
		},
		{
			name: "OAuth token after Basic authentication",
			opts: []Option{WithBasicAuth("bob", "pass"), WithOAuthToken("token1")},
			want: http.Header{"Authorization": {"Bearer token1"}},
		},
	}
	keys := []string{"X-Cybozu-Api-Token", "X-Cybozu-Authorization", "Authorization"}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{}`)
			c := newTestClient(t, ts, tc.opts...)
			if _, err := c.do[struct{}](t.Context(), http.MethodGet, "app.json", nil); err != nil {
				t.Fatalf("do returned error: %v", err)
			}
			for _, key := range keys {
				if got, want := got.Header.Values(key), tc.want.Values(key); !reflect.DeepEqual(got, want) {
					t.Errorf("header %s: got %q, want %q", key, got, want)
				}
			}
		})
	}
}

func TestClientUserAgent(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		ts, got := newTestServer(t, http.StatusOK, `{}`)
		c := newTestClient(t, ts)
		if _, err := c.do[struct{}](t.Context(), http.MethodGet, "app.json", nil); err != nil {
			t.Fatalf("do returned error: %v", err)
		}
		if got, want := got.Header.Get("User-Agent"), userAgent(); got != want {
			t.Errorf("User-Agent: got %q, want %q", got, want)
		}
		if ua := userAgent(); ua != "kintone-go" && !strings.HasPrefix(ua, "kintone-go/") {
			t.Errorf("userAgent() returned unexpected value: %q", ua)
		}
	})
	t.Run("custom", func(t *testing.T) {
		ts, got := newTestServer(t, http.StatusOK, `{}`)
		c := newTestClient(t, ts, WithUserAgent("my-app/1.0"))
		if _, err := c.do[struct{}](t.Context(), http.MethodGet, "app.json", nil); err != nil {
			t.Fatalf("do returned error: %v", err)
		}
		if got, want := got.Header.Get("User-Agent"), "my-app/1.0"; got != want {
			t.Errorf("User-Agent: got %q, want %q", got, want)
		}
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestClientHTTPClient(t *testing.T) {
	var called bool
	httpClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			called = true
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"name":"test"}`)),
				Request:    req,
			}, nil
		}),
	}
	c := newClient(t, "https://example.cybozu.com", WithHTTPClient(httpClient))
	got, err := c.do[struct {
		Name string `json:"name"`
	}](t.Context(), http.MethodGet, "app.json", nil)
	if err != nil {
		t.Fatalf("do returned error: %v", err)
	}
	if !called {
		t.Error("the HTTP client was not used")
	}
	if got.Name != "test" {
		t.Errorf("name: got %q, want %q", got.Name, "test")
	}
}

func TestClientDo(t *testing.T) {
	type params struct {
		App int64 `json:"app"`
	}
	testCases := []struct {
		name       string
		method     string
		params     any
		wantMethod string
		wantHeader http.Header
		wantBody   string
	}{
		{
			name:       "GET with parameters",
			method:     http.MethodGet,
			params:     params{App: 1},
			wantMethod: http.MethodPost,
			wantHeader: http.Header{
				"Content-Type":           {"application/json"},
				"X-Http-Method-Override": {http.MethodGet},
			},
			wantBody: `{"app":1}`,
		},
		{
			name:       "GET without parameters",
			method:     http.MethodGet,
			wantMethod: http.MethodGet,
			wantHeader: http.Header{},
		},
		{
			name:       "POST",
			method:     http.MethodPost,
			params:     params{App: 1},
			wantMethod: http.MethodPost,
			wantHeader: http.Header{"Content-Type": {"application/json"}},
			wantBody:   `{"app":1}`,
		},
		{
			name:       "PUT",
			method:     http.MethodPut,
			params:     params{App: 1},
			wantMethod: http.MethodPut,
			wantHeader: http.Header{"Content-Type": {"application/json"}},
			wantBody:   `{"app":1}`,
		},
		{
			name:       "DELETE",
			method:     http.MethodDelete,
			params:     params{App: 1},
			wantMethod: http.MethodDelete,
			wantHeader: http.Header{"Content-Type": {"application/json"}},
			wantBody:   `{"app":1}`,
		},
	}
	keys := []string{"Content-Type", "X-Http-Method-Override"}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{"id":"1","revision":"2"}`)
			c := newTestClient(t, ts)
			res, err := c.do[revision](t.Context(), tc.method, "record.json", tc.params)
			if err != nil {
				t.Fatalf("do returned error: %v", err)
			}
			if want := (revision{ID: 1, Revision: 2}); res != want {
				t.Errorf("result: got %+v, want %+v", res, want)
			}
			if got.Method != tc.wantMethod {
				t.Errorf("method: got %q, want %q", got.Method, tc.wantMethod)
			}
			if want := "/k/v1/record.json"; got.Path != want {
				t.Errorf("path: got %q, want %q", got.Path, want)
			}
			for _, key := range keys {
				if got, want := got.Header.Values(key), tc.wantHeader.Values(key); !reflect.DeepEqual(got, want) {
					t.Errorf("header %s: got %q, want %q", key, got, want)
				}
			}
			if got.Body != tc.wantBody {
				t.Errorf("body: got %q, want %q", got.Body, tc.wantBody)
			}
		})
	}
}

type revision struct {
	ID       int64 `json:"id,string"`
	Revision int64 `json:"revision,string"`
}

func TestClientDoIgnoreResponse(t *testing.T) {
	ts, _ := newTestServer(t, http.StatusOK, `not JSON`)
	c := newTestClient(t, ts)
	if _, err := c.do[struct{}](t.Context(), http.MethodDelete, "records.json", nil); err != nil {
		t.Errorf("do returned error: %v", err)
	}
}

func TestClientDoDecodeError(t *testing.T) {
	ts, _ := newTestServer(t, http.StatusOK, `not JSON`)
	c := newTestClient(t, ts)
	_, err := c.do[map[string]any](t.Context(), http.MethodGet, "app.json", nil)
	if err == nil || !strings.HasPrefix(err.Error(), "kintone: decode response: ") {
		t.Errorf("do should return decode error but got: %v", err)
	}
}

func TestClientDoMarshalError(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{}`)
	c := newTestClient(t, ts)
	_, err := c.do[struct{}](t.Context(), http.MethodPost, "record.json", map[string]any{"x": func() {}})
	if _, ok := errors.AsType[*json.SemanticError](err); !ok {
		t.Errorf("do should return *json.SemanticError but got: %v", err)
	}
	if got.Method != "" {
		t.Errorf("request should not be sent but got %s", got.Method)
	}
}

func TestClientDoContextCanceled(t *testing.T) {
	ts, _ := newTestServer(t, http.StatusOK, `{}`)
	c := newTestClient(t, ts)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := c.do[struct{}](ctx, http.MethodGet, "app.json", nil)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("do should return context.Canceled but got: %v", err)
	}
}

func TestClientDoError(t *testing.T) {
	testCases := []struct {
		name    string
		status  int
		body    string
		want    *Error
		wantMsg string
	}{
		{
			name:   "kintone error",
			status: http.StatusNotFound,
			body:   `{"code":"GAIA_RE01","id":"abc123","message":"The specified record is not found."}`,
			want: &Error{
				StatusCode: http.StatusNotFound,
				Code:       "GAIA_RE01",
				ID:         "abc123",
				Message:    "The specified record is not found.",
			},
			wantMsg: "kintone: The specified record is not found. (status: 404, code: GAIA_RE01, id: abc123)",
		},
		{
			name:   "validation error",
			status: http.StatusBadRequest,
			body: `{"code":"TEST_VALIDATION","id":"abc123","message":"Test validation error.",` +
				`"errors":{"record.title.value":{"messages":["Required field."]}}}`,
			want: &Error{
				StatusCode: http.StatusBadRequest,
				Code:       "TEST_VALIDATION",
				ID:         "abc123",
				Message:    "Test validation error.",
				Errors: map[string]ErrorDetail{
					"record.title.value": {Messages: []string{"Required field."}},
				},
			},
			wantMsg: "kintone: Test validation error. (status: 400, code: TEST_VALIDATION, id: abc123)",
		},
		{
			name:    "non-JSON body",
			status:  http.StatusBadGateway,
			body:    "<html>Bad Gateway</html>\n",
			want:    &Error{StatusCode: http.StatusBadGateway, Message: "<html>Bad Gateway</html>"},
			wantMsg: "kintone: <html>Bad Gateway</html> (status: 502)",
		},
		{
			name:    "long non-JSON body",
			status:  http.StatusBadGateway,
			body:    strings.Repeat("あ", 250),
			want:    &Error{StatusCode: http.StatusBadGateway, Message: strings.Repeat("あ", 200) + "..."},
			wantMsg: "kintone: " + strings.Repeat("あ", 200) + "... (status: 502)",
		},
		{
			name:    "empty body",
			status:  http.StatusServiceUnavailable,
			want:    &Error{StatusCode: http.StatusServiceUnavailable},
			wantMsg: "kintone: Service Unavailable (status: 503)",
		},
		{
			name:    "unknown status",
			status:  599,
			want:    &Error{StatusCode: 599},
			wantMsg: "kintone: unknown error (status: 599)",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, _ := newTestServer(t, tc.status, tc.body)
			c := newTestClient(t, ts)
			_, err := c.do[map[string]any](t.Context(), http.MethodGet, "record.json", nil)
			e, ok := errors.AsType[*Error](err)
			if !ok {
				t.Fatalf("do should return *Error but got: %v", err)
			}
			got := *e
			got.Header = nil
			if !reflect.DeepEqual(&got, tc.want) {
				t.Errorf("error: got %+v, want %+v", &got, tc.want)
			}
			if got := err.Error(); got != tc.wantMsg {
				t.Errorf("error message: got %q, want %q", got, tc.wantMsg)
			}
		})
	}
}

func TestClientDoErrorHeader(t *testing.T) {
	testCases := []struct {
		name string
		body string
	}{
		{"JSON body", `{"code": "TEST_ERROR", "id": "abc123", "message": "Test error."}`},
		{"non-JSON body", "Service Unavailable"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-ConcurrencyLimit-Limit", "100")
				w.Header().Set("X-ConcurrencyLimit-Running", "100")
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = io.WriteString(w, tc.body)
			}))
			c := newTestClient(t, ts)
			_, err := c.do[struct{}](t.Context(), http.MethodGet, "app.json", nil)
			e, ok := errors.AsType[*Error](err)
			if !ok {
				t.Fatalf("do should return *Error but got: %v", err)
			}
			if got := e.Header.Get("X-ConcurrencyLimit-Limit"); got != "100" {
				t.Errorf("X-ConcurrencyLimit-Limit: got %q, want %q", got, "100")
			}
			if got := e.Header.Get("X-ConcurrencyLimit-Running"); got != "100" {
				t.Errorf("X-ConcurrencyLimit-Running: got %q, want %q", got, "100")
			}
		})
	}
}

func TestClientDoBulkError(t *testing.T) {
	ts := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-ConcurrencyLimit-Limit", "100")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"results": [{}, {"code": "TEST_ERROR", "id": "abc123", "message": "Test error."}, {}]}`)
	}))
	c := newTestClient(t, ts)
	_, err := c.do[struct{}](t.Context(), http.MethodPost, "bulkRequest.json", map[string]any{})
	bulkErr, ok := errors.AsType[*BulkError](err)
	if !ok {
		t.Fatalf("do should return *BulkError but got: %v", err)
	}
	if bulkErr.Index != 1 {
		t.Errorf("index: got %d, want %d", bulkErr.Index, 1)
	}
	got := *bulkErr.Err
	got.Header = nil
	if want := (Error{StatusCode: http.StatusNotFound, Code: "TEST_ERROR", ID: "abc123", Message: "Test error."}); !reflect.DeepEqual(got, want) {
		t.Errorf("error: got %+v, want %+v", got, want)
	}
	if bulkErr.Err.Header.Get("X-ConcurrencyLimit-Limit") != "100" {
		t.Errorf("header should be set: %v", bulkErr.Err.Header)
	}
	if want := "kintone: bulk request 1: Test error. (status: 404, code: TEST_ERROR, id: abc123)"; err.Error() != want {
		t.Errorf("error message: got %q, want %q", err.Error(), want)
	}
	if e, ok := errors.AsType[*Error](err); !ok || e != bulkErr.Err {
		t.Errorf("errors.AsType should find *Error in *BulkError but got: %v", e)
	}
	if !IsErrorCode(err, "TEST_ERROR") {
		t.Error("IsErrorCode should find the code in *BulkError")
	}
}

func TestClientDoBulkErrorWithoutError(t *testing.T) {
	ts, _ := newTestServer(t, http.StatusBadRequest, `{"results": [{}, {}]}`)
	c := newTestClient(t, ts)
	_, err := c.do[struct{}](t.Context(), http.MethodPost, "bulkRequest.json", map[string]any{})
	if _, ok := errors.AsType[*BulkError](err); ok {
		t.Errorf("do should not return *BulkError without an error result but got: %v", err)
	}
	if e, ok := errors.AsType[*Error](err); !ok || e.StatusCode != http.StatusBadRequest {
		t.Errorf("do should return *Error but got: %v", err)
	}
}

func TestIsErrorCode(t *testing.T) {
	err := &Error{StatusCode: http.StatusNotFound, Code: "GAIA_RE01"}
	testCases := []struct {
		name string
		err  error
		code string
		want bool
	}{
		{"same code", err, "GAIA_RE01", true},
		{"different code", err, "TEST_ERROR", false},
		{"wrapped error", fmt.Errorf("get record: %w", err), "GAIA_RE01", true},
		{"other error", errors.New("GAIA_RE01"), "GAIA_RE01", false},
		{"nil error", nil, "", false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsErrorCode(tc.err, tc.code); got != tc.want {
				t.Errorf("IsErrorCode(%v, %q): got %v, want %v", tc.err, tc.code, got, tc.want)
			}
		})
	}
}

func TestClientMethods(t *testing.T) {
	params := map[string]any{"app": 1}
	testCases := []struct {
		name         string
		call         func(*Client) (map[string]any, error)
		wantMethod   string
		wantOverride string
	}{
		{
			name: "get",
			call: func(c *Client) (map[string]any, error) {
				return c.get[map[string]any](t.Context(), "app.json", params)
			},
			wantMethod:   http.MethodPost,
			wantOverride: http.MethodGet,
		},
		{
			name: "post",
			call: func(c *Client) (map[string]any, error) {
				return c.post[map[string]any](t.Context(), "app.json", params)
			},
			wantMethod: http.MethodPost,
		},
		{
			name: "put",
			call: func(c *Client) (map[string]any, error) {
				return c.put[map[string]any](t.Context(), "app.json", params)
			},
			wantMethod: http.MethodPut,
		},
		{
			name: "delete",
			call: func(c *Client) (map[string]any, error) {
				return c.delete[map[string]any](t.Context(), "app.json", params)
			},
			wantMethod: http.MethodDelete,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{"name":"test"}`)
			c := newTestClient(t, ts)
			res, err := tc.call(c)
			if err != nil {
				t.Fatalf("%s returned error: %v", tc.name, err)
			}
			if got.Method != tc.wantMethod {
				t.Errorf("method: got %q, want %q", got.Method, tc.wantMethod)
			}
			if override := got.Header.Get("X-HTTP-Method-Override"); override != tc.wantOverride {
				t.Errorf("X-HTTP-Method-Override: got %q, want %q", override, tc.wantOverride)
			}
			if want := "/k/v1/app.json"; got.Path != want {
				t.Errorf("path: got %q, want %q", got.Path, want)
			}
			if want := `{"app":1}`; got.Body != want {
				t.Errorf("body: got %q, want %q", got.Body, want)
			}
			if want := map[string]any{"name": "test"}; !reflect.DeepEqual(res, want) {
				t.Errorf("result: got %v, want %v", res, want)
			}
		})
	}
}

func TestClientRedirect(t *testing.T) {
	var redirected bool
	ts := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/elsewhere" {
			redirected = true
			return
		}
		http.Redirect(w, r, "http://other.example.com/elsewhere", http.StatusFound)
	}))
	for _, opts := range [][]Option{
		{WithAPIToken("secret")},
		{WithAPIToken("secret"), WithHTTPClient(&http.Client{Transport: ts.Client().Transport})},
	} {
		redirected = false
		c := newTestClient(t, ts, opts...)
		_, err := c.do[struct{}](t.Context(), http.MethodGet, "app.json", nil)
		if e, ok := errors.AsType[*Error](err); !ok || e.StatusCode != http.StatusFound {
			t.Errorf("do should return error of status 302 but got: %v", err)
		}
		if redirected {
			t.Error("the redirect should not be followed")
		}
	}
}

func TestClientHTTPClientNotModified(t *testing.T) {
	httpClient := &http.Client{}
	newClient(t, "https://example.cybozu.com", WithHTTPClient(httpClient))
	if httpClient.CheckRedirect != nil {
		t.Error("the HTTP client should not be modified")
	}
	if http.DefaultClient.CheckRedirect != nil {
		t.Error("the default HTTP client should not be modified")
	}
}

func TestClientNilHTTPClient(t *testing.T) {
	// http.DefaultClient reaches only a server listening on loopback
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	}))
	t.Cleanup(ts.Close)
	c := newClient(t, ts.URL, WithHTTPClient(nil))
	if _, err := c.do[struct{}](t.Context(), http.MethodGet, "app.json", nil); err != nil {
		t.Errorf("do returned error: %v", err)
	}
}

func TestClientHeaderNotShared(t *testing.T) {
	var tokens []string
	httpClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			tokens = append(tokens, req.Header.Get("X-Cybozu-API-Token"))
			req.Header["X-Cybozu-Api-Token"][0] = "modified"
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{}`)),
				Request:    req,
			}, nil
		}),
	}
	c := newClient(t, "https://example.cybozu.com", WithAPIToken("token1"), WithHTTPClient(httpClient))
	for range 2 {
		if _, err := c.do[struct{}](t.Context(), http.MethodGet, "app.json", nil); err != nil {
			t.Fatalf("do returned error: %v", err)
		}
	}
	if want := []string{"token1", "token1"}; !reflect.DeepEqual(tokens, want) {
		t.Errorf("tokens: got %q, want %q", tokens, want)
	}
}

func setValue[T any](t *testing.T, p *T, v T) {
	t.Helper()
	old := *p
	*p = v
	t.Cleanup(func() { *p = old })
}
