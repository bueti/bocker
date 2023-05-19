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
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restores a Posgres database",
	Run: func(cmd *cobra.Command, args []string) {
		err := tui.InitRestoreTui(opts)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	restoreCmd.Flags().StringVarP(&opts.Owner, "db-owner", "o", "", "Database user ")
	restoreCmd.Flags().StringVarP(&opts.Source, "db-source", "s", "", "Source database name")
	restoreCmd.Flags().StringVarP(&opts.Target, "db-target", "t", "", "Target database name")
	restoreCmd.Flags().StringVar(&opts.Host, "db-host", "localhost", "Hostname of the database host")
	restoreCmd.Flags().StringVar(&opts.Tag, "tag", "", "Tag of the image with the backup in it")
	restoreCmd.Flags().StringVarP(&opts.Container, "container-id", "c", "", "ID of container running PostgreSQL")
	restoreCmd.Flags().BoolVar(&opts.ExportRoles, "import-roles", false, "Create roles from backup")
	restoreCmd.Flags().StringVarP(&opts.Namespace, "namespace", "n", "bueti", "Docker Namespace")
	restoreCmd.Flags().StringVarP(&opts.Repository, "repository", "r", "", "Docker Repository")

	restoreCmd.MarkFlagRequired("tag")
	restoreCmd.MarkFlagRequired("db-owner")
	restoreCmd.MarkFlagRequired("db-source")
	restoreCmd.MarkFlagRequired("db-target")
	rootCmd.MarkPersistentFlagRequired("repository")
}
