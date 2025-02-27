package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"bocker.software-services.dev/pkg/config"
)

type HTTPClient struct {
	httpClient http.Client
	apiHost    string
	token      string
}

type AuthResp struct {
	Token string
}

func NewHTTPClient(app config.Application) (*HTTPClient, error) {
	if strings.HasPrefix(app.Config.Docker.Password, "dckr_oat") {
		return nil, fmt.Errorf("cannot use a docker organization token to list repositories")
	}

	c := http.Client{Timeout: 3 * time.Second}
	path := "/v2/users/login"
	body := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: app.Config.Docker.Username,
		Password: app.Config.Docker.Password,
	}
	out, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, app.Config.Docker.Host+path, bytes.NewBuffer(out))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		if res.StatusCode == 401 {
			return nil, fmt.Errorf("authentication failed, status code: %d", res.StatusCode)
		} else {
			return nil, fmt.Errorf("docker API error, status code: %d", res.StatusCode)
		}
	}
	decoder := json.NewDecoder(res.Body)
	resp := &AuthResp{}
	err = decoder.Decode(resp)
	if err != nil {
		return nil, err
	}

	return &HTTPClient{
		httpClient: http.Client{Timeout: 3 * time.Second},
		token:      resp.Token,
		apiHost:    app.Config.Docker.Host,
	}, nil
}

// DoRequest makes a request to the Docker Hub API, caller is responsible to close response body
func (c *HTTPClient) DoRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.apiHost+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))
	return c.httpClient.Do(req)
}
