# Phase 1: Data Model

**Feature**: OpenAI Responses Integration  
**Date**: 2025-10-16  
**Prerequisites**: research.md complete

## Entity Definitions

This document defines the 7 key entities from the specification with Go type mappings, validation rules, and relationships. All types cite governing documentation from `docs/providers/openai.md`.

---

### 1. ClientConfig

**Purpose**: Initialization options for SDK client; validates configuration at startup per constitution engineering standards

**Go Type Definition**:
```go
package aisdk

import "time"

// ClientConfig configures the OpenAI SDK client.
// Reference: docs/providers/openai.md lines 8-931 (API key and endpoint configuration)
type ClientConfig struct {
    // APIKey is the OpenAI API key (required).
    // Can be provided directly or loaded from OPENAI_API_KEY env var via NewFromEnv().
    APIKey string
    
    // BaseURL is the OpenAI API base URL (default: "https://api.openai.com/v1").
    BaseURL string
    
    // Timeout is the HTTP request timeout (default: 60s).
    // Applied to non-streaming requests; streaming respects context cancellation.
    Timeout time.Duration
    
    // MaxRetries is the maximum retry attempts for transient failures (default: 3).
    // Set to 0 to disable automatic retries.
    MaxRetries int
    
    // Logger is an optional structured logger interface for telemetry.
    // If nil, logging is disabled.
    Logger Logger
}

// Validate checks ClientConfig for required fields and constraint violations.
// Returns descriptive error per SC-006 (actionable error messages).
func (c *ClientConfig) Validate() error {
    if c.APIKey == "" {
        return ErrMissingAPIKey // "API key required; set OPENAI_API_KEY environment variable or provide via ClientConfig.APIKey"
    }
    if c.Timeout <= 0 {
        return ErrInvalidTimeout // "Timeout must be positive duration"
    }
    if c.MaxRetries < 0 {
        return ErrInvalidMaxRetries // "MaxRetries cannot be negative"
    }
    return nil
}
```

**Validation Rules**:
- `APIKey`: Required, non-empty string
- `BaseURL`: Optional, defaults to `https://api.openai.com/v1`
- `Timeout`: Required, must be > 0 (default 60s)
- `MaxRetries`: Required, must be >= 0 (default 3)
- `Logger`: Optional, nil disables logging

**Relationships**: Used by `Client` constructor; validated once at initialization

**Documentation Reference**: `docs/providers/openai.md` lines 8-931

---

### 2. Request (CreateResponseRequest)

**Purpose**: Represents complete call to OpenAI Responses API; strongly typed per FR-001

