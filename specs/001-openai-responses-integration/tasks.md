# Tasks: OpenAI Responses Integration

**Input**: Design documents from `/specs/001-openai-responses-integration/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅, quickstart.md ✅

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Initialize Go module with `go mod init github.com/amannhq/go-ai-sdk` and create basic directory structure (pkg/aisdk/, pkg/providers/, pkg/middleware/, internal/, examples/, tests/)
- [X] T002 [P] Create docs/README.md with SDK overview, design philosophy (Go-first, strongly typed, extensible), and quick links to examples
- [X] T003 [P] Create docs/architecture.md documenting provider interface pattern, middleware layering, and stdlib-only approach per research.md decisions
- [ ] T004 [P] Configure GitHub Actions workflow for `gofmt`, `go vet`, `golint`, and unit test execution on PR

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Core Types & Errors

- [X] T005 [P] Create pkg/aisdk/errors.go with APIError struct (Status int, Code string, Message string, CorrelationID string) and constructor functions
- [X] T006 [P] Create pkg/aisdk/errors.go RateLimitError extending APIError with RateLimitInfo (Limit, Remaining, ResetAt, RetryAfter fields per data-model.md)
- [X] T007 [P] Implement error wrapping helpers (WrapError, IsRetryable, IsRateLimitError) using errors.Is/As for classification

### HTTP Client Infrastructure

- [X] T008 Create internal/http/client.go with HTTPClient struct wrapping http.Client with connection pooling config (MaxIdleConns=100, IdleConnTimeout=90s per research.md)
- [X] T009 Implement internal/http/request.go DoRequest method accepting context.Context, handling timeouts (60s default), and propagating cancellation (<100ms termination per research.md)
- [X] T010 Add internal/http/headers.go extractRateLimitHeaders function parsing X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset into RateLimitInfo struct

### Middleware Layer

- [X] T011 Create pkg/middleware/retry.go with RetryConfig struct (MaxRetries=3, BaseDelay=1s, MaxDelay=60s per research.md)
- [X] T012 Implement pkg/middleware/retry.go exponentialBackoff function: `min(base * 2^attempt, 60s) + jitter` with rand.Intn for jitter calculation
- [X] T013 Implement pkg/middleware/retry.go isRetryable function classifying errors by HTTP status (429, 5xx retryable; 4xx non-retryable per research.md)
- [X] T014 [P] Create pkg/middleware/telemetry.go with TelemetryHooks interface (OnRequestStart, OnRetry, OnResponse, OnError callbacks)
- [X] T015 [P] Implement pkg/middleware/telemetry.go correlation ID injection via context.Context values (generate UUID v4 per request)

### Provider Interface

- [X] T016 Create pkg/providers/provider.go with Provider interface: `CreateResponse(ctx, req) (Response, error)` and `StreamResponse(ctx, req) (StreamReader, error)` per plan.md
- [X] T017 [P] Create pkg/providers/openai/config.go with OpenAIConfig struct (APIKey string, BaseURL string default "https://api.openai.com/v1", Timeout time.Duration)
- [X] T018 [P] Implement pkg/providers/openai/config.go ValidateConfig method checking APIKey non-empty, BaseURL valid URL, Timeout positive

### Schema Conversion (for FR-011 Structured Outputs)

- [X] T019 Create internal/schema/converter.go with StructToJSONSchema function using reflect.TypeOf to extract struct fields
- [X] T020 Implement internal/schema/converter.go field processing: extract json struct tags, handle nested structs recursively, support basic types (string, int, float, bool)
- [X] T021 Add internal/schema/converter.go validation for required fields via `validate:"required"` tag, generate JSON Schema with "required" array

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Developer Onboarding with First API Call (Priority: P1) 🎯 MVP

**Goal**: Enable developers to generate text within 5 minutes via simple client initialization and single method call

**Independent Test**: Install SDK, set OPENAI_API_KEY env var, call client.CreateResponse with text prompt, verify Response.OutputText contains generated text

### Core Client Implementation (US1)

- [X] T022 [P] [US1] Create pkg/aisdk/config.go with ClientConfig struct (APIKey, BaseURL, Timeout, MaxRetries, Logger fields per data-model.md)
- [X] T023 [P] [US1] Implement pkg/aisdk/config.go NewFromEnv constructor reading OPENAI_API_KEY env var, returning error if missing per FR-013
- [X] T024 [US1] Create pkg/aisdk/client.go with Client struct holding ClientConfig, HTTPClient, Provider, TelemetryHooks
- [X] T025 [US1] Implement pkg/aisdk/client.go New(config) constructor validating config, initializing HTTPClient with retry middleware, creating OpenAI provider

### Request Types (US1)

- [X] T026 [P] [US1] Create pkg/aisdk/request.go with CreateResponseRequest struct (Model, Input, Instructions, Temperature, MaxTokens, Stream, TextFormat fields per data-model.md)
- [X] T027 [P] [US1] Implement pkg/aisdk/request.go Validate method checking Model non-empty, Input non-empty, Temperature 0.0-2.0, MaxTokens positive per FR-003
- [X] T028 [US1] Create pkg/providers/openai/types.go with openAIRequest struct mapping CreateResponseRequest to OpenAI wire format per contracts/openai-responses-v1.json

### Response Types (US1)

- [X] T029 [P] [US1] Create pkg/aisdk/response.go with Response struct (ID, Output []OutputItem, Usage TokenUsage, RateLimitInfo per data-model.md)
- [X] T030 [P] [US1] Create pkg/aisdk/response.go OutputItem struct (ID, Type, Role, Content []ContentPart per data-model.md)
- [X] T031 [P] [US1] Create pkg/aisdk/response.go ContentPart struct (Type, Text, Annotations per data-model.md)
- [X] T032 [US1] Implement pkg/aisdk/response.go OutputText convenience method aggregating all ContentPart.Text into single string per FR-012
- [X] T033 [US1] Create pkg/providers/openai/types.go with openAIResponse struct mapping OpenAI wire format to Response per contracts/openai-responses-v1.json

### OpenAI Provider (US1)

- [X] T034 [US1] Create pkg/providers/openai/client.go implementing Provider interface with openAIProvider struct holding config and HTTPClient
- [X] T035 [US1] Implement pkg/providers/openai/client.go CreateResponse method: validate request, marshal to openAIRequest JSON, POST to /v1/responses, deserialize openAIResponse
- [X] T036 [US1] Add pkg/providers/openai/auth.go addAuthHeaders function injecting "Authorization: Bearer {api_key}" header per FR-004
- [X] T037 [US1] Implement pkg/providers/openai/errors.go mapOpenAIError function converting HTTP status + OpenAI error JSON to APIError or RateLimitError per FR-005

### Integration & Polish (US1)

- [X] T038 [US1] Wire pkg/providers/openai/client.go into pkg/aisdk/client.go CreateResponse method with telemetry hooks (OnRequestStart, OnResponse, OnError)
- [X] T039 [US1] Add pkg/aisdk/client.go context cancellation handling: check ctx.Err() before request, wrap errors with context.Cause per FR-009
- [X] T040 [US1] Create examples/quickstart/main.go demonstrating NewFromEnv, CreateResponse call, OutputText extraction per quickstart.md (5 lines of code)
- [ ] T041 [US1] Add godoc comments to all pkg/aisdk/ exported types citing docs/providers/openai.md line ranges per Principle VI

**Checkpoint**: At this point, User Story 1 should be fully functional - developers can generate text with <5 min setup

---

## Phase 4: User Story 2 - Structured Output Workflows (Priority: P2)

**Goal**: Enable schema-based output extraction with automatic validation, eliminating custom retry loops for malformed JSON

**Independent Test**: Define Go struct with json tags, pass to CreateResponseRequest.TextFormat via schema converter, verify Response.Output validates against struct schema

### Schema Integration (US2)

- [ ] T042 [P] [US2] Create pkg/aisdk/schema.go with TextFormat struct (Type="json_schema", JSONSchema string, Strict=true fields per data-model.md)
- [ ] T043 [P] [US2] Add pkg/aisdk/request.go WithStructuredOutput(schema interface{}) option converting Go struct to TextFormat via internal/schema/converter.go
- [ ] T044 [US2] Update pkg/providers/openai/types.go openAIRequest to include text_format field mapping TextFormat to OpenAI wire format per contracts/openai-responses-v1.json

### Structured Response Handling (US2)

- [ ] T045 [P] [US2] Add pkg/aisdk/response.go ParsedOutput field to Response struct for storing deserialized JSON matching provided schema
- [ ] T046 [US2] Implement pkg/aisdk/response.go UnmarshalOutput(dest interface{}) method extracting OutputText, unmarshaling to dest via json.Unmarshal
- [ ] T047 [US2] Add pkg/providers/openai/errors.go handleRefusal function detecting ContentPart.Type="refusal", returning structured error with refusal message per FR-002

### Examples & Documentation (US2)

- [ ] T048 [US2] Create examples/structured/main.go defining CalendarEvent struct, calling CreateResponse with WithStructuredOutput, unmarshaling result
- [ ] T049 [US2] Update docs/README.md with structured outputs section citing docs/providers/openai.md lines 2193-4038
- [ ] T050 [US2] Add godoc example to pkg/aisdk/schema.go demonstrating struct tag → JSON Schema conversion

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - structured outputs validate automatically

---

## Phase 5: User Story 3 - Multi-Turn Conversation Management (Priority: P3)

**Goal**: Simplify conversation state tracking via SDK helper while supporting manual message array management

**Independent Test**: Create conversation session, make initial request, make follow-up referencing "it" from first turn, verify response shows context awareness

### Conversation State (US3)

- [ ] T051 [P] [US3] Create pkg/aisdk/conversation.go with Conversation struct (ID, Messages []Message, Config ConversationConfig per data-model.md)
- [ ] T052 [P] [US3] Create pkg/aisdk/conversation.go Message struct (Role enum: user/assistant/tool, Content string, ResponseID string per data-model.md)
- [ ] T053 [US3] Implement pkg/aisdk/conversation.go NewConversation(config) constructor initializing empty Messages array
- [ ] T054 [US3] Add pkg/aisdk/conversation.go AddUserMessage(text) and AddAssistantMessage(response) methods appending to Messages array

### Request Integration (US3)

- [ ] T055 [US3] Update pkg/aisdk/request.go to support Input as either string or []Message (use interface{} with type assertion)
- [ ] T056 [US3] Add pkg/aisdk/conversation.go CreateResponse(ctx, text) method: append user message, call client.CreateResponse with Messages array, append assistant response
- [ ] T057 [US3] Implement pkg/aisdk/conversation.go context window truncation: if Messages exceeds model max (check Model field), drop oldest messages keeping system prompt

### Examples & Documentation (US3)

- [ ] T058 [US3] Create examples/conversation/main.go demonstrating NewConversation, AddUserMessage, CreateResponse, follow-up turn with pronoun reference
- [ ] T059 [US3] Update docs/README.md with conversation management section citing docs/providers/openai.md lines 7095-7400
- [ ] T060 [US3] Add godoc comments to pkg/aisdk/conversation.go explaining manual vs. SDK-managed state tradeoffs

**Checkpoint**: All user stories 1-3 should now be independently functional - conversations maintain context automatically

---

## Phase 6: User Story 4 - Streaming Response Handling (Priority: P4)

**Goal**: Enable incremental output display via typed event stream, improving perceived responsiveness for long-form content

**Independent Test**: Enable streaming via CreateResponseRequest.Stream=true, capture StreamEvent deltas as they arrive, verify aggregated text matches non-streaming result

### Streaming Infrastructure (US4)

- [ ] T061 [P] [US4] Create pkg/aisdk/stream.go with StreamEvent struct (Type enum, Delta string, ItemID string, Error *APIError per data-model.md covering 20+ event types)
- [ ] T062 [P] [US4] Create pkg/aisdk/stream.go StreamReader interface with Next() (StreamEvent, error) and Close() error methods
- [ ] T063 [US4] Implement pkg/providers/openai/streaming.go openAIStreamReader struct holding bufio.Scanner with custom split function for SSE parsing per research.md
- [ ] T064 [US4] Implement pkg/providers/openai/streaming.go sseSplitFunc scanning "data: " prefixed lines, handling "data: [DONE]" terminator

### Event Parsing & Deserialization (US4)

- [ ] T065 [US4] Create pkg/providers/openai/streaming.go parseStreamEvent function deserializing SSE data JSON to StreamEvent per contracts/openai-responses-v1.json
- [ ] T066 [US4] Add pkg/providers/openai/streaming.go error event handling: detect "type":"error" events, extract error details into StreamEvent.Error
- [ ] T067 [US4] Implement pkg/providers/openai/client.go StreamResponse method: validate request with Stream=true, POST to /v1/responses, return openAIStreamReader wrapping http.Response.Body

### Client Integration (US4)

- [ ] T068 [US4] Add pkg/aisdk/client.go StreamResponse(ctx, req) method wrapping provider.StreamResponse with telemetry hooks
- [ ] T069 [US4] Implement pkg/aisdk/stream.go context cancellation: check ctx.Done() in Next() loop, call Close() immediately, return context.Canceled error per FR-009
- [ ] T070 [US4] Add pkg/aisdk/stream.go goroutine cleanup: ensure http.Response.Body.Close() called in Close() method, prevent resource leaks per research.md

### Examples & Documentation (US4)

- [ ] T071 [US4] Create examples/streaming/main.go demonstrating StreamResponse call, Next() loop aggregating deltas, context cancellation handling
- [ ] T072 [US4] Update docs/README.md with streaming section citing docs/providers/openai.md lines 7618-7751, highlighting <2s first chunk target
- [ ] T073 [US4] Add godoc example to pkg/aisdk/stream.go showing proper Close() defer pattern for resource safety

**Checkpoint**: All user stories 1-4 should now be independently functional - streaming delivers incremental output

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

### Testing (if requested)

- [ ] T074 [P] Create tests/unit/request_test.go with table-driven tests for CreateResponseRequest.Validate covering edge cases (empty model, temperature out of range)
- [ ] T075 [P] Create tests/unit/response_test.go testing Response.OutputText aggregation across multiple OutputItems and ContentParts
- [ ] T076 [P] Create tests/unit/schema_test.go testing StructToJSONSchema with nested structs, arrays, required field validation
- [ ] T077 [P] Create tests/integration/openai_test.go using httptest.Server to mock OpenAI /v1/responses endpoint with success/error/rate-limit responses
- [ ] T078 [P] Create tests/integration/streaming_test.go mocking SSE stream with chunk events, done event, error mid-stream scenarios
- [ ] T079 [P] Create tests/e2e/retry_test.go simulating 429 rate limit → retry-after → success flow, validating exponential backoff with jitter

### Performance & Benchmarking

- [ ] T080 [P] Create tests/benchmark/request_bench.go benchmarking CreateResponseRequest marshal/unmarshal for zero-allocation paths
- [ ] T081 [P] Create tests/benchmark/concurrent_bench.go testing 100 concurrent CreateResponse calls, measuring latency degradation
- [ ] T082 [P] Create tests/benchmark/streaming_bench.go measuring time-to-first-chunk for streaming responses (target <2s p95)

### Documentation & Examples

- [ ] T083 [P] Update README.md in repository root with installation instructions, quick links to examples/, badges for Go version/build status
- [ ] T084 [P] Create docs/error-handling.md explaining APIError vs. RateLimitError, error wrapping patterns, retry logic
- [ ] T085 [P] Create docs/configuration.md documenting ClientConfig fields, environment variables, default values, validation rules
- [ ] T086 [P] Update examples/quickstart/README.md with step-by-step walkthrough matching quickstart.md acceptance criteria

### Security & Compliance

- [ ] T087 Review all logging/telemetry hooks to ensure API keys never logged (redact Authorization headers)
- [ ] T088 [P] Add docs/security.md with credential storage best practices (env vars, secret management systems, rotation policies)
- [ ] T089 Validate constitution compliance checklist: confirm stdlib-only, no globals, explicit config, context-aware HTTP per Principle VI

### Final Validation

- [ ] T090 Run quickstart.md validation: time installation → first successful API call, verify <5 min completion (SC-001)
- [ ] T091 Execute all examples/: quickstart, structured, conversation, streaming - confirm all run without modification (SC-008)
- [ ] T092 Run full test suite: unit, integration, e2e, benchmarks - verify 100% pass rate
- [ ] T093 Generate coverage report: ensure core SDK types (pkg/aisdk/, pkg/providers/openai/) achieve >80% line coverage
- [ ] T094 Update CHANGELOG.md with initial release notes: supported features (4 user stories), constitution compliance, known limitations

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Phase 2 completion - No dependencies on other user stories
- **User Story 2 (Phase 4)**: Depends on Phase 2 completion - May integrate with US1 but independently testable
- **User Story 3 (Phase 5)**: Depends on Phase 2 completion - May integrate with US1 but independently testable
- **User Story 4 (Phase 6)**: Depends on Phase 2 completion - May integrate with US1 but independently testable
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1 - Onboarding)**: Foundation only - can start immediately after Phase 2
- **US2 (P2 - Structured Outputs)**: Uses US1 CreateResponse flow + adds schema conversion - shares Response types
- **US3 (P3 - Conversations)**: Extends US1 with Message array input - wraps CreateResponse method
- **US4 (P4 - Streaming)**: Parallel path to US1 - shares Request/Error types but separate response handling

### Within Each User Story

- Types before services (request/response types before client methods)
- Provider implementation before client integration
- Core functionality before examples
- Implementation before documentation

### Critical Path (MVP = US1 Only)

```
Phase 1 Setup (4 tasks)
  ↓
