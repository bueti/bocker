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

	"bocker.software-services.dev/pkg/config"
	tui "bocker.software-services.dev/pkg/config/tui/setup"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set Registry Configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if username == "" || password == "" {
			if _, err := tea.NewProgram(tui.InitialModel()).Run(); err != nil {
				fmt.Printf("could not start bocker: %s\n", err)
				os.Exit(1)
			}
		} else {
			err := config.Write(username, password)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configSetCmd.Flags().StringVarP(&username, "username", "u", "", "Docker Hub Username")
	configSetCmd.Flags().StringVarP(&password, "password", "p", "", "Docker Hub Password")
}
