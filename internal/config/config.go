package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name          string          `yaml:"name"`
	Version       string          `yaml:"version"`
	Description   string          `yaml:"description"`
	Model         string          `yaml:"model"` // "provider/model-id"
	ModelParams   ModelParams     `yaml:"model_params"`
	Env           []EnvVar        `yaml:"env"`
	SystemPrompts []PromptSource  `yaml:"system_prompts"`
	Context       []ContextSource `yaml:"context"`
	FileAccess    FileAccess      `yaml:"file_access"`
	Commands      []Command       `yaml:"commands"`
	ToolUse       *ToolUse        `yaml:"tool_use"`
	OutputMode    string          `yaml:"output_mode"`
}

// ContextSource is one entry in context. Exactly one field should be set.
// Content is loaded at runtime from the working directory and injected into the system prompt.
type ContextSource struct {
	Path string `yaml:"path"` // relative to working directory; silently skipped if missing
	URL  string `yaml:"url"`  // fetched at runtime; silently skipped on error
}

type ModelParams struct {
	MaxTokens   int64   `yaml:"max_tokens"`
	Temperature float64 `yaml:"temperature"`
}

type EnvVar struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Default     string `yaml:"default"`
}

// PromptSource is one entry in system_prompts. Exactly one field should be set.
type PromptSource struct {
	Text string `yaml:"text"`
	File string `yaml:"file"`
	URL  string `yaml:"url"`
}

type FileAccess struct {
	Read  []FilePattern `yaml:"read"`
	Write []FilePattern `yaml:"write"`
}

// FilePattern is one entry in file_access.read or file_access.write.
// Exactly one of Glob or Dir should be set.
type FilePattern struct {
	Glob string `yaml:"glob"`
	Dir  string `yaml:"dir"`
}

type Command struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Prompt      string `yaml:"prompt"`
	Args        []Arg  `yaml:"args"`
	Flags       []Flag `yaml:"flags"`
	Session     bool   `yaml:"session"`
}

type Arg struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
}

type Flag struct {
	Name        string `yaml:"name"`
	Short       string `yaml:"short"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"` // string, bool, int, string_slice
	Default     any    `yaml:"default"`
}

type ToolUse struct {
	Enabled bool        `yaml:"enabled"`
	Shell   []ShellTool `yaml:"shell"`
	TUI     []string    `yaml:"tui"`
	Web     []string    `yaml:"web"`
}

type ShellTool struct {
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
}

// Provider returns the provider prefix from a "provider/model-id" model string.
func (c *Config) Provider() string {
	p, _, _ := strings.Cut(c.Model, "/")
	return p
}

// ModelID returns the model identifier portion of the "provider/model-id" string.
func (c *Config) ModelID() string {
	_, m, ok := strings.Cut(c.Model, "/")
	if !ok {
		return c.Model
	}
	return m
}

// OutputModeOrDefault returns the configured output mode, defaulting to "confirm".
func (c *Config) OutputModeOrDefault() string {
	if c.OutputMode == "" {
		return "confirm"
	}
	return c.OutputMode
}

// Command returns the command with the given name, or nil if not found.
func (c *Config) Command(name string) *Command {
	for i := range c.Commands {
		if c.Commands[i].Name == name {
			return &c.Commands[i]
		}
	}
	return nil
}

// DefaultCommand returns the sole command if it is named "default", or nil otherwise.
func (c *Config) DefaultCommand() *Command {
	if len(c.Commands) == 1 && c.Commands[0].Name == "default" {
		return &c.Commands[0]
	}
	return nil
}

// Load reads and parses a config file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if c.Model == "" {
		return fmt.Errorf("model is required")
	}
	if !strings.Contains(c.Model, "/") {
		return fmt.Errorf("model must be in 'provider/model-id' format (got %q)", c.Model)
	}
	if len(c.SystemPrompts) == 0 {
		return fmt.Errorf("at least one system_prompt is required")
	}
	if len(c.Commands) == 0 {
		return fmt.Errorf("at least one command is required")
	}
	for _, cmd := range c.Commands {
		if cmd.Name == "" {
			return fmt.Errorf("each command must have a name")
		}
	}
	switch c.OutputModeOrDefault() {
	case "confirm", "interactive", "direct":
	default:
		return fmt.Errorf("output_mode must be confirm, interactive, or direct")
	}
	if c.ToolUse != nil {
		known := map[string]bool{"list_select": true, "confirm": true, "text_input": true, "text_editor": true, "show_diff": true}
		for _, name := range c.ToolUse.TUI {
			if !known[name] {
				return fmt.Errorf("unknown tui tool %q", name)
			}
		}
		for _, name := range c.ToolUse.Web {
			if name != "fetch" {
				return fmt.Errorf("tool_use.web entry %q is not valid (only \"fetch\" is supported)", name)
			}
		}
	}
	return nil
}

// CheckEnv validates that all required env vars are set and returns resolved values.
// Returns an error listing all missing required vars.
func (c *Config) CheckEnv() (map[string]string, error) {
	resolved := make(map[string]string, len(c.Env))
	var missing []string

	for _, e := range c.Env {
		val := os.Getenv(e.Name)
		if val == "" {
			if e.Required {
				missing = append(missing, fmt.Sprintf("  %s — %s", e.Name, e.Description))
			} else {
				val = e.Default
			}
		}
		resolved[e.Name] = val
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables:\n%s", strings.Join(missing, "\n"))
	}
	return resolved, nil
}
