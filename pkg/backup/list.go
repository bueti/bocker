package backup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"bocker.software-services.dev/pkg/config"
	"bocker.software-services.dev/pkg/docker"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Layer struct {
	Digest      string `json:"digest"`
	Size        int    `json:"size"`
	Instruction string `json:"instruction"`
}

type Images struct {
	Architecture string  `json:"architecture"`
	Features     string  `json:"features"`
	Variant      string  `json:"variant,omitempty"`
	Digest       string  `json:"digest"`
	Layers       []Layer `json:"layers"`
	OS           string  `json:"os"`
	OSFeatures   string  `json:"os_features"`
	OSVersion    string  `json:"os_version,omitempty"`
	Size         int     `json:"size"`
	Status       string  `json:"status"`
	LastPulled   string  `json:"last_pulled,omitempty"`
	LastPushed   string  `json:"last_pushed"`
}

type Response struct {
	ID                  int      `json:"id"`
	Images              []Images `json:"images"`
	Creator             int      `json:"creator"`
	LastUpdated         string   `json:"last_updated"`
	LastUpdater         int      `json:"last_updater"`
	LastUpdaterUsername string   `json:"last_updater_username"`
	Name                string   `json:"name"`
	Repository          int      `json:"repository"`
	FullSize            int      `json:"full_size"`
	V2                  bool     `json:"v2"`
	Status              string   `json:"status"`
	TagLastPulled       string   `json:"tag_last_pulled,omitempty"`
	TagLastPushed       string   `json:"tag_last_pushed"`
}

type ListTagsResponse struct {
	Count    int        `json:"count"`
	Next     string     `json:"next,omitempty"`
	Previous string     `json:"previous,omitempty"`
	Results  []Response `json:"results"`
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table table.Model
}

func List(app config.Application) error {
	c, err := docker.NewHTTPClient(app)
	if err != nil {
		app.ErrorLog.Fatal(err)
	}

	path := fmt.Sprintf("/v2/namespaces/%s/repositories/%s/tags", app.Config.Docker.Namespace, app.Config.Docker.Repository)
	resp, err := c.DoRequest(http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode == 200 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		var tags ListTagsResponse
		err = json.Unmarshal(bodyBytes, &tags)
		if err != nil {
			return err
		}

		columns := []table.Column{
			{Title: "ID", Width: 10},
			{Title: "Name", Width: 20},
			{Title: "Last Updated", Width: 25},
			{Title: "Size", Width: 10},
		}

		var rows []table.Row

		for _, v := range tags.Results {
			size := float64(v.FullSize) / (1 << 20)
			sizeStr := fmt.Sprintf("%.2f MiB", size)

			dateTime, err := time.Parse(time.RFC3339, v.LastUpdated)

			if err != nil {
				return fmt.Errorf("cannot parse timestamp: %v", err)
			}

			rows = append(rows, []string{strconv.Itoa(v.ID), v.Name, dateTime.Format("02 Jan 2006 15:04 MST"), sizeStr})
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(7),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)
		t.SetStyles(s)
		m := model{t}
		if _, err := tea.NewProgram(m).Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}

	} else {
		app.ErrorLog.Println(resp.StatusCode)
	}

	return nil
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}
