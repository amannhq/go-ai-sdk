# Go AI SDK

A Go-first SDK for OpenAI's Responses API that delivers typed contracts, extensible provider architecture, and production-grade resilience.

## Design Philosophy

### Go-First Developer Experience
- **Idiomatic Go**: Context-first signatures, error-return tuples, options pattern
- **Zero-config defaults**: Environment-based API keys, production-safe timeouts
- **Fast onboarding**: First successful API call within 5 minutes

### Strongly Typed Provider Contracts
- **Type safety**: Compile-time validation of requests and responses
- **Structured outputs**: Go struct tags → JSON Schema conversion
- **Rich error types**: APIError, RateLimitError with actionable messages

### Extensible Provider Architecture
- **Minimal interface**: Easy to add new providers (Anthropic, Gemini)
- **Shared middleware**: Retry logic, rate limiting, telemetry - provider-agnostic
- **Pluggable**: Custom HTTP clients, observability hooks

### Performance & Resilience
- **Connection pooling**: Optimized HTTP client with 100 concurrent requests support
- **Automatic retries**: Exponential backoff with jitter for transient failures
- **Rate limit handling**: Extracts rate limit state, respects Retry-After headers
- **Context cancellation**: <100ms termination response time

### Standard Library Only
- **No external dependencies**: Uses only Go stdlib (`net/http`, `encoding/json`, `context`, etc.)
- **Lightweight**: Minimal footprint, easy to audit
- **Portable**: CGO-free builds for cross-platform compatibility

## Quick Start

See [../examples/quickstart/main.go](../examples/quickstart/main.go) for a complete example.

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/amannhq/go-ai-sdk/pkg/aisdk"
    "github.com/amannhq/go-ai-sdk/pkg/providers/openai"
)

func main() {
    // Initialize from OPENAI_API_KEY environment variable
    client, err := openai.NewFromEnv()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create request
    req := aisdk.CreateResponseRequest{
        Model: "gpt-5",
        Input: "Explain quantum computing in one sentence.",
    }
    
    // Execute with context
    resp, err := client.CreateResponse(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Response:", resp.OutputText())
}
```

## Features

### User Story 1: Developer Onboarding (MVP)
- **Simple text generation**: Single method call with typed request/response
- **Environment-based config**: No hardcoded API keys
- **Error handling**: Typed errors with actionable messages
- **Examples**: [examples/quickstart/](../examples/quickstart/)

### User Story 2: Structured Outputs
- **Schema-based extraction**: Go structs → JSON Schema conversion
- **Automatic validation**: OpenAI validates output against schema
- **Type-safe unmarshaling**: Direct to Go structs
- **Examples**: [examples/structured/](../examples/structured/)

### User Story 3: Multi-Turn Conversations
- **Conversation helpers**: SDK tracks message history
- **Manual mode**: Full control over message arrays
- **Context window management**: Automatic truncation
- **Examples**: [examples/conversation/](../examples/conversation/)

### User Story 4: Streaming Responses
- **SSE parsing**: stdlib-based Server-Sent Events handling
- **Typed events**: 20+ event types with type safety
- **Context cancellation**: Immediate termination on cancel
- **Examples**: [examples/streaming/](../examples/streaming/)

## Documentation

- [Architecture](./architecture.md) - Provider interface, middleware layering
- [Error Handling](./error-handling.md) - Error types, retry logic, classification
- [Configuration](./configuration.md) - ClientConfig fields, environment variables
- [Security](./security.md) - Credential storage, best practices

## Examples

All examples are runnable with `go run`:

```bash
# Simple text generation (User Story 1)
go run examples/quickstart/main.go

# Structured outputs (User Story 2)
go run examples/structured/main.go

# Multi-turn conversations (User Story 3)
go run examples/conversation/main.go

# Streaming responses (User Story 4)
go run examples/streaming/main.go
```

## Constitution Compliance

This SDK adheres to strict engineering principles:

- **Standard library only**: No external dependencies
- **Explicit configuration**: No global state, all config validated
- **Context-aware HTTP**: All requests respect context cancellation
- **Strongly typed**: Compile-time safety for all API interactions
- **Doc-led implementation**: All types cite governing documentation

See [../specs/001-openai-responses-integration/plan.md](../specs/001-openai-responses-integration/plan.md) for full constitution compliance details.

## Contributing

This project follows the SpecKit methodology. See contribution guidelines for implementation workflow.

## License

[License details to be added]
