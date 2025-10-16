package http

import (
	"net"
	"net/http"
	"time"
)

// HTTPClient wraps the standard library http.Client with optimized configuration.
// Reference: research.md decision #1 (connection pooling config)
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient creates a new HTTPClient with production-optimized settings.
// Connection pooling: MaxIdleConns=100, IdleConnTimeout=90s per research.md
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
			},
		},
	}
}

// Client returns the underlying http.Client
func (c *HTTPClient) Client() *http.Client {
	return c.client
}
