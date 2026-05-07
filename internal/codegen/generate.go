package codegen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"text/template"
)

const toolBuilderModule = "github.com/rayjohnson/tool-builder"

var mainTmpl = template.Must(template.New("main").Parse(`package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/rayjohnson/tool-builder/pkg/runtime"
)

//go:embed tool.yaml
var configYAML []byte
{{range .Prompts}}
//go:embed {{.Path}}
var {{.VarName}} []byte
{{end}}
func main() {
	embeds := runtime.Embeds{
		Config: configYAML,
		Prompts: map[string][]byte{
{{- range .Prompts}}
			{{printf "%q" .Path}}: {{.VarName}},
{{- end}}
		},
	}
	if err := runtime.Run(embeds, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
`))

var goModTmpl = template.Must(template.New("gomod").Parse(`module {{.ModuleName}}

go {{.GoVersion}}

require {{.ToolBuilderModule}} {{.RequireVersion}}
{{- if .LocalPath}}

replace {{.ToolBuilderModule}} => {{.LocalPath}}
{{- end}}
`))

type promptEntry struct {
	Path    string
	VarName string
}

// Generate writes a temp directory ready for `go build` and returns its path.
// The caller must clean up the directory.
//
// configBytes is the raw YAML of the tool config.
// prompts maps the relative path (as written in the config file: source) to content.
// toolName is used as the generated module name.
// version is the tool-builder version (e.g. "v0.1.1" or "dev").
// localPath, when non-empty, causes a `replace` directive pointing to the local
// tool-builder checkout (used for dev builds).
func Generate(configBytes []byte, prompts map[string][]byte, toolName, version, localPath string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "tool-build-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}

	if err := writeFiles(tmpDir, configBytes, prompts, toolName, version, localPath); err != nil {
		return tmpDir, err
	}
	return tmpDir, nil
}

func writeFiles(tmpDir string, configBytes []byte, prompts map[string][]byte, toolName, version, localPath string) error {
	// Write tool.yaml.
	if err := os.WriteFile(filepath.Join(tmpDir, "tool.yaml"), configBytes, 0o600); err != nil {
		return fmt.Errorf("writing tool.yaml: %w", err)
	}

	// Write prompt files, preserving relative paths.
	// Sort keys so generated var names are deterministic across builds.
	keys := make([]string, 0, len(prompts))
	for path := range prompts {
		keys = append(keys, path)
	}
	sort.Strings(keys)

	entries := make([]promptEntry, 0, len(prompts))
	for i, path := range keys {
		dest := filepath.Join(tmpDir, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
			return fmt.Errorf("creating dir for %q: %w", path, err)
		}
		if err := os.WriteFile(dest, prompts[path], 0o600); err != nil {
			return fmt.Errorf("writing %q: %w", path, err)
		}
		entries = append(entries, promptEntry{Path: path, VarName: fmt.Sprintf("prompt%d", i)})
	}

	// Generate and write main.go.
	mainSrc, err := renderMain(entries)
	if err != nil {
		return fmt.Errorf("generating main.go: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), mainSrc, 0o600); err != nil {
		return fmt.Errorf("writing main.go: %w", err)
	}

	// Generate and write go.mod.
	goModSrc, err := renderGoMod(toolName, version, localPath)
	if err != nil {
		return fmt.Errorf("generating go.mod: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), goModSrc, 0o600); err != nil {
		return fmt.Errorf("writing go.mod: %w", err)
	}

	return nil
}

func renderMain(prompts []promptEntry) ([]byte, error) {
	var buf bytes.Buffer
	if err := mainTmpl.Execute(&buf, struct{ Prompts []promptEntry }{prompts}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func renderGoMod(toolName, version, localPath string) ([]byte, error) {
	requireVersion := version
	if localPath != "" {
		requireVersion = "v0.0.0"
	}
	// toolName is used as the module name; it must be a valid module path segment
	// (lowercase, no spaces or special characters).
	data := struct {
		ModuleName        string
		GoVersion         string
		ToolBuilderModule string
		RequireVersion    string
		LocalPath         string
	}{
		ModuleName:        strings.ToLower(toolName),
		GoVersion:         strings.TrimPrefix(runtime.Version(), "go"),
		ToolBuilderModule: toolBuilderModule,
		RequireVersion:    requireVersion,
		LocalPath:         localPath,
	}
	var buf bytes.Buffer
	if err := goModTmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
