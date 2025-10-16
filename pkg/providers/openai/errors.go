package openai

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/amannhq/go-ai-sdk/pkg/aisdk"
)

// mapOpenAIError converts an HTTP error response to an APIError.
// Reference: FR-005 (error handling), research.md decision #6
func mapOpenAIError(resp *http.Response, correlationID string) *aisdk.APIError {
	// Try to parse OpenAI error format
	var oaiErr openAIError
	body, err := io.ReadAll(resp.Body)
	if err == nil {
		json.Unmarshal(body, &oaiErr)
	}

	code := oaiErr.Error.Code
	message := oaiErr.Error.Message

	// Fallback to status text if no error details
	if code == "" {
		code = http.StatusText(resp.StatusCode)
	}
	if message == "" {
		message = "Request failed with status " + resp.Status
	}

	return aisdk.NewAPIError(resp.StatusCode, code, message, correlationID)
}
