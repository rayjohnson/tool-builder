package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"slices"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/toolrunner"
	"github.com/rayjohnson/tool-builder/internal/config"
)

type shellInput struct {
	Args []string `json:"args" jsonschema:"required,description=Arguments to pass to the command"`
}

// buildShellTools creates one BetaTool per allowed shell command in the config.
func buildShellTools(cfg *config.Config, workDir string) ([]anthropic.BetaTool, error) {
	if cfg.ToolUse == nil || !cfg.ToolUse.Enabled {
		return nil, nil
	}

	tools := make([]anthropic.BetaTool, 0, len(cfg.ToolUse.Shell))
	for _, sh := range cfg.ToolUse.Shell {
		if _, err := exec.LookPath(sh.Command); err != nil {
			return nil, fmt.Errorf("required command %q not found in PATH — install it before running this tool", sh.Command)
		}
		name := shellToolName(sh.Command)
		desc := fmt.Sprintf(
			"Run: %s %s\nOnly these subcommands/args are permitted: %s",
			sh.Command,
			strings.Join(sh.Args, ", "),
			strings.Join(sh.Args, ", "),
		)

		tool, err := toolrunner.NewBetaToolFromJSONSchema(
			name,
			desc,
			func(ctx context.Context, input shellInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
				return execShell(ctx, sh, input.Args, workDir)
			},
		)
		if err != nil {
			return nil, fmt.Errorf("building shell tool %q: %w", sh.Command, err)
		}
		tools = append(tools, tool)
	}
	return tools, nil
}

// shellToolName converts a command name to a valid tool name.
// e.g. "golangci-lint" → "run_golangci_lint"
func shellToolName(command string) string {
	safe := strings.NewReplacer("-", "_", " ", "_", "/", "_").Replace(command)
	return "run_" + safe
}

func execShell(
	ctx context.Context,
	sh config.ShellTool,
	args []string,
	workDir string,
) (anthropic.BetaToolResultBlockParamContentUnion, error) {
	// Validate: first arg must be in the allowed list.
	if err := validateArgs(sh, args); err != nil {
		return errResult(err.Error())
	}

	cmd := exec.CommandContext(ctx, sh.Command, args...) //nolint:gosec // command and first arg are allowlisted
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()

	var result strings.Builder
	if stdout.Len() > 0 {
		result.WriteString(stdout.String())
	}
	if stderr.Len() > 0 {
		if result.Len() > 0 {
			result.WriteString("\n--- stderr ---\n")
		}
		result.WriteString(stderr.String())
	}
	if runErr != nil {
		fmt.Fprintf(&result, "\n--- exit error: %v ---\n", runErr)
	}
	if result.Len() == 0 {
		result.WriteString("(no output)")
	}

	return anthropic.BetaToolResultBlockParamContentUnion{
		OfText: &anthropic.BetaTextBlockParam{Text: result.String()},
	}, nil
}

// validateArgs checks that args[0] (the subcommand) is in the allowlist.
func validateArgs(sh config.ShellTool, args []string) error {
	if len(sh.Args) == 0 {
		return nil // no restriction on subcommands
	}
	if len(args) == 0 {
		return fmt.Errorf("%s requires at least one argument (allowed: %s)",
			sh.Command, strings.Join(sh.Args, ", "))
	}
	if slices.Contains(sh.Args, args[0]) {
		return nil
	}
	return fmt.Errorf("%s %q is not allowed (permitted first args: %s)",
		sh.Command, args[0], strings.Join(sh.Args, ", "))
}
