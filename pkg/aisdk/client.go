package aisdk

import (
	"context"
)

// Provider defines the contract that all AI providers must implement.
// This interface is defined here to avoid circular dependencies.
// Reference: architecture.md (Provider Interface Pattern)
type Provider interface {
	// CreateResponse makes a non-streaming request to the provider.
	CreateResponse(ctx context.Context, req *CreateResponseRequest) (*Response, error)

	// StreamResponse makes a streaming request to the provider.
	StreamResponse(ctx context.Context, req *CreateResponseRequest) (StreamReader, error)
}

// Client is the main SDK client for making AI API requests.
// Reference: architecture.md (Public API Layer)
type Client struct {
	config   *ClientConfig
	provider Provider
}

// New creates a new Client with the given configuration and provider.
// Validates config and initializes the client.
// Reference: data-model.md Entity #1
func New(config *ClientConfig, provider Provider) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if provider == nil {
		return nil, NewAPIError(0, "invalid_config", "provider is required", "")
	}

	return &Client{
		config:   config,
		provider: provider,
	}, nil
}

// CreateResponse makes a non-streaming request to the AI provider.
// Reference: Provider interface pattern
func (c *Client) CreateResponse(ctx context.Context, req *CreateResponseRequest) (*Response, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, WrapError(err, "invalid request")
	}

	// Delegate to provider
	return c.provider.CreateResponse(ctx, req)
}

// StreamResponse makes a streaming request to the AI provider.
// Reference: Provider interface pattern
func (c *Client) StreamResponse(ctx context.Context, req *CreateResponseRequest) (StreamReader, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, WrapError(err, "invalid request")
	}

	// Delegate to provider
	return c.provider.StreamResponse(ctx, req)
}
