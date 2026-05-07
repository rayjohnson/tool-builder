package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pmezard/go-difflib/difflib"
)

type diffState int

const (
	stateViewing  diffState = iota
	stateFeedback diffState = iota
)

var (
	styleAdded   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleRemoved = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleHunk    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	styleHeader  = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	styleDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

type diffModel struct {
	viewport  viewport.Model
	input     textinput.Model
	state     diffState
	filename  string
	rendered  string
	result    string
	done      bool
	ready     bool
	height    int
	width     int
}

func newDiffModel(filename, rendered string) diffModel {
	ti := textinput.New()
	ti.Placeholder = "Type feedback..."
	ti.Width = 60

	return diffModel{
		filename: filename,
		rendered: rendered,
		input:    ti,
		state:    stateViewing,
	}
}

func (m diffModel) Init() tea.Cmd {
	return nil
}

func (m diffModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		vpHeight := max(msg.Height-6, 1)
		if !m.ready {
			m.viewport = viewport.New(msg.Width, vpHeight)
			m.viewport.SetContent(m.rendered)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = vpHeight
		}
		return m, nil

	case tea.KeyMsg:
		if m.state == stateFeedback {
			switch msg.String() {
			case "enter":
				val := m.input.Value()
				if val == "" {
					m.result = "reject"
				} else {
					m.result = val
				}
				m.done = true
				return m, tea.Quit
			case "esc":
				m.state = stateViewing
				m.input.Reset()
				m.input.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
		}

		// stateViewing
		switch msg.String() {
		case "y", "enter":
			m.result = "accept"
			m.done = true
			return m, tea.Quit
		case "n", "q", "esc":
			m.result = "reject"
			m.done = true
			return m, tea.Quit
		case "f":
			m.state = stateFeedback
			m.input.Focus()
			return m, textinput.Blink
		case "up", "k":
			m.viewport.ScrollUp(1)
		case "down", "j":
			m.viewport.ScrollDown(1)
		case "pgup":
			m.viewport.HalfPageUp()
		case "pgdown":
			m.viewport.HalfPageDown()
		case "home":
			m.viewport.GotoTop()
		case "end":
			m.viewport.GotoBottom()
		}
		return m, nil
	}

	return m, nil
}

func (m diffModel) View() string {
	titleBar := titleStyle.Render(" show_diff: " + m.filename + " ")
	titleBar += "\n"

	if !m.ready {
		return titleBar + "Initializing...\n"
	}

	var footer string
	if m.state == stateFeedback {
		footer = "\n  Feedback: " + m.input.View() + "  [enter] submit  [esc] cancel  "
	} else {
		footer = "\n  [y] accept  [n] reject  [f] feedback  [↑↓] scroll  "
	}

	return titleBar + m.viewport.View() + footer
}

func computeUnifiedDiff(filename, oldContent, newContent string) string {
	ud := difflib.UnifiedDiff{
		A:        difflib.SplitLines(oldContent),
		B:        difflib.SplitLines(newContent),
		FromFile: filename + " (original)",
		ToFile:   filename + " (proposed)",
		Context:  3,
	}
	text, _ := difflib.GetUnifiedDiffString(ud)
	return strings.TrimRight(text, "\n")
}

func renderDiff(raw string) string {
	lines := strings.Split(raw, "\n")
	var sb strings.Builder
	for i, line := range lines {
		var rendered string
		switch {
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			rendered = styleHeader.Render(line)
		case strings.HasPrefix(line, "@@"):
			rendered = styleHunk.Render(line)
		case strings.HasPrefix(line, "+"):
			rendered = styleAdded.Render(line)
		case strings.HasPrefix(line, "-"):
			rendered = styleRemoved.Render(line)
		default:
			rendered = styleDim.Render(line)
		}
		sb.WriteString(rendered)
		if i < len(lines)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// ShowDiff displays a scrollable colored diff between oldContent and newContent.
// filename is shown in the header.
// Returns:
//
//	"accept"       — user pressed y or enter
//	"reject"       — user pressed n, q, or esc (or aborted)
//	"<feedback>"   — user pressed f, typed text, and pressed enter
func ShowDiff(filename, oldContent, newContent string) (string, error) {
	raw := computeUnifiedDiff(filename, oldContent, newContent)
	if raw == "" {
		return "accept", nil
	}
	rendered := renderDiff(raw)
	m := newDiffModel(filename, rendered)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return "", err
	}
	result := final.(diffModel).result //nolint:forcetypeassert // we always set diffModel as the model
	if result == "" {
		return "reject", nil
	}
	return result, nil
}
