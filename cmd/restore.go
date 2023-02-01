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
		imagePath := fmt.Sprintf("%s/%s:%s", app.Config.Docker.Namespace, app.Config.Docker.Repository, app.Config.Docker.Tag)
		app.InfoLog.Printf("Pulling image (%s) from registry...", imagePath)
		pullArgs := []string{"pull", imagePath}
		pullCmd := exec.Command(dockerBin, pullArgs...)
		err = pullCmd.Run()
		if err != nil {
			app.ErrorLog.Fatal(err)
		}

		app.InfoLog.Println("Extracting backup from Docker Image...")
		outputFile := "output.tar"
		outputFilePath := filepath.Join(app.Config.TmpDir, outputFile)
		saveArgs := []string{"save", imagePath, "--output", outputFilePath}
		saveCmd := exec.Command(dockerBin, saveArgs...)
		err = saveCmd.Run()
		if err != nil {
			app.ErrorLog.Fatal(err)
		}

		// unpack layer
		manifestFile := "manifest.json"
		err = helpers.Untar(outputFilePath, manifestFile, app.Config.TmpDir)
		if err != nil {
			app.ErrorLog.Fatalf("Couldn't unpack %s", manifestFile)
		}

		// read manifest.json and extract layer with backup in it
		file, err := os.ReadFile(filepath.Join(app.Config.TmpDir, manifestFile))
		if err != nil {
			app.ErrorLog.Fatal(err)
		}
		var manifest []DockerImage
		err = json.Unmarshal(file, &manifest)
		if err != nil {
			app.ErrorLog.Fatal(err)
		}

		backupLayerTar := manifest[0].Layers[len(manifest[0].Layers)-1]

		err = helpers.Untar(filepath.Join(app.Config.TmpDir, outputFile), backupLayerTar, app.Config.TmpDir)
		if err != nil {
			app.ErrorLog.Fatal(err)
		}

		backupFile := fmt.Sprintf("%s_%s_backup.psql", app.Config.DB.Name, app.Config.Docker.Tag)

		err = helpers.Untar(filepath.Join(app.Config.TmpDir, backupLayerTar), backupFile, app.Config.TmpDir)
		if err != nil {
			app.ErrorLog.Fatal(err)
		}

		// restore backup
		app.InfoLog.Println("Creating database")
		pgsqlBin, err := exec.LookPath("psql")
		if err == nil {
			pgsqlBin, _ = filepath.Abs(pgsqlBin)
		}
		stmt := fmt.Sprintf("CREATE DATABASE %s OWNER %s ENCODING UTF8", app.Config.DB.Name, app.Config.DB.Owner)
		psqlArgs := []string{"-U", app.Config.DB.Owner, "-d", "postgres", "-c", stmt}

		psqlCmd := exec.Command(pgsqlBin, psqlArgs...)
		var outb, errb bytes.Buffer
		psqlCmd.Stdout = &outb
		psqlCmd.Stderr = &errb
		err = psqlCmd.Run()
		if err != nil {
			if strings.Contains(errb.String(), "already exists") {
				app.InfoLog.Println("Database already exists, skipping creation...")
			} else {
				app.ErrorLog.Fatal(errb.String())
			}
		}

		// 2. restore backup
		app.InfoLog.Println("Restoring database")
		pgRestoreBin, err := exec.LookPath("pg_restore")
		if err == nil {
			pgRestoreBin, _ = filepath.Abs(pgRestoreBin)
		}
		pgRestoreArgs := []string{"-U", app.Config.DB.Owner, "-F", "c", "-c", "-v", fmt.Sprintf("--dbname=%s", app.Config.DB.Name), "-h", app.Config.DB.Host, filepath.Join(app.Config.TmpDir, backupFile)}

		pgRestoreCmd := exec.Command(pgRestoreBin, pgRestoreArgs...)
		err = pgRestoreCmd.Run()
		if err != nil {
			app.InfoLog.Print(err)
		}
		app.InfoLog.Println("Database successfully restored.")
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	restoreCmd.Flags().StringVarP(&app.Config.DB.Name, "dbName", "d", "ioverlander_production", "Database name")
	restoreCmd.Flags().StringVarP(&app.Config.DB.Owner, "dbOwner", "o", "ioverlander", "Database user")
	restoreCmd.Flags().StringVarP(&app.Config.Docker.Namespace, "Namespace", "n", "bueti", "Docker Namespace")
	restoreCmd.Flags().StringVarP(&app.Config.Docker.Repository, "Repository", "r", "", "Docker Repository")
	restoreCmd.Flags().StringVarP(&app.Config.Docker.Tag, "tag", "t", "", "Tag of the image with the backup in it")
}

type DockerImage struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}
