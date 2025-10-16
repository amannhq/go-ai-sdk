# Implementation Plan: OpenAI Responses Integration

**Branch**: `001-openai-responses-integration` | **Date**: 2025-10-16 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-openai-responses-integration/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build a Go-first SDK for OpenAI's Responses API that delivers typed contracts, extensible provider architecture, and production-grade resilience. Core requirement: enable developers to generate text, structured outputs, and streaming responses within 5 minutes of installation while maintaining compile-time safety and automatic retry/rate-limit handling. Technical approach centers on a minimal provider interface, shared HTTP middleware for auth/retry/observability, and Go struct tags for schema-to-JSON-Schema translation, ensuring future providers (Anthropic, Gemini) integrate without refactoring shared infrastructure.

## Technical Context

**Language/Version**: Go 1.21+ (for generics support in typed responses and context patterns)  
**Primary Dependencies**: Go standard library only (`net/http`, `encoding/json`, `context`, `time`, `io`); SSE parsing may require lightweight stdlib-compatible helper  
**Storage**: N/A (stateless SDK; conversation state tracked in-memory or caller-managed)  
**Testing**: Go testing stdlib (`testing`, `net/http/httptest` for mocked servers), table-driven tests for typed contracts, benchmark suite for allocation tracking  
**Target Platform**: Cross-platform (Linux, macOS, Windows); server-side and CLI applications; must support CGO-free builds for portability  
**Project Type**: Single library package with multi-provider support (initial focus: OpenAI; architecture supports Anthropic, Gemini extensions)  
**Performance Goals**: 
  - 100 concurrent requests without degradation beyond network I/O
  - <2 second time-to-first-chunk for streaming (p95)
  - <100ms context cancellation response time
  - Zero-allocation paths for hot request/response serialization where feasible  
**Constraints**: 
  - Standard library only (constitution compliance; governance exception required for any dependency)
  - Typed contracts must cover all OpenAI response shapes (output items, content parts, stream events)
  - Rate limit state must be thread-safe for concurrent client usage
  - Memory-safe streaming with bounded goroutine lifecycle  
**Scale/Scope**: 
  - Initial scope: ~3,000 LOC for core SDK (types, client, retry middleware, streaming)
  - Support 4 user scenarios (onboarding, structured outputs, conversations, streaming)
  - 13 functional requirements mapped to constitution principles
  - Extensibility target: add new provider with <500 LOC adapter, zero shared code changes

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Go-First Developer Experience ✅
- **API Design**: Client initialized with `openai.New(config)` or `openai.NewFromEnv()`; request methods follow `client.CreateResponse(ctx, request)` pattern mirroring stdlib conventions
- **Onboarding**: Zero-config defaults (env-based API key, production-safe timeouts); first call within 5 lines of code
- **Idiomatic Go**: Context-first signatures, error-return tuples, options pattern for configuration, typed enums via `iota` or string constants
- **Documentation**: Inline godoc for all exported types; runnable examples in package docs; quickstart in README

### Principle II: Strongly Typed Provider Contracts ✅
- **Request Types**: `CreateResponseRequest` struct with fields matching OpenAI spec (Model, Input, Instructions, Temperature, MaxTokens, Stream, TextFormat); validation via `Validate() error` method
- **Response Types**: `Response` struct with Output []OutputItem, Usage TokenUsage, ID string; OutputItem/ContentPart/StreamEvent types mirror OpenAI schema
- **Error Contracts**: `APIError` type wrapping HTTP status, OpenAI error code, message; `RateLimitError` extending with retry-after metadata
- **Serialization**: Centralized `MarshalRequest`/`UnmarshalResponse` funcs; JSON struct tags define wire format; unknown fields handled via `json.RawMessage` for forward compatibility

### Principle III: Extensible Provider Architecture ✅
- **Provider Interface**: `type Provider interface { CreateResponse(ctx, req) (resp, err); StreamResponse(ctx, req) (stream, err) }` – minimal contract
- **OpenAI Adapter**: Implements Provider; handles OpenAI-specific auth (Bearer token), endpoint routing (`/v1/responses`), error mapping
- **Shared Middlewares**: Retry logic via `RetryableHTTPClient` wrapper; rate limit tracking via `RateLimitMiddleware`; telemetry hooks via `ObservabilityMiddleware` – all provider-agnostic
- **Extension Path**: Future Anthropic provider implements same interface; registers with different config (base URL, auth scheme); reuses all middleware

### Principles IV & V: Performance, Resilience, Compliance ✅
- **Performance**: Connection pooling via shared `http.Client`; streaming uses `io.Reader` + line-buffered parsing to avoid buffering full responses; benchmark targets: 100 concurrent, <2s first chunk
- **Retry/Backoff**: Exponential backoff with jitter for 429/5xx; max 3 retries default; respects `Retry-After` header; context deadline honored
- **Rate Limits**: Extract `X-RateLimit-*` headers into `RateLimitInfo` attached to responses/errors; surface to caller; used by retry middleware
- **Telemetry**: Hooks for `OnRequestStart`, `OnRetry`, `OnResponse`, `OnError` accepting func callbacks; include correlation IDs via `context.Context` values; log level configurable

