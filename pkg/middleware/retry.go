package middleware

import (
	"math"
	"math/rand"
	"time"
)

// RetryConfig configures the retry behavior for transient failures.
// Reference: research.md decision #5 (exponential backoff with jitter)
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts (default: 3)
	MaxRetries int
	
	// BaseDelay is the initial backoff delay (default: 1s)
	BaseDelay time.Duration
	
	// MaxDelay is the maximum backoff delay cap (default: 60s)
	MaxDelay time.Duration
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		MaxDelay:   60 * time.Second,
	}
}

// ExponentialBackoff calculates the backoff delay for the given attempt using
// the formula: min(base * 2^attempt, maxDelay) + jitter
// Reference: research.md decision #5
func (c *RetryConfig) ExponentialBackoff(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	
	// Calculate exponential delay: base * 2^attempt
	exp := math.Pow(2, float64(attempt))
	delay := time.Duration(float64(c.BaseDelay) * exp)
	
	// Cap at maxDelay
	if delay > c.MaxDelay {
		delay = c.MaxDelay
	}
	
	// Add jitter: ±20% of delay
	jitterRange := float64(delay) * 0.4 // Total range is 40% (±20%)
	jitter := (rand.Float64() * jitterRange) - (jitterRange / 2)
	
	finalDelay := delay + time.Duration(jitter)
	if finalDelay < 0 {
		finalDelay = c.BaseDelay
	}
	
	return finalDelay
}

// IsRetryableStatus determines if an HTTP status code represents a retryable error.
// Retryable: 429 (rate limit), 5xx (server errors)
// Non-retryable: 4xx (client errors except 429)
// Reference: research.md decision #6 (error classification strategy)
func IsRetryableStatus(statusCode int) bool {
	switch statusCode {
	case 429: // Rate limit
		return true
	case 500, 502, 503, 504: // Server errors
		return true
	case 401, 403, 400, 404: // Client errors
		return false
	default:
		// Default to false for unknown codes
		return statusCode >= 500
	}
}