**Go Type Definition**:
```go
package aisdk

// CreateResponseRequest represents a request to the OpenAI Responses API.
// Reference: docs/providers/openai.md lines 8-931, 934-1344
type CreateResponseRequest struct {
    // Model is the OpenAI model ID (required, e.g., "gpt-5", "gpt-4o").
    Model string `json:"model"`
    
    // Input is the prompt text or structured message array (required).
    // String for simple prompts; []Message for multi-turn conversations.
    Input interface{} `json:"input"` // string or []Message
    
    // Instructions provides high-level behavior guidance (optional).
    // Takes priority over user messages per OpenAI model spec.
    Instructions string `json:"instructions,omitempty"`
    
    // Temperature controls randomness (optional, 0.0-2.0, default model-specific).
    Temperature *float64 `json:"temperature,omitempty"`
    
    // MaxTokens limits response length (optional, model-dependent maximum).
    MaxTokens *int `json:"max_tokens,omitempty"`
    
    // Stream enables streaming mode (optional, default false).
    // If true, use StreamResponse() instead of CreateResponse().
    Stream bool `json:"stream,omitempty"`
    
    // TextFormat specifies structured output schema (optional).
    // Requires JSON Schema with strict mode per FR-011.
    TextFormat *TextFormat `json:"text,omitempty"`
    
    // PreviousResponseID links to prior conversation turn (optional).
    // Reference: docs/providers/openai.md lines 7095-7400
    PreviousResponseID string `json:"previous_response_id,omitempty"`
    
    // ReasoningEffort controls o-series reasoning depth (optional, "low"/"medium"/"high").
    ReasoningEffort string `json:"reasoning,omitempty"`
}

// Message represents a single message in multi-turn input.
type Message struct {
    Role    string      `json:"role"`    // "user", "assistant", "developer"
    Content interface{} `json:"content"` // string or []ContentPart
}

// TextFormat defines structured output schema.
type TextFormat struct {
    Type   string                 `json:"type"`   // "json_schema"
    Name   string                 `json:"name"`   // Schema name
    Schema map[string]interface{} `json:"schema"` // JSON Schema object
    Strict bool                   `json:"strict"` // Must be true for structured outputs
}

// Validate checks CreateResponseRequest for required fields and constraints (FR-003).
func (r *CreateResponseRequest) Validate() error {
    if r.Model == "" {
        return ErrMissingModel // "Model is required (e.g., 'gpt-5')"
    }
    if r.Input == nil {
        return ErrMissingInput // "Input is required (string or []Message)"
    }
    if r.Temperature != nil && (*r.Temperature < 0.0 || *r.Temperature > 2.0) {
        return ErrInvalidTemperature // "Temperature must be between 0.0 and 2.0"
    }
    if r.MaxTokens != nil && *r.MaxTokens <= 0 {
        return ErrInvalidMaxTokens // "MaxTokens must be positive"
    }
    if r.TextFormat != nil && r.TextFormat.Type == "json_schema" && !r.TextFormat.Strict {
        return ErrInvalidTextFormat // "Structured outputs require strict: true"
    }
    return nil
}
```

**Validation Rules**:
- `Model`: Required, non-empty string
- `Input`: Required, string or []Message
- `Temperature`: Optional, 0.0-2.0 if provided
- `MaxTokens`: Optional, >0 if provided
- `TextFormat`: If json_schema, Strict must be true
- `Stream`: If true, caller must use StreamResponse() method

**Relationships**: 
- Input can contain []Message with embedded ContentPart
- TextFormat embeds JSON Schema for FR-011
- PreviousResponseID links to prior Response.ID for conversations

**Documentation Reference**: `docs/providers/openai.md` lines 8-931, 934-1344, 2193-4038

---

### 3. Response

**Purpose**: Represents complete output from OpenAI; typed per FR-002

**Go Type Definition**:
```go
package aisdk

// Response represents the complete output from OpenAI Responses API.
// Reference: docs/providers/openai.md lines 934-1344
type Response struct {
    // ID is the unique response identifier (used for conversation chaining).
    ID string `json:"id"`
    
    // Object is the response type (always "response").
    Object string `json:"object"`
    
    // Output contains the model's generated content (array of OutputItem).
    // May include multiple items for tool calls, reasoning, etc.
    Output []OutputItem `json:"output"`
    
    // Usage tracks token consumption.
    Usage TokenUsage `json:"usage"`
    
    // Model is the model that generated the response.
    Model string `json:"model"`
    
    // Created is the Unix timestamp of response creation.
    Created int64 `json:"created"`
    
    // RateLimitInfo contains rate limit state (extracted from headers).
    RateLimitInfo *RateLimitInfo `json:"-"` // Not in JSON response
}

// TokenUsage tracks token consumption for billing/monitoring.
type TokenUsage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}

// OutputText is a convenience method aggregating all text content (FR-012).
// Handles fragmented output across multiple OutputItem and ContentPart.
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
```

**Validation Rules**:
- ID must be non-empty
- Output array can be empty for refusals
- Usage fields are non-negative

**Relationships**:
- Contains []OutputItem (see Entity 4)
- Response.ID used in next request's PreviousResponseID for conversations
- RateLimitInfo attached from HTTP headers (see Entity 6)

**Documentation Reference**: `docs/providers/openai.md` lines 934-1344

---

### 4. OutputItem

**Purpose**: Represents single item in response output array; supports multi-step reasoning and tool calls

