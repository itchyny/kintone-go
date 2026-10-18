// Package kintone provides a client for the kintone REST API.
package kintone

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"sync"
)

// Client is a kintone REST API client.
type Client struct {
	baseURL      *url.URL
	httpClient   *http.Client
	header       http.Header
	guestSpaceID int64
}

// Option configures a [Client].
type Option func(*Client)

// NewClient creates a client for the kintone environment at baseURL.
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("kintone: invalid base URL: %w", err)
	}
	if u.Scheme != "https" && u.Scheme != "http" || u.Host == "" {
		return nil, fmt.Errorf("kintone: invalid base URL: %q", baseURL)
	}
	if u.User != nil {
		return nil, errors.New("kintone: invalid base URL: user information is not allowed")
	}
	u.Path = strings.TrimSuffix(u.Path, "/")
	u.RawPath, u.RawQuery, u.Fragment = "", "", ""
	c := &Client{
		baseURL:    u,
		httpClient: http.DefaultClient,
		header:     http.Header{"User-Agent": {userAgent()}},
	}
	for _, opt := range opts {
		opt(c)
	}
	httpClient := *c.httpClient
	httpClient.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	c.httpClient = &httpClient
	return c, nil
}

const modulePath = "github.com/itchyny/kintone-go"

// userAgent returns the default User-Agent with the module version.
var userAgent = sync.OnceValue(func() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "kintone-go"
	}
	mods := append([]*debug.Module{&bi.Main}, bi.Deps...)
	for _, m := range mods {
		if m.Path != modulePath {
			continue
		}
		if m.Replace != nil {
			m = m.Replace
		}
		if m.Version != "" && m.Version != "(devel)" {
			return "kintone-go/" + m.Version
		}
	}
	return "kintone-go"
})

// WithAPIToken authenticates with API tokens. Multiple tokens can be
// specified to access multiple apps (for example on lookup fields).
func WithAPIToken(tokens ...string) Option {
	return func(c *Client) {
		c.header.Set("X-Cybozu-API-Token", strings.Join(tokens, ","))
	}
}

// WithPassword authenticates with the login name and password.
func WithPassword(username, password string) Option {
	return func(c *Client) {
		c.header.Set("X-Cybozu-Authorization", basicCredential(username, password))
	}
}

// WithOAuthToken authenticates with the OAuth access token.
func WithOAuthToken(token string) Option {
	return func(c *Client) {
		c.header.Set("Authorization", "Bearer "+token)
	}
}

// WithBasicAuth sets the credential of the Basic authentication,
// which is required when the domain enables it.
func WithBasicAuth(username, password string) Option {
	return func(c *Client) {
		c.header.Set("Authorization", "Basic "+basicCredential(username, password))
	}
}

func basicCredential(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

// WithHTTPClient sets the HTTP client used for requests. The client does not
// follow redirects, so that the credentials are not sent to other hosts.
// If httpClient is nil, [http.DefaultClient] is used.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient == nil {
			httpClient = http.DefaultClient
		}
		c.httpClient = httpClient
	}
}

// WithGuestSpace makes the client access apps in the guest space.
func WithGuestSpace(id int64) Option {
	return func(c *Client) {
		c.guestSpaceID = id
	}
}

// WithUserAgent sets the User-Agent header.
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.header.Set("User-Agent", userAgent)
	}
}

func (c *Client) withoutGuestSpace() *Client {
	return c.withGuestSpace(0)
}

func (c *Client) withGuestSpace(id int64) *Client {
	d := *c
	d.guestSpaceID = id
	return &d
}

func (c *Client) endpoint(path string) *url.URL {
	u := *c.baseURL
	if c.guestSpaceID != 0 {
		u.Path += "/k/guest/" + strconv.FormatInt(c.guestSpaceID, 10) + "/v1/" + path
	} else {
		u.Path += "/k/v1/" + path
	}
	return &u
}

func (c *Client) get[T any](ctx context.Context, path string, params any) (T, error) {
	return c.do[T](ctx, http.MethodGet, path, params)
}

func (c *Client) post[T any](ctx context.Context, path string, params any) (T, error) {
	return c.do[T](ctx, http.MethodPost, path, params)
}

func (c *Client) put[T any](ctx context.Context, path string, params any) (T, error) {
	return c.do[T](ctx, http.MethodPut, path, params)
}

func (c *Client) delete[T any](ctx context.Context, path string, params any) (T, error) {
	return c.do[T](ctx, http.MethodDelete, path, params)
}

// do a JSON request and decodes the JSON response as T.
func (c *Client) do[T any](ctx context.Context, method, path string, params any) (T, error) {
	result, _, err := c.doHeader[T](ctx, method, path, params)
	return result, err
}

