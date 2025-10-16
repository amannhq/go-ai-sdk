package openai

import (
	"errors"
	"net/url"
	"time"
)

// Config holds the configuration for the OpenAI provider.
// Reference: data-model.md Entity #1 (ClientConfig)
type Config struct {
	// APIKey is the OpenAI API key (required)
	APIKey string

	// BaseURL is the OpenAI API base URL (default: https://api.openai.com/v1)
	BaseURL string

	// Timeout is the HTTP request timeout (default: 60s)
	Timeout time.Duration
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		BaseURL: "https://api.openai.com/v1",
		Timeout: 60 * time.Second,
	}
}

// Validate checks the Config for required fields and constraints.
// Returns descriptive error per SC-006 (actionable error messages).
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return errors.New("API key required; set OPENAI_API_KEY environment variable or provide via Config.APIKey")
	}

	if c.BaseURL == "" {
		return errors.New("BaseURL cannot be empty")
	}

	// Validate BaseURL is a valid URL
	_, err := url.Parse(c.BaseURL)
	if err != nil {
		return errors.New("BaseURL must be a valid URL")
	}

	if c.Timeout <= 0 {
		return errors.New("Timeout must be positive duration")
	}

	return nil
}
