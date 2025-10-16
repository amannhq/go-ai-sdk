package quickstart

import (
	"context"
	"fmt"
	"log"
	"time"
	
	"github.com/amannhq/go-ai-sdk/pkg/aisdk"
	"github.com/amannhq/go-ai-sdk/pkg/providers/openai"
)

func main() {
	// Initialize client from OPENAI_API_KEY environment variable
	client, err := openai.NewFromEnv()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	
	// Create request
	req := &aisdk.CreateResponseRequest{
		Model: "gpt-4o",
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
