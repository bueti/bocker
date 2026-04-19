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
	"context"
	"os"
	"os/signal"
	"syscall"

	"bocker.software-services.dev/pkg/config"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var (
	app     = &config.Application{}
	rootCmd = &cobra.Command{
		Use:   "bocker",
		Short: "Create Postgresql backups and store them in Docker images",
		Long: `Bocker is a command line tool which creates a backup from a PostgreSQL database,
wraps it in a Docker image, and uploads it to Docker Hub. Of course, Bocker will also do the
reverse and restore your database from a backup in Docker Hub.`,
		// Returned errors are already reported by Execute; skip cobra's usage
		// banner so a failing pg_dump doesn't dump the full --help on the user.
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

// Execute wires Ctrl+C into a cancellable context and runs the root command.
// It returns whatever error the selected subcommand produced so main can decide
// the exit code.
func Execute() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&app.Config.Docker.Namespace, "namespace", "n", "bueti", "Docker Namespace")
	rootCmd.PersistentFlags().StringVarP(&app.Config.Docker.Repository, "repository", "r", "", "Docker Repository")
}