### Principle VI: Doc-Led Implementation Guardrails ✅
- **Documentation References**: All type definitions cite `docs/providers/openai.md` line ranges in godoc comments
  - Request types → lines 8-931, 934-1344
  - Structured outputs → lines 2193-4038
  - Streaming events → lines 7618-7751
  - Conversation state → lines 7095-7400
- **Reference Index Compliance**: Plan phases explicitly call out which capabilities (from OpenAI Reference Index) are implemented in each phase

### Engineering Standards Compliance ✅
- **Standard Library Only**: No external dependencies; SSE parsing implemented using `bufio.Scanner` + custom split func if needed
- **Explicit Configuration**: `ClientConfig` struct with validation; no globals; env binding via explicit `NewFromEnv()` constructor
- **Context-Aware HTTP**: All requests accept `context.Context`; timeouts/cancellation propagate immediately
- **Wrapped Errors**: Errors include operation, provider, correlation ID via `fmt.Errorf` wrapping; comparable via `errors.Is`/`As`
- **Testing Coverage**: Unit tests for types (marshal/unmarshal); integration tests with `httptest.Server` mocking OpenAI; E2E tests for rate limit scenarios

**GATE RESULT**: ✅ **PASSED** – All principles satisfied; no governance exceptions required; proceed to Phase 0 research

## Project Structure

### Documentation (this feature)

```
specs/001-openai-responses-integration/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── openai-responses-v1.json  # OpenAPI schema for request/response types
├── checklists/
│   └── requirements.md  # Quality validation (already created)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
# Go SDK structure (single library package)
pkg/
├── aisdk/               # Main SDK package (public API surface)
│   ├── client.go        # Client initialization, config validation
│   ├── request.go       # CreateResponseRequest type + validation
│   ├── response.go      # Response, OutputItem, ContentPart types
│   ├── stream.go        # StreamEvent types, streaming client
│   ├── errors.go        # APIError, RateLimitError, error wrapping
│   └── options.go       # Functional options for client config
├── providers/           # Provider interface + implementations
│   ├── provider.go      # Provider interface definition
│   └── openai/          # OpenAI-specific adapter
│       ├── client.go    # OpenAI provider implementation
│       ├── types.go     # OpenAI-specific request/response mappings
│       ├── auth.go      # Bearer token auth logic
│       └── streaming.go # SSE parsing for OpenAI streams
├── middleware/          # Shared HTTP middleware (provider-agnostic)
│   ├── retry.go         # Exponential backoff, transient error detection
│   ├── ratelimit.go     # Rate limit header extraction, state tracking
│   └── telemetry.go     # Observability hooks, correlation IDs
└── internal/            # Internal utilities (not exported)
    ├── http/            # HTTP client wrapper, connection pooling
    ├── schema/          # Go struct → JSON Schema converter (FR-011)
    └── validation/      # Request validation helpers

examples/
├── quickstart/          # Simple text generation (P1 user story)
├── structured/          # Schema-based extraction (P2 user story)
├── conversation/        # Multi-turn context (P3 user story)
└── streaming/           # Incremental output (P4 user story)

tests/
├── unit/                # Type marshaling, validation logic, schema conversion
├── integration/         # httptest.Server mocking OpenAI endpoints
└── e2e/                 # Rate limit scenarios, concurrent requests, cancellation

docs/
├── providers/
│   └── openai.md        # OpenAI documentation reference (already exists)
└── architecture/        # Design decisions from research.md
```

**Structure Decision**: Single Go library package following stdlib conventions. Core SDK types (`pkg/aisdk/`) expose high-level API; provider implementations (`pkg/providers/openai/`) adapt to OpenAI specifics; shared middleware (`pkg/middleware/`) enables reuse across future providers. Examples demonstrate all user stories with runnable code. Tests organized by scope (unit/integration/e2e) per constitution testing standards.

## Complexity Tracking

*No violations detected – all constitution principles satisfied without exceptions.*

This section intentionally left minimal per template guidance. All design decisions align with constitution principles:
- Standard library only (no dependency exceptions)
- Single library package structure (no multi-project complexity)
- Minimal provider interface (no abstraction over-engineering)
- Direct HTTP client usage (no framework dependencies)

---

## Phase Completion Summary

### Phase 0: Research & Architecture ✅ COMPLETE

**Artifacts Generated**:
- `research.md`: 7 research questions resolved with standard-library solutions

