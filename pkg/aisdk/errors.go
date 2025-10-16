package aisdk

import (
	"errors"
	"fmt"
	"time"
)

// Common error variables
var (
	// ErrMissingAPIKey indicates that no API key was provided
	ErrMissingAPIKey = errors.New("API key required; set OPENAI_API_KEY environment variable or provide via ClientConfig.APIKey")

	// ErrMissingModel indicates that no model was specified in the request
	ErrMissingModel = errors.New("Model is required (e.g., 'gpt-5')")

	// ErrMissingInput indicates that no input was provided in the request
	ErrMissingInput = errors.New("Input is required (string or []Message)")

	// ErrInvalidTemperature indicates that temperature is out of range
	ErrInvalidTemperature = errors.New("Temperature must be between 0.0 and 2.0")

	// ErrInvalidMaxTokens indicates that max tokens is invalid
	ErrInvalidMaxTokens = errors.New("MaxTokens must be positive")

	// ErrInvalidTimeout indicates that timeout is invalid
	ErrInvalidTimeout = errors.New("Timeout must be positive duration")

	// ErrInvalidMaxRetries indicates that max retries is invalid
	ErrInvalidMaxRetries = errors.New("MaxRetries cannot be negative")

	// ErrInvalidTextFormat indicates that text format configuration is invalid
	ErrInvalidTextFormat = errors.New("Structured outputs require strict: true")

	// ErrInvalidReasoningEffort indicates that reasoning effort is invalid
	ErrInvalidReasoningEffort = errors.New("Reasoning effort must be 'low', 'medium', or 'high'")
)

// APIError represents an error returned by an AI provider's API.
// Reference: docs/providers/openai.md lines 8-931
type APIError struct {
	// StatusCode is the HTTP status code
	StatusCode int

	// Code is the provider-specific error code
	Code string

	// Message is the human-readable error message
	Message string

	// CorrelationID is the request correlation ID for tracing
	CorrelationID string
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.CorrelationID != "" {
		return fmt.Sprintf("API error (status=%d, code=%s, correlation_id=%s): %s",
			e.StatusCode, e.Code, e.CorrelationID, e.Message)
	}
	return fmt.Sprintf("API error (status=%d, code=%s): %s", e.StatusCode, e.Code, e.Message)
}

// RateLimitInfo contains rate limit state extracted from HTTP headers.
// Reference: docs/providers/openai.md lines 8-931 (implied by HTTP rate limit headers)
type RateLimitInfo struct {
	// Limit is the maximum requests allowed in the time window
	Limit int

	// Remaining is the requests left in the current window
	Remaining int

	// ResetAt is when the rate limit window resets
	ResetAt time.Time

	// RetryAfter is the delay before retrying (from Retry-After header)
	// Only populated on 429 responses
	RetryAfter time.Duration
}

// RateLimitError represents a rate limit error with additional rate limit details.
// Reference: docs/providers/openai.md lines 8-931
type RateLimitError struct {
	*APIError

	// RateLimitInfo contains the parsed rate limit state
	RateLimitInfo *RateLimitInfo
}

// Error implements the error interface
func (e *RateLimitError) Error() string {
	baseErr := e.APIError.Error()
	if e.RateLimitInfo != nil && e.RateLimitInfo.RetryAfter > 0 {
		return fmt.Sprintf("%s (retry_after=%v)", baseErr, e.RateLimitInfo.RetryAfter)
	}
	if e.RateLimitInfo != nil && e.RateLimitInfo.Remaining >= 0 {
		return fmt.Sprintf("%s (remaining=%d/%d)", baseErr, e.RateLimitInfo.Remaining, e.RateLimitInfo.Limit)
	}
	return baseErr
}

// Unwrap returns the underlying APIError for errors.Is/As
func (e *RateLimitError) Unwrap() error {
	return e.APIError
}

// NewAPIError creates a new APIError with the given details
func NewAPIError(statusCode int, code, message, correlationID string) *APIError {
	return &APIError{
		StatusCode:    statusCode,
		Code:          code,
		Message:       message,
		CorrelationID: correlationID,
	}
}

// NewRateLimitError creates a new RateLimitError with the given details
func NewRateLimitError(statusCode int, code, message, correlationID string, info *RateLimitInfo) *RateLimitError {
	return &RateLimitError{
		APIError: &APIError{
			StatusCode:    statusCode,
			Code:          code,
			Message:       message,
			CorrelationID: correlationID,
		},
		RateLimitInfo: info,
	}
}

// WrapError wraps an error with context information
func WrapError(err error, operation string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}

// IsRetryable determines if an error is retryable based on its type and status code.
// Reference: research.md decision #6 (Error Classification Strategy)
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Rate limit errors are always retryable
	var rateLimitErr *RateLimitError
	if errors.As(err, &rateLimitErr) {
		return true
	}

	// Check API errors by status code
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case 429: // Rate limit
			return true
		case 500, 502, 503, 504: // Server errors
			return true
		case 401, 403, 400, 404: // Client errors
			return false
		default:
			return false
		}
	}

	// Unknown errors are not retryable by default
	return false
}

// IsRateLimitError checks if an error is a rate limit error
func IsRateLimitError(err error) bool {
	var rateLimitErr *RateLimitError
	return errors.As(err, &rateLimitErr)
}
