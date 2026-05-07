package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rayjohnson/tool-builder/internal/config"
	"github.com/stretchr/testify/require"
)

func TestLoad_Valid(t *testing.T) {
	yaml := `
name: mytool
version: 1.0.0
description: A test tool
model: anthropic/claude-opus-4-7
system_prompts:
  - text: You are a helpful assistant.
commands:
  - name: default
    description: Do something
`
	cfg := loadFromString(t, yaml)
	require.Equal(t, "mytool", cfg.Name)
	require.Equal(t, "anthropic", cfg.Provider())
	require.Equal(t, "claude-opus-4-7", cfg.ModelID())
	require.Equal(t, "confirm", cfg.OutputModeOrDefault())
}

func TestLoad_MissingName(t *testing.T) {
	yaml := `
version: 1.0.0
model: anthropic/claude-opus-4-7
system_prompts:
  - text: hi
commands:
  - name: default
    description: do it
`
	_, err := loadFromStringErr(t, yaml)
	require.ErrorContains(t, err, "name is required")
}

func TestLoad_MissingModel(t *testing.T) {
	yaml := `
name: mytool
version: 1.0.0
system_prompts:
  - text: hi
commands:
  - name: default
    description: do it
`
	_, err := loadFromStringErr(t, yaml)
	require.ErrorContains(t, err, "model is required")
}

func TestLoad_ModelMissingProvider(t *testing.T) {
	yaml := `
name: mytool
version: 1.0.0
model: claude-opus-4-7
system_prompts:
  - text: hi
commands:
  - name: default
    description: do it
`
	_, err := loadFromStringErr(t, yaml)
	require.ErrorContains(t, err, "provider/model-id")
}

func TestLoad_NoSystemPrompts(t *testing.T) {
	yaml := `
name: mytool
version: 1.0.0
model: anthropic/claude-opus-4-7
commands:
  - name: default
    description: do it
`
	_, err := loadFromStringErr(t, yaml)
	require.ErrorContains(t, err, "system_prompt")
}

func TestLoad_NoCommands(t *testing.T) {
	yaml := `
name: mytool
version: 1.0.0
model: anthropic/claude-opus-4-7
system_prompts:
  - text: hi
`
	_, err := loadFromStringErr(t, yaml)
	require.ErrorContains(t, err, "command")
}

func TestLoad_InvalidOutputMode(t *testing.T) {
	yaml := `
name: mytool
version: 1.0.0
model: anthropic/claude-opus-4-7
system_prompts:
  - text: hi
commands:
  - name: default
    description: do it
output_mode: bogus
`
	_, err := loadFromStringErr(t, yaml)
	require.ErrorContains(t, err, "output_mode")
}

func TestDefaultCommand(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		wantDefault bool
	}{
		{
			name: "single default command",
			yaml: `
name: t
version: 1.0.0
model: anthropic/m
system_prompts:
  - text: hi
commands:
  - name: default
    description: do it
`,
			wantDefault: true,
		},
		{
			name: "single non-default command",
			yaml: `
name: t
version: 1.0.0
model: anthropic/m
system_prompts:
  - text: hi
commands:
  - name: generate
    description: do it
`,
			wantDefault: false,
		},
		{
			name: "multiple commands",
			yaml: `
name: t
version: 1.0.0
model: anthropic/m
system_prompts:
  - text: hi
commands:
  - name: default
    description: do it
  - name: fix
    description: fix it
`,
			wantDefault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := loadFromString(t, tt.yaml)
			got := cfg.DefaultCommand()
			if tt.wantDefault {
				require.NotNil(t, got)
				require.Equal(t, "default", got.Name)
			} else {
				require.Nil(t, got)
			}
		})
	}
}

func TestCheckEnv_MissingRequired(t *testing.T) {
	yaml := `
name: mytool
version: 1.0.0
model: anthropic/m
system_prompts:
  - text: hi
commands:
  - name: default
    description: do it
env:
  - name: MY_SECRET_TOKEN
    description: A required token
    required: true
`
	cfg := loadFromString(t, yaml)
	t.Setenv("MY_SECRET_TOKEN", "")
	_, err := cfg.CheckEnv()
	require.ErrorContains(t, err, "MY_SECRET_TOKEN")
}

func TestCheckEnv_Default(t *testing.T) {
	yaml := `
name: mytool
version: 1.0.0
model: anthropic/m
system_prompts:
  - text: hi
commands:
  - name: default
    description: do it
env:
  - name: MY_OPTIONAL_VAR
    description: An optional var
    required: false
    default: fallback
`
	cfg := loadFromString(t, yaml)
	t.Setenv("MY_OPTIONAL_VAR", "")
	resolved, err := cfg.CheckEnv()
	require.NoError(t, err)
	require.Equal(t, "fallback", resolved["MY_OPTIONAL_VAR"])
}

func TestLoad_InvalidTUITool(t *testing.T) {
	yaml := `
name: mytool
version: 1.0.0
model: anthropic/m
system_prompts:
  - text: hi
commands:
  - name: default
    description: do it
tool_use:
  enabled: true
  tui: [bogus]
`
	_, err := loadFromStringErr(t, yaml)
	require.ErrorContains(t, err, "unknown tui tool")
}

func TestLoad_InvalidWebTool(t *testing.T) {
	yaml := `
name: mytool
version: 1.0.0
model: anthropic/m
system_prompts:
  - text: hi
commands:
  - name: default
    description: do it
tool_use:
  enabled: true
  web: [bogus]
`
	_, err := loadFromStringErr(t, yaml)
	require.ErrorContains(t, err, "bogus")
}

func TestLoad_ValidWebTools(t *testing.T) {
	yaml := `
name: mytool
version: 1.0.0
model: anthropic/m
system_prompts:
  - text: hi
commands:
  - name: default
    description: do it
tool_use:
  enabled: true
  web: [fetch]
`
	cfg := loadFromString(t, yaml)
	require.Equal(t, []string{"fetch"}, cfg.ToolUse.Web)
}

func TestLoad_Context(t *testing.T) {
	yaml := `
name: mytool
version: 1.0.0
model: anthropic/m
system_prompts:
  - text: hi
commands:
  - name: default
    description: do it
context:
  - path: CLAUDE.md
  - url: https://example.com/standards.md
`
	cfg := loadFromString(t, yaml)
	require.Len(t, cfg.Context, 2)
	require.Equal(t, "CLAUDE.md", cfg.Context[0].Path)
	require.Equal(t, "https://example.com/standards.md", cfg.Context[1].URL)
}

func TestLoad_SessionOnCommand(t *testing.T) {
	yaml := `
name: mytool
version: 1.0.0
model: anthropic/m
system_prompts:
  - text: hi
commands:
  - name: default
    description: do it
    session: true
`
	cfg := loadFromString(t, yaml)
	require.True(t, cfg.Commands[0].Session)
}

// loadFromString writes yaml to a temp file and loads it, requiring success.
func loadFromString(t *testing.T, yaml string) *config.Config {
	t.Helper()
	cfg, err := loadFromStringErr(t, yaml)
	require.NoError(t, err)
	return cfg
}

func loadFromStringErr(t *testing.T, yaml string) (*config.Config, error) {
	t.Helper()
	f := filepath.Join(t.TempDir(), "tool.yaml")
	require.NoError(t, os.WriteFile(f, []byte(yaml), 0o600))
	return config.Load(f)
}
