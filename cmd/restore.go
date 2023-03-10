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
	"fmt"
	"os"
	"path/filepath"

	"bocker.software-services.dev/pkg/db"
	"bocker.software-services.dev/pkg/docker"
	"github.com/spf13/cobra"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restores a Posgres database",
	Run: func(cmd *cobra.Command, args []string) {
		app := app.Setup()

		tmpDir, err := os.MkdirTemp("", "")
		if err != nil {
			app.ErroLog.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)
		app.Config.TmpDir = tmpDir

		app.Config.Docker.ImagePath = fmt.Sprintf("%s/%s:%s", app.Config.Docker.Namespace, app.Config.Docker.Repository, app.Config.Docker.Tag)

		err = docker.Pull(*app)
		if err != nil {
			app.ErroLog.Fatal(err.Error())
		}

		app.InfoLog.Info("Extracting backup from Docker image...")
		err = docker.Unpack(*app)
		if err != nil {
			app.ErroLog.Fatal(err)
		}

		app.InfoLog.Info("Creating database...")
		err = db.CreateDB(*app)
		if err != nil {
			app.ErroLog.Fatal(err)
		}

		backupFile := filepath.Join(app.Config.TmpDir, fmt.Sprintf("%s_%s_backup.psql", app.Config.DB.SourceName, app.Config.Docker.Tag))
		if app.Config.Docker.ContainerID != "" {
			err = docker.CopyTo(app.Config.Docker.ContainerID, backupFile)
			if err != nil {
				app.ErroLog.Fatal(err)
			}
		}

		if app.Config.DB.ImportRoles {
			rolesFile := filepath.Join(app.Config.TmpDir, fmt.Sprintf("%s_%s_roles_backup.sql", app.Config.DB.SourceName, app.Config.DB.DateTime))
			if app.Config.Docker.ContainerID != "" {
				err = docker.CopyTo(app.Config.Docker.ContainerID, rolesFile)
				if err != nil {
					app.ErroLog.Fatal(err)
				}
			}
		}

		app.InfoLog.Info("Restoring database...")
		err = db.Restore(*app)
		if err != nil {
			app.ErroLog.Fatal(err)
		}
		fmt.Println("Database successfully restored.")
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	restoreCmd.Flags().StringVarP(&app.Config.DB.Owner, "db-owner", "o", "", "Database user ")
	restoreCmd.Flags().StringVarP(&app.Config.DB.SourceName, "db-source", "s", "", "Source database name")
	restoreCmd.Flags().StringVarP(&app.Config.DB.TargetName, "db-target", "t", "", "Target database name")
	restoreCmd.Flags().StringVar(&app.Config.DB.Host, "db-host", "localhost", "Hostname of the database host")
	restoreCmd.Flags().StringVar(&app.Config.Docker.Tag, "tag", "", "Tag of the image with the backup in it")
	restoreCmd.Flags().StringVarP(&app.Config.Docker.ContainerID, "container-id", "c", "", "ID of container running PostgreSQL")
	restoreCmd.Flags().BoolVar(&app.Config.DB.ImportRoles, "import-roles", false, "Create roles from backup")

	restoreCmd.MarkFlagRequired("tag")
	restoreCmd.MarkFlagRequired("db-owner")
	restoreCmd.MarkFlagRequired("db-source")
	restoreCmd.MarkFlagRequired("db-target")
	rootCmd.MarkPersistentFlagRequired("repository")
}
