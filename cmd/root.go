package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	toolBuilderVersion string
	toolBuilderModDir  string
)

var rootCmd = &cobra.Command{
	Use:   "tool-builder",
	Short: "Build AI-powered CLI tools from a YAML config",
	Long: `tool-builder builds AI-powered command-line tools defined by a YAML config file.

Each config file describes an agentic tool: its domain knowledge (system prompts),
the files it can read and write, optional shell tools, and how it interacts with the user.
Use 'build' to compile the config into a self-contained binary.

Example:
  tool-builder build --config commit-msg/tool.yaml -o ./bin/commit-msg`,
}

func Execute(version, buildTime, moduleDir string) {
	toolBuilderVersion = version
	toolBuilderModDir = moduleDir
	rootCmd.Version = fmt.Sprintf("%s (built %s)", version, buildTime)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
