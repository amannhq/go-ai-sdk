package aisdk

import (
	"strings"
)

// Response represents the complete output from an AI provider.
// Reference: docs/providers/openai.md lines 934-1344, data-model.md Entity #3
type Response struct {
	// ID is the unique response identifier (used for conversation chaining)
	ID string `json:"id"`

	// Object is the response type (always "response")
	Object string `json:"object"`

	// Output contains the model's generated content (array of OutputItem)
	Output []OutputItem `json:"output"`

	// Usage tracks token consumption
	Usage TokenUsage `json:"usage"`

	// Model is the model that generated the response
	Model string `json:"model"`

	// Created is the Unix timestamp of response creation
	Created int64 `json:"created"`

	// RateLimitInfo contains rate limit state (extracted from headers)
	RateLimitInfo *RateLimitInfo `json:"-"` // Not in JSON response
}

// TokenUsage tracks token consumption for billing/monitoring.
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OutputItem represents a single item in the Response.Output array.
// Reference: docs/providers/openai.md lines 934-1344, data-model.md Entity #4
type OutputItem struct {
	// ID is the unique item identifier
	ID string `json:"id"`

	// Type identifies the item kind ("message", "function_call", etc.)
	Type string `json:"type"`

	// Role is the message role ("assistant", "tool")
	Role string `json:"role"`

	// Content contains the item's content parts
	Content []ContentPart `json:"content"`
}

// ContentPart represents a fragment of content within an OutputItem.
// Reference: docs/providers/openai.md lines 934-1344, data-model.md Entity #5
type ContentPart struct {
	// Type identifies the content kind ("output_text", "refusal", etc.)
	Type string `json:"type"`

	// Text contains the text content (present for output_text type)
	Text string `json:"text,omitempty"`

	// Annotations contains inline citations or other metadata
	Annotations []Annotation `json:"annotations,omitempty"`

	// Refusal contains refusal reason (present for refusal type)
	Refusal string `json:"refusal,omitempty"`
}

// Annotation represents inline metadata (citations, warnings).
type Annotation struct {
	Type       string `json:"type"`
	Text       string `json:"text"`
	StartIndex int    `json:"start_index"`
	EndIndex   int    `json:"end_index"`
}

// OutputText is a convenience method aggregating all text content (FR-012).
// Handles fragmented output across multiple OutputItem and ContentPart.
// Reference: data-model.md Entity #3
func (r *Response) OutputText() string {
	var buf strings.Builder
	for _, item := range r.Output {
		for _, part := range item.Content {
			if part.Type == "output_text" {
				buf.WriteString(part.Text)
			}
		}
	}
	return buf.String()
}
