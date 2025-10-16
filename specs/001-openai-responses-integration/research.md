# Phase 0: Research & Architecture Decisions

**Feature**: OpenAI Responses Integration  
**Date**: 2025-10-16  
**Status**: Complete

## Research Tasks

### 1. SSE (Server-Sent Events) Parsing in Go stdlib

**Question**: How to parse OpenAI's streaming response format using only Go standard library?

**Decision**: Use `bufio.Scanner` with custom `SplitFunc` to parse SSE format (lines prefixed with `data: `, `event: `, etc.)

**Rationale**:
- OpenAI streams follow SSE spec: `data: {json}\n\n` delimiter pattern
- `bufio.Scanner` provides line-by-line reading with minimal allocations
- Custom split function handles multi-line events and `data:` prefix stripping
- Avoids third-party SSE libraries while maintaining performance

**Alternatives Considered**:
- **Manual `bufio.Reader` + state machine**: More control but error-prone; scanner abstraction reduces bugs
- **Third-party `sse` package**: Violates constitution standard-library-only principle; governance exception not justified
- **Buffering full response**: Memory inefficient for long streams; defeats streaming purpose

**Implementation Notes**:
```go
// Custom split function for SSE format
func scanSSE(data []byte, atEOF bool) (advance int, token []byte, err error) {
    // Look for double newline (event boundary)
    // Strip "data: " prefix and parse JSON
}
```

**Documentation Reference**: `docs/providers/openai.md` lines 7618-7751 (SSE event format specification)

---

### 2. Go Struct Tags → JSON Schema Conversion

**Question**: How to convert Go structs to JSON Schema for structured outputs (FR-011) without external libraries?

**Decision**: Implement reflection-based schema generator using `reflect` package to introspect struct tags and generate OpenAI-compatible JSON Schema

**Rationale**:
- OpenAI structured outputs require `strict: true` JSON Schema with exact field types
- Go struct tags (`json:"field_name"`, custom `jsonschema:"required,description=..."`) map cleanly to JSON Schema properties
- Reflection is stdlib; performance acceptable for schema generation (happens once per request type)
- Provides type-safe schema definition in Go (constitution Principle II)

**Alternatives Considered**:
- **Manual JSON Schema strings**: Error-prone, no compile-time validation, violates strongly-typed principle
- **Code generation tool**: Adds build complexity; reflection approach simpler for SDK use case
- **Third-party schema library (e.g., `jsonschema`)**: Violates standard-library-only constraint

**Implementation Notes**:
```go
// Example struct with tags
type CalendarEvent struct {
    Name         string   `json:"name" jsonschema:"required,description=Event name"`
    Date         string   `json:"date" jsonschema:"required"`
    Participants []string `json:"participants"`
}

// Reflection-based converter
func StructToJSONSchema(v interface{}) (map[string]interface{}, error) {
    // Use reflect.TypeOf to inspect fields, tags
    // Generate {"type":"object","properties":{...},"required":[...]}
}
```

**Documentation Reference**: `docs/providers/openai.md` lines 2193-4038 (JSON Schema format requirements for structured outputs)

---

### 3. Rate Limit State Thread Safety

**Question**: How to safely track rate limit state across concurrent goroutines without external sync primitives?

**Decision**: Use `sync.RWMutex` to protect `RateLimitInfo` struct; reads are frequent (every response), writes rare (rate limit updates)

**Rationale**:
- Go's `sync` package is stdlib, optimized for low-contention scenarios
- Read-heavy workload (check remaining quota) benefits from `RLock()`
- Write lock only on 429 responses or header updates (rare)
- Avoids channels/atomics complexity; clear ownership model

**Alternatives Considered**:
- **Atomic values (`sync/atomic`)**: Complex for struct updates; requires encoding state as single int64
- **Channel-based actor**: Over-engineered for simple state; adds goroutine overhead
- **No synchronization**: Race conditions would corrupt rate limit state under concurrent load

