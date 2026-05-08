package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rayjohnson/tool-builder/internal/config"
	"github.com/rayjohnson/tool-builder/internal/runner"
	"github.com/spf13/cobra"
)

// Embeds carries the pre-loaded config and prompt files embedded in a generated binary.
type Embeds struct {
	Config  []byte            // tool.yaml content
	Prompts map[string][]byte // relative path (as written in config) -> file content
}

// Run is the entry point called by generated binaries. It extracts embedded assets
// to a temp directory, builds a Cobra CLI from the config, and executes it.
func Run(embeds Embeds, args []string) error {
	tmpDir, err := os.MkdirTemp("", "tool-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := extractEmbeds(tmpDir, embeds); err != nil {
		return err
	}

	cfg, err := config.Load(filepath.Join(tmpDir, "tool.yaml"))
	if err != nil {
		return err
	}

	if _, err := cfg.CheckEnv(); err != nil {
		return err
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	root := buildRoot(cfg, tmpDir, workDir)
	root.SetArgs(args)
	return root.ExecuteContext(context.Background())
}

func extractEmbeds(dir string, embeds Embeds) error {
	if err := os.WriteFile(filepath.Join(dir, "tool.yaml"), embeds.Config, 0o600); err != nil {
		return fmt.Errorf("writing embedded config: %w", err)
	}
	for path, content := range embeds.Prompts {
		dest := filepath.Join(dir, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
			return fmt.Errorf("creating dir for embedded file %q: %w", path, err)
		}
		if err := os.WriteFile(dest, content, 0o600); err != nil {
			return fmt.Errorf("writing embedded file %q: %w", path, err)
		}
	}
	return nil
}

func buildRoot(cfg *config.Config, configDir, workDir string) *cobra.Command {
	root := &cobra.Command{
		Use:           cfg.Name,
		Short:         cfg.Description,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	if def := cfg.DefaultCommand(); def != nil {
		wireCommand(root, cfg, configDir, workDir, def)
	} else {
		for i := range cfg.Commands {
			cmd := &cfg.Commands[i]
			sub := &cobra.Command{
				Use:           cmd.Name,
				Short:         cmd.Description,
				SilenceUsage:  true,
				SilenceErrors: true,
			}
			wireCommand(sub, cfg, configDir, workDir, cmd)
			root.AddCommand(sub)
		}
	}

	return root
}

func wireCommand(cobraCmd *cobra.Command, cfg *config.Config, configDir, workDir string, toolCmd *config.Command) {
	registerFlags(cobraCmd, toolCmd)
	cobraCmd.Args = cobra.ArbitraryArgs
	cobraCmd.RunE = func(cmd *cobra.Command, args []string) error {
		argFiles, userPrompt := splitArgs(args)

		if extra := buildFlagContext(cmd, toolCmd); extra != "" {
			if userPrompt != "" {
				userPrompt = extra + "\n" + userPrompt
			} else {
				userPrompt = extra
			}
		}

		r := runner.New(cfg, configDir, workDir, os.Stdin, os.Stdout)
		return r.Run(cmd.Context(), toolCmd, argFiles, userPrompt)
	}
}

func registerFlags(cobraCmd *cobra.Command, toolCmd *config.Command) {
	for _, f := range toolCmd.Flags {
		switch f.Type {
		case "bool":
			def, _ := f.Default.(bool)
			if f.Short != "" {
				cobraCmd.Flags().BoolP(f.Name, f.Short, def, f.Description)
			} else {
				cobraCmd.Flags().Bool(f.Name, def, f.Description)
			}
		case "int":
			def, _ := f.Default.(int)
			if f.Short != "" {
				cobraCmd.Flags().IntP(f.Name, f.Short, def, f.Description)
			} else {
				cobraCmd.Flags().Int(f.Name, def, f.Description)
			}
		default: // "string"
			def, _ := f.Default.(string)
			if f.Short != "" {
				cobraCmd.Flags().StringP(f.Name, f.Short, def, f.Description)
			} else {
				cobraCmd.Flags().String(f.Name, def, f.Description)
			}
		}
	}
}

// buildFlagContext returns a formatted string of flags the user explicitly set,
// for injection into the agent's initial message. Only flags the user actually
// passed on the command line are included — flags at their default value are
// omitted so the agent doesn't confuse default state with user intent.
func buildFlagContext(cmd *cobra.Command, toolCmd *config.Command) string {
	var lines []string
	for _, f := range toolCmd.Flags {
		if !cmd.Flags().Changed(f.Name) {
			continue
		}
		switch f.Type {
		case "bool":
			val, _ := cmd.Flags().GetBool(f.Name)
			lines = append(lines, fmt.Sprintf("--%s: %v", f.Name, val))
		case "int":
			val, _ := cmd.Flags().GetInt(f.Name)
			lines = append(lines, fmt.Sprintf("--%s: %d", f.Name, val))
		default: // "string"
			val, _ := cmd.Flags().GetString(f.Name)
			lines = append(lines, fmt.Sprintf("--%s: %s", f.Name, val))
		}
	}
	return strings.Join(lines, "\n")
}

func splitArgs(args []string) (files []string, userPrompt string) {
	if len(args) == 0 {
		return nil, ""
	}
	last := args[len(args)-1]
	looksLikeFile := len(args) > 1 ||
		filepath.Base(last) != last ||
		strings.ContainsRune(last, '.')
	if !looksLikeFile {
		return args[:len(args)-1], last
	}
	return args, ""
}
