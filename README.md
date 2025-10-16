# go-ai-sdk

A Go-first SDK for OpenAI's Responses API that delivers strongly typed contracts, extensible provider architecture, and production-grade resilience.

## Features

- **5-Minute Onboarding**: Generate text with just 5 lines of code
- **Strongly Typed**: Compile-time safety for all requests and responses
- **Structured Outputs**: Schema-based extraction with automatic validation
- **Conversation Management**: Multi-turn context tracking built-in
- **Streaming Support**: Incremental output for responsive UIs
- **Production Ready**: Automatic retries, rate limiting, and context cancellation
- **Zero Dependencies**: Standard library only - no external packages

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/amannhq/go-ai-sdk/pkg/aisdk"
)

func main() {
    // Initialize client from OPENAI_API_KEY environment variable
    client, err := aisdk.NewFromEnv()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create a simple text generation request
    req := aisdk.CreateResponseRequest{
        Model: "gpt-4o",
        Input: "Explain quantum computing in one sentence",
    }
    
    resp, err := client.CreateResponse(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(resp.OutputText())
}
```

## Installation

```bash
go get github.com/amannhq/go-ai-sdk
```

## Documentation

- [Architecture Overview](docs/architecture.md) - Design philosophy and provider patterns
- [Error Handling](docs/error-handling.md) - Working with APIError and RateLimitError
- [Configuration](docs/configuration.md) - ClientConfig options and environment variables
- [Security Best Practices](docs/security.md) - Credential management and compliance

## Examples

Explore working examples in the `examples/` directory:

- **[quickstart/](examples/quickstart/)** - Simple text generation (5 minutes)
- **[structured/](examples/structured/)** - Schema-based data extraction
- **[conversation/](examples/conversation/)** - Multi-turn dialogues
- **[streaming/](examples/streaming/)** - Incremental response handling

## Development Status

**Current Phase**: Planning Complete ✅

| Phase | Status | Deliverable |
|-------|--------|-------------|
| Specification | ✅ Complete | [spec.md](specs/001-openai-responses-integration/spec.md) |
| Planning | ✅ Complete | [plan.md](specs/001-openai-responses-integration/plan.md) |
| Research | ✅ Complete | [research.md](specs/001-openai-responses-integration/research.md) |
| Data Model | ✅ Complete | [data-model.md](specs/001-openai-responses-integration/data-model.md) |
| API Contracts | ✅ Complete | [contracts/](specs/001-openai-responses-integration/contracts/) |
| Tasks | ✅ Complete | [tasks.md](specs/001-openai-responses-integration/tasks.md) |
| Implementation | 🚧 In Progress | 94 tasks across 7 phases |

See [tasks.md](specs/001-openai-responses-integration/tasks.md) for detailed implementation roadmap.

## Architecture

```
pkg/
├── aisdk/              # Public API surface
│   ├── client.go       # Client initialization
│   ├── request.go      # Request types
│   ├── response.go     # Response types
│   ├── stream.go       # Streaming support
│   ├── errors.go       # Error types
│   └── conversation.go # Conversation helpers
├── providers/          # Provider implementations
│   ├── provider.go     # Provider interface
│   └── openai/         # OpenAI adapter
└── middleware/         # Shared HTTP middleware
    ├── retry.go        # Exponential backoff
    ├── ratelimit.go    # Rate limit tracking
    └── telemetry.go    # Observability hooks
```

## Design Principles

This SDK follows six core principles defined in our [constitution](/.specify/memory/constitution.md):

1. **Go-First Developer Experience** - Idiomatic Go patterns, context-first signatures
2. **Strongly Typed Contracts** - Compile-time safety for all provider interactions
3. **Extensible Providers** - Minimal interface enabling future Anthropic/Gemini support
4. **Performance & Resilience** - Automatic retries, connection pooling, bounded goroutines
5. **Responsible Compliance** - Structured telemetry, credential hygiene, rate limit awareness
6. **Doc-Led Implementation** - All types cite official OpenAI documentation

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for:

- Development workflow and branch conventions
- Testing requirements (unit, integration, e2e)
- Code style and documentation standards
- Constitution compliance checklist

## Roadmap

### v0.1.0 - MVP (Phase 3)
- ✅ Core types and client initialization
- ✅ Simple text generation
- ✅ Error handling with retries
- 🚧 User Story 1: Developer Onboarding (20 tasks)

### v0.2.0 - Structured Outputs (Phase 4)
- 🔲 Schema conversion (Go structs → JSON Schema)
- 🔲 Validated responses with refusal handling
- 🔲 User Story 2: Structured Workflows (9 tasks)

### v0.3.0 - Advanced Features (Phases 5-6)
- 🔲 Conversation state management
- 🔲 Server-sent events streaming
- 🔲 User Stories 3-4: Conversations + Streaming (23 tasks)

### v1.0.0 - Production Release (Phase 7)
- 🔲 Full test coverage (unit, integration, e2e)
- 🔲 Performance benchmarks (100 concurrent, <2s streaming)
- 🔲 Security hardening and compliance validation

## License

[Add your license here - commonly MIT or Apache 2.0]

## Support

- **Issues**: [GitHub Issues](https://github.com/amannhq/go-ai-sdk/issues)
- **Discussions**: [GitHub Discussions](https://github.com/amannhq/go-ai-sdk/discussions)
- **Documentation**: [docs/](docs/)

---

Built with ❤️ following [SpecKit](https://github.com/speckit/speckit) structured development workflow.
