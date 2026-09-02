package soroban

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Northbound-Dev/soroban-go-toolkit/internal/jsonrpc"
)

const (
	// TestnetURL is the Stellar-operated public RPC endpoint for testnet, and
	// the default target of a client built with no options.
	TestnetURL = "https://soroban-testnet.stellar.org"

	// FuturenetURL is the Stellar-operated public RPC endpoint for futurenet.
	FuturenetURL = "https://rpc-futurenet.stellar.org"

	// DefaultTimeout bounds each request when the caller supplies neither
	// WithTimeout nor WithHTTPClient.
	DefaultTimeout = 30 * time.Second
)

// Client talks to a single Stellar RPC endpoint. It is safe for concurrent use
// by multiple goroutines.
type Client struct {
	rpc *jsonrpc.Client
	url string
}

type config struct {
	url        string
	httpClient *http.Client
	timeout    time.Duration
}

// Option customises a Client built by New.
type Option func(*config)

// WithURL targets a specific RPC endpoint instead of the testnet default. The
// URL must be absolute and use http or https.
func WithURL(endpoint string) Option {
	return func(c *config) { c.url = endpoint }
}

// WithHTTPClient supplies the http.Client used for every request, for callers
// who need custom transports, proxies, or instrumentation. Because the supplied
// client carries its own timeout, this option overrides WithTimeout.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *config) { c.httpClient = hc }
}

// WithTimeout bounds each request. It is ignored when WithHTTPClient is also
// given. Per-call deadlines can be set with the context instead.
func WithTimeout(d time.Duration) Option {
	return func(c *config) { c.timeout = d }
}

// New returns a Client. With no options it targets TestnetURL with
// DefaultTimeout. The endpoint is validated eagerly, so a malformed URL is
// reported here rather than on the first call.
func New(opts ...Option) (*Client, error) {
	cfg := config{url: TestnetURL, timeout: DefaultTimeout}
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.url == "" {
		return nil, fmt.Errorf("soroban: endpoint URL is empty")
	}
	parsed, err := url.Parse(cfg.url)
	if err != nil {
		return nil, fmt.Errorf("soroban: parse endpoint %q: %w", cfg.url, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("soroban: endpoint %q must use http or https, got %q", cfg.url, parsed.Scheme)
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("soroban: endpoint %q has no host", cfg.url)
	}

	hc := cfg.httpClient
	if hc == nil {
		hc = &http.Client{Timeout: cfg.timeout}
	}

	return &Client{rpc: jsonrpc.NewClient(cfg.url, hc), url: cfg.url}, nil
}

// Endpoint reports the RPC URL this client was built with.
func (c *Client) Endpoint() string { return c.url }

// call performs the RPC and normalises transport errors into this package's
// exported error types, so that callers never receive an internal type they
// cannot assert on.
func (c *Client) call(ctx context.Context, method string, params, out any) error {
	return translateError(c.rpc.Call(ctx, method, params, out))
}
