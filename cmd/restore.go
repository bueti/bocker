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

var restoreOpts struct {
	DBOwner, DBSource, DBTarget, DBHost, Tag, ContainerID string
	ImportRoles                                           bool
}

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore a Postgresql database",
	RunE: func(cmd *cobra.Command, args []string) error {
		app.Config.DB.Owner = restoreOpts.DBOwner
		app.Config.DB.SourceName = restoreOpts.DBSource
		app.Config.DB.TargetName = restoreOpts.DBTarget
		app.Config.DB.Host = restoreOpts.DBHost
		app.Config.Docker.Tag = restoreOpts.Tag
		app.Config.Docker.ContainerID = restoreOpts.ContainerID
		app.Config.DB.ImportRoles = restoreOpts.ImportRoles
		return tui.InitRestoreTui(cmd.Context(), app)
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	restoreCmd.Flags().StringVarP(&restoreOpts.DBOwner, "db-owner", "o", "", "Database user ")
	restoreCmd.Flags().StringVarP(&restoreOpts.DBSource, "db-source", "s", "", "Source database name")
	restoreCmd.Flags().StringVarP(&restoreOpts.DBTarget, "db-target", "t", "", "Target database name")
	restoreCmd.Flags().StringVar(&restoreOpts.DBHost, "db-host", "localhost", "Hostname of the database host")
	restoreCmd.Flags().StringVar(&restoreOpts.Tag, "tag", "", "Tag of the image with the backup in it")
	restoreCmd.Flags().StringVarP(&restoreOpts.ContainerID, "container-id", "c", "", "ID of container running PostgreSQL")
	restoreCmd.Flags().BoolVar(&restoreOpts.ImportRoles, "import-roles", false, "Create roles from backup")

	_ = restoreCmd.MarkFlagRequired("tag")
	_ = restoreCmd.MarkFlagRequired("db-owner")
	_ = restoreCmd.MarkFlagRequired("db-source")
	_ = restoreCmd.MarkFlagRequired("db-target")
	_ = rootCmd.MarkPersistentFlagRequired("repository")
}
