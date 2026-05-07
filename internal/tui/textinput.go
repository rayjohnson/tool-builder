package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type textInputModel struct {
	label   string
	input   textinput.Model
	aborted bool
}

func (m textInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m textInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.aborted = true
			return m, tea.Quit
		case "enter":
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m textInputModel) View() string {
	return titleStyle.Render(m.label) + "\n" + m.input.View() + "\n"
}

// Input shows a single-line text input and returns what the user typed.
// Returns an empty string if the user cancels with Esc or ctrl+c.
func Input(label, placeholder string) (string, error) {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Width = 60
	ti.Focus()

	m := textInputModel{label: label, input: ti}
	p := tea.NewProgram(m, tea.WithInput(os.Stdin), tea.WithOutput(os.Stdout))
	final, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("text input error: %w", err)
	}
	result, ok := final.(textInputModel)
	if !ok {
		return "", fmt.Errorf("unexpected model type from bubbletea")
	}
	if result.aborted {
		return "", nil
	}
	return result.input.Value(), nil
}
