package systemprompt

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rayjohnson/tool-builder/internal/config"
)

// urlCache caches URL-fetched prompt content for the process lifetime.
var urlCache sync.Map // map[string]string

// Load assembles the full system prompt from a list of sources.
// configDir is the directory containing the config file, used to resolve relative file paths.
func Load(sources []config.PromptSource, configDir string) (string, error) {
	parts := make([]string, 0, len(sources))
	for i, src := range sources {
		text, err := resolve(src, configDir)
		if err != nil {
			return "", fmt.Errorf("system_prompts[%d]: %w", i, err)
		}
		if t := strings.TrimSpace(text); t != "" {
			parts = append(parts, t)
		}
	}
	return strings.Join(parts, "\n\n"), nil
}

func resolve(src config.PromptSource, configDir string) (string, error) {
	switch {
	case src.Text != "":
		return src.Text, nil

	case src.File != "":
		path := src.File
		if !filepath.IsAbs(path) {
			path = filepath.Join(configDir, path)
		}
		data, err := os.ReadFile(path) //nolint:gosec // path is from trusted config file
		if err != nil {
			return "", fmt.Errorf("reading prompt file %q: %w", src.File, err)
		}
		return string(data), nil

	case src.URL != "":
		return fetchURL(src.URL)

	default:
		return "", fmt.Errorf("empty prompt source (set text, file, or url)")
	}
}

func fetchURL(url string) (string, error) {
	if cached, ok := urlCache.Load(url); ok {
		return cached.(string), nil //nolint:forcetypeassert
	}

	resp, err := http.Get(url) //nolint:gosec // URL is from trusted config file
	if err != nil {
		return "", fmt.Errorf("fetching prompt URL %q: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetching prompt URL %q: HTTP %d", url, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading prompt URL %q: %w", url, err)
	}

	content := string(data)
	urlCache.Store(url, content)
	return content, nil
}
