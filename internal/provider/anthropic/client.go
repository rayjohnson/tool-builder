package anthropic

import (
	"context"
	"fmt"
	"os"
	"strings"

	sdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/rayjohnson/tool-builder/internal/provider"
)

const defaultMaxTokens = 8096

// Client wraps the Anthropic SDK and implements provider.Provider.
type Client struct {
	inner     *sdk.Client
	model     string
	maxTokens int64
}

// New creates an Anthropic client. Requires ANTHROPIC_API_KEY to be set.
func New(model string, maxTokens int64) (*Client, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}
	if maxTokens == 0 {
		maxTokens = defaultMaxTokens
	}
	c := sdk.NewClient(option.WithAPIKey(apiKey))
	return &Client{inner: &c, model: model, maxTokens: maxTokens}, nil
}

// Send performs a single non-streaming request and returns the full response text.
func (c *Client) Send(ctx context.Context, systemPrompt string, messages []provider.Message) (string, error) {
	var msgs []sdk.MessageParam
	for _, m := range messages {
		switch m.Role {
		case "user":
			msgs = append(msgs, sdk.NewUserMessage(sdk.NewTextBlock(m.Content)))
		case "assistant":
			msgs = append(msgs, sdk.NewAssistantMessage(sdk.NewTextBlock(m.Content)))
		}
	}

	resp, err := c.inner.Messages.New(ctx, sdk.MessageNewParams{
		Model:     sdk.Model(c.model),
		MaxTokens: c.maxTokens,
		System: []sdk.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: msgs,
	})
	if err != nil {
		return "", fmt.Errorf("anthropic API error: %w", err)
	}

	var result strings.Builder
	for _, block := range resp.Content {
		if block.Type == "text" {
			result.WriteString(block.Text)
		}
	}
	return result.String(), nil
}
