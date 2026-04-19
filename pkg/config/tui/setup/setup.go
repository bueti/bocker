package tui

import (
	"fmt"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

var (
	focusedColor = compat.AdaptiveColor{Light: lipgloss.Color("236"), Dark: lipgloss.Color("248")}
	blurredColor = compat.AdaptiveColor{Light: lipgloss.Color("238"), Dark: lipgloss.Color("246")}

	focusedStyle = lipgloss.NewStyle().Foreground(focusedColor)
	blurredStyle = lipgloss.NewStyle().Foreground(blurredColor)
	noStyle      = lipgloss.NewStyle()

	focusedButton = focusedStyle.Bold(true).Render("[ Save ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Save"))
)

func InitialModel() Model {
	m := Model{
		inputs: make([]textinput.Model, 2),
	}

	styles := textinput.DefaultStyles(compat.HasDarkBackground)
	styles.Focused.Text = focusedStyle
	styles.Focused.Prompt = focusedStyle
	styles.Blurred.Text = noStyle
	styles.Blurred.Prompt = noStyle

	var t textinput.Model

	for i := range m.inputs {
		t = textinput.New()
		t.SetStyles(styles)
		t.CharLimit = 255
		t.Prompt = ""

		switch i {
		case 0:
			t.Placeholder = "Username"
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
