package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"bocker.software-services.dev/pkg/backup/tui"
	"bocker.software-services.dev/pkg/config"
	"bocker.software-services.dev/pkg/docker"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
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

func List(ctx context.Context, app config.Application) error {
	c, err := docker.NewHTTPClient(app)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/v2/namespaces/%s/repositories/%s/tags", app.Config.Docker.Namespace, app.Config.Docker.Repository)
	resp, err := c.DoRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("docker hub returned status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var tags ListTagsResponse
	if err := json.Unmarshal(bodyBytes, &tags); err != nil {
		return err
	}

	columns := []table.Column{
		{Title: "ID", Width: 10},
		{Title: "Tag", Width: 20},
		{Title: "Last Updated", Width: 25},
		{Title: "Size", Width: 10},
	}

	rows := make([]table.Row, 0, len(tags.Results))
	for _, v := range tags.Results {
		size := float64(v.FullSize) / (1 << 20)
		sizeStr := fmt.Sprintf("%.2f MiB", size)

		dateTime, err := time.Parse(time.RFC3339, v.LastUpdated)
		if err != nil {
			return fmt.Errorf("cannot parse timestamp: %w", err)
		}

		rows = append(rows, []string{strconv.Itoa(v.ID), v.Name, dateTime.Format("02 Jan 2006 15:04 MST"), sizeStr})
	}

	m := tui.NewModel(columns, rows)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		return fmt.Errorf("could not start backup list tui: %w", err)
	}
	return nil
}
