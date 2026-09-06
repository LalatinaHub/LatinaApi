package httputil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// DefaultClient is a high-performance, pooled HTTP client.
var DefaultClient = NewClient(5 * time.Second)

// Client wraps http.Client with connection pooling and retry capabilities.
type Client struct {
	httpClient *http.Client
}

// NewClient returns a new Client with a pooled Transport and specific timeout.
func NewClient(timeout time.Duration) *Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   20,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: timeout,
	}

	return &Client{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   timeout,
		},
	}
}

// Get performs an HTTP GET request with context.
func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "LatinaApi/1.0")
	return c.httpClient.Do(req)
}

// GetJSON performs an HTTP GET and decodes the JSON response body into target.
func (c *Client) GetJSON(ctx context.Context, url string, target any) error {
	resp, err := c.Get(ctx, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("request to %s failed with status %d: %s", url, resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

// GetWithRetry performs an HTTP GET with max retries and exponential backoff.
func (c *Client) GetWithRetry(ctx context.Context, url string, maxRetries int, initialBackoff time.Duration) (*http.Response, error) {
	var (
		lastErr error
		backoff = initialBackoff
	)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				backoff *= 2
			}
		}

		resp, err := c.Get(ctx, url)
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}

		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("server error status %d", resp.StatusCode)
			_ = resp.Body.Close()
		}
	}

	return nil, fmt.Errorf("after %d retries: %w", maxRetries, lastErr)
}
