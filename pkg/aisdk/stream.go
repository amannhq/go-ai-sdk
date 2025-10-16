package aisdk

// StreamEvent represents a single event in a streaming response.
// Reference: docs/providers/openai.md lines 7618-7751, data-model.md Entity #7
type StreamEvent struct {
	// Type identifies the event kind (20+ types per OpenAI streaming spec)
	Type string `json:"type"`

	// ResponseID links the event to its parent response
	ResponseID string `json:"response_id,omitempty"`

	// ItemID links the event to its parent output item
	ItemID string `json:"item_id,omitempty"`

	// Delta contains incremental content for text/refusal deltas
	Delta string `json:"delta,omitempty"`

	// Error contains error details for error events
	Error *StreamError `json:"error,omitempty"`

	// Output contains completed output item for item_done events
	Output *OutputItem `json:"output,omitempty"`

	// Usage contains token usage for response_completed events
	Usage *TokenUsage `json:"usage,omitempty"`
}

// StreamError represents an error during streaming.
type StreamError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Common event types (constants for type safety)
const (
	EventResponseCreated    = "response.created"
	EventResponseInProgress = "response.in_progress"
	EventResponseCompleted  = "response.completed"
	EventResponseFailed     = "response.failed"
	EventOutputItemAdded    = "response.output_item.added"
	EventOutputItemDone     = "response.output_item.done"
	EventContentPartAdded   = "response.content_part.added"
	EventContentPartDone    = "response.content_part.done"
	EventOutputTextDelta    = "response.output_text.delta"
	EventOutputTextDone     = "response.output_text.done"
	EventRefusalDelta       = "response.refusal.delta"
	EventRefusalDone        = "response.refusal.done"
	EventError              = "error"
)

// StreamReader provides an interface for reading streaming events.
// Reference: architecture.md (Streaming flow)
type StreamReader interface {
	// Next returns the next event from the stream.
	// Returns io.EOF when the stream is complete.
	Next() (*StreamEvent, error)

	// Close terminates the stream and cleans up resources.
	Close() error
}
