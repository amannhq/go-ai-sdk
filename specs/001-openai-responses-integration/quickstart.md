# Quickstart Guide: OpenAI Responses Integration

**Feature**: 001-openai-responses-integration  
**User Story**: P1 - Developer Onboarding with First API Call  
**Goal**: Generate text from OpenAI within 5 minutes of installation

## Prerequisites

- Go 1.21+ installed
- OpenAI API key (sign up at https://platform.openai.com/)

## Installation

```bash
# Once published (Phase 2+ implementation)
go get github.com/your-org/aisdk-go
```

For development/testing, clone the repository:

```bash
git clone https://github.com/your-org/aisdk-go.git
cd aisdk-go
```

## Step 1: Set Your API Key

Export your OpenAI API key as an environment variable:

```bash
export OPENAI_API_KEY="sk-your-api-key-here"
```

**Security Note**: Never hardcode API keys in source code. Use environment variables or secure configuration management.

## Step 2: Simple Text Generation

Create a file `main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/your-org/aisdk-go/pkg/aisdk"
    "github.com/your-org/aisdk-go/pkg/providers/openai"
)

func main() {
    // Initialize client from environment variable
    client, err := openai.NewFromEnv()
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    
    // Create request
    req := aisdk.CreateResponseRequest{
        Model: "gpt-5",
        Input: "Write a one-sentence bedtime story about a unicorn.",
    }
    
    // Execute with 30-second timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    resp, err := client.CreateResponse(ctx, req)
    if err != nil {
        log.Fatalf("Request failed: %v", err)
    }
    
    // Print generated text
    fmt.Println("Response:", resp.OutputText())
    fmt.Printf("Tokens used: %d\n", resp.Usage.TotalTokens)
}
```

Run the program:

```bash
go run main.go
```

**Expected Output**:
```
Response: Under the soft glow of the moon, Luna the unicorn danced through fields of twinkling stardust, leaving trails of dreams for every child asleep.
Tokens used: 42
```

## Step 3: Handling Errors

Add error handling for common scenarios:

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "time"
    
    "github.com/your-org/aisdk-go/pkg/aisdk"
    "github.com/your-org/aisdk-go/pkg/providers/openai"
)

func main() {
    client, err := openai.NewFromEnv()
    if err != nil {
        // Clear error message per SC-006
        if errors.Is(err, aisdk.ErrMissingAPIKey) {
            log.Fatal("API key required. Set OPENAI_API_KEY environment variable or provide via ClientConfig.APIKey")
        }
        log.Fatalf("Failed to create client: %v", err)
    }
    
    req := aisdk.CreateResponseRequest{
        Model: "gpt-5",
        Input: "Explain quantum computing in one sentence.",
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    resp, err := client.CreateResponse(ctx, req)
    if err != nil {
        // Typed error handling
        var apiErr *aisdk.APIError
        if errors.As(err, &apiErr) {
            fmt.Printf("API error: %s (code: %s)\n", apiErr.Message, apiErr.Code)
            
            // Check for rate limits
            var rateLimitErr *aisdk.RateLimitError
            if errors.As(err, &rateLimitErr) {
                fmt.Printf("Rate limited. Retry after: %v\n", rateLimitErr.RateLimitInfo.RetryAfter)
            }
            return
        }
        
        // Context cancellation
        if errors.Is(err, context.DeadlineExceeded) {
            log.Fatal("Request timeout - consider increasing context deadline")
        }
        
        log.Fatalf("Request failed: %v", err)
    }
    
    fmt.Println("Response:", resp.OutputText())
}
```

## Step 4: Custom Configuration

Override defaults with explicit configuration:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/your-org/aisdk-go/pkg/aisdk"
    "github.com/your-org/aisdk-go/pkg/providers/openai"
)

func main() {
    // Custom configuration
    config := aisdk.ClientConfig{
        APIKey:     "sk-your-api-key-here", // Or load from secure vault
        BaseURL:    "https://api.openai.com/v1", // Can point to proxy/mock
        Timeout:    60 * time.Second, // Longer timeout for complex requests
        MaxRetries: 5, // More retries for flaky networks
        Logger:     nil, // Set to enable telemetry hooks
    }
    
    client, err := openai.New(config)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    
    // Request with temperature control
    req := aisdk.CreateResponseRequest{
        Model:       "gpt-5",
        Input:       "List 5 creative business names for a bakery.",
        Temperature: ptr(1.2), // Higher creativity
        MaxTokens:   ptr(150), // Limit response length
    }
    
    ctx := context.Background()
    resp, err := client.CreateResponse(ctx, req)
    if err != nil {
        log.Fatalf("Request failed: %v", err)
    }
    
    fmt.Println("Response:", resp.OutputText())
}

// Helper for pointer fields
func ptr[T any](v T) *T {
    return &v
}
```

## Key Concepts

### Context Management
- Always pass `context.Context` to control timeouts and cancellation
- Use `context.WithTimeout()` for bounded execution
- Use `context.WithCancel()` for user-initiated cancellation
- SDK respects context cancellation within 100ms (SC-009)

### Error Handling
- All errors are typed for programmatic handling
- Use `errors.Is()` and `errors.As()` for error classification
- Rate limit errors include `RateLimitInfo` with retry guidance
- Validation errors occur before network requests (fail-fast)

### Rate Limits
- SDK automatically retries 429 responses with exponential backoff
- Default: 3 retries with jitter (max ~7s total wait)
- Rate limit info available on responses via `RateLimitInfo` field
- Respects `Retry-After` header when provided

### Configuration
- `NewFromEnv()`: Zero-config initialization from `OPENAI_API_KEY`
- `New(config)`: Explicit configuration with validation
- Defaults: 60s timeout, 3 max retries, no logging
- Standard library only - no external dependencies

## Next Steps

### Structured Outputs (P2 User Story)
Learn how to extract typed data with schema enforcement:
```bash
go run examples/structured/main.go
```

### Conversations (P3 User Story)
Build multi-turn dialogues with context management:
```bash
go run examples/conversation/main.go
```

### Streaming (P4 User Story)
Display incremental output as it generates:
```bash
go run examples/streaming/main.go
```

## Troubleshooting

### "API key required" Error
- Ensure `OPENAI_API_KEY` environment variable is set
- Check for typos in the key (should start with `sk-`)
- Verify key is active in OpenAI dashboard

### "Request timeout" Error
- Increase `context.WithTimeout()` duration for longer responses
- Check network connectivity to `api.openai.com`
- Consider using streaming mode for long-form content

### "Rate limit exceeded" Error
- SDK automatically retries with backoff (no action needed)
- For persistent rate limits, reduce request frequency
- Upgrade OpenAI plan for higher quotas

### "Invalid model" Error
- Verify model name is correct (e.g., `gpt-5`, not `gpt5`)
- Check model availability for your OpenAI account tier
- See OpenAI documentation for supported models

## Success Criteria Validation

This quickstart satisfies:
- **SC-001**: First API call within 5 minutes ✅
  - Steps 1-2 take <5 minutes for experienced Go developers
- **SC-006**: Actionable error messages ✅
  - All errors include remediation guidance
- **SC-007**: No external dependencies ✅
  - Standard library only (validated in Project Structure)

## Documentation References

- **Request/Response Types**: See `data-model.md` for full type definitions
- **API Contracts**: See `contracts/openai-responses-v1.json` for OpenAPI schema
- **OpenAI Documentation**: `docs/providers/openai.md` lines 8-931 (quickstart patterns)
- **Constitution Principles**: All code follows Go-First Developer Experience (Principle I)
