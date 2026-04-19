package tui

import (
	"strings"
	"time"

	"bocker.software-services.dev/pkg/logger"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Stage struct {
	Name           string
	Action         func() error
	Error          error
	IsActive       bool
	IsComplete     bool
	IsCompleteFunc func() bool
	Reset          func() error
}
type model struct {
	Error      error
	stages     []Stage
	stageIndex int
	spinner    spinner.Model
}

type startDeployMsg struct{}

func startDeployCmd() tea.Msg {
	return startDeployMsg{}
}

// runStageCmd runs the given stage's Action on bubbletea's Cmd goroutine and
// ferries the result back as a stageCompleteMsg. It does not touch model state
// directly — that's Update's job — which is what lets View read the model
// concurrently without a data race.
func runStageCmd(stage Stage, idx int) tea.Cmd {
	return func() tea.Msg {
		if stage.IsCompleteFunc() {
			return stageCompleteMsg{idx: idx}
		}
		return stageCompleteMsg{idx: idx, err: stage.Action()}
	}
}

type stageCompleteMsg struct {
	idx int
	err error
}

type errMsg struct{ err error }

// For messages that contain errors it's often handy to also implement the
// error interface on the message.
func (e errMsg) Error() string { return e.err.Error() }

func (m *model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, startDeployCmd)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case stageCompleteMsg:
		m.stages[msg.idx].Error = msg.err
		if msg.err != nil {
			m.Error = msg.err
			logger.WriteCommandLogFile(m.Error)
			return m, tea.Quit
		}
		m.stages[msg.idx].IsComplete = true
		m.stages[msg.idx].IsActive = false
		if m.stageIndex+1 >= len(m.stages) {
			return m, tea.Quit
		}
		m.stageIndex++
		m.stages[m.stageIndex].IsActive = true
		return m, runStageCmd(m.stages[m.stageIndex], m.stageIndex)

	case errMsg:
		m.Error = msg
		return m, tea.Quit

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case startDeployMsg:
		m.stages[m.stageIndex].IsActive = true
		return m, runStageCmd(m.stages[m.stageIndex], m.stageIndex)
	}

	var spinnerCmd tea.Cmd
	m.spinner, spinnerCmd = m.spinner.Update(msg)
	return m, spinnerCmd
}

func (m *model) View() tea.View {
	sb := strings.Builder{}

	for _, stage := range m.stages {
		sb.WriteString(renderCheckbox(stage) + " " + renderWorkingStatus(*m, stage) + "\n")
	}
	return tea.NewView(sb.String())
}

func newModel(stages []Stage) model {
	s := spinner.New()
	clock := spinner.Spinner{
		Frames: []string{"🕐 ", "🕑 ", "🕒 ", "🕓 ", "🕔 ", "🕕 ", "🕖 ", "🕗 ", "🕘 ", "🕙 ", "🕚 ", "🕛 "},
		FPS:    time.Second / 8, //nolint:gomnd
	}
	s.Spinner = clock
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		spinner: s,
		stages:  stages,
	}
}

func renderCheckbox(s Stage) string {
	sb := strings.Builder{}
	if s.Error != nil {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("  ❌ "))
	} else if s.IsComplete {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render("  ✅ "))
	} else if s.IsActive {
		sb.WriteString(" ")
	} else {
		sb.WriteString("  ⏳ ")
	}
	return sb.String()
}

func renderWorkingStatus(m model, s Stage) string {
	sb := strings.Builder{}
	if !s.IsComplete && s.IsActive {
		if s.Error == nil {
			sb.WriteString(m.spinner.View())
			sb.WriteString(" ")
		}
	}
	sb.WriteString(s.Name)
	return sb.String()
}
