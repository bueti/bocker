package docker

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"io"

	"bocker.software-services.dev/pkg/config"
	"bocker.software-services.dev/pkg/logger"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type APIClient struct {
	docker client.Client
}

type Status struct {
	Status string `json:"status"`
	ID     string `json:"id,omitempty"`
}

func NewClient() (*APIClient, error) {
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &APIClient{docker: *c}, nil
}

func (c *APIClient) Authentication(app config.Application) (string, error) {

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

func (c *APIClient) ParseOutput(app config.Application, out io.ReadCloser) error {
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
		logger.LogCommand(v.Status)
	}
	return nil
}