Phase 2 Foundational (17 tasks) ← BLOCKS EVERYTHING
  ↓
Phase 3 US1 Implementation (20 tasks)
  ↓
Phase 7 US1 Validation (quickstart, basic tests)
```

**Minimum Deliverable**: T001-T041 + T090 = Fully functional text generation within 5 minutes

### Parallel Opportunities

**Phase 1** (all can run in parallel):
- T002 (docs/README.md), T003 (docs/architecture.md), T004 (GitHub Actions) - different files

**Phase 2** (parallel groups):
- Group A: T005-T007 (errors.go) - sequential within file
- Group B: T008-T010 (internal/http/*) - sequential within package
- Group C: T011-T013 (middleware/retry.go) - sequential within file
- Group D: T014-T015 (middleware/telemetry.go) - sequential within file
- Group E: T016-T018 (providers/*) - sequential within package
- Group F: T019-T021 (internal/schema/*) - sequential within package
- **Groups A-F can all run in parallel** (different packages)

**Phase 3 US1** (parallel groups):
- Group A: T022-T023 (config.go) + T026-T027 (request.go) + T029-T031 (response.go types) - different files, parallel
- Group B: T024-T025 (client.go) - depends on Group A completion
- Group C: T028 (openai/types.go request), T033 (openai/types.go response) - sequential within file
- Group D: T034-T037 (openai provider) - depends on Groups A+C
- Group E: T038-T041 (integration + examples) - depends on Group D

**Phase 4 US2** (parallel tasks):
- T042-T043 (schema.go), T044 (openai/types.go), T045-T047 (response.go extensions) - all parallel after US1 complete

**Phase 5 US3** (parallel tasks):
- T051-T053 (conversation.go types), T055 (request.go update) - parallel

**Phase 6 US4** (parallel tasks):
- T061-T062 (stream.go types), T063-T064 (openai/streaming.go) - parallel groups

**Phase 7** (highly parallel):
- All tests (T074-T079) can run in parallel (different files)
- All benchmarks (T080-T082) can run in parallel
- All docs (T083-T086) can run in parallel
- Security review (T087-T089) can run in parallel

---

## Parallel Example: Foundational Phase

```bash
# After Phase 1 completes, launch all Foundational groups in parallel:

