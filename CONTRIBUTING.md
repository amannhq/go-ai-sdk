# Contributing to go-ai-sdk

Thank you for your interest in contributing! This document outlines our development workflow, quality standards, and how to get your contributions merged.

## Table of Contents

- [Development Workflow](#development-workflow)
- [Branch Conventions](#branch-conventions)
- [Testing Requirements](#testing-requirements)
- [Code Style](#code-style)
- [Constitution Compliance](#constitution-compliance)
- [Pull Request Process](#pull-request-process)
- [Community Guidelines](#community-guidelines)

## Development Workflow

This project follows the [SpecKit](https://github.com/speckit/speckit) structured development methodology:

1. **Specification** - Define user stories and requirements in `specs/`
2. **Planning** - Create implementation plan with architecture decisions
3. **Research** - Resolve technical unknowns and design questions
4. **Tasks** - Break down work into concrete, testable tasks
5. **Implementation** - Execute tasks with tests and documentation
6. **Validation** - Verify against success criteria

### Getting Started

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/go-ai-sdk.git
   cd go-ai-sdk
   ```

2. **Install Go 1.21 or later**
   ```bash
   go version  # Should be 1.21+
   ```

3. **Review the specification**
   - Read [specs/001-openai-responses-integration/spec.md](specs/001-openai-responses-integration/spec.md)
   - Understand the 4 user stories (onboarding, structured outputs, conversations, streaming)
   - Check [tasks.md](specs/001-openai-responses-integration/tasks.md) for available work

4. **Set up development environment**
   ```bash
   # Install development tools
   go install golang.org/x/tools/cmd/goimports@latest
   go install golang.org/x/lint/golint@latest
   
   # Set up OpenAI API key for testing
   export OPENAI_API_KEY="your-test-key"
   ```

## Branch Conventions

### Feature Branches

All work happens on feature branches following this pattern:

```
<feature-number>-<user-story>-<brief-description>
```

Examples:
- `001-us1-client-initialization` - User Story 1 (Onboarding) tasks
- `001-us2-schema-conversion` - User Story 2 (Structured Outputs) tasks
- `001-us3-conversation-state` - User Story 3 (Conversations) tasks
- `001-us4-streaming-events` - User Story 4 (Streaming) tasks

### Branch Workflow

1. Create branch from `main`:
   ```bash
   git checkout main
   git pull origin main
   git checkout -b 001-us1-request-types
   ```

2. Work on your tasks (commit frequently):
   ```bash
   git add pkg/aisdk/request.go
   git commit -m "feat(us1): add CreateResponseRequest type with validation"
   ```

3. Push and create pull request:
   ```bash
   git push origin 001-us1-request-types
   # Open PR on GitHub with reference to spec/tasks
   ```

## Testing Requirements

### Test Categories

All contributions must include appropriate tests:

1. **Unit Tests** (`tests/unit/`) - REQUIRED for all new code
   - Test individual functions and types in isolation
   - Use table-driven tests for multiple scenarios
   - Mock external dependencies (HTTP calls, time, randomness)

2. **Integration Tests** (`tests/integration/`) - REQUIRED for provider implementations
   - Use `net/http/httptest` to mock OpenAI endpoints
   - Test complete request/response flows
   - Validate error handling and retry logic

3. **E2E Tests** (`tests/e2e/`) - OPTIONAL but recommended
   - Test against real OpenAI API (requires valid key)
   - Validate rate limiting and cancellation scenarios
   - Run manually or in CI with secrets

### Running Tests

```bash
# Run all tests
go test ./...

# Run unit tests only
go test ./tests/unit/...

# Run with coverage
go test -cover ./...

# Run specific test
go test -v ./tests/unit -run TestCreateResponseRequest_Validate

# Run benchmarks
go test -bench=. ./tests/benchmark/...
```

### Writing Tests

Example unit test structure:

```go
// tests/unit/request_test.go
package unit

import (
    "testing"
    "github.com/amannhq/go-ai-sdk/pkg/aisdk"
)

func TestCreateResponseRequest_Validate(t *testing.T) {
    tests := []struct {
        name    string
        req     aisdk.CreateResponseRequest
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid request",
            req: aisdk.CreateResponseRequest{
                Model: "gpt-4o",
                Input: "test prompt",
            },
            wantErr: false,
        },
        {
            name: "missing model",
            req: aisdk.CreateResponseRequest{
                Input: "test prompt",
            },
            wantErr: true,
            errMsg:  "model is required",
        },
        // Add more test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.req.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
            if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
                t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.errMsg)
            }
        })
    }
}
```

## Code Style

### Go Conventions

Follow standard Go idioms:

- **Formatting**: Use `gofmt` (enforced in CI)
- **Imports**: Use `goimports` to organize imports
- **Naming**: Use clear, descriptive names (avoid abbreviations)
- **Errors**: Return errors, don't panic (except for initialization failures)
- **Context**: Accept `context.Context` as first parameter for I/O operations

### Documentation

All exported types and functions MUST have godoc comments:

```go
// CreateResponseRequest represents a request to generate a response from a model.
// It contains all parameters supported by the OpenAI Responses API.
//
// Required fields: Model, Input
// Optional fields: Instructions, Temperature, MaxTokens, Stream, TextFormat
//
// See docs/providers/openai.md lines 8-931 for detailed parameter descriptions.
type CreateResponseRequest struct {
    Model        string      `json:"model"`
    Input        interface{} `json:"input"`
    Instructions string      `json:"instructions,omitempty"`
    // ...
}
```

### File Organization

- One primary type per file (e.g., `request.go` for CreateResponseRequest)
- Group related helpers in the same file as the type
- Keep files under 500 lines (split into multiple files if needed)

### Error Handling

Use wrapped errors with context:

```go
if err := validateConfig(cfg); err != nil {
    return fmt.Errorf("invalid client config: %w", err)
}
```

## Constitution Compliance

All contributions MUST satisfy the six principles in [.specify/memory/constitution.md](.specify/memory/constitution.md):

### Checklist for Pull Requests

- [ ] **Principle I: Go-First DX**
  - Uses idiomatic Go patterns (context-first, error returns)
  - Provides sensible defaults (env-based API key, production timeouts)
  - Includes godoc comments and runnable examples

- [ ] **Principle II: Strongly Typed Contracts**
  - Defines request/response types with JSON struct tags
  - Validates inputs before network dispatch
  - Uses type-safe enums (not magic strings)

- [ ] **Principle III: Extensible Providers**
  - Changes to shared middleware don't require provider updates
  - Provider-specific logic stays in `pkg/providers/openai/`
  - New capabilities use Provider interface methods

- [ ] **Principle IV: Performance & Resilience**
  - Respects context cancellation (<100ms termination)
  - Implements retry logic for transient failures only
  - Cleans up goroutines and connections properly

- [ ] **Principle V: Responsible Compliance**
  - Never logs API keys or sensitive data
  - Surfaces rate limit information to callers
  - Includes correlation IDs for distributed tracing

- [ ] **Principle VI: Doc-Led Implementation**
  - Types cite `docs/providers/openai.md` line ranges in godoc
  - Implementation matches OpenAI documentation behavior
  - Examples demonstrate documented patterns

### Exceptions

If you need to violate a principle (e.g., add an external dependency):

1. Document the exception in your PR description
2. Explain why standard library is insufficient
3. Get maintainer approval before implementing

## Pull Request Process

### Before Submitting

1. **Run quality checks**:
   ```bash
   gofmt -w .
   goimports -w .
   go vet ./...
   golint ./...
   go test ./...
   ```

2. **Update documentation**:
   - Add godoc comments to new types/functions
   - Update relevant docs in `docs/` directory
   - Add examples to `examples/` if introducing new features

3. **Check task completion**:
   - Mark completed tasks in `tasks.md` with `[x]`
   - Verify task acceptance criteria met
   - Update phase checkpoints if story complete

### PR Template

Use this template for your pull request:

```markdown
## Summary
Brief description of what this PR accomplishes

Closes: [Task IDs from tasks.md, e.g., T022, T023, T024]
Related to: [User Story from spec.md, e.g., US1 - Developer Onboarding]

## Changes
- Added CreateResponseRequest type with validation
- Implemented ClientConfig with environment variable support
- Added unit tests for request validation

## Testing
- [ ] Unit tests pass (`go test ./tests/unit/...`)
- [ ] Integration tests pass (if applicable)
- [ ] Manual testing completed with real API key

## Constitution Compliance
- [x] Principle I: Go-First DX (idiomatic patterns used)
- [x] Principle II: Strongly Typed (request structs with validation)
- [x] Principle III: Extensible (no provider-specific logic in shared code)
- [x] Principle IV: Performance (context-aware, no resource leaks)
- [x] Principle V: Compliance (no credentials logged)
- [x] Principle VI: Doc-Led (godoc cites openai.md line ranges)

## Documentation
- [ ] Godoc comments added/updated
- [ ] Examples added/updated (if applicable)
- [ ] README.md updated (if public API changed)
```

### Review Process

1. **Automated Checks** - CI must pass (formatting, linting, tests)
2. **Constitution Review** - Maintainer verifies principle compliance
3. **Code Review** - At least one maintainer approval required
4. **Testing Verification** - Reviewer runs tests locally
5. **Merge** - Squash and merge with clean commit message

### Commit Message Format

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]
[optional footer]
```

Types:
- `feat` - New feature (maps to user story task)
- `fix` - Bug fix
- `docs` - Documentation only
- `test` - Test additions/updates
- `refactor` - Code restructuring (no behavior change)
- `perf` - Performance improvement
- `chore` - Maintenance (dependencies, tooling)

Examples:
```
feat(us1): add CreateResponseRequest type with validation

Implements T026-T027 from tasks.md. Includes Model, Input,
Instructions, Temperature, MaxTokens fields with JSON tags
and Validate() method checking required fields and ranges.

Refs: specs/001-openai-responses-integration/tasks.md#T026
```

## Community Guidelines

### Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on the code, not the person
- Welcome newcomers and help them succeed

### Getting Help

- **Questions**: Open a [GitHub Discussion](https://github.com/amannhq/go-ai-sdk/discussions)
- **Bugs**: Open an [Issue](https://github.com/amannhq/go-ai-sdk/issues) with reproduction steps
- **Features**: Check [spec.md](specs/001-openai-responses-integration/spec.md) first, then discuss

### Recognition

Contributors are recognized in:
- CHANGELOG.md release notes
- README.md contributors section (coming soon)
- GitHub contributor graph

---

Thank you for contributing to go-ai-sdk! Your efforts help build a better Go ecosystem for AI development. 🚀
