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
	"log"
	"os"
	"time"

	"bocker.software-services.dev/pkg/bocker/config"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var (
	app = &config.Application{
		ErrorLog: log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		InfoLog:  log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
	}

	rootCmd = &cobra.Command{
		Use:   "bocker",
		Short: "Create Postgresql backups and store them in Docker images",
		Long: `Bocker is a command line tool which creates a backup from a PostgreSQL database, 
wraps it in a Docker image, and uploads it to Docker Hub. 
Of course, Bocker will also do the reverse and restore your database from a backup in Docker Hub.`,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// This seems wrong...
	// TODO: What would be the idiomatic way to do this?
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		app.ErrorLog.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	app.Config.TmpDir = tmpDir

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&app.Config.Docker.Namespace, "Namespace", "n", "bueti", "Docker Namespace")
	rootCmd.PersistentFlags().StringVarP(&app.Config.Docker.Repository, "Repository", "r", "", "Docker Repository")

	username, ok := os.LookupEnv("DOCKER_USERNAME")
	if !ok {
		app.ErrorLog.Fatal("DOCKER_USERNAME not set")
	}
	app.Config.Docker.Username = username

	password, ok := os.LookupEnv("DOCKER_PAT")
	if !ok {
		app.ErrorLog.Fatal("DOCKER_PAT not set")
	}
	app.Config.Docker.Password = password

	host, ok := os.LookupEnv("DOCKER_HOST")
	if !ok {
		app.Config.Docker.Host = "https://hub.docker.com"
	} else {
		app.Config.Docker.Host = host
	}

	dt := time.Now()
	app.Config.DB.DateTime = dt.Format("2006-01-02_15-04-05")
}
