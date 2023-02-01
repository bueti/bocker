package docker

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"

	"bocker.software-services.dev/pkg/bocker/config"
)

func Build(app config.Application) error {
	var outb, errb bytes.Buffer

	dockerBin, err := exec.LookPath("docker")
	if err == nil {
		dockerBin, _ = filepath.Abs(dockerBin)
	}

	var buildArgs []string
	if app.Config.DB.ExportRoles {
		buildArgs = []string{"build",
			"--build-arg", fmt.Sprintf("backup_file=%s", app.Config.DB.BackupFileName),
			"--build-arg", fmt.Sprintf("roles_file=%s", app.Config.DB.RolesFileName),
			"-t", app.Config.Docker.Tag, "-f", "internal/Dockerfile.backup", app.Config.TmpDir}
	} else {
		buildArgs = []string{"build",
			"--build-arg", fmt.Sprintf("backup_file=%s", app.Config.DB.BackupFileName),
			"-t", app.Config.Docker.Tag, "-f", "internal/Dockerfile.backup", app.Config.TmpDir}
	}

	buildCmd := exec.Command(dockerBin, buildArgs...)
	buildCmd.Stdout = &outb
	buildCmd.Stderr = &errb
	err = buildCmd.Run()
	if err != nil {
		return fmt.Errorf(errb.String())
	}
	return nil
}

func Push(app config.Application) error {
	var outb, errb bytes.Buffer

	dockerBin, err := exec.LookPath("docker")
	if err == nil {
		dockerBin, _ = filepath.Abs(dockerBin)
	}

	pushArgs := []string{"push", app.Config.Docker.Tag}
	pushCmd := exec.Command(dockerBin, pushArgs...)
	pushCmd.Stdout = &outb
	pushCmd.Stderr = &errb
	err = pushCmd.Run()
	if err != nil {
		return fmt.Errorf(errb.String(), app.Config.Docker.Tag)
	}

	return nil
}