// doHeader is the same as do, but also returns the response header.
func (c *Client) doHeader[T any](ctx context.Context, method, path string, params any) (T, http.Header, error) {
	var result T
	var body io.Reader
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			return result, nil, fmt.Errorf("kintone: %w", err)
		}
		body = bytes.NewReader(b)
	}
	httpMethod := method
	if method == http.MethodGet && params != nil {
		httpMethod = http.MethodPost
	}
	req, err := http.NewRequestWithContext(ctx, httpMethod, c.endpoint(path).String(), body)
	if err != nil {
		return result, nil, fmt.Errorf("kintone: %w", err)
	}
	if params != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if httpMethod != method {
		req.Header.Set("X-HTTP-Method-Override", method)
	}
	res, err := c.send(req)
	if err != nil {
		return result, nil, err
	}
	defer res.Body.Close()
	if _, ok := any(result).(struct{}); ok {
		_, _ = io.Copy(io.Discard, res.Body)
		return result, res.Header, nil
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return result, nil, fmt.Errorf("kintone: decode response: %w", err)
	}
	return result, res.Header, nil
}

// apiRequest is a request of an API. The response is decoded by decode, or
// ignored if decode is nil.
type apiRequest[T any] struct {
	method string
	path   string
	params any
	decode func(json.RawMessage) (T, error)
}

// call sends the API request and decodes the response.
func (c *Client) call[T any](ctx context.Context, r apiRequest[T]) (T, error) {
	var result T
	if r.decode == nil {
		_, err := c.do[struct{}](ctx, r.method, r.path, r.params)
		return result, err
	}
	raw, err := c.do[json.RawMessage](ctx, r.method, r.path, r.params)
	if err != nil {
		return result, err
	}
	return r.decode(raw)
}

// decodeWith returns a function to decode the response as W, and convert it
// with f.
func decodeWith[W, T any](f func(W) (T, error)) func(json.RawMessage) (T, error) {
	return func(b json.RawMessage) (T, error) {
		var w W
		if err := json.Unmarshal(b, &w); err != nil {
			var result T
			return result, fmt.Errorf("kintone: decode response: %w", err)
		}
		return f(w)
	}
}

func decodeRevision(r struct {
	Revision int64 `json:"revision,string"`
}) (int64, error) {
	return r.Revision, nil
}

// send the request with authentication headers. The caller must close
// the response body. Non-2xx responses are returned as *Error.
func (c *Client) send(req *http.Request) (*http.Response, error) {
	for k, vs := range c.header {
		req.Header[k] = slices.Clone(vs)
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kintone: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		defer res.Body.Close()
		return nil, parseError(res)
	}
	return res, nil
}

// Error is an error response from kintone. Header is the response header.
type Error struct {
	StatusCode int                    `json:"-"`
	Header     http.Header            `json:"-"`
	Code       string                 `json:"code"`
	ID         string                 `json:"id"`
	Message    string                 `json:"message"`
	Errors     map[string]ErrorDetail `json:"errors,omitempty"`
}

// ErrorDetail is the error detail of a specific field or parameter.
type ErrorDetail struct {
	Messages []string `json:"messages"`
}

func (e *Error) Error() string {
	var sb strings.Builder
	sb.WriteString("kintone: ")
	if e.Message != "" {
		sb.WriteString(e.Message)
	} else if text := http.StatusText(e.StatusCode); text != "" {
		sb.WriteString(text)
	} else {
		sb.WriteString("unknown error")
	}
	fmt.Fprintf(&sb, " (status: %d", e.StatusCode)
	if e.Code != "" {
		fmt.Fprintf(&sb, ", code: %s", e.Code)
	}
	if e.ID != "" {
		fmt.Fprintf(&sb, ", id: %s", e.ID)
	}
	sb.WriteString(")")
	return sb.String()
}

func parseError(res *http.Response) error {
	b, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err == nil {
		var bulk struct {
			Results []json.RawMessage `json:"results"`
		}
		if json.Unmarshal(b, &bulk) == nil {
			for i, r := range bulk.Results {
				e := &Error{StatusCode: res.StatusCode, Header: res.Header}
				if json.Unmarshal(r, e) == nil && (e.Code != "" || e.Message != "") {
					return &BulkError{Index: i, Err: e}
				}
			}
		}
	}
	e := &Error{StatusCode: res.StatusCode, Header: res.Header}
	if err == nil && json.Unmarshal(b, e) != nil {
		*e = Error{StatusCode: res.StatusCode, Header: res.Header, Message: truncate(strings.TrimSpace(string(b)), 200)}
	}
	return e
}

func truncate(s string, n int) string {
	var i int
	for j := range s {
		if i == n {
			return s[:j] + "..."
		}
		i++
	}
	return s
}

// BulkError is an error of a bulk request. Index is the index of the request
// that failed, and Err is its error.
type BulkError struct {
	Index int
	Err   *Error
}

func (e *BulkError) Error() string {
	return fmt.Sprintf("kintone: bulk request %d: %s", e.Index, strings.TrimPrefix(e.Err.Error(), "kintone: "))
}

// Unwrap returns the error of the request that failed.
func (e *BulkError) Unwrap() error {
	return e.Err
}

// IsErrorCode reports whether err is an *Error with the code.
func IsErrorCode(err error, code string) bool {
	var e *Error
	return errors.As(err, &e) && e.Code == code
}
