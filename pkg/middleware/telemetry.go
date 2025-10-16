package middleware

import (
	"context"
	"fmt"
	"time"
)

// TelemetryHooks defines callbacks for observability integration.
// All hooks are optional and will be called if non-nil.
// Reference: architecture.md (Middleware Layering section)
type TelemetryHooks struct {
	// OnRequestStart is called before sending a request
	OnRequestStart func(ctx context.Context, method, url string)

	// OnRetry is called when a request is retried
	OnRetry func(ctx context.Context, attempt int, err error)

	// OnResponse is called after receiving a successful response
	OnResponse func(ctx context.Context, statusCode int, duration float64)

	// OnError is called when a request fails after all retries
	OnError func(ctx context.Context, err error)
}

// correlationIDKey is the context key for correlation IDs
type correlationIDKey struct{}

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey{}, correlationID)
}

// GetCorrelationID retrieves the correlation ID from the context
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDKey{}).(string); ok {
		return id
	}
	return ""
}

// GenerateCorrelationID generates a simple correlation ID
// In production, consider using UUID v4
func GenerateCorrelationID() string {
	// Simple timestamp-based ID for now
	// TODO: Replace with proper UUID v4 generation using crypto/rand
	return fmt.Sprintf("corr-%d", time.Now().UnixNano())
}

// Note: For proper UUID v4 generation without external dependencies,
// we would implement RFC 4122 using crypto/rand. For now, using simple
// timestamp-based IDs to maintain stdlib-only constraint.
