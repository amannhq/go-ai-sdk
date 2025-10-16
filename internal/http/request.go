package http

import (
	"context"
	"net/http"
	"time"
)

// DoRequest executes an HTTP request with context support.
// Propagates context cancellation and enforces timeouts.
// Reference: research.md decision #4 (context cancellation propagation)
func (c *HTTPClient) DoRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Check if context is already cancelled before making request
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Create request with context
	reqWithContext := req.WithContext(ctx)

	// Execute request
	resp, err := c.client.Do(reqWithContext)
	if err != nil {
		// Wrap error with context if available
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	return resp, nil
}

// DoRequestWithRetry is a helper that provides a simplified interface for retrying requests
func (c *HTTPClient) DoRequestWithRetry(ctx context.Context, req *http.Request, maxRetries int, backoffFunc func(int) time.Duration) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context before attempting
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Clone request for retry (body can only be read once)
		reqClone := req.Clone(ctx)

		resp, err := c.DoRequest(ctx, reqClone)
		if err == nil && resp.StatusCode < 500 && resp.StatusCode != 429 {
			return resp, nil
		}

		// Store error for potential return
		if err != nil {
			lastErr = err
		} else {
			resp.Body.Close() // Close before retry
			lastErr = nil
		}

		// Don't backoff on last attempt
		if attempt < maxRetries {
			delay := backoffFunc(attempt)
			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	// Retry one more time to get the final response
	return c.DoRequest(ctx, req)
}
