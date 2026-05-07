package provider

import "context"

// Message is a single turn in a conversation.
type Message struct {
	Role    string // "user" or "assistant"
	Content string
}

// Provider is the interface all LLM backends must implement.
type Provider interface {
	// Send performs a single non-streaming request and returns the full response.
	Send(ctx context.Context, systemPrompt string, messages []Message) (string, error)
}
