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
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"bocker.software-services.dev/pkg/bocker/helpers"
	"github.com/spf13/cobra"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restores a Posgres database",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		dockerBin, err := exec.LookPath("docker")
		if err == nil {
			dockerBin, _ = filepath.Abs(dockerBin)
		}

		// pull image from registry
		imagePath := fmt.Sprintf("%s/%s:%s", app.config.docker.namespace, app.config.docker.repository, app.config.docker.tag)
		pullArgs := []string{"pull", imagePath}
		pullCmd := exec.Command(dockerBin, pullArgs...)
		err = pullCmd.Run()
		if err != nil {
			app.errorLog.Fatal(err)
		}

		outputFile := "output.tar"
		defer os.Remove(outputFile)

		saveArgs := []string{"save", imagePath, "--output", outputFile}
		saveCmd := exec.Command(dockerBin, saveArgs...)
		err = saveCmd.Run()
		if err != nil {
			app.errorLog.Fatal(err)
		}

		// unpack layer
		manifestFile := "manifest.json"
		defer os.Remove(manifestFile)

		err = helpers.Untar(outputFile, manifestFile)
		if err != nil {
			app.errorLog.Fatalf("Couldn't unpack %s", manifestFile)
		}

		// read manifest.json and extract layer with backup in it
		file, err := os.ReadFile(manifestFile)
		if err != nil {
			app.errorLog.Fatal(err)
		}
		var manifest []DockerImage
		err = json.Unmarshal(file, &manifest)
		if err != nil {
			app.errorLog.Fatal(err)
		}

		backupLayerTar := manifest[0].Layers[len(manifest[0].Layers)-1]
		defer os.Remove(backupLayerTar)

		err = helpers.Untar(outputFile, backupLayerTar)
		if err != nil {
			app.errorLog.Fatal(err)
		}

		backupFile := fmt.Sprintf("%s_%s_backup.psql", app.config.db.name, app.config.docker.tag)
		defer os.Remove(backupFile)

		err = helpers.Untar(backupLayerTar, backupFile)
		if err != nil {
			app.errorLog.Fatal(err)
		}

		// restore backup
		// 1. create database
		pgsqlBin, err := exec.LookPath("psql")
		if err == nil {
			pgsqlBin, _ = filepath.Abs(pgsqlBin)
		}
		stmt := fmt.Sprintf("CREATE DATABASE %s OWNER %s ENCODING UTF8", app.config.db.name, app.config.db.owner)
		psqlArgs := []string{"-U", app.config.db.owner, "-d", "postgres", "-c", stmt}

		psqlCmd := exec.Command(pgsqlBin, psqlArgs...)
		var outb, errb bytes.Buffer
		psqlCmd.Stdout = &outb
		psqlCmd.Stderr = &errb
		err = psqlCmd.Run()
		if err != nil {
			if strings.Contains(errb.String(), "already exists") {
				app.infoLog.Println("Database already exists, skipping creation...")
			} else {
				app.errorLog.Fatal(errb.String())
			}
		}

		// 2. restore backup
		pgRestoreBin, err := exec.LookPath("pg_restore")
		if err == nil {
			pgRestoreBin, _ = filepath.Abs(pgRestoreBin)
		}
		pgRestoreArgs := []string{"-U", app.config.db.owner, "-F", "c", "-c", "-v", fmt.Sprintf("--dbname=%s", app.config.db.name), "-h", app.config.db.host, backupFile}

		pgRestoreCmd := exec.Command(pgRestoreBin, pgRestoreArgs...)
		err = pgRestoreCmd.Run()
		if err != nil {
			app.infoLog.Print(err)
		}
		app.infoLog.Println("Database successfully restored.")
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	restoreCmd.Flags().StringVarP(&app.config.db.name, "dbName", "d", "ioverlander_production", "Database name")
	restoreCmd.Flags().StringVarP(&app.config.db.owner, "dbOwner", "o", "ioverlander", "Database user")
	restoreCmd.Flags().StringVarP(&app.config.docker.namespace, "namespace", "n", "bueti", "Docker Namespace")
	restoreCmd.Flags().StringVarP(&app.config.docker.repository, "repository", "r", "", "Docker Repository")
	restoreCmd.Flags().StringVarP(&app.config.docker.tag, "tag", "t", "", "Tag of the image with the backup in it")
}

type DockerImage struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}
