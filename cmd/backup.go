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
	"bocker.software-services.dev/pkg/config"
	"bocker.software-services.dev/pkg/tui"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var opts config.Options

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

		err := tui.InitTui(opts)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().StringVarP(&opts.Username, "db-user", "u", "", "Database user name (required)")
	backupCmd.Flags().StringVarP(&opts.Host, "db-host", "", "localhost", "Hostname of the database host")
	backupCmd.Flags().StringVarP(&opts.Source, "db-source", "s", "", "Source database name")
	backupCmd.Flags().StringVarP(&opts.Container, "container-id", "c", "", "ID of container running PostgreSQL")
	backupCmd.Flags().StringVarP(&opts.Namespace, "namespace", "n", "bueti", "Docker Namespace")
	backupCmd.Flags().StringVarP(&opts.Repository, "repository", "r", "", "Docker Repository")
	backupCmd.Flags().BoolVar(&opts.ExportRoles, "export-roles", false, "Include roles in backup")

	backupCmd.MarkFlagRequired("db-name")
	backupCmd.MarkFlagRequired("db-user")
	backupCmd.MarkFlagRequired("db-source")
	rootCmd.MarkPersistentFlagRequired("repository")
}