**Key Decisions**:
1. **SSE Parsing**: `bufio.Scanner` with custom split function for streaming events
2. **Schema Conversion**: Reflection-based Go struct → JSON Schema using `reflect` package
3. **Rate Limit Sync**: `sync.RWMutex` for thread-safe rate limit state tracking
4. **Context Cancellation**: Standard `context.Context` propagation with <100ms termination
5. **Exponential Backoff**: `min(base * 2^attempt, 60s) + jitter` formula with 3 max retries
6. **Error Classification**: HTTP status + OpenAI error codes determine retryability
7. **Conversation State**: Hybrid SDK helper + manual option for flexibility

**Governance Compliance**: Zero external dependencies required; all solutions use Go standard library

---

### Phase 1: Design & Contracts ✅ COMPLETE

**Artifacts Generated**:
- `data-model.md`: 7 key entity definitions with Go types, validation rules, relationships
- `contracts/openai-responses-v1.json`: OpenAPI 3.1 schema for request/response contracts
- `quickstart.md`: P1 user story walkthrough with runnable examples
- `.github/copilot-instructions.md`: Updated with Go 1.21+ and stdlib-only technology choices

**Entity Coverage**:
1. **ClientConfig**: Initialization options with validation (FR-013)
2. **Request**: CreateResponseRequest with 13 parameters (FR-001)
3. **Response**: Complete output structure with OutputItem array (FR-002)
4. **OutputItem**: Single response item with content parts
5. **ContentPart**: Text/refusal fragments with annotations
6. **RateLimitInfo**: Rate limit state from headers (FR-007)
7. **StreamEvent**: Streaming event types (20+ event kinds) (FR-008)

**Contract Validation**:
- All 13 functional requirements mapped to types
- Documentation references cite `docs/providers/openai.md` line ranges
- OpenAPI schema validates request/response shapes
- Quickstart demonstrates P1 user story (<5 min time-to-value)

**Constitution Re-Check**: ✅ **PASSED** post-design
- All principles remain satisfied
- Type definitions enforce compile-time safety (Principle II)
- Extensibility demonstrated through provider interface (Principle III)
- Performance considerations embedded in design (Principle IV)

---

## Next Steps

### Phase 2: Implementation Tasks ⏭️ PENDING

**Command**: Run `/speckit.tasks` to generate `tasks.md`

**Expected Deliverables**:
- Detailed implementation tasks for each Go package
- Test plan mapping to 13 functional requirements
- Dependency order (e.g., types → client → middleware → provider)
- Effort estimates and acceptance criteria
- Risk mitigation strategies

**Task Breakdown Preview**:
1. **Core Types** (pkg/aisdk/): Request, Response, errors, validation
2. **HTTP Client** (internal/http/): Connection pooling, context propagation
3. **Retry Middleware** (middleware/retry.go): Exponential backoff, error classification
4. **Rate Limit Middleware** (middleware/ratelimit.go): Header parsing, state tracking
5. **OpenAI Provider** (providers/openai/): Auth, endpoint routing, SSE parsing
6. **Schema Converter** (internal/schema/): Reflection-based JSON Schema generation
7. **Streaming Client** (pkg/aisdk/stream.go): Event parsing, goroutine lifecycle
8. **Examples**: 4 runnable examples for user stories P1-P4
9. **Tests**: Unit, integration, E2E test suites
10. **Documentation**: Godoc, README, architecture diagrams

### Success Verification

Implementation complete when:
- [ ] All 13 functional requirements (FR-001 through FR-013) implemented
- [ ] All 10 success criteria (SC-001 through SC-010) validated
- [ ] 4 user stories (P1-P4) demonstrated in examples
- [ ] 6 edge cases handled in tests
- [ ] Constitution compliance checklist passed
- [ ] Zero external dependencies (stdlib only)
- [ ] All tests passing (unit, integration, E2E)
- [ ] Benchmarks meet performance targets (100 concurrent, <2s first chunk)

---

## Implementation Plan Status

| Phase | Status | Artifacts | Gate Result |
|-------|--------|-----------|-------------|
| Phase 0: Research | ✅ Complete | research.md | All decisions stdlib-compatible |
| Phase 1: Design | ✅ Complete | data-model.md, contracts/, quickstart.md | Constitution re-check passed |
| Phase 2: Tasks | ⏭️ Pending | tasks.md (awaiting /speckit.tasks) | N/A - next step |

**Current Branch**: `001-openai-responses-integration`  
**Plan Path**: `/Users/amannhq/Desktop/Me/go-ai-sdk/specs/001-openai-responses-integration/plan.md`  
**Spec Path**: `/Users/amannhq/Desktop/Me/go-ai-sdk/specs/001-openai-responses-integration/spec.md`

**Readiness**: ✅ **READY FOR PHASE 2** - All planning complete; proceed with `/speckit.tasks` to generate implementation tasks

