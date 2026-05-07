package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type confirmModel struct {
	question   string
	defaultYes bool
	answer     bool
	done       bool
}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "y", "Y":
			m.answer = true
			m.done = true
			return m, tea.Quit
		case "n", "N", "ctrl+c", "esc":
			m.answer = false
			m.done = true
			return m, tea.Quit
		case "enter":
			m.answer = m.defaultYes
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	if m.done {
		answer := hintStyle.Render("No")
		if m.answer {
			answer = checkStyle.Render("Yes")
		}
		return titleStyle.Render(m.question) + " → " + answer + "\n"
	}
	hint := "[y/N]"
	if m.defaultYes {
		hint = "[Y/n]"
	}
	return titleStyle.Render(m.question) + " " + hintStyle.Render(hint) + " "
}

// Confirm shows a yes/no prompt and returns the user's answer.
// A single keypress is sufficient — no Enter needed.
// Returns false if the user cancels with Esc or ctrl+c.
func Confirm(question string, defaultYes bool) (bool, error) {
	m := confirmModel{question: question, defaultYes: defaultYes}
	p := tea.NewProgram(m, tea.WithInput(os.Stdin), tea.WithOutput(os.Stdout))
	final, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("confirm error: %w", err)
	}
	result, ok := final.(confirmModel)
	if !ok {
		return false, fmt.Errorf("unexpected model type from bubbletea")
	}
	return result.answer, nil
}
