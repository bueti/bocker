package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedColor = lipgloss.AdaptiveColor{Light: "236", Dark: "248"}
	blurredColor = lipgloss.AdaptiveColor{Light: "238", Dark: "246"}

	focusedStyle = lipgloss.NewStyle().Foreground(focusedColor)
	blurredStyle = lipgloss.NewStyle().Foreground(blurredColor)
	cursorStyle  = focusedStyle.Copy()
	noStyle      = lipgloss.NewStyle()

	focusedButton = focusedStyle.Copy().Bold(true).Render("[ Save ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Save"))
)

func InitialModel() Model {
	m := Model{
		inputs: make([]textinput.Model, 2),
	}

	var t textinput.Model

	for i := range m.inputs {
		t = textinput.New()
		t.CursorStyle = cursorStyle
		t.CharLimit = 255
		t.Prompt = ""

		switch i {
		case 0:
			t.Placeholder = "Username"
			t.TextStyle = focusedStyle
			t.Focus()
		case 1:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		m.inputs[i] = t
	}

	return m
}
