package runner

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/rayjohnson/tool-builder/internal/config"
	"github.com/rayjohnson/tool-builder/internal/systemprompt"
)

// Runner drives the agent loop for a single tool invocation.
type Runner struct {
	cfg       *config.Config
	configDir string // directory of the config file (for relative prompt paths)
	workDir   string // working directory (for file_access patterns)
	in        io.Reader
	out       io.Writer
}

// New creates a Runner for the given config.
// configDir is the directory containing the config file.
// workDir is the current working directory (where the tool is being run).
func New(cfg *config.Config, configDir, workDir string, in io.Reader, out io.Writer) *Runner {
	return &Runner{
		cfg:       cfg,
		configDir: configDir,
		workDir:   workDir,
		in:        in,
		out:       out,
	}
}

// Run executes the given command with the provided positional args and optional user prompt.
// argFiles are files named as positional args (read and injected into the initial message).
// userPrompt is any additional instruction the user provided on the command line.
func (r *Runner) Run(ctx context.Context, cmd *config.Command, argFiles []string, userPrompt string) error {
	// 1. Assemble system prompt.
	systemPrompt, err := systemprompt.Load(r.cfg.SystemPrompts, r.configDir)
	if err != nil {
		return fmt.Errorf("loading system prompts: %w", err)
	}

	// Append runtime context (from context: entries in config).
	runtimeCtx := loadRuntimeContext(r.cfg.Context, r.workDir)
	if runtimeCtx != "" {
		systemPrompt = systemPrompt + "\n\n" + runtimeCtx
	}

	// 2. Build initial user message.
	initialMsg, err := r.buildInitialMessage(cmd, argFiles, userPrompt)
	if err != nil {
		return fmt.Errorf("building initial message: %w", err)
	}

	// 3. Assemble tools.
	var mu sync.Mutex
	fileTools, err := buildFileTools(r.cfg, r.workDir, r.in, r.out, &mu)
	if err != nil {
		return err
	}
	shellTools, err := buildShellTools(r.cfg, r.workDir)
	if err != nil {
		return err
	}
	tuiTools, err := buildTUITools(r.cfg, &mu)
	if err != nil {
		return err
	}
	webTools, err := buildWebTools(r.cfg)
	if err != nil {
		return err
	}
	tools := append(fileTools, append(shellTools, append(tuiTools, webTools...)...)...)

	// 4. Create Anthropic client.
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	// 5. Configure max_tokens.
	maxTokens := r.cfg.ModelParams.MaxTokens
	if maxTokens == 0 {
		maxTokens = 8096
	}

	// 6. Session loop — runs once for non-session commands, loops for session mode.
	messages := []anthropic.BetaMessageParam{
		anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(initialMsg)),
	}

	for {
		params := anthropic.BetaToolRunnerParams{
			BetaMessageNewParams: anthropic.BetaMessageNewParams{
				Model:     anthropic.Model(r.cfg.ModelID()),
				MaxTokens: maxTokens,
				System: []anthropic.BetaTextBlockParam{
					{Text: systemPrompt},
				},
				Messages: messages,
			},
		}

		streamRunner := client.Beta.Messages.NewToolRunnerStreaming(tools, params)

		// Stream conversation, printing text tokens as they arrive.
		for eventSeq, err := range streamRunner.AllStreaming(ctx) {
			if err != nil {
				return fmt.Errorf("agent stream error: %w", err)
			}
			for event, err := range eventSeq {
				if err != nil {
					return fmt.Errorf("agent stream event error: %w", err)
				}
				if e, ok := event.AsAny().(anthropic.BetaRawContentBlockDeltaEvent); ok {
					if td, ok := e.Delta.AsAny().(anthropic.BetaTextDelta); ok {
						fmt.Fprint(r.out, td.Text)
					}
				}
			}
			// Print a newline after each assistant turn that produced text.
			fmt.Fprintln(r.out)
		}

		if err := streamRunner.Err(); err != nil {
			return err
		}

		if !cmd.Session {
			break
		}

		// Session mode: prompt the user for the next message.
		fmt.Fprint(r.out, "\n> ")
		line, err := readline(r.in)
		if err != nil || strings.TrimSpace(line) == "" {
			break
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "exit" || trimmed == "quit" {
			break
		}

		// Accumulate history and append the new user message.
		messages = streamRunner.Messages()
		messages = append(messages, anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(trimmed)))
	}

	return nil
}

// loadRuntimeContext loads content from the given context sources and formats it for the system prompt.
// path entries are read from the working directory; missing files are silently skipped.
// url entries are fetched via HTTP; errors are silently skipped.
// Returns an empty string if no sources produce content.
func loadRuntimeContext(sources []config.ContextSource, workDir string) string {
	if len(sources) == 0 {
		return ""
	}

	type item struct {
		name    string
		content string
	}
	var items []item

	httpClient := &http.Client{Timeout: 30 * time.Second}

	for _, src := range sources {
		switch {
		case src.Path != "":
			absPath := src.Path
			if !filepath.IsAbs(src.Path) {
				absPath = filepath.Join(workDir, src.Path)
			}
			data, err := os.ReadFile(absPath)
			if err != nil {
				continue // silently skip missing files
			}
			items = append(items, item{name: src.Path, content: string(data)})

		case src.URL != "":
			resp, err := httpClient.Get(src.URL) //nolint:noctx
			if err != nil {
				continue // silently skip errors
			}
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				continue
			}
			items = append(items, item{name: src.URL, content: string(body)})
		}
	}

	if len(items) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Project context\n")
	for _, it := range items {
		fmt.Fprintf(&sb, "\n<%s>\n%s\n</%s>", it.name, it.content, it.name)
	}
	return sb.String()
}

// buildInitialMessage assembles the first user turn: arg files + command prompt + user prompt.
func (r *Runner) buildInitialMessage(cmd *config.Command, argFiles []string, userPrompt string) (string, error) {
	var parts []string

	// Inject arg files.
	if len(argFiles) > 0 {
		var fileParts []string
		for _, f := range argFiles {
			absPath := f
			if !filepath.IsAbs(f) {
				absPath = filepath.Join(r.workDir, f)
			}
			data, err := os.ReadFile(absPath)
			if err != nil {
				return "", fmt.Errorf("reading argument file %q: %w", f, err)
			}
			fileParts = append(fileParts, fmt.Sprintf("<file path=%q>\n%s\n</file>", f, string(data)))
		}
		parts = append(parts, strings.Join(fileParts, "\n"))
	}

	// Add command-level prompt.
	if cmd.Prompt != "" {
		parts = append(parts, strings.TrimSpace(cmd.Prompt))
	}

	// Add user-supplied prompt.
	if userPrompt != "" {
		parts = append(parts, strings.TrimSpace(userPrompt))
	}

	if len(parts) == 0 {
		return "Hello — how can I help?", nil
	}
	return strings.Join(parts, "\n\n"), nil
}
