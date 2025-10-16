# Architecture

## Overview

The Go AI SDK follows a layered architecture with clear separation of concerns:

1. **Public API Layer** (`pkg/aisdk/`): High-level SDK interface for developers
2. **Provider Layer** (`pkg/providers/`): Provider-specific implementations (OpenAI, future: Anthropic, Gemini)
3. **Middleware Layer** (`pkg/middleware/`): Cross-cutting concerns (retry, rate limiting, telemetry)
4. **Internal Layer** (`internal/`): Shared utilities (HTTP client, schema conversion)

## Provider Interface Pattern

The SDK is built around a minimal provider interface that enables extensibility without coupling:

```go
// Provider defines the contract that all AI providers must implement.
type Provider interface {
    // CreateResponse makes a non-streaming request to the provider.
    CreateResponse(ctx context.Context, req CreateResponseRequest) (*Response, error)
    
    // StreamResponse makes a streaming request to the provider.
    StreamResponse(ctx context.Context, req CreateResponseRequest) (StreamReader, error)
}
```

### Design Principles

1. **Minimal Surface Area**: Only two methods required
2. **Provider-Agnostic Types**: `CreateResponseRequest` and `Response` are not tied to OpenAI
3. **Shared Middleware**: Retry, rate limiting, telemetry work across all providers
4. **Easy Extension**: Adding a new provider requires ~500 LOC adapter with zero shared code changes

## Middleware Layering

Middleware wraps the HTTP client to provide cross-cutting functionality:

```
User Request
    ↓
Client (pkg/aisdk/client.go)
    ↓
Telemetry Middleware (correlation ID, hooks)
    ↓
Retry Middleware (exponential backoff)
    ↓
Rate Limit Middleware (state tracking)
    ↓
Provider Implementation (pkg/providers/openai/)
    ↓
HTTP Client (internal/http/client.go)
    ↓
OpenAI API
```

### Middleware Responsibilities

**Telemetry Middleware** (`pkg/middleware/telemetry.go`):
- Injects correlation IDs via context
- Calls hooks: OnRequestStart, OnRetry, OnResponse, OnError
- Propagates context values for distributed tracing

**Retry Middleware** (`pkg/middleware/retry.go`):
- Classifies errors (retryable vs permanent)
- Implements exponential backoff with jitter
- Respects Retry-After headers
- Max 3 retries by default

**Rate Limit Middleware** (`pkg/middleware/ratelimit.go`):
- Extracts X-RateLimit-* headers
- Tracks remaining quota
- Attaches RateLimitInfo to responses/errors
- Thread-safe for concurrent clients

## Component Interactions

### Request Flow (Non-Streaming)

```
1. User creates CreateResponseRequest
2. Client.CreateResponse() validates request
3. Telemetry middleware adds correlation ID
4. Retry middleware wraps call with backoff logic
5. Provider.CreateResponse() serializes to OpenAI format
6. HTTP client sends POST to /v1/responses
7. Provider deserializes OpenAI response
8. Rate limit middleware extracts headers
9. Telemetry middleware calls OnResponse hook
10. Client returns Response to user
```

### Request Flow (Streaming)

```
1. User creates CreateResponseRequest with Stream=true
2. Client.StreamResponse() validates request
3. Provider.StreamResponse() opens SSE connection
4. StreamReader wraps http.Response.Body
5. User calls Next() in loop to read events
6. SSE parser (bufio.Scanner) yields StreamEvent
7. On ctx.Done(), stream.Close() terminates connection
8. Telemetry middleware calls OnResponse with final state
```

## Stdlib-Only Approach

All functionality uses only Go standard library:

- **HTTP Client**: `net/http` with custom `Transport` for connection pooling
- **SSE Parsing**: `bufio.Scanner` with custom `SplitFunc`
- **JSON Schema**: `reflect` package for struct introspection
- **Concurrency**: `sync.RWMutex` for rate limit state
- **Cancellation**: `context.Context` for timeout/cancel propagation

### Decision Rationale

From `research.md`:

> All research questions resolved with standard-library solutions. No governance exceptions required.

Benefits:
- **Zero dependency risk**: No supply chain vulnerabilities
- **Easy auditing**: All code is in-repo or stdlib
- **Portability**: CGO-free builds work everywhere
- **Simplicity**: Fewer abstractions, clearer code paths

