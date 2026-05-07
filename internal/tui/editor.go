package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Edit opens $EDITOR (falling back to $VISUAL, vi, then nano) with the provided
// content pre-populated and returns the saved result. filename is used only as a
// suffix on the temp file to give the editor a syntax-highlighting hint.
// Returns the original content unchanged if the editor exits with an error.
func Edit(content, filename string) (string, error) {
	editor, err := resolveEditor()
	if err != nil {
		return "", err
	}

	suffix := filepath.Base(filename)
	if suffix == "" || suffix == "." {
		suffix = "edit.txt"
	}
	tmp, err := os.CreateTemp("", "tool-builder-*-"+suffix)
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		return "", fmt.Errorf("writing temp file: %w", err)
	}
	tmp.Close()

	cmd := exec.Command(editor, tmpName) //nolint:gosec // editor is resolved from $EDITOR or well-known fallbacks
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return content, fmt.Errorf("editor exited with error: %w", err)
	}

	edited, err := os.ReadFile(tmpName)
	if err != nil {
		return "", fmt.Errorf("reading edited file: %w", err)
	}
	return string(edited), nil
}

func resolveEditor() (string, error) {
	for _, env := range []string{"EDITOR", "VISUAL"} {
		if val := os.Getenv(env); val != "" {
			return val, nil
		}
	}
	for _, candidate := range []string{"vi", "nano", "notepad"} {
		if path, err := exec.LookPath(candidate); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no editor found: set $EDITOR")
}
