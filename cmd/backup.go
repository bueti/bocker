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
	"bocker.software-services.dev/pkg/tui"
	"github.com/spf13/cobra"
)

// backupOpts holds the backup-subcommand's own flag state so it can't collide
// with restore's bindings to the same config fields.
var backupOpts struct {
	DBUser, DBHost, DBSource, ContainerID string
	ExportRoles, DaemonMode               bool
}

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
	RunE: func(cmd *cobra.Command, args []string) error {
		app.Config.DB.User = backupOpts.DBUser
		app.Config.DB.Host = backupOpts.DBHost
		app.Config.DB.SourceName = backupOpts.DBSource
		app.Config.Docker.ContainerID = backupOpts.ContainerID
		app.Config.DB.ExportRoles = backupOpts.ExportRoles
		app.Config.DaemonMode = backupOpts.DaemonMode
		return tui.InitBackupTui(cmd.Context(), app)
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().StringVarP(&backupOpts.DBUser, "db-user", "u", "", "Database user name (required)")
	backupCmd.Flags().StringVar(&backupOpts.DBHost, "db-host", "localhost", "Hostname of the database host")
	backupCmd.Flags().StringVarP(&backupOpts.DBSource, "db-source", "s", "", "Source database name")
	backupCmd.Flags().StringVarP(&backupOpts.ContainerID, "container-id", "c", "", "ID of container running PostgreSQL")
	backupCmd.Flags().BoolVar(&backupOpts.ExportRoles, "export-roles", false, "Include roles in backup")
	backupCmd.Flags().BoolVarP(&backupOpts.DaemonMode, "daemon", "d", false, "Run in daemon mode (no TTY required)")

	_ = backupCmd.MarkFlagRequired("db-user")
	_ = backupCmd.MarkFlagRequired("db-source")
	_ = rootCmd.MarkPersistentFlagRequired("repository")
}