## Error Handling Architecture

Errors flow through three layers:

1. **Validation Errors**: Caught at request creation time (fail-fast)
2. **HTTP Errors**: Classified by retry middleware (retryable vs permanent)
3. **Provider Errors**: Mapped to APIError or RateLimitError with details

```go
// APIError is the base error type for all provider errors
type APIError struct {
    StatusCode    int    // HTTP status code
    Code          string // Provider-specific error code
    Message       string // Human-readable error message
    CorrelationID string // Correlation ID for tracing
}

// RateLimitError extends APIError with rate limit details
type RateLimitError struct {
    *APIError
    RateLimitInfo *RateLimitInfo // Parsed from headers
}
```

Users check errors with `errors.Is()` and `errors.As()`:

```go
var rateLimitErr *aisdk.RateLimitError
if errors.As(err, &rateLimitErr) {
    fmt.Printf("Rate limited. Retry after: %v\n", rateLimitErr.RateLimitInfo.RetryAfter)
}
```

## Extension Points

### Adding a New Provider

To add a new provider (e.g., Anthropic):

1. Create `pkg/providers/anthropic/` directory
2. Implement `Provider` interface in `client.go`
3. Create `types.go` to map Anthropic wire format to SDK types
4. Add provider-specific config in `config.go`
5. Register with SDK via constructor (e.g., `anthropic.New(config)`)

**No changes required** to:
- Shared types (`pkg/aisdk/`)
- Middleware (`pkg/middleware/`)
- Internal utilities (`internal/`)

### Custom Middleware

Users can wrap the HTTP client with custom middleware:

```go
type CustomMiddleware struct {
    next http.RoundTripper
}

func (m *CustomMiddleware) RoundTrip(req *http.Request) (*http.Response, error) {
    // Custom logic before request
    resp, err := m.next.RoundTrip(req)
    // Custom logic after response
    return resp, err
}
```

Then inject via `ClientConfig`:

```go
config := aisdk.ClientConfig{
    HTTPClient: &http.Client{
        Transport: &CustomMiddleware{next: http.DefaultTransport},
    },
}
```

## Performance Considerations

### Connection Pooling

From `research.md`:

> HTTP client wrapping http.Client with connection pooling config (MaxIdleConns=100, IdleConnTimeout=90s)

Benefits:
- Reuses TCP connections across requests
- Reduces TLS handshake overhead
- Supports 100 concurrent requests without degradation

### Zero-Allocation Paths

Hot paths aim for zero allocations:
- Request/response marshal/unmarshal reuses buffers where possible
- StreamEvent parsing avoids intermediate string copies
- Error wrapping uses `fmt.Errorf` (stack-allocated format strings)

Benchmark targets (from plan.md):
- 100 concurrent requests without degradation beyond network I/O
- <2 second time-to-first-chunk for streaming (p95)
- <100ms context cancellation response time

### Memory Safety

Streaming goroutines are bounded:
- `StreamReader.Close()` terminates read loop
- `defer Close()` pattern ensures cleanup
- Context cancellation breaks loop immediately

## Testing Strategy

### Unit Tests (`tests/unit/`)
- Request/response marshal/unmarshal
- Schema converter (struct → JSON Schema)
- Error wrapping and classification
- Rate limit header parsing

### Integration Tests (`tests/integration/`)
- `httptest.Server` mocking OpenAI endpoints
- Success, error, rate limit scenarios
- SSE streaming with chunk/error events
- Concurrent request handling

### E2E Tests (`tests/e2e/`)
- Full retry flow (429 → backoff → success)
- Context cancellation propagation
- Multi-turn conversations
- Streaming with mid-stream errors

## Documentation References

All architecture decisions trace to governing documents:

- **Constitution Principles**: `../specs/001-openai-responses-integration/plan.md`
- **Research Decisions**: `../specs/001-openai-responses-integration/research.md`
- **Type Definitions**: `../specs/001-openai-responses-integration/data-model.md`
- **OpenAI Documentation**: `docs/providers/openai.md`

Per Principle VI (Doc-Led Implementation):
> All type definitions cite docs/providers/openai.md line ranges in godoc comments
