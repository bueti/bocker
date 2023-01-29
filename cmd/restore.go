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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var tag string
var dbOwner string

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restores a Posgres database",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		dockerBin, err := exec.LookPath("docker")
		if err == nil {
			dockerBin, _ = filepath.Abs(dockerBin)
		}
		tarBin, err := exec.LookPath("tar")
		if err == nil {
			tarBin, _ = filepath.Abs(tarBin)
		}

		imagePath := fmt.Sprintf("%s/%s:%s", namespace, repo, tag)
		pullArgs := []string{"pull", imagePath}
		pullCmd := exec.Command(dockerBin, pullArgs...)
		pullCmd.Stdin = os.Stdin
		pullCmd.Stdout = os.Stdout
		pullCmd.Stderr = os.Stderr
		err = pullCmd.Run()
		if err != nil {
			log.Fatal(err)
		}

		outputFile := "output.tar"
		// defer os.Remove(outputFile)

		saveArgs := []string{"save", imagePath, "--output", outputFile}
		saveCmd := exec.Command(dockerBin, saveArgs...)
		saveCmd.Stdin = os.Stdin
		saveCmd.Stdout = os.Stdout
		saveCmd.Stderr = os.Stderr
		err = saveCmd.Run()
		if err != nil {
			log.Fatal(err)
		}

		// unpack layer
		// tar -xf layer.tar production_backup_${DATE}.psql
		// psqlFile := fmt.Sprintf("%s_%s_backup.psql", dbName, tag)
		unpackArgs := []string{"-xf", outputFile, "manifest.json"}
		unpackCmd := exec.Command(tarBin, unpackArgs...)
		unpackCmd.Stdin = os.Stdin
		unpackCmd.Stdout = os.Stdout
		unpackCmd.Stderr = os.Stderr
		err = unpackCmd.Run()
		if err != nil {
			log.Fatal(err)
		}

		// read manifest.json
		file, err := os.ReadFile("manifest.json")
		if err != nil {
			log.Fatal(err)
		}
		var manifest []DockerImage
		err = json.Unmarshal(file, &manifest)
		if err != nil {
			log.Fatal(err)
		}
		// first layer contains backup
		fmt.Println(manifest[0].Layers[0])
		// extract there will be the file
		// sl[len(sl)-1]
		last := len(manifest[0].Layers) - 1
		unpackArgs = []string{"-xf", outputFile, manifest[0].Layers[last]}
		unpackCmd = exec.Command(tarBin, unpackArgs...)
		unpackCmd.Stdin = os.Stdin
		unpackCmd.Stdout = os.Stdout
		unpackCmd.Stderr = os.Stderr
		err = unpackCmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		// backupFile := fmt.Sprintf("%s_%s_backup.psql", dbName, tag)
		backupFile := fmt.Sprintf("production_backup_%s.psql", tag)
		unpackArgs = []string{"-xf", manifest[0].Layers[last], backupFile}
		unpackCmd = exec.Command(tarBin, unpackArgs...)
		unpackCmd.Stdin = os.Stdin
		unpackCmd.Stdout = os.Stdout
		unpackCmd.Stderr = os.Stderr
		err = unpackCmd.Run()
		if err != nil {
			log.Fatal(err)
		}

		// os.RemoveAll(manifest[0].Layers[last])
		// os.RemoveAll("manifest.json")
		// os.RemoveAll(outputFile)

		// restore backup
		// 1. create database
		pgsqlBin, err := exec.LookPath("psql")
		if err == nil {
			pgsqlBin, _ = filepath.Abs(pgsqlBin)
		}
		// psql -U ioverlander -d postgres -c 'CREATE DATABASE ioverlander_production OWNER ioverlander ENCODING UTF8'
		stmt := fmt.Sprintf("CREATE DATABASE %s OWNER %s ENCODING UTF8", dbName, dbOwner)
		psqlArgs := []string{"-U", dbOwner, "-d", "postgres", "-c", stmt}

		psqlCmd := exec.Command(pgsqlBin, psqlArgs...)
		psqlCmd.Stdin = os.Stdin
		psqlCmd.Stdout = os.Stdout
		psqlCmd.Stderr = os.Stderr
		err = psqlCmd.Run()
		if err != nil {
			log.Print(err)
		}

		// 2. restore backup
		// -h localhost /production_backup_${DATE}.psql
		pgRestoreBin, err := exec.LookPath("pg_restore")
		if err == nil {
			pgRestoreBin, _ = filepath.Abs(pgRestoreBin)
		}
		// pg_restore -c -F c -v -U ioverlander --dbname=ioverlander_production -h localhost /production_backup_${DATE}.psql
		pgRestoreArgs := []string{"-U", dbOwner, "-F", "c", "-c", "-v", fmt.Sprintf("--dbname=%s", dbName), "-h", hostName, backupFile}

		pgRestoreCmd := exec.Command(pgRestoreBin, pgRestoreArgs...)
		pgRestoreCmd.Stdin = os.Stdin
		pgRestoreCmd.Stdout = os.Stdout
		pgRestoreCmd.Stderr = os.Stderr
		err = pgRestoreCmd.Run()
		if err != nil {
			log.Print(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	restoreCmd.Flags().StringVarP(&dbName, "dbName", "d", "ioverlander_production", "Database name")
	restoreCmd.Flags().StringVarP(&dbOwner, "dbOwner", "o", "ioverlander", "Database user")
	restoreCmd.Flags().StringVarP(&namespace, "namespace", "n", "bueti", "Docker Namespace")
	restoreCmd.Flags().StringVarP(&repo, "repository", "r", "", "Docker Repository")
	restoreCmd.Flags().StringVarP(&tag, "tag", "t", "", "Tag of the image with the backup in it")
}

type DockerImage struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}
