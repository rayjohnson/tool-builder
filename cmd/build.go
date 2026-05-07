package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rayjohnson/tool-builder/internal/codegen"
	"github.com/rayjohnson/tool-builder/internal/config"
	"github.com/spf13/cobra"
)

var (
	buildConfigPath string
	buildOutput     string
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a standalone binary from a config file",
	Long: `Build a self-contained binary from a tool config file.

The binary embeds the config and all prompt files. End users need only the
binary and an API key — no dependency on tool-builder at runtime.

For development builds, run 'make build' so the module directory is set
correctly in the generated go.mod.

Examples:
  tool-builder build --config commit-msg/tool.yaml
  tool-builder build --config gotest.yaml --output ./bin/gotest`,
	Args: cobra.NoArgs,
	RunE: runBuild,
}

func init() {
	buildCmd.Flags().StringVarP(&buildConfigPath, "config", "c", "", "path to tool config YAML (required)")
	_ = buildCmd.MarkFlagRequired("config")
	buildCmd.Flags().StringVarP(&buildOutput, "output", "o", "", "output binary path (default: ./<tool-name>)")
}

func runBuild(_ *cobra.Command, _ []string) error {
	if toolBuilderVersion == "dev" && toolBuilderModDir == "" {
		return fmt.Errorf("cannot determine tool-builder module path for dev build\n" +
			"Run 'make build' instead of 'go build .' to set the module directory via ldflags")
	}

	cfg, err := config.Load(buildConfigPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	configDir, err := filepath.Abs(filepath.Dir(buildConfigPath))
	if err != nil {
		return fmt.Errorf("resolving config dir: %w", err)
	}

	configBytes, err := os.ReadFile(buildConfigPath) //nolint:gosec // user-supplied path
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	prompts, err := collectPrompts(cfg, configDir)
	if err != nil {
		return err
	}

	output := buildOutput
	if output == "" {
		output = "./" + cfg.Name
	}
	absOutput, err := filepath.Abs(output)
	if err != nil {
		return fmt.Errorf("resolving output path: %w", err)
	}

	tmpDir, err := codegen.Generate(configBytes, prompts, cfg.Name, toolBuilderVersion, toolBuilderModDir)
	if err != nil {
		_ = os.RemoveAll(tmpDir) //nolint:gosec // tmpDir comes from os.MkdirTemp
		return fmt.Errorf("generating build directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Fprintf(os.Stdout, "Building %s...\n", cfg.Name)
	//nolint:gosec // arguments are constructed internally, not from user input
	goCmd := exec.Command("go", "build", "-mod=mod", "-o", absOutput, ".")
	goCmd.Dir = tmpDir
	goCmd.Stdout = os.Stdout
	goCmd.Stderr = os.Stderr
	if err := goCmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Built: %s\n", absOutput)
	return nil
}

// collectPrompts reads all file: prompt sources into a path->content map.
// URL sources are not yet supported and will return an error.
func collectPrompts(cfg *config.Config, configDir string) (map[string][]byte, error) {
	prompts := make(map[string][]byte)
	for _, src := range cfg.SystemPrompts {
		switch {
		case src.File != "":
			path := src.File
			if !filepath.IsAbs(path) {
				path = filepath.Join(configDir, src.File)
			}
			data, err := os.ReadFile(path) //nolint:gosec // path from trusted config file
			if err != nil {
				return nil, fmt.Errorf("reading prompt file %q: %w", src.File, err)
			}
			prompts[src.File] = data
		case src.URL != "":
			return nil, fmt.Errorf("URL prompt sources are not yet supported by build: %q", src.URL)
		}
	}
	return prompts, nil
}
