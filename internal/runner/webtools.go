package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
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

type webSearchInput struct {
	Query string `json:"query" jsonschema:"required,description=Search query"`
	Count int    `json:"count" jsonschema:"description=Number of results to return (default 5, max 10)"`
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
		case "search":
			t, err = buildWebSearchTool()
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

func buildWebSearchTool() (anthropic.BetaTool, error) {
	return toolrunner.NewBetaToolFromJSONSchema(
		"web_search",
		"Search the web and return a list of results with titles, URLs, and descriptions. Requires BRAVE_API_KEY environment variable.",
		func(_ context.Context, input webSearchInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			return execWebSearch(input)
		},
	)
}

func execWebSearch(input webSearchInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
	apiKey := os.Getenv("BRAVE_API_KEY")
	if apiKey == "" {
		return errResult("web_search: BRAVE_API_KEY not set")
	}

	count := input.Count
	if count == 0 {
		count = 5
	}
	if count > 10 {
		count = 10
	}

	searchURL := fmt.Sprintf(
		"https://api.search.brave.com/res/v1/web/search?q=%s&count=%d",
		url.QueryEscape(input.Query),
		count,
	)

	req, err := http.NewRequest(http.MethodGet, searchURL, nil)
	if err != nil {
		return errResult(fmt.Sprintf("web_search: building request: %v", err))
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return errResult(fmt.Sprintf("web_search: request failed: %v", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errResult(fmt.Sprintf("web_search: reading response: %v", err))
	}

	var result struct {
		Web struct {
			Results []struct {
				Title       string `json:"title"`
				URL         string `json:"url"`
				Description string `json:"description"`
			} `json:"results"`
		} `json:"web"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return errResult(fmt.Sprintf("web_search: parsing response: %v", err))
	}

	var sb strings.Builder
	for i, r := range result.Web.Results {
		fmt.Fprintf(&sb, "%d. %s\n   %s\n   %s\n\n", i+1, r.Title, r.URL, r.Description)
	}

	if sb.Len() == 0 {
		return okResult("No results found.")
	}
	return okResult(strings.TrimRight(sb.String(), "\n"))
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
