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
bocker -c <container id> -H <host> -n <db name> -u <db user> -o <output file name>`,
	Run: func(cmd *cobra.Command, args []string) {
		dt := time.Now()
		dateTime := dt.Format("2006-01-02_15-04-05")
		backupFile := fmt.Sprintf("%s_%s_backup.psql", dbName, dateTime)
		defer os.Remove(backupFile)

		pg_bin, err := exec.LookPath("pg_dump")
		if err == nil {
			pg_bin, _ = filepath.Abs(pg_bin)
		}
		bkpCmd := exec.Command(pg_bin, "-F", "c", "-U", dbName, "-h", hostName, dbName, "-f", backupFile)
		bkpCmd.Stdin = os.Stdin
		bkpCmd.Stdout = os.Stdout
		bkpCmd.Stderr = os.Stderr
		err = bkpCmd.Run()
		if err != nil {
			log.Panic(err)
		}

		// create image
		docker_bin, err := exec.LookPath("docker")
		if err == nil {
			docker_bin, _ = filepath.Abs(docker_bin)
		}

		tag := fmt.Sprintf("%s:%s", repo, dateTime)
		buildArgs := []string{"build", "--build-arg", "backup_file=" + backupFile,
			"-t", tag, "-f", "internal/Dockerfile.backup", "."}

		buildCmd := exec.Command(docker_bin, buildArgs...)
		buildCmd.Stdin = os.Stdin
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		err = buildCmd.Run()
		if err != nil {
			log.Panic(err)
		}

		// push it
		pushArgs := []string{"push", tag}
		pushCmd := exec.Command(docker_bin, pushArgs...)
		pushCmd.Stdin = os.Stdin
		pushCmd.Stdout = os.Stdout
		pushCmd.Stderr = os.Stderr
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

	backupCmd.MarkFlagRequired("name")
	backupCmd.MarkFlagRequired("user")
}