**Implementation Notes**:
```go
type RateLimitTracker struct {
    mu   sync.RWMutex
    info RateLimitInfo
}

func (t *RateLimitTracker) Update(headers http.Header) {
    t.mu.Lock()
    defer t.mu.Unlock()
    // Parse headers, update t.info
}

func (t *RateLimitTracker) Get() RateLimitInfo {
    t.mu.RLock()
    defer t.mu.RUnlock()
    return t.info
}
```

**Documentation Reference**: Constitution Principle IV (thread-safe concurrent requests); Go concurrency best practices

---

### 4. Context Cancellation Propagation

**Question**: How to ensure context cancellation terminates HTTP requests + streaming goroutines within 100ms (SC-009)?

**Decision**: Pass `context.Context` to `http.Request.WithContext()`; for streaming, use `context.Done()` channel select in read loop

**Rationale**:
- `http.Client` respects request context; cancels in-flight connections immediately
- Streaming goroutine checks `ctx.Done()` on each event read; breaks loop on cancellation
- `defer` cleanup ensures resources released (connections closed, goroutines exit)
- Standard Go pattern; no custom cancellation logic needed

**Alternatives Considered**:
- **Manual timeout tracking**: Error-prone; stdlib context is battle-tested
- **Separate cancel channel**: Duplicates context functionality
- **No cancellation support**: Violates constitution resilience principle

**Implementation Notes**:
```go
func (c *Client) StreamResponse(ctx context.Context, req Request) (*Stream, error) {
    httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, body)
    resp, err := c.httpClient.Do(httpReq) // Cancelled if ctx.Done()
    
    stream := &Stream{
        scanner: bufio.NewScanner(resp.Body),
        ctx:     ctx,
    }
    
    go stream.readLoop() // Checks ctx.Done() in loop
    return stream, nil
}

func (s *Stream) readLoop() {
    defer s.Close()
    for {
        select {
        case <-s.ctx.Done():
            return // Immediate termination
        default:
            if !s.scanner.Scan() { return }
            // Process event
        }
    }
}
```

**Documentation Reference**: Go `context` package docs; constitution SC-009 (100ms cancellation response time)

---

### 5. Exponential Backoff with Jitter

**Question**: What backoff formula ensures predictable retry behavior without thundering herd issues?

**Decision**: Use `min(base * 2^attempt, maxDelay) + jitter` where base=1s, maxDelay=60s, jitter=±20%

**Rationale**:
- Exponential: 1s → 2s → 4s → 8s → 16s → 32s → 60s (capped)
- Jitter prevents synchronized retries across multiple clients
- 3 retries default (total ~7s max wait) balances responsiveness vs success rate
- Aligns with industry best practices (AWS SDK, Google Cloud SDK use similar formulas)

**Alternatives Considered**:
- **Linear backoff**: Too slow to recover; doesn't back off enough under load
- **No jitter**: Causes thundering herd when many clients hit rate limits simultaneously
- **Fibonacci sequence**: Unnecessarily complex vs exponential with cap

**Implementation Notes**:
```go
func calculateBackoff(attempt int) time.Duration {
    base := 1 * time.Second
    maxDelay := 60 * time.Second
    
    delay := base * (1 << uint(attempt)) // 2^attempt
    if delay > maxDelay {
        delay = maxDelay
    }
    
    jitter := time.Duration(rand.Float64() * 0.4 * float64(delay)) // ±20%
    return delay + jitter - (delay / 5) // Center jitter around delay
}
```

**Documentation Reference**: Constitution Principle IV (deterministic backoff policies); RFC 7230 (retry guidance)

---

### 6. Error Classification Strategy

**Question**: How to distinguish retryable vs permanent errors across HTTP status codes and OpenAI error types?

**Decision**: Classify by HTTP status + OpenAI error code:
- **Retryable**: 429 (rate limit), 5xx (server error), network timeouts
- **Permanent**: 400 (bad request), 401 (auth), 403 (forbidden), 404 (not found)
- **Special**: 429 with `Retry-After` header → use that delay, not exponential backoff

