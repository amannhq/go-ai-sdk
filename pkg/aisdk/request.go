package aisdk

// CreateResponseRequest represents a request to an AI provider's API.
// Reference: docs/providers/openai.md lines 8-931, 934-1344, data-model.md Entity #2
type CreateResponseRequest struct {
	// Model is the AI model ID (required, e.g., "gpt-5", "gpt-4o")
	Model string `json:"model"`

	// Input is the prompt text or structured message array (required)
	// String for simple prompts; []Message for multi-turn conversations
	Input interface{} `json:"input"` // string or []Message

	// Instructions provides high-level behavior guidance (optional)
	Instructions string `json:"instructions,omitempty"`

	// Temperature controls randomness (optional, 0.0-2.0)
	Temperature *float64 `json:"temperature,omitempty"`

	// MaxTokens limits response length (optional)
	MaxTokens *int `json:"max_tokens,omitempty"`

	// Stream enables streaming mode (optional, default false)
	Stream bool `json:"stream,omitempty"`

	// TextFormat specifies structured output schema (optional)
	TextFormat *TextFormat `json:"text,omitempty"`

	// PreviousResponseID links to prior conversation turn (optional)
	PreviousResponseID string `json:"previous_response_id,omitempty"`

	// Reasoning controls o-series reasoning depth (optional)
	Reasoning *ReasoningConfig `json:"reasoning,omitempty"`
}

// TextFormat defines structured output schema.
// Reference: data-model.md Entity #2
type TextFormat struct {
	Type   string                 `json:"type"`   // "json_schema"
	Name   string                 `json:"name"`   // Schema name
	Schema map[string]interface{} `json:"schema"` // JSON Schema object
	Strict bool                   `json:"strict"` // Must be true for structured outputs
}

// ReasoningConfig controls reasoning behavior for o-series models.
type ReasoningConfig struct {
	// Effort specifies reasoning depth: "low", "medium", or "high"
	Effort string `json:"effort"`
}

// Validate checks CreateResponseRequest for required fields and constraints (FR-003).
// Reference: data-model.md Entity #2
func (r *CreateResponseRequest) Validate() error {
	if r.Model == "" {
		return ErrMissingModel
	}
	if r.Input == nil {
		return ErrMissingInput
	}
	if r.Temperature != nil && (*r.Temperature < 0.0 || *r.Temperature > 2.0) {
		return ErrInvalidTemperature
	}
	if r.MaxTokens != nil && *r.MaxTokens <= 0 {
		return ErrInvalidMaxTokens
	}
	if r.TextFormat != nil && r.TextFormat.Type == "json_schema" && !r.TextFormat.Strict {
		return ErrInvalidTextFormat
	}
	if r.Reasoning != nil {
		effort := r.Reasoning.Effort
		if effort != "low" && effort != "medium" && effort != "high" {
			return ErrInvalidReasoningEffort
		}
	}
	return nil
}
