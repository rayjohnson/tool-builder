package systemprompt_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rayjohnson/tool-builder/internal/config"
	"github.com/rayjohnson/tool-builder/internal/systemprompt"
	"github.com/stretchr/testify/require"
)

func TestLoad_TextSource(t *testing.T) {
	sources := []config.PromptSource{
		{Text: "You are a helpful assistant."},
	}
	got, err := systemprompt.Load(sources, "")
	require.NoError(t, err)
	require.Equal(t, "You are a helpful assistant.", got)
}

func TestLoad_MultipleTextSources(t *testing.T) {
	sources := []config.PromptSource{
		{Text: "First part."},
		{Text: "Second part."},
	}
	got, err := systemprompt.Load(sources, "")
	require.NoError(t, err)
	require.Contains(t, got, "First part.")
	require.Contains(t, got, "Second part.")
}

func TestLoad_FileSource(t *testing.T) {
	dir := t.TempDir()
	promptPath := filepath.Join(dir, "system.md")
	require.NoError(t, os.WriteFile(promptPath, []byte("# From file"), 0o600))

	sources := []config.PromptSource{
		{File: "system.md"},
	}
	got, err := systemprompt.Load(sources, dir)
	require.NoError(t, err)
	require.Equal(t, "# From file", got)
}

func TestLoad_FileSource_NotFound(t *testing.T) {
	sources := []config.PromptSource{
		{File: "nonexistent.md"},
	}
	_, err := systemprompt.Load(sources, t.TempDir())
	require.ErrorContains(t, err, "nonexistent.md")
}

func TestLoad_EmptySource(t *testing.T) {
	sources := []config.PromptSource{{}}
	_, err := systemprompt.Load(sources, "")
	require.ErrorContains(t, err, "empty prompt source")
}
