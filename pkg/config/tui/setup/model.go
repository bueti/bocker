package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

type Model struct {
	focusIndex int
	inputs     []textinput.Model
	Username   string
	Password   string
	Done       bool
	width      int
	height     int
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, store values provided
			if s == "enter" && m.focusIndex == len(m.inputs) {
				m.Username = m.inputs[0].Value()
				m.Password = m.inputs[1].Value()
				m.Done = true
				return m, tea.Quit
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					continue
				}
				m.inputs[i].Blur()
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m Model) View() tea.View {
	if m.width == 0 {
		return tea.NewView("loading...")
	}

	boderColor := compat.AdaptiveColor{Light: lipgloss.Color("22"), Dark: lipgloss.Color("42")}
	style := lipgloss.NewStyle().
		BorderForeground(boderColor).
		BorderStyle(lipgloss.NormalBorder()).
		Width(80).
		BorderBottom(true)

	title := "\nPlease provide your Docker Username and PAT:\n"
	titleStyle := lipgloss.NewStyle().Foreground(boderColor).Bold(true)
	s := titleStyle.Render(title)

	var b strings.Builder
	for i := range m.inputs {
		b.WriteString(style.Render(m.inputs[i].View()))
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	return tea.NewView(lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Left,
		lipgloss.JoinVertical(
			lipgloss.Left,
			s,
			b.String(),
		),
	))
}
