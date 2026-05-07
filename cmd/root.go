package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tool-builder",
	Short: "Build AI-powered CLI tools from a YAML config",
	Long: `tool-builder runs AI-powered command-line tools defined by a YAML config file.

Each config file describes an agentic tool: its domain knowledge (system prompts),
the files it can read and write, optional shell tools, and how it interacts with the user.

Examples:
  tool-builder run --config gotest.yaml ./pkg/foo.go
  tool-builder run --config commit-msg.yaml`,
}

func Execute(version, buildTime string) {
	rootCmd.Version = fmt.Sprintf("%s (built %s)", version, buildTime)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
}
