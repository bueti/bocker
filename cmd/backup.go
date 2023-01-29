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
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var (
	hostName string
	dbName   string
	user     string
	repo     string
	roles    bool
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
		backupFile := fmt.Sprintf("%s_%s_backup.psql", dbName, dateTime)
		rolesFile := ""
		defer os.Remove(backupFile)
		defer os.Remove(rolesFile)

		log.Println("Creating backup...")
		pgDumpBin, err := exec.LookPath("pg_dump")
		if err == nil {
			pgDumpBin, _ = filepath.Abs(pgDumpBin)
		}
		bkpCmd := exec.Command(pgDumpBin, "-F", "c", "-U", "postgres", "-h", hostName, dbName, "-f", backupFile)
		err = bkpCmd.Run()
		if err != nil {
			log.Fatal(err)
		}

		if roles {
			log.Println("Exporting roles...")
			rolesFile = fmt.Sprintf("%s_%s_roles_backup.sql", dbName, dateTime)
			defer os.Remove(rolesFile)

			pgDumallBin, err := exec.LookPath("pg_dumpall")
			if err == nil {
				pgDumallBin, _ = filepath.Abs(pgDumallBin)
			}
			bkpCmd := exec.Command(pgDumallBin, "--clean", "--if-exists", "--no-comments", "--globals-only", fmt.Sprintf("--file=%s", rolesFile))
			err = bkpCmd.Run()
			if err != nil {
				log.Fatal(err)
			}
		}

		// create image
		log.Println("Building image...")
		dockerBin, err := exec.LookPath("docker")
		if err == nil {
			dockerBin, _ = filepath.Abs(dockerBin)
		}

		tag := fmt.Sprintf("%s:%s", repo, dateTime)
		var buildArgs []string
		if roles {
			buildArgs = []string{"build",
				"--build-arg", "backup_file=" + backupFile,
				"--build-arg", fmt.Sprintf("roles_file=%s", rolesFile),
				"-t", tag, "-f", "internal/Dockerfile.backup", "."}
		} else {
			buildArgs = []string{"build",
				"--build-arg", "backup_file=" + backupFile,
				"-t", tag, "-f", "internal/Dockerfile.backup", "."}
		}

		buildCmd := exec.Command(dockerBin, buildArgs...)
		err = buildCmd.Run()
		if err != nil {
			log.Panic(err)
		}

		// push it
		log.Println("Pushing image...")
		pushArgs := []string{"push", tag}
		pushCmd := exec.Command(dockerBin, pushArgs...)
		err = pushCmd.Run()
		if err != nil {
			log.Panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().StringVarP(&hostName, "host", "H", "localhost", "Hostname of the database host")
	backupCmd.Flags().StringVarP(&dbName, "name", "n", "", "Database name (required)")
	backupCmd.Flags().StringVarP(&user, "user", "u", "", "Database user name (required)")
	backupCmd.Flags().StringVarP(&repo, "repository", "r", "bueti/ioverlander_backup", "Repository to push image to")
	backupCmd.Flags().BoolVar(&roles, "export-roles", false, "Include roles in backup")

	backupCmd.MarkFlagRequired("name")
	backupCmd.MarkFlagRequired("user")
}
