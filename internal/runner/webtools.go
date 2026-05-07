package runner

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/toolrunner"
	"github.com/rayjohnson/tool-builder/internal/config"
)

type webFetchInput struct {
	URL      string `json:"url"       jsonschema:"required,description=The URL to fetch"`
	MaxBytes int    `json:"max_bytes" jsonschema:"description=Maximum response bytes to return (default 20000)"`
}

// buildWebTools returns tools for each name in cfg.ToolUse.Web.
func buildWebTools(cfg *config.Config) ([]anthropic.BetaTool, error) {
	if cfg.ToolUse == nil {
		return nil, nil
	}

	var tools []anthropic.BetaTool
	for _, name := range cfg.ToolUse.Web {
		var (
			t   anthropic.BetaTool
			err error
		)
		switch name {
		case "fetch":
			t, err = buildWebFetchTool()
		default:
			return nil, fmt.Errorf("unknown web tool %q", name)
		}
		if err != nil {
			return nil, fmt.Errorf("building web tool %q: %w", name, err)
		}
		tools = append(tools, t)
	}
	return tools, nil
}

func buildWebFetchTool() (anthropic.BetaTool, error) {
	return toolrunner.NewBetaToolFromJSONSchema(
		"web_fetch",
		"Fetch the content of a URL and return it as text. HTML pages are converted to plain text. Use this to read documentation, API specs, web pages, or any URL-accessible content.",
		func(_ context.Context, input webFetchInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			return execWebFetch(input)
		},
	)
}

func execWebFetch(input webFetchInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
	maxBytes := input.MaxBytes
	if maxBytes <= 0 {
		maxBytes = 20000
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(input.URL) //nolint:noctx
	if err != nil {
		return errResult(fmt.Sprintf("web_fetch: request failed: %v", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, int64(maxBytes)+1))
	if err != nil {
		return errResult(fmt.Sprintf("web_fetch: reading response: %v", err))
	}

	text := string(body)
	truncated := false
	if len(text) > maxBytes {
		text = text[:maxBytes]
		truncated = true
	}

	ct := resp.Header.Get("Content-Type")
	if strings.Contains(ct, "text/html") {
		text = htmlToText(text)
	}

	if truncated {
		text += fmt.Sprintf("\n\n[truncated at %d bytes]", maxBytes)
	}

	return okResult(text)
}

// HTML stripping helpers.
var (
	reScript   = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	reStyle    = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	reBlock    = regexp.MustCompile(`(?i)<(p|div|br|h[1-6]|li|tr|td|th)[^>]*>`)
	reTag      = regexp.MustCompile(`<[^>]+>`)
	reSpaces   = regexp.MustCompile(`[ \t]+`)
	reNewlines = regexp.MustCompile(`\n{3,}`)
)

func htmlToText(s string) string {
	s = reScript.ReplaceAllString(s, "")
	s = reStyle.ReplaceAllString(s, "")
	s = reBlock.ReplaceAllString(s, "\n")
	s = reTag.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	s = reSpaces.ReplaceAllString(s, " ")
	s = reNewlines.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}
