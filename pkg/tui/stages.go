package tui

import (
	"strings"
	"time"

	"bocker.software-services.dev/pkg/config"
	"bocker.software-services.dev/pkg/logger"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var App config.Application

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
	status     int
	Error      error
	stages     []Stage
	stageIndex int
	spinner    spinner.Model
}

type startDeployMsg struct{}

func startDeployCmd() tea.Msg {
	return startDeployMsg{}
}

func (m *model) runStage() tea.Msg {
	if !m.stages[m.stageIndex].IsCompleteFunc() {
		// Run the current stage, and record its result status
		m.stages[m.stageIndex].Error = m.stages[m.stageIndex].Action()
	}
	return stageCompleteMsg{}
}

type stageCompleteMsg struct{}

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
		// If we have an error, then set the error so that the views can properly update
		if m.stages[m.stageIndex].Error != nil {
			m.Error = m.stages[m.stageIndex].Error
			logger.WriteCommandLogFile(m.Error)
			return m, tea.Quit
		}
		// Otherwise, mark the current stage as complete and move to the next stage
		m.stages[m.stageIndex].IsComplete = true
		m.stages[m.stageIndex].IsActive = false
		// If we've reached the end of the defined stages, we're done
		if m.stageIndex+1 >= len(m.stages) {
			return m, tea.Quit
		}
		m.stageIndex++
		// set next stage to active
		m.stages[m.stageIndex].IsActive = true
		return m, m.runStage

	case errMsg:
		m.Error = msg
		return m, tea.Quit

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

	case startDeployMsg:
		m.stages[m.stageIndex].IsActive = true
		return m, m.runStage
	}

	var spinnerCmd tea.Cmd
	m.spinner, spinnerCmd = m.spinner.Update(msg)
	return m, spinnerCmd
}

func (m *model) View() string {
	sb := strings.Builder{}

	//sb.WriteString(fmt.Sprintf("Current stage: %s\n", m.stages[m.stageIndex].Name))

	for _, stage := range m.stages {
		sb.WriteString(renderCheckbox(stage) + " " + renderWorkingStatus(*m, stage) + "\n")
	}
	return sb.String()
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
		sb.WriteString(m.spinner.View())
		sb.WriteString(" ")
	}
	sb.WriteString(s.Name)
	return sb.String()
}
