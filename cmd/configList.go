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

	"bocker.software-services.dev/pkg/config"
	"github.com/spf13/cobra"
)

var showPassword bool

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Registry Configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetUsername()
		if err != nil {
			return err
		}
		fmt.Printf("Username: %s\n", cfg.Username)

		if !showPassword {
			fmt.Println("Password: (hidden; pass --show-password to reveal)")
			return nil
		}

		password, err := config.GetKey(config.AppName)
		if err != nil {
			return err
		}
		fmt.Printf("Password: %s\n", password)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configListCmd.Flags().BoolVar(&showPassword, "show-password", false, "Print the stored Docker Hub password to stdout")
}
