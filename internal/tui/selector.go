package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true)
	hintStyle   = lipgloss.NewStyle().Faint(true)
	checkStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
)

type selectorModel struct {
	title    string
	items    []string
	selected map[int]bool
	cursor   int
	single   bool
	aborted  bool
}

func (m selectorModel) Init() tea.Cmd { return nil }

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.aborted = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case " ":
			if !m.single {
				m.selected[m.cursor] = !m.selected[m.cursor]
			}
		case "a":
			if !m.single {
				for i := range m.items {
					m.selected[i] = true
				}
			}
		case "n":
			if !m.single {
				m.selected = make(map[int]bool)
			}
		case "enter":
			if m.single {
				m.selected[m.cursor] = true
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m selectorModel) View() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render(m.title))
	sb.WriteString("\n")
	if m.single {
		sb.WriteString(hintStyle.Render("↑/↓: move  enter: select  q: cancel"))
	} else {
		sb.WriteString(hintStyle.Render("↑/↓: move  space: toggle  a: all  n: none  enter: confirm  q: cancel"))
	}
	sb.WriteString("\n\n")

	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("> ")
		}

		var line string
		if m.single {
			line = cursor + item
		} else {
			box := "[ ] "
			if m.selected[i] {
				box = checkStyle.Render("[x] ")
			}
			line = cursor + box + item
		}
		sb.WriteString(line + "\n")
	}
	return sb.String()
}

// Select shows an interactive selection list and returns the items the user chose.
// If single is true, the user picks exactly one item and it returns immediately on enter.
// Returns nil (not an empty slice) if the user aborted with q or ctrl+c.
func Select(title string, items []string, single bool) ([]string, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no items to select from")
	}

	m := selectorModel{
		title:    title,
		items:    items,
		selected: make(map[int]bool),
		single:   single,
	}

	p := tea.NewProgram(m, tea.WithInput(os.Stdin), tea.WithOutput(os.Stdout))
	final, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("selector error: %w", err)
	}

	result, ok := final.(selectorModel)
	if !ok {
		return nil, fmt.Errorf("unexpected model type from bubbletea")
	}
	if result.aborted {
		return nil, nil
	}

	var chosen []string
	for i, item := range items {
		if result.selected[i] {
			chosen = append(chosen, item)
		}
	}
	return chosen, nil
}
