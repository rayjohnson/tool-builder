package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestConfirmModel_yes(t *testing.T) {
	m := confirmModel{question: "Continue?"}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = next.(confirmModel)
	assert.True(t, m.answer)
	assert.True(t, m.done)
}

func TestConfirmModel_no(t *testing.T) {
	m := confirmModel{question: "Continue?", defaultYes: true}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = next.(confirmModel)
	assert.False(t, m.answer)
	assert.True(t, m.done)
}

func TestConfirmModel_enter_defaultNo(t *testing.T) {
	m := confirmModel{question: "Continue?", defaultYes: false}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(confirmModel)
	assert.False(t, m.answer)
}

func TestConfirmModel_enter_defaultYes(t *testing.T) {
	m := confirmModel{question: "Continue?", defaultYes: true}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(confirmModel)
	assert.True(t, m.answer)
}

func TestConfirmModel_cancel(t *testing.T) {
	m := confirmModel{question: "Continue?"}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = next.(confirmModel)
	assert.False(t, m.answer)
	assert.True(t, m.done)
}

func TestConfirmModel_view_pending(t *testing.T) {
	m := confirmModel{question: "Delete everything?", defaultYes: false}
	view := m.View()
	assert.Contains(t, view, "Delete everything?")
	assert.Contains(t, view, "[y/N]")
}

func TestConfirmModel_view_defaultYes(t *testing.T) {
	m := confirmModel{question: "Proceed?", defaultYes: true}
	view := m.View()
	assert.Contains(t, view, "[Y/n]")
}

func TestConfirmModel_view_done(t *testing.T) {
	m := confirmModel{question: "Proceed?", answer: true, done: true}
	view := m.View()
	assert.Contains(t, view, "Yes")

	m2 := confirmModel{question: "Proceed?", answer: false, done: true}
	view2 := m2.View()
	assert.Contains(t, view2, "No")
}