**Go Type Definition**:
```go
package aisdk

// OutputItem represents a single item in the Response.Output array.
// Reference: docs/providers/openai.md lines 934-1344
type OutputItem struct {
    // ID is the unique item identifier.
    ID string `json:"id"`
    
    // Type identifies the item kind ("message", "function_call", etc.).
    Type string `json:"type"`
    
    // Role is the message role ("assistant", "tool").
    Role string `json:"role"`
    
    // Content contains the item's content parts.
    Content []ContentPart `json:"content"`
}
```

**Validation Rules**:
- ID must be non-empty
- Type must be valid enum (message, function_call, etc.)
- Role must be valid enum (assistant, tool)
- Content array length >= 0

**Relationships**:
- Owned by Response
- Contains []ContentPart (see Entity 5)
- Multiple OutputItem for tool calls or reasoning steps

**Documentation Reference**: `docs/providers/openai.md` lines 934-1344

---

### 5. ContentPart

**Purpose**: Represents fragment of content within OutputItem; handles text, refusals, annotations

**Go Type Definition**:
```go
package aisdk

// ContentPart represents a fragment of content within an OutputItem.
// Reference: docs/providers/openai.md lines 934-1344
type ContentPart struct {
    // Type identifies the content kind ("output_text", "refusal", etc.).
    Type string `json:"type"`
    
    // Text contains the text content (present for output_text type).
    Text string `json:"text,omitempty"`
    
    // Annotations contains inline citations or other metadata.
    Annotations []Annotation `json:"annotations,omitempty"`
    
    // Refusal contains refusal reason (present for refusal type).
    Refusal string `json:"refusal,omitempty"`
}

// Annotation represents inline metadata (citations, warnings).
type Annotation struct {
    Type      string `json:"type"`
    Text      string `json:"text"`
    StartIndex int   `json:"start_index"`
    EndIndex   int   `json:"end_index"`
}
```

**Validation Rules**:
- Type must be non-empty
- Text present if Type == "output_text"
- Refusal present if Type == "refusal"
- Annotations indices must be valid (StartIndex <= EndIndex)

**Relationships**:
- Owned by OutputItem
- Single OutputItem can contain multiple ContentPart

**Documentation Reference**: `docs/providers/openai.md` lines 934-1344

---

### 6. RateLimitInfo

**Purpose**: Represents rate limit state from HTTP headers; enables informed backoff (FR-007)

**Go Type Definition**:
```go
package aisdk

import "time"

// RateLimitInfo represents rate limit state from OpenAI response headers.
// Reference: docs/providers/openai.md lines 8-931 (implied by HTTP rate limit headers)
type RateLimitInfo struct {
    // Limit is the maximum requests allowed in the time window.
    Limit int
    
    // Remaining is the requests left in the current window.
    Remaining int
    
    // ResetAt is when the rate limit window resets (Unix timestamp).
    ResetAt time.Time
    
    // RetryAfter is the delay before retrying (from Retry-After header).
    // Only populated on 429 responses.
    RetryAfter time.Duration
}

// Extract parses RateLimitInfo from HTTP response headers.
func ExtractRateLimitInfo(headers http.Header) *RateLimitInfo {
    info := &RateLimitInfo{}
    
    if limit := headers.Get("X-RateLimit-Limit"); limit != "" {
        info.Limit, _ = strconv.Atoi(limit)
    }
    if remaining := headers.Get("X-RateLimit-Remaining"); remaining != "" {
        info.Remaining, _ = strconv.Atoi(remaining)
    }
    if reset := headers.Get("X-RateLimit-Reset"); reset != "" {
        timestamp, _ := strconv.ParseInt(reset, 10, 64)
        info.ResetAt = time.Unix(timestamp, 0)
    }
    if retryAfter := headers.Get("Retry-After"); retryAfter != "" {
        seconds, _ := strconv.Atoi(retryAfter)
        info.RetryAfter = time.Duration(seconds) * time.Second
    }
    
    return info
}
```

**Validation Rules**:
- All fields optional (headers may be absent)
- Remaining should be <= Limit
- ResetAt should be future timestamp
- RetryAfter only present on 429 responses

