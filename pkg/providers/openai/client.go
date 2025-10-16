package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	internalhttp "github.com/amannhq/go-ai-sdk/internal/http"
	"github.com/amannhq/go-ai-sdk/pkg/aisdk"
	"github.com/amannhq/go-ai-sdk/pkg/middleware"
)

// Client implements the Provider interface for OpenAI.
// Reference: architecture.md (Provider Interface Pattern)
type Client struct {
	config      *Config
	httpClient  *internalhttp.HTTPClient
	retryConfig *middleware.RetryConfig
}

// New creates a new OpenAI client with the given configuration.
func New(config *Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Client{
		config:      config,
		httpClient:  internalhttp.NewHTTPClient(config.Timeout),
		retryConfig: middleware.DefaultRetryConfig(),
	}, nil
}

// NewFromEnv creates a new OpenAI client loading configuration from environment.
func NewFromEnv() (*Client, error) {
	config, err := NewConfigFromEnv()
	if err != nil {
		return nil, err
	}
	return New(config)
}

// CreateResponse implements Provider.CreateResponse for OpenAI.
// Reference: data-model.md Entity #2, #3
func (c *Client) CreateResponse(ctx context.Context, req *aisdk.CreateResponseRequest) (*aisdk.Response, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, aisdk.WrapError(err, "openai.CreateResponse")
	}

	// Convert to OpenAI format
	oaiReq := toOpenAIRequest(req)

	// Marshal request
	body, err := json.Marshal(oaiReq)
	if err != nil {
		return nil, aisdk.WrapError(err, "marshal request")
	}

	// Create HTTP request
	url := c.config.BaseURL + "/responses"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, aisdk.WrapError(err, "create http request")
	}

	// Add headers
	addAuthHeaders(httpReq, c.config.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute with retry
	var httpResp *http.Response
	var lastErr error

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		// Execute request
		httpResp, err = c.httpClient.DoRequest(ctx, httpReq)
		if err != nil {
			lastErr = err
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			// Retry on network errors
			if attempt < c.retryConfig.MaxRetries {
				backoff := c.retryConfig.ExponentialBackoff(attempt)
				select {
				case <-time.After(backoff):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			break
		}

		// Check status code
		if httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
			// Success
			break
		}

		// Extract rate limit info
		rateLimitInfo := internalhttp.ExtractRateLimitHeaders(httpResp.Header)

		// Handle error response
		apiErr := mapOpenAIError(httpResp, middleware.GetCorrelationID(ctx))
		httpResp.Body.Close()

		// Check if retryable
		if !middleware.IsRetryableStatus(httpResp.StatusCode) {
			// Non-retryable error
			if httpResp.StatusCode == 429 {
				return nil, aisdk.NewRateLimitError(httpResp.StatusCode, apiErr.Code, apiErr.Message, apiErr.CorrelationID, convertRateLimitInfo(rateLimitInfo))
			}
			return nil, apiErr
		}

		// Retry with backoff
		if attempt < c.retryConfig.MaxRetries {
			var backoff time.Duration
			if httpResp.StatusCode == 429 && rateLimitInfo.RetryAfter > 0 {
				// Use server-provided retry-after
				backoff = rateLimitInfo.RetryAfter
			} else {
				backoff = c.retryConfig.ExponentialBackoff(attempt)
			}
			select {
			case <-time.After(backoff):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		// Max retries exceeded
		if httpResp.StatusCode == 429 {
			return nil, aisdk.NewRateLimitError(httpResp.StatusCode, apiErr.Code, apiErr.Message, apiErr.CorrelationID, convertRateLimitInfo(rateLimitInfo))
		}
		return nil, apiErr
	}

	if lastErr != nil {
		return nil, aisdk.WrapError(lastErr, "openai.CreateResponse")
	}

	if httpResp == nil {
		return nil, aisdk.NewAPIError(0, "unknown", "no response received", middleware.GetCorrelationID(ctx))
	}

	defer httpResp.Body.Close()

	// Parse response
	var oaiResp openAIResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&oaiResp); err != nil {
		return nil, aisdk.WrapError(err, "decode response")
	}

	// Convert to SDK format
	resp := toAISDKResponse(&oaiResp)

	// Attach rate limit info
	resp.RateLimitInfo = convertRateLimitInfo(internalhttp.ExtractRateLimitHeaders(httpResp.Header))

	return resp, nil
}

// convertRateLimitInfo converts internal RateLimitInfo to aisdk.RateLimitInfo
func convertRateLimitInfo(info *internalhttp.RateLimitInfo) *aisdk.RateLimitInfo {
	if info == nil {
		return nil
	}
	return &aisdk.RateLimitInfo{
		Limit:      info.Limit,
		Remaining:  info.Remaining,
		ResetAt:    info.ResetAt,
		RetryAfter: info.RetryAfter,
	}
} // StreamResponse implements Provider.StreamResponse for OpenAI.
// Reference: data-model.md Entity #7
func (c *Client) StreamResponse(ctx context.Context, req *aisdk.CreateResponseRequest) (aisdk.StreamReader, error) {
	// TODO: Implement streaming (T065-T067)
	return nil, fmt.Errorf("streaming not yet implemented")
}
