package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/furuya-3150/fam-diary-log/pkg/errors"
)

// ClientConfig represents HTTP client configuration
type ClientConfig struct {
	Timeout      time.Duration
	MaxRetries   int
	RetryBackoff time.Duration
}

// DefaultConfig returns default HTTP client configuration
func DefaultConfig() ClientConfig {
	return ClientConfig{
		Timeout:      10 * time.Second,
		MaxRetries:   3,
		RetryBackoff: 1 * time.Second,
	}
}

// Client provides common HTTP operations with retry and error handling
type Client struct {
	httpClient *http.Client
	config     ClientConfig
}

// NewClient creates a new HTTP client with default configuration
func NewClient() *Client {
	return NewClientWithConfig(DefaultConfig())
}

// NewClientWithConfig creates a new HTTP client with custom configuration
func NewClientWithConfig(config ClientConfig) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		config: config,
	}
}

// Do executes an HTTP request with retry logic
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		// Add context to request
		req = req.WithContext(ctx)

		// Execute request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt < c.config.MaxRetries {
				time.Sleep(c.config.RetryBackoff * time.Duration(attempt+1))
				continue
			}
			return nil, &errors.ExternalAPIError{
				Message: "http request failed",
				Cause: lastErr,
			}
		}

		// Check status code
		if resp.StatusCode >= 500 {
			// Retry on server errors
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: status=%d", resp.StatusCode)
			if attempt < c.config.MaxRetries {
				time.Sleep(c.config.RetryBackoff * time.Duration(attempt+1))
				continue
			}
			return nil, &errors.ExternalAPIError{
				Message: fmt.Sprintf("server error: status=%d", resp.StatusCode),
				Cause: lastErr,
			}
		}

		if resp.StatusCode >= 400 {
			// Don't retry on client errors
			return resp, nil
		}

		// Success
		return resp, nil
	}

	return nil, lastErr
}