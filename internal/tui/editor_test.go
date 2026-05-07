package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveEditor_EDITOR(t *testing.T) {
	t.Setenv("EDITOR", "myed")
	t.Setenv("VISUAL", "")
	got, err := resolveEditor()
	require.NoError(t, err)
	assert.Equal(t, "myed", got)
}

func TestResolveEditor_VISUAL_fallback(t *testing.T) {
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "myvisual")
	got, err := resolveEditor()
	require.NoError(t, err)
	assert.Equal(t, "myvisual", got)
}

func TestEdit_noChange(t *testing.T) {
	// "true" is always in PATH and exits 0 without touching its argument.
	t.Setenv("EDITOR", "true")
	content := "original content\nline two\n"
	result, err := Edit(content, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, content, result)
}

func TestEdit_noEditor(t *testing.T) {
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "")
	// With no env vars and none of the fallback editors likely present in a
	// minimal CI environment, this should return an error. If vi/nano is
	// present the test is skipped to avoid hanging on a real editor.
	_, vi := resolveEditor()
	if vi == nil {
		t.Skip("a fallback editor (vi/nano) is present; skipping no-editor test")
	}
	_, err := Edit("content", "test.txt")
	require.Error(t, err)
}
