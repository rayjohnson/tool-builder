package runtime

import (
	"testing"

	"github.com/rayjohnson/tool-builder/internal/config"
	"github.com/spf13/cobra"
)

func TestBuildFlagContext(t *testing.T) {
	tests := []struct {
		name     string
		flags    []config.Flag
		setFlags map[string]string // name → string value to pass to Flags().Set()
		want     string
	}{
		{
			name: "string flag with non-empty default not set by user",
			flags: []config.Flag{
				{Name: "path", Type: "string", Default: "./..."},
			},
			setFlags: nil,
			want:     "", // must NOT appear — this was the bug
		},
		{
			name: "string flag with non-empty default explicitly set by user",
			flags: []config.Flag{
				{Name: "path", Type: "string", Default: "./..."},
			},
			setFlags: map[string]string{"path": "./internal/..."},
			want:     "--path: ./internal/...",
		},
		{
			name: "string flag with empty default not set by user",
			flags: []config.Flag{
				{Name: "config", Type: "string"},
			},
			setFlags: nil,
			want:     "",
		},
		{
			name: "bool flag not set by user",
			flags: []config.Flag{
				{Name: "verbose", Type: "bool"},
			},
			setFlags: nil,
			want:     "",
		},
		{
			name: "bool flag set by user",
			flags: []config.Flag{
				{Name: "verbose", Type: "bool"},
			},
			setFlags: map[string]string{"verbose": "true"},
			want:     "--verbose: true",
		},
		{
			name: "int flag not set by user",
			flags: []config.Flag{
				{Name: "count", Type: "int", Default: 10},
			},
			setFlags: nil,
			want:     "",
		},
		{
			name: "int flag set by user",
			flags: []config.Flag{
				{Name: "count", Type: "int", Default: 10},
			},
			setFlags: map[string]string{"count": "42"},
			want:     "--count: 42",
		},
		{
			name: "multiple flags — only set ones appear",
			flags: []config.Flag{
				{Name: "path", Type: "string", Default: "./..."},
				{Name: "only", Type: "string"},
				{Name: "verbose", Type: "bool"},
			},
			setFlags: map[string]string{"only": "misspell"},
			want:     "--only: misspell",
		},
		{
			name: "multiple flags — all set",
			flags: []config.Flag{
				{Name: "path", Type: "string", Default: "./..."},
				{Name: "only", Type: "string"},
			},
			setFlags: map[string]string{"path": "./cmd/...", "only": "gosec"},
			want:     "--path: ./cmd/...\n--only: gosec",
		},
		{
			name: "no flags defined",
			flags: nil,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			toolCmd := &config.Command{Flags: tt.flags}
			registerFlags(cmd, toolCmd)

			for name, val := range tt.setFlags {
				if err := cmd.Flags().Set(name, val); err != nil {
					t.Fatalf("setting flag %q=%q: %v", name, val, err)
				}
			}

			got := buildFlagContext(cmd, toolCmd)
			if got != tt.want {
				t.Errorf("buildFlagContext() = %q, want %q", got, tt.want)
			}
		})
	}
}