**Relationships**:
- Attached to Response (non-serialized)
- Attached to RateLimitError
- Used by retry middleware

**Documentation Reference**: Standard HTTP rate limit headers; OpenAI API behavior

---

### 7. StreamEvent

**Purpose**: Represents single event in streaming response; typed per FR-008

**Go Type Definition**:
```go
package aisdk

// StreamEvent represents a single event in a streaming response.
// Reference: docs/providers/openai.md lines 7618-7751
type StreamEvent struct {
    // Type identifies the event kind (20+ types per OpenAI streaming spec).
    Type string `json:"type"`
    
    // ResponseID links the event to its parent response.
    ResponseID string `json:"response_id,omitempty"`
    
    // ItemID links the event to its parent output item.
    ItemID string `json:"item_id,omitempty"`
    
    // Delta contains incremental content for text/refusal deltas.
    Delta string `json:"delta,omitempty"`
    
    // Error contains error details for error events.
    Error *StreamError `json:"error,omitempty"`
    
    // Output contains completed output item for item_done events.
    Output *OutputItem `json:"output,omitempty"`
    
    // Usage contains token usage for response_completed events.
    Usage *TokenUsage `json:"usage,omitempty"`
}

// StreamError represents an error during streaming.
type StreamError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

// Common event types (constants for type safety)
const (
    EventResponseCreated       = "response.created"
    EventResponseInProgress    = "response.in_progress"
    EventResponseCompleted     = "response.completed"
    EventResponseFailed        = "response.failed"
    EventOutputItemAdded       = "response.output_item.added"
    EventOutputItemDone        = "response.output_item.done"
    EventContentPartAdded      = "response.content_part.added"
    EventContentPartDone       = "response.content_part.done"
    EventOutputTextDelta       = "response.output_text.delta"
    EventOutputTextDone        = "response.output_text.done"
    EventRefusalDelta          = "response.refusal.delta"
    EventRefusalDone           = "response.refusal.done"
    EventError                 = "error"
)
```

**Validation Rules**:
- Type must be non-empty and valid enum
- Delta present for delta events
- Output present for item_done events
- Error present for error events
- Event types match OpenAI streaming spec

**Relationships**:
- Emitted by StreamResponse() method
- ResponseID/ItemID provide correlation
- Accumulating deltas reconstructs full response

**Documentation Reference**: `docs/providers/openai.md` lines 7618-7751

---

## Entity Relationship Diagram

```
ClientConfig
    └─> Client (1:1)
         ├─> CreateResponseRequest (1:N requests)
         │    ├─> Message[] (0:N for multi-turn)
         │    └─> TextFormat (0:1 for structured)
         │
         ├─> Response (1:1 per request)
         │    ├─> OutputItem[] (1:N items)
         │    │    └─> ContentPart[] (1:N parts)
         │    ├─> TokenUsage (1:1)
         │    └─> RateLimitInfo (0:1)
         │
         └─> StreamEvent[] (streaming mode)
              └─> OutputItem (0:1 on item_done)
              └─> StreamError (0:1 on error)
```

---

## State Transitions

### Request Lifecycle
1. **Created**: ClientConfig validated, Client initialized
2. **Validated**: CreateResponseRequest.Validate() passed
3. **Serialized**: Request marshaled to JSON (FR-004)
4. **Sent**: HTTP POST to `/v1/responses`
5. **Retrying**: On 429/5xx with backoff (FR-006)
6. **Completed**: Response deserialized (FR-005)
7. **Error**: APIError or RateLimitError returned

### Streaming Lifecycle
1. **Connected**: SSE stream opened
2. **Events**: StreamEvent emitted per delta/item
3. **Accumulating**: Client buffers deltas
4. **Completed**: response.completed event
5. **Error**: error event or connection failure
6. **Closed**: Stream resources cleaned up

---

## Next Steps

Phase 1 data-model.md complete ✅

Proceed to:
1. Generate `contracts/openai-responses-v1.json` (OpenAPI schema)
2. Generate `quickstart.md` (example using these types)
3. Run update-agent-context script
