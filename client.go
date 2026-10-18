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
	d := *c
	d.guestSpaceID = 0
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
	var result T
	var body io.Reader
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			return result, fmt.Errorf("kintone: %w", err)
		}
		body = bytes.NewReader(b)
	}
	httpMethod := method
	if method == http.MethodGet && params != nil {
		httpMethod = http.MethodPost
	}
	req, err := http.NewRequestWithContext(ctx, httpMethod, c.endpoint(path).String(), body)
	if err != nil {
		return result, fmt.Errorf("kintone: %w", err)
	}
	if params != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if httpMethod != method {
		req.Header.Set("X-HTTP-Method-Override", method)
	}
	res, err := c.send(req)
	if err != nil {
		return result, err
	}
	defer res.Body.Close()
	if _, ok := any(result).(struct{}); ok {
		_, _ = io.Copy(io.Discard, res.Body)
		return result, nil
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("kintone: decode response: %w", err)
	}
	return result, nil
}

// send the request with authentication headers. The caller must close
// the response body. Non-2xx responses are returned as *Error.
func (c *Client) send(req *http.Request) (*http.Response, error) {
	for k, vs := range c.header {
		req.Header[k] = vs
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

// Error is an error response from kintone.
type Error struct {
	StatusCode int                    `json:"-"`
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
	} else {
		sb.WriteString(http.StatusText(e.StatusCode))
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
	e := &Error{StatusCode: res.StatusCode}
	if err == nil && json.Unmarshal(b, e) != nil {
		*e = Error{StatusCode: res.StatusCode, Message: strings.TrimSpace(string(b))}
	}
	return e
}

// IsErrorCode reports whether err is an *Error with the code.
func IsErrorCode(err error, code string) bool {
	var e *Error
	return errors.As(err, &e) && e.Code == code
}
