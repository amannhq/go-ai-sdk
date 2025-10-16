package openai

import (
	"net/http"
)

// addAuthHeaders adds OpenAI authentication headers to the request.
// Reference: FR-004 (authentication)
func addAuthHeaders(req *http.Request, apiKey string) {
	req.Header.Set("Authorization", "Bearer "+apiKey)
}
