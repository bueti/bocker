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
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup a Postgresql Database",
	Long: `This command creates a Postgresql database backup with pg_dump. 
The resulting file is wrapped in a Docker image. 
Finally, this Docker image is uploaded to a Docker registy.

Requires:
- Docker installed and configured
- pg_dump installed

Example:
bocker -H <host> -n <db name> -u <db user> -o <output file name>`,
	Run: func(cmd *cobra.Command, args []string) {
		dt := time.Now()
		dateTime := dt.Format("2006-01-02_15-04-05")
		backupFileName := fmt.Sprintf("%s_%s_backup.psql", app.config.db.name, dateTime)
		backupFilePath := filepath.Join(app.config.tmpDir, backupFileName)
		rolesFileName := ""

		app.infoLog.Println("Creating backup...")
		pgDumpBin, err := exec.LookPath("pg_dump")
		if err == nil {
			pgDumpBin, _ = filepath.Abs(pgDumpBin)
		}
		var outb, errb bytes.Buffer
		bkpCmd := exec.Command(pgDumpBin, "-F", "c", "-U", "postgres", "-h", app.config.db.host, app.config.db.name, "-f", backupFilePath)
		bkpCmd.Stdout = &outb
		bkpCmd.Stderr = &errb
		err = bkpCmd.Run()
		if err != nil {
			app.errorLog.Fatal(errb.String())
		}

		if app.config.db.exportRoles {
			app.infoLog.Println("Exporting roles...")
			rolesFileName = fmt.Sprintf("%s_%s_roles_backup.sql", app.config.db.name, dateTime)
			rolesFilePath := filepath.Join(app.config.tmpDir, rolesFileName)

			pgDumallBin, err := exec.LookPath("pg_dumpall")
			if err == nil {
				pgDumallBin, _ = filepath.Abs(pgDumallBin)
			}
			bkpCmd := exec.Command(pgDumallBin, "--clean", "--if-exists", "--no-comments", "--globals-only", fmt.Sprintf("--file=%s", rolesFilePath))
			err = bkpCmd.Run()
			if err != nil {
				app.errorLog.Fatal(err)
			}
		}

		// create image
		app.infoLog.Println("Building image...")
		dockerBin, err := exec.LookPath("docker")
		if err == nil {
			dockerBin, _ = filepath.Abs(dockerBin)
		}

		tag := fmt.Sprintf("%s/%s:%s", app.config.docker.namespace, app.config.docker.repository, dateTime)
		var buildArgs []string
		if app.config.db.exportRoles {
			buildArgs = []string{"build",
				"--build-arg", fmt.Sprintf("backup_file=%s", backupFileName),
				"--build-arg", fmt.Sprintf("roles_file=%s", rolesFileName),
				"-t", tag, "-f", "internal/Dockerfile.backup", app.config.tmpDir}
		} else {
			buildArgs = []string{"build",
				"--build-arg", fmt.Sprintf("backup_file=%s", backupFileName),
				"-t", tag, "-f", "internal/Dockerfile.backup", app.config.tmpDir}
		}

		buildCmd := exec.Command(dockerBin, buildArgs...)
		buildCmd.Stdout = &outb
		buildCmd.Stderr = &errb
		err = buildCmd.Run()
		if err != nil {
			app.errorLog.Fatal(errb.String())
		}

		// push it
		app.infoLog.Println("Pushing image...")
		pushArgs := []string{"push", tag}
		pushCmd := exec.Command(dockerBin, pushArgs...)
		pushCmd.Stdout = &outb
		pushCmd.Stderr = &errb
		err = pushCmd.Run()
		if err != nil {
			app.errorLog.Fatal(errb.String(), tag)
		}
		fmt.Printf("Published image %s\n", tag)
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().StringVarP(&app.config.db.host, "host", "", "localhost", "Hostname of the database host")
	backupCmd.Flags().StringVarP(&app.config.db.name, "name", "n", "", "Database name (required)")
	backupCmd.Flags().StringVarP(&app.config.db.user, "user", "u", "", "Database user name (required)")
	backupCmd.Flags().StringVarP(&app.config.docker.namespace, "namespace", "", "buet", "Repository to push image to")
	backupCmd.Flags().StringVarP(&app.config.docker.repository, "repository", "r", "ioverlander_backup", "Repository to push image to")
	backupCmd.Flags().BoolVar(&app.config.db.exportRoles, "export-roles", false, "Include roles in backup")

	backupCmd.MarkFlagRequired("name")
	backupCmd.MarkFlagRequired("user")
}
