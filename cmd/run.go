package cmd

import (
	"context"
	"fmt"

	"github.com/rayjohnson/tool-builder/internal/config"
	"github.com/rayjohnson/tool-builder/internal/provider"
	"github.com/rayjohnson/tool-builder/internal/provider/anthropic"
	"github.com/spf13/cobra"
)

var configPath string

var runCmd = &cobra.Command{
	Use:   "run [args...]",
	Short: "Run a config-defined tool",
	Long: `Run an AI-powered tool defined by a YAML config file.

Any arguments after the flags are passed to the tool as defined by its config.

Examples:
  tool-builder run --config gotest.yaml ./pkg/foo.go
  tool-builder run --config commit-msg.yaml`,
	Args: cobra.ArbitraryArgs,
	RunE: runTool,
}

func init() {
	runCmd.Flags().StringVarP(&configPath, "config", "c", "", "path to tool config YAML (required)")
	_ = runCmd.MarkFlagRequired("config")
}

func runTool(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config %q: %w", configPath, err)
	}

	if _, err := cfg.CheckEnv(); err != nil {
		return err
	}

	p, err := newProvider(cfg)
	if err != nil {
		return err
	}

	// Placeholder: print config summary until the runner is implemented.
	_ = ctx
	_ = p
	_ = args
	fmt.Printf("tool: %s %s\n", cfg.Name, cfg.Version)
	fmt.Printf("model: %s\n", cfg.Model)
	fmt.Printf("commands: %d\n", len(cfg.Commands))
	fmt.Println("(runner not yet implemented)")

	return nil
}

// newProvider constructs the LLM provider for the given config.
// Additional providers are added here as they are implemented.
func newProvider(cfg *config.Config) (provider.Provider, error) {
	switch cfg.Provider() {
	case "anthropic":
		return anthropic.New(cfg.ModelID(), cfg.ModelParams.MaxTokens)
	default:
		return nil, fmt.Errorf("unknown provider %q (supported: anthropic)", cfg.Provider())
	}
}
