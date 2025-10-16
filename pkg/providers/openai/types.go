package openai

import (
	"github.com/amannhq/go-ai-sdk/pkg/aisdk"
)

// openAIRequest represents the OpenAI wire format for requests.
// Maps from aisdk.CreateResponseRequest to OpenAI API format.
// Reference: contracts/openai-responses-v1.json
type openAIRequest struct {
	Model              string      `json:"model"`
	Input              interface{} `json:"input"` // string or []Message
	Instructions       string      `json:"instructions,omitempty"`
	Temperature        *float64    `json:"temperature,omitempty"`
	MaxTokens          *int        `json:"max_tokens,omitempty"`
	Stream             bool        `json:"stream,omitempty"`
	Text               *textFormat `json:"text,omitempty"`
	PreviousResponseID string      `json:"previous_response_id,omitempty"`
	Reasoning          *reasoning  `json:"reasoning,omitempty"`
}

// textFormat represents the OpenAI text format configuration
type textFormat struct {
	Type   string                 `json:"type"`
	Name   string                 `json:"name,omitempty"`
	Schema map[string]interface{} `json:"json_schema,omitempty"`
	Strict bool                   `json:"strict,omitempty"`
}

// reasoning represents the OpenAI reasoning configuration
type reasoning struct {
	Effort string `json:"effort"`
}

// toOpenAIRequest converts aisdk.CreateResponseRequest to openAIRequest
func toOpenAIRequest(req *aisdk.CreateResponseRequest) *openAIRequest {
	oaiReq := &openAIRequest{
		Model:              req.Model,
		Input:              req.Input,
		Instructions:       req.Instructions,
		Temperature:        req.Temperature,
		MaxTokens:          req.MaxTokens,
		Stream:             req.Stream,
		PreviousResponseID: req.PreviousResponseID,
	}

	// Convert TextFormat if present
	if req.TextFormat != nil {
		oaiReq.Text = &textFormat{
			Type:   req.TextFormat.Type,
			Name:   req.TextFormat.Name,
			Schema: req.TextFormat.Schema,
			Strict: req.TextFormat.Strict,
		}
	}

	// Convert Reasoning if present
	if req.Reasoning != nil {
		oaiReq.Reasoning = &reasoning{
			Effort: req.Reasoning.Effort,
		}
	}

	return oaiReq
}

// openAIResponse represents the OpenAI wire format for responses.
// Maps from OpenAI API format to aisdk.Response.
// Reference: contracts/openai-responses-v1.json
type openAIResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Output  []openAIOutputItem `json:"output"`
	Usage   openAIUsage        `json:"usage"`
	Model   string             `json:"model"`
	Created int64              `json:"created"`
}

// openAIOutputItem represents an output item in OpenAI format
type openAIOutputItem struct {
	ID      string              `json:"id"`
	Type    string              `json:"type"`
	Role    string              `json:"role"`
	Content []openAIContentPart `json:"content"`
}

// openAIContentPart represents a content part in OpenAI format
type openAIContentPart struct {
	Type        string             `json:"type"`
	Text        string             `json:"text,omitempty"`
	Annotations []openAIAnnotation `json:"annotations,omitempty"`
	Refusal     string             `json:"refusal,omitempty"`
}

// openAIAnnotation represents an annotation in OpenAI format
type openAIAnnotation struct {
	Type       string `json:"type"`
	Text       string `json:"text"`
	StartIndex int    `json:"start_index"`
	EndIndex   int    `json:"end_index"`
}

// openAIUsage represents token usage in OpenAI format
type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// toAISDKResponse converts openAIResponse to aisdk.Response
func toAISDKResponse(oaiResp *openAIResponse) *aisdk.Response {
	resp := &aisdk.Response{
		ID:      oaiResp.ID,
		Object:  oaiResp.Object,
		Model:   oaiResp.Model,
		Created: oaiResp.Created,
		Usage: aisdk.TokenUsage{
			PromptTokens:     oaiResp.Usage.PromptTokens,
			CompletionTokens: oaiResp.Usage.CompletionTokens,
			TotalTokens:      oaiResp.Usage.TotalTokens,
		},
		Output: make([]aisdk.OutputItem, len(oaiResp.Output)),
	}

	// Convert output items
	for i, oaiItem := range oaiResp.Output {
		resp.Output[i] = aisdk.OutputItem{
			ID:      oaiItem.ID,
			Type:    oaiItem.Type,
			Role:    oaiItem.Role,
			Content: make([]aisdk.ContentPart, len(oaiItem.Content)),
		}

		// Convert content parts
		for j, oaiPart := range oaiItem.Content {
			part := aisdk.ContentPart{
				Type:    oaiPart.Type,
				Text:    oaiPart.Text,
				Refusal: oaiPart.Refusal,
			}

			// Convert annotations if present
			if len(oaiPart.Annotations) > 0 {
				part.Annotations = make([]aisdk.Annotation, len(oaiPart.Annotations))
				for k, oaiAnnot := range oaiPart.Annotations {
					part.Annotations[k] = aisdk.Annotation{
						Type:       oaiAnnot.Type,
						Text:       oaiAnnot.Text,
						StartIndex: oaiAnnot.StartIndex,
						EndIndex:   oaiAnnot.EndIndex,
					}
				}
			}

			resp.Output[i].Content[j] = part
		}
	}

	return resp
}

// openAIError represents an error response from OpenAI
type openAIError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}
