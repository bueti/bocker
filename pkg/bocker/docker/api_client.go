package docker

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"io"

	"bocker.software-services.dev/pkg/bocker/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type Status struct {
	Status string `json:"status"`
	ID     string `json:"id,omitempty"`
}

func NewClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

func Authentication(app config.Application) (string, error) {

	authConfig := types.AuthConfig{
		Username: app.Config.Docker.Username,
		Password: app.Config.Docker.Password,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(encodedJSON), nil
}

func ParseOutput(app config.Application, out io.ReadCloser) error {
	var stati []Status

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		var status Status
		if err := json.Unmarshal(scanner.Bytes(), &status); err != nil {
			return err
		}
		stati = append(stati, status)
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	for _, v := range stati {
		app.InfoLog.Println(v.Status)
	}
	return nil
}
