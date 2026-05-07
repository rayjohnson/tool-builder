package runner_test

import (
	"testing"

	"github.com/rayjohnson/tool-builder/internal/config"
	"github.com/rayjohnson/tool-builder/internal/runner"
	"github.com/stretchr/testify/require"
)

func TestIsReadAllowed_NoPatterns(t *testing.T) {
	access := config.FileAccess{}
	require.False(t, runner.IsReadAllowed("foo.go", access, "/work"))
}

func TestIsReadAllowed_GlobMatch(t *testing.T) {
	access := config.FileAccess{
		Read: []config.FilePattern{{Glob: "**/*.go"}},
	}
	require.True(t, runner.IsReadAllowed("pkg/foo.go", access, "/work"))
	require.True(t, runner.IsReadAllowed("main.go", access, "/work"))
	require.False(t, runner.IsReadAllowed("README.md", access, "/work"))
}

func TestIsReadAllowed_DirMatch(t *testing.T) {
	access := config.FileAccess{
		Read: []config.FilePattern{{Dir: "internal"}},
	}
	require.True(t, runner.IsReadAllowed("/work/internal/foo.go", access, "/work"))
	require.True(t, runner.IsReadAllowed("/work/internal/sub/bar.go", access, "/work"))
	require.False(t, runner.IsReadAllowed("/work/cmd/root.go", access, "/work"))
}

func TestIsWriteAllowed_GlobMatch(t *testing.T) {
	access := config.FileAccess{
		Write: []config.FilePattern{{Glob: "**/*_test.go"}},
	}
	require.True(t, runner.IsWriteAllowed("pkg/foo_test.go", access, "/work"))
	require.False(t, runner.IsWriteAllowed("pkg/foo.go", access, "/work"))
}

func TestIsReadAllowed_Absolute(t *testing.T) {
	access := config.FileAccess{
		Read: []config.FilePattern{{Glob: "**/*.go"}},
	}
	require.True(t, runner.IsReadAllowed("/work/internal/foo.go", access, "/work"))
	require.False(t, runner.IsReadAllowed("/other/foo.go", access, "/work"))
}
