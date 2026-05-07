package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeUnifiedDiff_Changes(t *testing.T) {
	old := "line one\nline two\nline three\n"
	new := "line one\nline TWO\nline three\n"
	result := computeUnifiedDiff("file.go", old, new)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "-line two")
	assert.Contains(t, result, "+line TWO")
}

func TestComputeUnifiedDiff_NoChanges(t *testing.T) {
	content := "line one\nline two\n"
	result := computeUnifiedDiff("file.go", content, content)
	assert.Empty(t, result)
}

func TestRenderDiff_DoesNotPanic(t *testing.T) {
	old := "hello world\n"
	newContent := "hello Go\n"
	raw := computeUnifiedDiff("test.go", old, newContent)
	require.NotEmpty(t, raw)
	result := renderDiff(raw)
	assert.NotEmpty(t, result)
	// Verify it contains content (ANSI codes around the text)
	assert.Contains(t, result, "hello")
}

func TestDiffModel_Accept(t *testing.T) {
	m := newDiffModel("file.go", "some diff content\n+added line\n-removed line")
	m.ready = true

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	result := next.(diffModel)
	assert.True(t, result.done)
	assert.Equal(t, "accept", result.result)
}

func TestDiffModel_AcceptEnter(t *testing.T) {
	m := newDiffModel("file.go", "diff content")
	m.ready = true

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := next.(diffModel)
	assert.True(t, result.done)
	assert.Equal(t, "accept", result.result)
}

func TestDiffModel_Reject(t *testing.T) {
	m := newDiffModel("file.go", "some diff content")
	m.ready = true

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	result := next.(diffModel)
	assert.True(t, result.done)
	assert.Equal(t, "reject", result.result)
}

func TestDiffModel_RejectQ(t *testing.T) {
	m := newDiffModel("file.go", "some diff content")
	m.ready = true

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	result := next.(diffModel)
	assert.True(t, result.done)
	assert.Equal(t, "reject", result.result)
}

func TestDiffModel_RejectEsc(t *testing.T) {
	m := newDiffModel("file.go", "some diff content")
	m.ready = true

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	result := next.(diffModel)
	assert.True(t, result.done)
	assert.Equal(t, "reject", result.result)
}

func TestDiffModel_Feedback(t *testing.T) {
	m := newDiffModel("file.go", "some diff content")
	m.ready = true

	// Press 'f' to enter feedback mode
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = next.(diffModel)
	assert.Equal(t, stateFeedback, m.state)

	// Type some characters
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = next.(diffModel)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	m = next.(diffModel)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = next.(diffModel)

	// Press enter to submit
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(diffModel)
	assert.True(t, m.done)
	assert.Equal(t, "abc", m.result)
}

func TestDiffModel_FeedbackEscCancels(t *testing.T) {
	m := newDiffModel("file.go", "some diff content")
	m.ready = true

	// Enter feedback mode
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = next.(diffModel)
	assert.Equal(t, stateFeedback, m.state)

	// Press esc to cancel feedback
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(diffModel)
	assert.Equal(t, stateViewing, m.state)
	assert.False(t, m.done)
}

func TestDiffModel_FeedbackEmptyReturnsReject(t *testing.T) {
	m := newDiffModel("file.go", "some diff content")
	m.ready = true

	// Enter feedback mode
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = next.(diffModel)

	// Press enter without typing anything
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(diffModel)
	assert.True(t, m.done)
	assert.Equal(t, "reject", m.result)
}

func TestDiffModel_WindowSizeMsg(t *testing.T) {
	m := newDiffModel("file.go", "content")
	assert.False(t, m.ready)

	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(diffModel)
	assert.True(t, m.ready)
	assert.Equal(t, 80, m.width)
	assert.Equal(t, 24, m.height)
}

func TestDiffModel_ViewNotReady(t *testing.T) {
	m := newDiffModel("myfile.go", "content")
	view := m.View()
	assert.Contains(t, view, "myfile.go")
	assert.Contains(t, view, "Initializing...")
}

func TestDiffModel_ViewReady(t *testing.T) {
	m := newDiffModel("myfile.go", "content")
	m.ready = true
	// Initialize a small viewport
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(diffModel)

	view := m.View()
	assert.Contains(t, view, "myfile.go")
	assert.Contains(t, view, "[y] accept")
	assert.Contains(t, view, "[n] reject")
	assert.Contains(t, view, "[f] feedback")
}

func TestDiffModel_ViewFeedbackState(t *testing.T) {
	m := newDiffModel("myfile.go", "content")
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(diffModel)

	// Enter feedback state
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = next.(diffModel)

	view := m.View()
	assert.Contains(t, view, "Feedback:")
	assert.Contains(t, view, "[enter] submit")
	assert.Contains(t, view, "[esc] cancel")
}