**Rationale**:
- Matches OpenAI API behavior (4xx client errors don't resolve with retries)
- `Retry-After` header provides optimal delay from server
- Prevents retry loops on misconfiguration (missing API key)
- Constitution Principle V: fail fast on misconfiguration

**Alternatives Considered**:
- **Retry all errors**: Wastes quota on permanent failures
- **Never retry 4xx**: Misses legitimate 429 rate limits
- **Fixed retry delay**: Ignores server guidance (`Retry-After`)

**Implementation Notes**:
```go
func (e *APIError) IsRetryable() bool {
    switch e.StatusCode {
    case 429:
        return true // Rate limit
    case 500, 502, 503, 504:
        return true // Server errors
    case 401, 403, 400, 404:
        return false // Client errors
    default:
        return false
    }
}

func (e *RateLimitError) RetryAfter() time.Duration {
    if e.RetryAfterSec > 0 {
        return time.Duration(e.RetryAfterSec) * time.Second
    }
    return 0 // Use exponential backoff
}
```

**Documentation Reference**: `docs/providers/openai.md` lines 8-931 (HTTP status codes and error responses)

---

### 7. Conversation State Management Approach

**Question**: Should conversation state be managed by SDK or caller (P3 user story)?

**Decision**: Hybrid approach – SDK provides `ConversationChain` helper that tracks `previous_response_id` but caller owns conversation lifecycle

**Rationale**:
- OpenAI Conversations API uses `previous_response_id` to link turns (lines 7095-7400)
- SDK helper reduces boilerplate but doesn't enforce single pattern
- Allows advanced users to manage state externally (e.g., database persistence)
- Aligns with Go principle: "provide mechanism, not policy"

**Alternatives Considered**:
- **SDK-managed state**: Opinionated; limits flexibility for advanced use cases
- **Caller-only management**: Too much boilerplate for common case; hurts P1 onboarding
- **Stateful client**: Breaks concurrent usage; violates constitution concurrency model

**Implementation Notes**:
```go
// Helper for simple cases
type ConversationChain struct {
    client           *Client
    previousResponseID string
}

func (c *ConversationChain) Send(ctx context.Context, input string) (*Response, error) {
    req := CreateResponseRequest{
        Model: "gpt-5",
        Input: input,
        PreviousResponseID: c.previousResponseID, // Link to prior turn
    }
    resp, err := c.client.CreateResponse(ctx, req)
    if err == nil {
        c.previousResponseID = resp.ID // Update for next turn
    }
    return resp, err
}

// Advanced users can still use CreateResponse directly with manual ID tracking
```

**Documentation Reference**: `docs/providers/openai.md` lines 7095-7400 (Conversation state and `previous_response_id` parameter)

---

## Architecture Decisions Summary

| Decision | Choice | Constitution Alignment |
|----------|--------|------------------------|
| SSE Parsing | `bufio.Scanner` + custom split | Standard library only (Eng. Standards) |
| Schema Conversion | Reflection on struct tags | Strongly typed contracts (Principle II) |
| Rate Limit Sync | `sync.RWMutex` | Thread-safe concurrent usage (Principle IV) |
| Cancellation | `context.Context` propagation | Resilience, resource cleanup (Principle IV) |
| Retry Backoff | Exponential with jitter, 3 max | Deterministic policies (Principle IV) |
| Error Classification | Status code + OpenAI error type | Fail fast on config errors (Principle V) |
| Conversation State | Optional SDK helper + manual option | Flexibility without over-engineering (Principle I) |

All research questions resolved with standard-library solutions. No governance exceptions required. Ready for Phase 1 design.

---

## Next Steps

Phase 0 Complete ✅

Proceed to Phase 1:
1. Generate `data-model.md` (type definitions for 7 key entities)
2. Generate `contracts/openai-responses-v1.json` (OpenAPI schema)
3. Generate `quickstart.md` (P1 user story walkthrough)
4. Run `.specify/scripts/bash/update-agent-context.sh copilot` to update agent context
