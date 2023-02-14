package backup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"bocker.software-services.dev/pkg/config"
	"bocker.software-services.dev/pkg/docker"
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

		for _, v := range tags.Results {
			fmt.Println(v.Name, v.FullSize, v.LastUpdaterUsername)
		}
	} else {
		app.ErrorLog.Println(resp.StatusCode)
	}

	return nil
}
