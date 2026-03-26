package gerritsdk

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	gogerrit "github.com/andygrunwald/go-gerrit"
)

// Client is a high-level Gerrit REST API client.
type Client struct {
	inner *gogerrit.Client
	host  string
}

// Option configures a Client.
type Option func(*clientConfig)

type clientConfig struct {
	httpClient *http.Client
	auth       AuthMethod
}

// WithHTTPClient sets a custom http.Client for the underlying transport.
func WithHTTPClient(c *http.Client) Option {
	return func(cfg *clientConfig) { cfg.httpClient = c }
}

// WithAuth sets the authentication method.
func WithAuth(a AuthMethod) Option {
	return func(cfg *clientConfig) { cfg.auth = a }
}

// NewClient creates a Gerrit client for the given host.
// The host should be a bare hostname like "bazel-review.googlesource.com".
func NewClient(ctx context.Context, host string, opts ...Option) (*Client, error) {
	cfg := &clientConfig{}
	for _, o := range opts {
		o(cfg)
	}

	httpClient := cfg.httpClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	if cfg.auth != nil {
		t, err := cfg.auth.apply(transportOrDefault(httpClient.Transport))
		if err != nil {
			return nil, fmt.Errorf("auth for %s: %w", host, err)
		}
		httpClient.Transport = t
	}

	// Gerrit authenticated endpoints are under /a/
	baseURL := fmt.Sprintf("https://%s/a/", host)

	inner, err := gogerrit.NewClient(ctx, baseURL, httpClient)
	if err != nil {
		return nil, fmt.Errorf("create gerrit client for %s: %w", host, err)
	}

	return &Client{inner: inner, host: host}, nil
}

// NewAnonymousClient creates an unauthenticated Gerrit client.
// Only public endpoints will be accessible.
func NewAnonymousClient(ctx context.Context, host string, opts ...Option) (*Client, error) {
	cfg := &clientConfig{}
	for _, o := range opts {
		o(cfg)
	}

	httpClient := cfg.httpClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	baseURL := fmt.Sprintf("https://%s/", host)

	inner, err := gogerrit.NewClient(ctx, baseURL, httpClient)
	if err != nil {
		return nil, fmt.Errorf("create anonymous gerrit client for %s: %w", host, err)
	}

	return &Client{inner: inner, host: host}, nil
}

// Host returns the Gerrit instance hostname.
func (c *Client) Host() string { return c.host }

// Inner returns the underlying go-gerrit client for advanced usage.
func (c *Client) Inner() *gogerrit.Client { return c.inner }

// changeURL builds the web URL for a change.
func (c *Client) changeURL(project string, number int) string {
	return fmt.Sprintf("https://%s/c/%s/+/%d", c.host, project, number)
}

func transportOrDefault(t http.RoundTripper) http.RoundTripper {
	if t != nil {
		return t
	}
	return http.DefaultTransport
}

// newClientFromURL creates a Client with an explicit base URL. Used for testing.
func newClientFromURL(ctx context.Context, baseURL string, httpClient *http.Client) (*Client, error) {
	inner, err := gogerrit.NewClient(ctx, baseURL, httpClient)
	if err != nil {
		return nil, err
	}
	host := hostFromURL(baseURL)
	return &Client{inner: inner, host: host}, nil
}

// hostFromURL extracts the host from a Gerrit URL prefix like "https://host/a/"
func hostFromURL(baseURL string) string {
	s := strings.TrimPrefix(baseURL, "https://")
	s = strings.TrimPrefix(s, "http://")
	if idx := strings.Index(s, "/"); idx > 0 {
		s = s[:idx]
	}
	return s
}
