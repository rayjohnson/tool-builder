package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/toolrunner"
	"github.com/rayjohnson/tool-builder/internal/config"
)

type readFileInput struct {
	Path string `json:"path" jsonschema:"required,description=Path to the file to read (relative to working directory)"`
}

type writeFileInput struct {
	Path    string `json:"path"    jsonschema:"required,description=Path to write (relative to working directory)"`
	Content string `json:"content" jsonschema:"required,description=Full file content to write"`
}

type editFileInput struct {
	Path      string `json:"path"       jsonschema:"required,description=Path to the file to edit (relative to working directory)"`
	OldString string `json:"old_string" jsonschema:"required,description=Exact text to find and replace — must match exactly once"`
	NewString string `json:"new_string" jsonschema:"required,description=Replacement text"`
}

// buildFileTools returns read_file and write_file tools scoped to the config's file_access rules.
// mu serializes confirm prompts when the LLM requests multiple writes in parallel.
func buildFileTools(
	cfg *config.Config,
	workDir string,
	in io.Reader,
	out io.Writer,
	mu *sync.Mutex,
) ([]anthropic.BetaTool, error) {
	readTool, err := toolrunner.NewBetaToolFromJSONSchema(
		"read_file",
		"Read a file from the working directory. Only files within the configured file_access.read scope are accessible.",
		func(ctx context.Context, input readFileInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			return execReadFile(input, cfg, workDir)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("building read_file tool: %w", err)
	}

	writeTool, err := toolrunner.NewBetaToolFromJSONSchema(
		"write_file",
		"Write content to a file in the working directory. Only files within the configured file_access.write scope may be written.",
		func(ctx context.Context, input writeFileInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			mu.Lock()
			defer mu.Unlock()
			return execWriteFile(input, cfg, workDir, in, out)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("building write_file tool: %w", err)
	}

	editTool, err := toolrunner.NewBetaToolFromJSONSchema(
		"edit_file",
		"Make a targeted edit to a file by replacing an exact string. old_string must appear exactly once in the file — use enough context (surrounding lines) to make it unique. Preferred over write_file for modifying existing files.",
		func(ctx context.Context, input editFileInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			mu.Lock()
			defer mu.Unlock()
			return execEditFile(input, cfg, workDir, in, out)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("building edit_file tool: %w", err)
	}

	return []anthropic.BetaTool{readTool, writeTool, editTool}, nil
}

func execReadFile(
	input readFileInput,
	cfg *config.Config,
	workDir string,
) (anthropic.BetaToolResultBlockParamContentUnion, error) {
	absPath := resolveRelative(input.Path, workDir)

	if !IsReadAllowed(absPath, cfg.FileAccess, workDir) {
		return errResult(fmt.Sprintf("read denied: %q is outside the configured file_access.read scope", input.Path))
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return errResult(fmt.Sprintf("read error: %v", err))
	}

	return anthropic.BetaToolResultBlockParamContentUnion{
		OfText: &anthropic.BetaTextBlockParam{Text: string(data)},
	}, nil
}

func execWriteFile(
	input writeFileInput,
	cfg *config.Config,
	workDir string,
	in io.Reader,
	out io.Writer,
) (anthropic.BetaToolResultBlockParamContentUnion, error) {
	absPath := resolveRelative(input.Path, workDir)

	if !IsWriteAllowed(absPath, cfg.FileAccess, workDir) {
		return errResult(fmt.Sprintf("write denied: %q is outside the configured file_access.write scope", input.Path))
	}

	mode := cfg.OutputModeOrDefault()
	switch mode {
	case "direct", "terse":
		if err := writeFile(absPath, input.Content); err != nil {
			return errResult(fmt.Sprintf("write error: %v", err))
		}
		return okResult(fmt.Sprintf("wrote %s", input.Path))

	case "confirm", "interactive":
		confirmed, err := confirmWrite(input.Path, input.Content, in, out)
		if err != nil {
			return errResult(fmt.Sprintf("confirm error: %v", err))
		}
		if !confirmed {
			return okResult(fmt.Sprintf("skipped %s (user declined)", input.Path))
		}
		if err := writeFile(absPath, input.Content); err != nil {
			return errResult(fmt.Sprintf("write error: %v", err))
		}
		return okResult(fmt.Sprintf("wrote %s", input.Path))

	default:
		return errResult(fmt.Sprintf("unknown output_mode %q", mode))
	}
}

func writeFile(absPath, content string) error {
	if err := os.MkdirAll(filepath.Dir(absPath), 0o750); err != nil {
		return err
	}
	return os.WriteFile(absPath, []byte(content), 0o600)
}

func resolveRelative(path, workDir string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(workDir, path)
}

func execEditFile(
	input editFileInput,
	cfg *config.Config,
	workDir string,
	in io.Reader,
	out io.Writer,
) (anthropic.BetaToolResultBlockParamContentUnion, error) {
	absPath := resolveRelative(input.Path, workDir)

	if !IsWriteAllowed(absPath, cfg.FileAccess, workDir) {
		return errResult(fmt.Sprintf("write denied: %q is outside the configured file_access.write scope", input.Path))
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return errResult(fmt.Sprintf("read error: %v", err))
	}
	content := string(data)

	count := strings.Count(content, input.OldString)
	switch {
	case count == 0:
		return errResult(fmt.Sprintf("edit_file: old_string not found in %s", input.Path))
	case count > 1:
		return errResult(fmt.Sprintf("edit_file: old_string matches %d times in %s — add more surrounding context to make it unique", count, input.Path))
	}

	newContent := strings.Replace(content, input.OldString, input.NewString, 1)

	mode := cfg.OutputModeOrDefault()
	switch mode {
	case "direct", "terse":
		if err := writeFile(absPath, newContent); err != nil {
			return errResult(fmt.Sprintf("write error: %v", err))
		}
	case "confirm", "interactive":
		confirmed, err := confirmWrite(input.Path, newContent, in, out)
		if err != nil {
			return errResult(fmt.Sprintf("confirm error: %v", err))
		}
		if !confirmed {
			return okResult(fmt.Sprintf("skipped %s (user declined)", input.Path))
		}
		if err := writeFile(absPath, newContent); err != nil {
			return errResult(fmt.Sprintf("write error: %v", err))
		}
	default:
		return errResult(fmt.Sprintf("unknown output_mode %q", mode))
	}

	return okResult(fmt.Sprintf("edited %s", input.Path))
}

func okResult(msg string) (anthropic.BetaToolResultBlockParamContentUnion, error) {
	return anthropic.BetaToolResultBlockParamContentUnion{
		OfText: &anthropic.BetaTextBlockParam{Text: msg},
	}, nil
}

func errResult(msg string) (anthropic.BetaToolResultBlockParamContentUnion, error) {
	return anthropic.BetaToolResultBlockParamContentUnion{
		OfText: &anthropic.BetaTextBlockParam{Text: msg},
	}, nil
}
