package runner

import (
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rayjohnson/tool-builder/internal/config"
)

// IsReadAllowed reports whether the given path (absolute or relative to workDir)
// is within the configured file_access.read scope.
func IsReadAllowed(path string, access config.FileAccess, workDir string) bool {
	return matchesPatterns(path, access.Read, workDir)
}

// IsWriteAllowed reports whether the given path is within file_access.write scope.
func IsWriteAllowed(path string, access config.FileAccess, workDir string) bool {
	return matchesPatterns(path, access.Write, workDir)
}

func matchesPatterns(path string, patterns []config.FilePattern, workDir string) bool {
	if len(patterns) == 0 {
		return false
	}

	// Resolve to absolute path for dir checks; keep relative for glob checks.
	absPath := path
	if !filepath.IsAbs(absPath) {
		absPath = filepath.Join(workDir, path)
	}
	// Canonical relative path from workDir (for glob matching).
	relPath, err := filepath.Rel(workDir, absPath)
	if err != nil {
		relPath = path
	}

	for _, p := range patterns {
		switch {
		case p.Dir != "":
			dirAbs := p.Dir
			if !filepath.IsAbs(dirAbs) {
				dirAbs = filepath.Join(workDir, p.Dir)
			}
			dirAbs = filepath.Clean(dirAbs)
			// Path is allowed if it equals the dir or is under it.
			if absPath == dirAbs || strings.HasPrefix(absPath, dirAbs+string(filepath.Separator)) {
				return true
			}

		case p.Glob != "":
			matched, err := doublestar.Match(filepath.ToSlash(p.Glob), filepath.ToSlash(relPath))
			if err == nil && matched {
				return true
			}
		}
	}
	return false
}
