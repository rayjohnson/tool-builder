package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelect_emptyItems(t *testing.T) {
	_, err := Select("title", nil, false)
	require.Error(t, err)
}

func TestSelectorModel_navigate(t *testing.T) {
	m := selectorModel{items: []string{"a", "b", "c"}, selected: make(map[int]bool)}

	// down
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = next.(selectorModel)
	assert.Equal(t, 1, m.cursor)

	// down again
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(selectorModel)
	assert.Equal(t, 2, m.cursor)

	// can't go past last item
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(selectorModel)
	assert.Equal(t, 2, m.cursor)

	// up
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = next.(selectorModel)
	assert.Equal(t, 1, m.cursor)

	// up with k
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = next.(selectorModel)
	assert.Equal(t, 0, m.cursor)

	// can't go before first item
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = next.(selectorModel)
	assert.Equal(t, 0, m.cursor)
}

func TestSelectorModel_toggle(t *testing.T) {
	m := selectorModel{items: []string{"a", "b"}, selected: make(map[int]bool)}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = next.(selectorModel)
	assert.True(t, m.selected[0], "space should select item")

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = next.(selectorModel)
	assert.False(t, m.selected[0], "space again should deselect item")
}

func TestSelectorModel_selectAll(t *testing.T) {
	m := selectorModel{items: []string{"a", "b", "c"}, selected: make(map[int]bool)}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = next.(selectorModel)
	assert.True(t, m.selected[0])
	assert.True(t, m.selected[1])
	assert.True(t, m.selected[2])
}

func TestSelectorModel_deselectAll(t *testing.T) {
	m := selectorModel{items: []string{"a", "b"}, selected: map[int]bool{0: true, 1: true}}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = next.(selectorModel)
	assert.Empty(t, m.selected)
}

func TestSelectorModel_singleMode_noToggle(t *testing.T) {
	m := selectorModel{items: []string{"a", "b"}, selected: make(map[int]bool), single: true}

	// space should not toggle in single mode
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = next.(selectorModel)
	assert.Empty(t, m.selected)
}

func TestSelectorModel_enter(t *testing.T) {
	m := selectorModel{items: []string{"a", "b"}, selected: map[int]bool{1: true}}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(selectorModel)
	assert.NotNil(t, cmd) // tea.Quit
	assert.False(t, m.aborted)
}

func TestSelectorModel_singleMode_enter(t *testing.T) {
	m := selectorModel{items: []string{"a", "b"}, selected: make(map[int]bool), single: true, cursor: 1}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(selectorModel)
	assert.True(t, m.selected[1], "enter in single mode should select cursor item")
}

func TestSelectorModel_quit(t *testing.T) {
	m := selectorModel{items: []string{"a"}, selected: make(map[int]bool)}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = next.(selectorModel)
	assert.True(t, m.aborted)

	m2 := selectorModel{items: []string{"a"}, selected: make(map[int]bool)}
	next, _ = m2.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m2 = next.(selectorModel)
	assert.True(t, m2.aborted)
}

func TestSelectorModel_view(t *testing.T) {
	m := selectorModel{
		title:    "Pick something",
		items:    []string{"alpha", "beta"},
		selected: map[int]bool{1: true},
	}
	view := m.View()
	assert.Contains(t, view, "Pick something")
	assert.Contains(t, view, "alpha")
	assert.Contains(t, view, "beta")
}
