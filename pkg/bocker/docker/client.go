package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Client struct {
	httpClient http.Client
	apiHost    string
	token      string
}

type AuthResp struct {
	Token string
}

func NewClient() *Client {
	username, ok := os.LookupEnv("DOCKER_USERNAME")
	if !ok {
		log.Fatal("DOCKER_USERNAME not set")
	}
	password, ok := os.LookupEnv("DOCKER_PAT")
	if !ok {
		log.Fatal("DOCKER_PAT not set")
	}

	apiHost := "https://hub.docker.com"
	c := http.Client{Timeout: 3 * time.Second}
	path := "/v2/users/login"
	body := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: username,
		Password: password,
	}
	out, err := json.Marshal(body)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, apiHost+path, bytes.NewBuffer(out))
	if err != nil {
		return nil
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode != 200 {
		if res.StatusCode == 401 {
			log.Fatalf("Authentication failed, status code: %d", res.StatusCode)
		} else {
			log.Fatalf("Docker API error, status code: %d", res.StatusCode)
		}
	}
	decoder := json.NewDecoder(res.Body)
	resp := &AuthResp{}
	err = decoder.Decode(resp)
	if err != nil {
		log.Fatal(err)
	}

	return &Client{
		httpClient: http.Client{Timeout: 3 * time.Second},
		token:      resp.Token,
		apiHost:    apiHost,
	}
}

// DoRequest makes a request to the Docker Hub API, caller is responsible to close response body
func (c *Client) DoRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.apiHost+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))
	return c.httpClient.Do(req)
}
