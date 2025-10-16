package openai

import (
	"os"

	"github.com/amannhq/go-ai-sdk/pkg/aisdk"
)

// NewConfigFromEnv creates a Config loading the API key from environment.
// Reads OPENAI_API_KEY environment variable.
// Reference: FR-013 (environment-based configuration)
func NewConfigFromEnv() (*Config, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, aisdk.ErrMissingAPIKey
	}

	config := DefaultConfig()
	config.APIKey = apiKey
	return config, nil
}
