package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rayjohnson/tool-builder/internal/config"
	"github.com/rayjohnson/tool-builder/internal/runner"
	"github.com/spf13/cobra"
)

var configPath string

var runCmd = &cobra.Command{
	Use:   "run [command] [args...]",
	Short: "Run a config-defined tool",
	Long: `Run an AI-powered tool defined by a YAML config file.

If the config defines a single command named "default", tool-builder runs it directly.
If the config defines multiple commands, specify the command name as the first argument.
Any remaining arguments are passed to the tool as positional args (files to read).

Examples:
  tool-builder run --config gotest.yaml generate ./pkg/foo.go
  tool-builder run --config lint-fixer.yaml
  tool-builder run --config commit-msg.yaml "focus on the auth changes"`,
	Args: cobra.ArbitraryArgs,
	RunE: runTool,
}

func init() {
	runCmd.Flags().StringVarP(&configPath, "config", "c", "", "path to tool config YAML (required)")
	_ = runCmd.MarkFlagRequired("config")
}

func runTool(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config %q: %w", configPath, err)
	}

	if _, err := cfg.CheckEnv(); err != nil {
		return err
	}

	// Resolve command and its positional args.
	toolCmd, toolArgs, err := resolveCommand(cfg, args)
	if err != nil {
		return err
	}

	// Any trailing non-file string args become the user prompt.
	// For now treat all remaining args as file paths; the last string-only arg
	// is treated as an extra user prompt if it doesn't look like a file path.
	argFiles, userPrompt := splitArgs(toolArgs)

	configDir := filepath.Dir(configPath)
	if !filepath.IsAbs(configDir) {
		abs, err := filepath.Abs(configDir)
		if err == nil {
			configDir = abs
		}
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	r := runner.New(cfg, configDir, workDir, os.Stdin, os.Stdout)
	return r.Run(cmd.Context(), toolCmd, argFiles, userPrompt)
}

// resolveCommand picks the right command from the config given the CLI args.
func resolveCommand(cfg *config.Config, args []string) (*config.Command, []string, error) {
	// Single default command — no subcommand needed.
	if def := cfg.DefaultCommand(); def != nil {
		return def, args, nil
	}

	// Multiple commands: first arg must name one of them.
	if len(args) == 0 {
		names := make([]string, len(cfg.Commands))
		for i, c := range cfg.Commands {
			names[i] = c.Name
		}
		return nil, nil, fmt.Errorf("specify a command: %v", names)
	}

	toolCmd := cfg.Command(args[0])
	if toolCmd == nil {
		names := make([]string, len(cfg.Commands))
		for i, c := range cfg.Commands {
			names[i] = c.Name
		}
		return nil, nil, fmt.Errorf("unknown command %q — available: %v", args[0], names)
	}
	return toolCmd, args[1:], nil
}

// splitArgs separates file-path args from a trailing free-text user prompt.
// A single non-file trailing arg with no path separators is treated as a user prompt.
func splitArgs(args []string) (files []string, userPrompt string) {
	if len(args) == 0 {
		return nil, ""
	}

	last := args[len(args)-1]
	// Heuristic: if the last arg contains a path separator or looks like a file
	// (contains a dot), treat it as a file. Otherwise treat it as a user prompt.
	looksLikeFile := len(args) > 1 ||
		filepath.Base(last) != last || // contains path separator
		containsDot(last)              // has an extension

	if !looksLikeFile {
		return args[:len(args)-1], last
	}
	return args, ""
}

func containsDot(s string) bool {
	for _, c := range s {
		if c == '.' {
			return true
		}
	}
	return false
}
