/*
Copyright © 2023 Benjamin Buetikofer <bbu@ik.me>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"bocker.software-services.dev/pkg/bocker/docker"
	"github.com/spf13/cobra"
)

var (
	namespace string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available backups",
	Run: func(cmd *cobra.Command, args []string) {
		dockerUsername, ok := os.LookupEnv("DOCKER_USERNAME")
		if !ok {
			log.Fatal("DOCKER_USERNAME not set")
		}
		dockerPAT, ok := os.LookupEnv("DOCKER_PAT")
		if !ok {
			log.Fatal("DOCKER_PAT not set")
		}

		c := docker.NewClient(dockerUsername, dockerPAT)

		path := fmt.Sprintf("/v2/namespaces/%s/repositories/%s/tags", namespace, repo)
		resp, err := c.DoRequest(http.MethodGet, path, nil)
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode == 200 {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			var tags ListTagsResponse
			err = json.Unmarshal(bodyBytes, &tags)
			if err != nil {
				log.Fatal(err)
			}

			for _, v := range tags.Results {
				fmt.Println(v.Name, v.FullSize, v.LastUpdaterUsername)
			}
		} else {
			log.Println(resp.StatusCode)
		}
	},
}

func init() {
	backupCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&namespace, "namespace", "n", "bueti", "Docker Namespace")
	listCmd.Flags().StringVarP(&repo, "repository", "r", "ioverlander_backup", "Docker Repository")
}

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
