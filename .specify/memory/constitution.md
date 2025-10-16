<!--
Sync Impact Report
Version change: 1.0.0 → 1.1.0
Modified principles: Added VI. Doc-Led Implementation Guardrails; updated Engineering Standards and Delivery Workflow for doc citations
Added sections: OpenAI Reference Index
Removed sections: None
Templates requiring updates:
- ✅ .specify/templates/plan-template.md
- ✅ .specify/templates/spec-template.md
- ✅ .specify/templates/tasks-template.md
Follow-up TODOs: None
-->

# AISDK-Go Constitution


### I. Go-First Developer Experience
- MUST expose a cohesive, idiomatic Go API that mirrors the mental model of the Vercel AI SDK while following Go naming, context usage, and error-return conventions.
- MUST keep developer onboarding friction low: provide inline documentation, runnable examples, and defaults that work without extra scaffolding.
- MUST avoid third-party abstractions that obscure behavior; all helper layers live in-repo and are reviewable.
*Rationale: A Go-native surface preserves familiarity for Go developers and ensures the SDK stays maintainable without external dependencies.*

### II. Strongly Typed Provider Contracts
- MUST model every request, response, and error with Go structs and enums, enforcing compile-time guarantees for provider capabilities.
- MUST keep serialization logic centralized so changes propagate consistently, and validate payloads before dispatch.
- MUST document response shape expectations to guide downstream consumers and tests.
*Rationale: Rich types prevent runtime surprises and enable tooling, linting, and refactors without breaking integrators.*

### III. Extensible Provider Architecture
- MUST define a minimal, provider-agnostic interface that each foundation model implementation satisfies without copy/paste.
- MUST isolate provider-specific concerns (auth, endpoints, rate semantics) behind adapters so new providers can register with configuration only.
- MUST ensure shared middlewares (retry, instrumentation, request shaping) are reusable across providers.
*Rationale: A clean extension seam lets the SDK scale beyond OpenAI to Anthropic, Gemini, and future providers with predictable effort.*

### IV. Performance and Resilience
- MUST optimize for low allocation counts, connection reuse, and streaming-first workflows to keep latencies competitive.
- MUST provide deterministic backoff, retry, and timeout policies that callers can tune yet remain safe by default.
- SHOULD benchmark critical code paths and capture regression metrics in CI when feasible.
*Rationale: Responsive, resilient behavior builds trust and keeps production workloads stable even as usage grows.*

### V. Responsible Usage Compliance
- MUST respect each provider’s published rate limits, SDK usage guidelines, and data-handling constraints.
- MUST surface rate-limit state, telemetry hooks, and audit-friendly logs so client teams can operate responsibly.
- MUST fail fast on misconfiguration (missing keys, unsupported models) with actionable error messaging.
*Rationale: Compliance safeguards prevent service disruptions and protect both the SDK maintainers and downstream users.*

### VI. Doc-Led Implementation Guardrails
- MUST consult `docs/providers/openai.md` before designing or altering functionality and capture cited line ranges in plans, specs, and review notes.
- MUST keep the OpenAI reference index accurate whenever the upstream documentation shifts.
- MUST block delivery work if required documentation excerpts are missing or outdated until alignment is restored.
*Rationale: Tight coupling between the SDK and authoritative documentation prevents drift as OpenAI evolves APIs.*

## Engineering Standards
- Source language is Go; only the standard library and provider HTTP APIs are permitted unless an explicit governance exception is recorded.
- Configuration MUST flow through explicit structs/env bindings with validation at startup, never implicit globals.
- HTTP communication MUST use shared clients with context-aware timeouts and support for streaming responses.
- Errors MUST wrap context (operation, provider, correlation identifiers) while remaining comparable for testing.
- Testing MUST cover typed contracts (unit), end-to-end provider simulations (integration or mocked HTTP), and rate-limit handling scenarios.
- Documentation MUST include quickstarts and migration notes whenever interfaces change, and cite governing OpenAI doc line ranges for every capability touched.
- Specs and plans MUST enumerate the OpenAI documentation segments (by line range) that justify each new or updated capability.

## OpenAI Reference Index
| Capability | `docs/providers/openai.md` line range | Highlights |
| --- | --- | --- |
| Quickstart onboarding | 8-931 | Responses API basics, SDK setup, first call walkthrough |
| Text generation | 934-1344 | Prompting patterns, structured outputs overview, sample code |
| Images and vision | 1347-1910 | Multimodal inputs, image and file analysis workflows |
| Audio and speech | 1913-2190 | Transcription, text-to-speech, streaming audio guidance |
| Structured model outputs | 2193-4038 | JSON schema enforcement, helpers, advanced patterns |
| Function calling | 4041-5094 | Tool definitions, invocation lifecycle, parallel calls |
| Using tools (built-ins) | 5097-5515 | Built-in tool usage, safety considerations |
| Connectors and MCP | 5517-6499 | Remote MCP integration, connector auth, security practices |
| Web search deep dive | 6501-6837 | Tool behavior, agentic search, limitations, pricing |
| Code Interpreter | 6839-7093 | Python tool usage, containers, file management |
| Conversation state | 7095-7400 | Conversations API, context chaining, window management |
| Background mode | 7403-7616 | Async execution lifecycle, polling, cancellation |
| Streaming API responses | 7618-7751 | Server-sent events usage, client examples, moderation notes |
| Webhooks | 7753-7990 | Event subscription, signature verification, sample servers |
| File inputs | 7992-8364 | PDF ingestion, URL/Base64 handling, model context behavior |

## Delivery Workflow
- Before Phase 0 planning, confirm the feature plan documents how it honors each Core Principle, especially typed contracts, extensibility, and compliance.
- Specifications MUST enumerate typed models, rate-limit expectations, and provider capabilities per user story.
- Tasks MUST separate shared infrastructure work from provider-specific implementations so additional providers can land incrementally.
- Code review checklists MUST include verification of type safety, DRY compliance, and adherence to the performance and compliance guardrails above.

- This constitution supersedes other contributing guidance when conflicts arise; teams may add stricter local rules but not weaker ones.
- Amendments require: draft rationale, mapping of impacted principles/sections, consensus from project maintainers, and updates to dependent templates before merge.
- Versioning follows semantic rules: MAJOR for breaking governance changes, MINOR for new principles/sections, PATCH for clarifications.
- Every release cycle MUST include a compliance review confirming principles, engineering standards, and workflow rules remain satisfied.
- Governance exceptions MUST be documented with expiry dates and reviewed quarterly.
- Any amendment that changes OpenAI documentation coverage MUST update the reference index and dependent templates in the same change set.

**Version**: 1.1.0 | **Ratified**: 2025-10-16 | **Last Amended**: 2025-10-16


