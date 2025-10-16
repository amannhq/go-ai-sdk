package http

import (
	"net/http"
	"strconv"
	"time"
)

// RateLimitInfo contains rate limit state extracted from HTTP headers.
type RateLimitInfo struct {
	Limit      int
	Remaining  int
	ResetAt    time.Time
	RetryAfter time.Duration
}

// ExtractRateLimitHeaders parses rate limit information from HTTP response headers.
// Extracts X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset, and Retry-After.
// Reference: data-model.md Entity #6 (RateLimitInfo)
func ExtractRateLimitHeaders(headers http.Header) *RateLimitInfo {
	info := &RateLimitInfo{}

	// Extract X-RateLimit-Limit
	if limit := headers.Get("X-RateLimit-Limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			info.Limit = val
		}
	}

	// Extract X-RateLimit-Remaining
	if remaining := headers.Get("X-RateLimit-Remaining"); remaining != "" {
		if val, err := strconv.Atoi(remaining); err == nil {
			info.Remaining = val
		}
	}

	// Extract X-RateLimit-Reset (Unix timestamp)
	if reset := headers.Get("X-RateLimit-Reset"); reset != "" {
		if timestamp, err := strconv.ParseInt(reset, 10, 64); err == nil {
			info.ResetAt = time.Unix(timestamp, 0)
		}
	}

	// Extract Retry-After (seconds)
	if retryAfter := headers.Get("Retry-After"); retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			info.RetryAfter = time.Duration(seconds) * time.Second
		}
	}

	return info
}