# Terminal 1 - Errors
Task T005: Create pkg/aisdk/errors.go with APIError struct
Task T006: Add RateLimitError to same file
Task T007: Add error helpers to same file

# Terminal 2 - HTTP Client
Task T008: Create internal/http/client.go
Task T009: Add DoRequest to same file
Task T010: Create internal/http/headers.go

# Terminal 3 - Retry Middleware
Task T011: Create pkg/middleware/retry.go with RetryConfig
Task T012: Add exponentialBackoff function
Task T013: Add isRetryable function

# Terminal 4 - Telemetry Middleware
Task T014: Create pkg/middleware/telemetry.go with TelemetryHooks
Task T015: Add correlation ID injection

# Terminal 5 - Provider Interface
Task T016: Create pkg/providers/provider.go
Task T017: Create pkg/providers/openai/config.go
Task T018: Add ValidateConfig method

# Terminal 6 - Schema Converter
Task T019: Create internal/schema/converter.go
Task T020: Add field processing logic
Task T021: Add validation for required fields

# All 6 terminals complete → Foundation checkpoint reached
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (4 tasks, ~2 hours)
2. Complete Phase 2: Foundational (17 tasks, ~1 day with parallelization)
3. Complete Phase 3: User Story 1 (20 tasks, ~2 days)
4. **STOP and VALIDATE**: Run T090 quickstart validation, verify <5 min onboarding
5. Deploy/demo MVP: Developers can generate text with typed responses

