package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/toolrunner"
	"github.com/rayjohnson/tool-builder/internal/config"
	"github.com/rayjohnson/tool-builder/internal/tui"
)

// buildTUITools registers interactive TUI tools declared in tool_use.tui.
// mu serializes TUI interactions with other terminal output.
func buildTUITools(cfg *config.Config, mu *sync.Mutex) ([]anthropic.BetaTool, error) {
	if cfg.ToolUse == nil {
		return nil, nil
	}
	var tools []anthropic.BetaTool
	for _, name := range cfg.ToolUse.TUI {
		switch name {
		case "list_select":
			t, err := buildListSelectTool(mu)
			if err != nil {
				return nil, fmt.Errorf("building list_select tool: %w", err)
			}
			tools = append(tools, t)
		default:
			return nil, fmt.Errorf("unknown TUI tool %q", name)
		}
	}
	return tools, nil
}

func buildListSelectTool(mu *sync.Mutex) (anthropic.BetaTool, error) {
	type input struct {
		Title  string   `json:"title"         jsonschema:"required,description=Heading shown above the list"`
		Items  []string `json:"items"         jsonschema:"required,description=The items to display"`
		Single bool     `json:"single_select" jsonschema:"description=If true the user picks exactly one item; default false (multi-select)"`
	}
	return toolrunner.NewBetaToolFromJSONSchema(
		"list_select",
		"Show an interactive selection list to the user and return their choices. "+
			"Call this whenever you have a set of options and want the user to decide which ones to act on — "+
			"files to include, people to notify, items to process, etc. "+
			"Set single_select to true when you need exactly one choice. "+
			"Returns a JSON array of the selected strings, or an empty array if the user selected nothing or cancelled.",
		func(_ context.Context, in input) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			if len(in.Items) == 0 {
				return errResult("list_select: no items provided")
			}
			mu.Lock()
			defer mu.Unlock()

			chosen, err := tui.Select(in.Title, in.Items, in.Single)
			if err != nil {
				return errResult(fmt.Sprintf("list_select error: %s", err))
			}
			if chosen == nil {
				chosen = []string{}
			}
			data, _ := json.Marshal(chosen)
			return okResult(string(data))
		},
	)
}
