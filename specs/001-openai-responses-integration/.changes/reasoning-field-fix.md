# Data Model Fix: Reasoning Field Structure

**Date**: 2025-10-16  
**Type**: Schema Correction  
**Impact**: Breaking change to CreateResponseRequest type definition

## Problem

The `CreateResponseRequest` type in data-model.md incorrectly defined the reasoning field as a simple string:

```go
ReasoningEffort string `json:"reasoning,omitempty"`
```

However, the OpenAPI contract in `contracts/openai-responses-v1.json` specifies that `reasoning` should be an object with an `effort` property:

```json
"reasoning": {
  "type": "object",
  "properties": {
    "effort": {
      "type": "string",
      "enum": ["low", "medium", "high"],
      "description": "Reasoning depth for o-series models"
    }
  }
}
```

## Solution

Updated the data model to match the OpenAPI contract:

### 1. Changed Field Definition

**Before**:
```go
ReasoningEffort string `json:"reasoning,omitempty"`
```

**After**:
```go
Reasoning *ReasoningConfig `json:"reasoning,omitempty"`
```

### 2. Added ReasoningConfig Type

```go
// ReasoningConfig controls reasoning behavior for o-series models.
type ReasoningConfig struct {
    // Effort specifies reasoning depth: "low", "medium", or "high".
    Effort string `json:"effort"`
}
```

### 3. Updated Validation Logic

Added validation in `CreateResponseRequest.Validate()`:

```go
if r.Reasoning != nil {
    effort := r.Reasoning.Effort
    if effort != "low" && effort != "medium" && effort != "high" {
        return ErrInvalidReasoningEffort // "Reasoning effort must be 'low', 'medium', or 'high'"
    }
}
```

### 4. Updated Validation Rules

Added to documentation:
- `Reasoning`: If provided, Effort must be "low", "medium", or "high"

## Files Modified

- `specs/001-openai-responses-integration/data-model.md`:
  - Lines ~115: Changed field definition
  - Lines ~125: Added ReasoningConfig type
  - Lines ~145-160: Updated Validate method
  - Lines ~167: Updated validation rules documentation

## Migration Guide

When implementing this in code, developers should:

### Before (Incorrect)
```go
req := aisdk.CreateResponseRequest{
    Model: "o1-preview",
    Input: "Solve this complex problem...",
    ReasoningEffort: "high", // WRONG
}
```

### After (Correct)
```go
req := aisdk.CreateResponseRequest{
    Model: "o1-preview",
    Input: "Solve this complex problem...",
    Reasoning: &aisdk.ReasoningConfig{
        Effort: "high",
    },
}
```

### Optional Field Handling

```go
// No reasoning (uses default)
req1 := aisdk.CreateResponseRequest{
    Model: "o1-preview",
    Input: "Question",
    Reasoning: nil, // or omit entirely
}

// With reasoning
req2 := aisdk.CreateResponseRequest{
    Model: "o1-preview",
    Input: "Question",
    Reasoning: &aisdk.ReasoningConfig{
        Effort: "medium",
    },
}
```

## Impact Assessment

### Breaking Changes
- ✅ **Specification Phase**: Fixed before implementation - no code exists yet
- ✅ **Type Safety**: Pointer struct provides better nil handling
- ✅ **Validation**: Enum validation ensures only valid values accepted

### Non-Breaking
- No existing code to migrate (still in planning phase)
- No examples or tests reference this field yet
- Tasks.md doesn't mention reasoning implementation

## Verification

- [x] OpenAPI contract matches new type definition
- [x] Validation logic handles nil pointer case
- [x] Validation logic enforces enum values
- [x] Documentation updated with new validation rule
- [x] No breaking changes to examples (none exist yet)
- [x] No breaking changes to tasks (not referenced)

## Next Steps

When implementing tasks T026-T027 (CreateResponseRequest type), ensure:

1. Use `Reasoning *ReasoningConfig` field definition
2. Implement ReasoningConfig struct with Effort field
3. Add validation for enum values in Validate() method
4. Add unit tests for reasoning validation scenarios:
   - nil Reasoning (valid)
   - Valid effort values: "low", "medium", "high"
   - Invalid effort values: "", "extreme", "1", etc.

## Related References

- OpenAPI Contract: `contracts/openai-responses-v1.json` lines 166-175
- Data Model: `data-model.md` section 2 (Request)
- Implementation Tasks: T026-T027 (when executed, follow this corrected schema)
- Constitution Principle II: Strongly Typed Contracts (validates this fix)
