package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/toolrunner"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rayjohnson/tool-builder/internal/config"
)

// buildMCPTools connects to all configured MCP servers, fetches their tool manifests,
// and returns them as Anthropic tools. The returned cleanup func closes all connections.
func buildMCPTools(ctx context.Context, cfg *config.Config) ([]anthropic.BetaTool, func(), error) {
	if cfg.ToolUse == nil || len(cfg.ToolUse.MCP) == 0 {
		return nil, func() {}, nil
	}

	var allTools []anthropic.BetaTool
	var closers []func()

	for _, srv := range cfg.ToolUse.MCP {
		tools, closer, err := connectMCPServer(ctx, srv)
		if err != nil {
			for _, c := range closers {
				c()
			}
			return nil, func() {}, fmt.Errorf("MCP server %q: %w", srv.Name, err)
		}
		allTools = append(allTools, tools...)
		closers = append(closers, closer)
	}

	return allTools, func() {
		for _, c := range closers {
			c()
		}
	}, nil
}

func connectMCPServer(ctx context.Context, srv config.MCPServer) ([]anthropic.BetaTool, func(), error) {
	var mcpClient *client.Client
	var err error

	if srv.URL != "" {
		mcpClient, err = client.NewStreamableHttpClient(srv.URL)
		if err != nil {
			return nil, nil, fmt.Errorf("creating HTTP client: %w", err)
		}
		if err = mcpClient.Start(ctx); err != nil {
			return nil, nil, fmt.Errorf("starting HTTP client: %w", err)
		}
	} else {
		mcpClient, err = client.NewStdioMCPClient(srv.Command, srv.Env, srv.Args...)
		if err != nil {
			return nil, nil, fmt.Errorf("starting subprocess %q: %w", srv.Command, err)
		}
	}

	closer := func() { _ = mcpClient.Close() }

	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{Name: "tool-builder", Version: "0.1"}
	if _, err := mcpClient.Initialize(ctx, initReq); err != nil {
		closer()
		return nil, nil, fmt.Errorf("initializing: %w", err)
	}

	toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		closer()
		return nil, nil, fmt.Errorf("listing tools: %w", err)
	}

	// Apply allowlist if configured.
	mcpTools := toolsResult.Tools
	if len(srv.Tools) > 0 {
		allowed := make(map[string]bool, len(srv.Tools))
		for _, name := range srv.Tools {
			allowed[name] = true
		}
		filtered := mcpTools[:0]
		for _, t := range mcpTools {
			if allowed[t.Name] {
				filtered = append(filtered, t)
			}
		}
		mcpTools = filtered
	}

	tools := make([]anthropic.BetaTool, 0, len(mcpTools))
	for _, mcpTool := range mcpTools {
		tool, err := wrapMCPTool(ctx, mcpClient, srv.Name, mcpTool)
		if err != nil {
			closer()
			return nil, nil, fmt.Errorf("wrapping tool %q: %w", mcpTool.Name, err)
		}
		tools = append(tools, tool)
	}

	return tools, closer, nil
}

func wrapMCPTool(
	ctx context.Context,
	mcpClient *client.Client,
	serverName string,
	mcpTool mcp.Tool,
) (anthropic.BetaTool, error) {
	// Prefix the tool name so the agent can tell which server it came from.
	prefixedName := serverName + "__" + mcpTool.Name

	// Marshal the MCP tool's input schema to bytes for the Anthropic tool definition.
	schemaBytes, err := json.Marshal(mcpTool.InputSchema)
	if err != nil {
		return nil, fmt.Errorf("marshaling input schema: %w", err)
	}

	originalName := mcpTool.Name // capture for closure

	tool, err := toolrunner.NewBetaToolFromBytes[json.RawMessage](
		prefixedName,
		mcpTool.Description,
		schemaBytes,
		func(callCtx context.Context, rawArgs json.RawMessage) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			// Unmarshal to any so mcp-go can re-marshal it correctly.
			var args any
			if len(rawArgs) > 0 && string(rawArgs) != "null" {
				if err := json.Unmarshal(rawArgs, &args); err != nil {
					return errResult(fmt.Sprintf("mcp %s: invalid arguments: %v", prefixedName, err))
				}
			}
			result, err := mcpClient.CallTool(callCtx, mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      originalName,
					Arguments: args,
				},
			})
			if err != nil {
				return errResult(fmt.Sprintf("mcp %s: call failed: %v", prefixedName, err))
			}
			return mcpResultToAnthropic(prefixedName, result)
		},
	)
	_ = ctx // ctx is captured in closure via callCtx; outer ctx only used at setup
	return tool, err
}

// mcpResultToAnthropic converts an MCP CallToolResult to an Anthropic tool result.
// Text content items are concatenated; other content types are summarised.
// If the MCP result carries IsError, the Anthropic result is an error block so the
// agent sees the failure and can report it rather than silently proceeding.
func mcpResultToAnthropic(toolName string, result *mcp.CallToolResult) (anthropic.BetaToolResultBlockParamContentUnion, error) {
	var sb strings.Builder
	for _, c := range result.Content {
		switch v := c.(type) {
		case mcp.TextContent:
			sb.WriteString(v.Text)
		default:
			// Non-text content (images, audio, embedded resources) — note it so the
			// agent knows something was returned even if it can't be rendered as text.
			fmt.Fprintf(&sb, "[non-text content from %s]", toolName)
		}
	}
	text := sb.String()
	if text == "" {
		text = "(no output)"
	}
	if result.IsError {
		return errResult(text)
	}
	return okResult(text)
}
