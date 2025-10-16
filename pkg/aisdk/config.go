package aisdk

import (
	"os"
	"time"

	"github.com/amannhq/go-ai-sdk/pkg/middleware"
)

// ClientConfig configures the AI SDK client.
// Reference: data-model.md Entity #1
type ClientConfig struct {
	// APIKey is the provider API key (required)
	// Can be provided directly or loaded from environment variable
	APIKey string

	// BaseURL is the provider API base URL
	BaseURL string

	// Timeout is the HTTP request timeout (default: 60s)
	Timeout time.Duration

	// MaxRetries is the maximum retry attempts for transient failures (default: 3)
	MaxRetries int

	// Logger is an optional structured logger interface for telemetry
	// If nil, logging is disabled
	Logger Logger

	// TelemetryHooks provides optional observability callbacks
	TelemetryHooks *middleware.TelemetryHooks
}

// Logger is a simple logging interface for telemetry
type Logger interface {
	// Log logs a message with optional key-value pairs
	Log(level string, message string, keyvals ...interface{})
}

// DefaultConfig returns a ClientConfig with default values
func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		BaseURL:    "https://api.openai.com/v1",
		Timeout:    60 * time.Second,
		MaxRetries: 3,
	}
}

// NewConfigFromEnv creates a ClientConfig loading the API key from environment.
// Reads OPENAI_API_KEY environment variable.
// Reference: FR-013 (environment-based configuration)
func NewConfigFromEnv() (*ClientConfig, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	config := DefaultConfig()
	config.APIKey = apiKey
	return config, nil
}

// Validate checks ClientConfig for required fields and constraint violations.
// Returns descriptive error per SC-006 (actionable error messages).
func (c *ClientConfig) Validate() error {
	if c.APIKey == "" {
		return ErrMissingAPIKey
	}
	if c.Timeout <= 0 {
		return ErrInvalidTimeout
	}
	if c.MaxRetries < 0 {
		return ErrInvalidMaxRetries
	}
	return nil
}
