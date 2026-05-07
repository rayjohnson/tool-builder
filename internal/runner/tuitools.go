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
		var (
			t   anthropic.BetaTool
			err error
		)
		switch name {
		case "list_select":
			t, err = buildListSelectTool(mu)
		case "confirm":
			t, err = buildConfirmTool(mu)
		case "text_input":
			t, err = buildTextInputTool(mu)
		case "text_editor":
			t, err = buildTextEditorTool(mu)
		case "show_diff":
			t, err = buildShowDiffTool(mu)
		default:
			return nil, fmt.Errorf("unknown TUI tool %q", name)
		}
		if err != nil {
			return nil, fmt.Errorf("building %s tool: %w", name, err)
		}
		tools = append(tools, t)
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

func buildConfirmTool(mu *sync.Mutex) (anthropic.BetaTool, error) {
	type input struct {
		Question   string `json:"question"    jsonschema:"required,description=The yes/no question to ask the user"`
		DefaultYes bool   `json:"default_yes" jsonschema:"description=If true pressing Enter defaults to yes; default false (Enter = no)"`
	}
	return toolrunner.NewBetaToolFromJSONSchema(
		"confirm",
		"Ask the user a yes/no question and return their answer. "+
			"Use this before any irreversible or significant action — staging files, running git commit, sending a message, deleting data. "+
			"A single keypress is sufficient; no Enter needed. "+
			"Returns \"yes\" or \"no\".",
		func(_ context.Context, in input) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			mu.Lock()
			defer mu.Unlock()

			answer, err := tui.Confirm(in.Question, in.DefaultYes)
			if err != nil {
				return errResult(fmt.Sprintf("confirm error: %s", err))
			}
			if answer {
				return okResult("yes")
			}
			return okResult("no")
		},
	)
}

func buildTextInputTool(mu *sync.Mutex) (anthropic.BetaTool, error) {
	type input struct {
		Prompt      string `json:"prompt"      jsonschema:"required,description=The question or label shown above the text field"`
		Placeholder string `json:"placeholder" jsonschema:"description=Placeholder text shown in the empty input field"`
	}
	return toolrunner.NewBetaToolFromJSONSchema(
		"text_input",
		"Prompt the user to type a short text value and return it. "+
			"Use this when you need free-form input — a name, email address, search query, custom message, etc. "+
			"Supports cursor movement and editing. "+
			"Returns the typed text, or an empty string if the user cancelled.",
		func(_ context.Context, in input) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			mu.Lock()
			defer mu.Unlock()

			value, err := tui.Input(in.Prompt, in.Placeholder)
			if err != nil {
				return errResult(fmt.Sprintf("text_input error: %s", err))
			}
			return okResult(value)
		},
	)
}

func buildShowDiffTool(mu *sync.Mutex) (anthropic.BetaTool, error) {
	type input struct {
		OldContent string `json:"old_content" jsonschema:"required,description=Original file content"`
		NewContent string `json:"new_content" jsonschema:"required,description=Proposed new content"`
		Filename   string `json:"filename"    jsonschema:"description=Filename shown in the diff header"`
	}
	return toolrunner.NewBetaToolFromJSONSchema(
		"show_diff",
		"Show the user a scrollable colored diff between original and proposed content. "+
			"Call this before writing a file to let the user review the change. "+
			"Returns \"accept\", \"reject\", or feedback text the user typed. "+
			"If feedback is returned, revise the proposal and call show_diff again.",
		func(_ context.Context, in input) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			mu.Lock()
			defer mu.Unlock()
			result, err := tui.ShowDiff(in.Filename, in.OldContent, in.NewContent)
			if err != nil {
				return errResult(fmt.Sprintf("show_diff error: %s", err))
			}
			return okResult(result)
		},
	)
}

func buildTextEditorTool(mu *sync.Mutex) (anthropic.BetaTool, error) {
	type input struct {
		Content  string `json:"content"  jsonschema:"required,description=The initial content to open in the editor"`
		Filename string `json:"filename" jsonschema:"description=Filename hint for editor syntax highlighting (e.g. commit-msg.txt or query.sql)"`
	}
	return toolrunner.NewBetaToolFromJSONSchema(
		"text_editor",
		"Open the user's preferred text editor ($EDITOR) with an initial draft and return the edited content. "+
			"Use this for longer content the user may want to revise before it is acted on — "+
			"a commit message body, an email draft, a config snippet, a document section. "+
			"Returns the saved text after the user closes the editor.",
		func(_ context.Context, in input) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			mu.Lock()
			defer mu.Unlock()

			edited, err := tui.Edit(in.Content, in.Filename)
			if err != nil {
				return errResult(fmt.Sprintf("text_editor error: %s", err))
			}
			return okResult(edited)
		},
	)
}
