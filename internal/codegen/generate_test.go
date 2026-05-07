//nolint:gosec // all file reads in this test are from temp dirs created by the test itself
package codegen_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rayjohnson/tool-builder/internal/codegen"
	"github.com/stretchr/testify/require"
)

var minimalConfig = []byte(`
name: mytool
version: 0.1.0
model: anthropic/m
system_prompts:
  - text: hi
commands:
  - name: default
    description: do it
`)

func TestGenerate_FilesCreated(t *testing.T) {
	prompts := map[string][]byte{
		"prompts/system.md": []byte("# System Prompt"),
	}

	tmpDir, err := codegen.Generate(minimalConfig, prompts, "mytool", "v0.1.0", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// tool.yaml written verbatim.
	got, err := os.ReadFile(filepath.Join(tmpDir, "tool.yaml"))
	require.NoError(t, err)
	require.Equal(t, minimalConfig, got)

	// Prompt file written at its relative path.
	got, err = os.ReadFile(filepath.Join(tmpDir, "prompts", "system.md"))
	require.NoError(t, err)
	require.Equal(t, []byte("# System Prompt"), got)

	// main.go exists and contains expected fragments.
	mainGo, err := os.ReadFile(filepath.Join(tmpDir, "main.go"))
	require.NoError(t, err)
	src := string(mainGo)
	require.Contains(t, src, `//go:embed tool.yaml`)
	require.Contains(t, src, `//go:embed prompts/system.md`)
	require.Contains(t, src, `runtime.Run(embeds`)
	require.Contains(t, src, `"prompts/system.md": prompt0`)

	// go.mod has the right module and version.
	goMod, err := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
	require.NoError(t, err)
	mod := string(goMod)
	require.Contains(t, mod, "module mytool")
	require.Contains(t, mod, "require github.com/rayjohnson/tool-builder v0.1.0")
	require.NotContains(t, mod, "replace")
}

func TestGenerate_DevBuild(t *testing.T) {
	tmpDir, err := codegen.Generate(minimalConfig, nil, "mytool", "dev", "/src/tool-builder")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	goMod, err := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
	require.NoError(t, err)
	mod := string(goMod)
	require.Contains(t, mod, "require github.com/rayjohnson/tool-builder v0.0.0")
	require.Contains(t, mod, "replace github.com/rayjohnson/tool-builder => /src/tool-builder")
}

func TestGenerate_NoPrompts(t *testing.T) {
	tmpDir, err := codegen.Generate(minimalConfig, nil, "mytool", "v0.1.0", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	mainGo, err := os.ReadFile(filepath.Join(tmpDir, "main.go"))
	require.NoError(t, err)
	src := string(mainGo)
	// No prompt embed directives, but the runtime.Run call is present.
	require.Contains(t, src, `runtime.Run(embeds`)
	require.NotContains(t, src, "//go:embed prompt")
	require.NotContains(t, src, "prompt0") // no indexed embed vars
}