**Estimated MVP Time**: 3-4 days with parallel execution

### Incremental Delivery

1. **Foundation (Phases 1-2)** → All infrastructure ready
2. **+US1 (Phase 3)** → MVP: Text generation working (deploy/demo checkpoint)
3. **+US2 (Phase 4)** → Structured outputs (deploy/demo checkpoint)
4. **+US3 (Phase 5)** → Conversations (deploy/demo checkpoint)
5. **+US4 (Phase 6)** → Streaming (deploy/demo checkpoint)
6. **Polish (Phase 7)** → Production-ready release

Each checkpoint delivers independently testable value.

### Parallel Team Strategy

With 3+ developers:

1. **Week 1**: All devs complete Phases 1-2 together (foundation)
2. **Week 2**:
   - Developer A: US1 (Phase 3) - MVP priority
   - Developer B: US2 (Phase 4) - starts after US1 types complete
   - Developer C: Documentation (docs/, examples/) - parallel to A/B
3. **Week 3**:
   - Developer A: US3 (Phase 5)
   - Developer B: US4 (Phase 6)
   - Developer C: Tests + benchmarks (Phase 7)
4. **Week 4**: Integration, polish, validation (all devs)

---

## Notes

- **[P] tasks**: Different files/packages, no dependencies - safe for parallel execution
- **[Story] labels**: Map tasks to user stories for traceability (US1=onboarding, US2=structured, US3=conversations, US4=streaming)
- **File path convention**: All paths use pkg/ (public API), internal/ (private utils), tests/ (all test types), examples/ (runnable demos)
- **Constitution compliance**: Every task respects stdlib-only constraint, explicit configuration, context-aware patterns per Principle VI
- **Documentation references**: Types cite docs/providers/openai.md line ranges per doc-led implementation principle
- **Independent testing**: Each user story has clear acceptance test in "Independent Test" section - validates story works standalone
- **Checkpoint validation**: After each phase, verify independently before proceeding - prevents compounding technical debt
- **MVP optimization**: T001-T041 deliver core value (text generation <5 min) - prioritize if time-constrained