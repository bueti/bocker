package docker

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"bocker.software-services.dev/pkg/bocker/config"
	"bocker.software-services.dev/pkg/bocker/helpers"
	"github.com/docker/docker/api/types"
)

//go:embed "Dockerfile"
var Dockerfile []byte

type DockerImage struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

// Copies a file from a running docker container to the app.Config.TmpDir
func CopyFrom(app config.Application) error {
	var outb, errb bytes.Buffer
	app.Config.DB.BackupFileName = fmt.Sprintf("%s_%s_backup.psql", app.Config.DB.SourceName, app.Config.DB.DateTime)

	dockerBin, err := exec.LookPath("docker")
	if err == nil {
		dockerBin, _ = filepath.Abs(dockerBin)
	}

	// docker cp ${DB_CONTAINER}:/${BACKUP_FILE_NAME} ${BACKUP_DIR}/
	cpArgs := []string{"cp", app.Config.Docker.ContainerID + ":/var/tmp/" + app.Config.DB.BackupFileName, app.Config.TmpDir}
	cpCmd := exec.Command(dockerBin, cpArgs...)
	cpCmd.Stdout = &outb
	cpCmd.Stderr = &errb
	err = cpCmd.Run()
	if err != nil {
		return fmt.Errorf(errb.String())
	}
	return nil
}

// Copies a file to a running docker container to /var/tmp
func CopyTo(container, filename string) error {
	var outb, errb bytes.Buffer

	dockerBin, err := exec.LookPath("docker")
	if err == nil {
		dockerBin, _ = filepath.Abs(dockerBin)
	}

	// docker cp <filename> <container id>:/var/tmp/<filename>
	cpArgs := []string{"cp", filename, container + ":/var/tmp/"}
	cpCmd := exec.Command(dockerBin, cpArgs...)
	cpCmd.Stdout = &outb
	cpCmd.Stderr = &errb
	err = cpCmd.Run()
	if err != nil {
		return fmt.Errorf(errb.String())
	}
	return nil
}

func Build(app config.Application) error {
	var outb, errb bytes.Buffer

	// write Dockerfile
	dockerfilePath := filepath.Join(app.Config.TmpDir, "Dockerfile")
	err := os.WriteFile(dockerfilePath, Dockerfile, 0755)
	if err != nil {
		return fmt.Errorf("unable to write file: %v", err)
	}

	app.Config.DB.BackupFileName = fmt.Sprintf("%s_%s_backup.psql", app.Config.DB.SourceName, app.Config.DB.DateTime)
	dockerBin, err := exec.LookPath("docker")
	if err == nil {
		dockerBin, _ = filepath.Abs(dockerBin)
	}

	var buildArgs []string
	if app.Config.DB.ExportRoles {
		buildArgs = []string{"build",
			"--build-arg", fmt.Sprintf("backup_file=%s", app.Config.DB.BackupFileName),
			"--build-arg", fmt.Sprintf("roles_file=%s", app.Config.DB.RolesFileName),
			"-t", app.Config.Docker.ImagePath,
			"-f", dockerfilePath, app.Config.TmpDir}
	} else {
		buildArgs = []string{"build",
			"--build-arg", fmt.Sprintf("backup_file=%s", app.Config.DB.BackupFileName),
			"-t", app.Config.Docker.ImagePath,
			"-f", dockerfilePath, app.Config.TmpDir}
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
	cli, err := NewClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	authStr, err := Authentication(app)
	if err != nil {
		return err
	}

	out, err := cli.ImagePush(app.Config.Context, app.Config.Docker.ImagePath, types.ImagePushOptions{RegistryAuth: authStr})
	if err != nil {
		return err
	}
	defer out.Close()

	err = ParseOutput(app, out)
	if err != nil {
		return err
	}

	return nil
}

func Pull(app config.Application) error {
	cli, err := NewClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	authStr, err := Authentication(app)
	if err != nil {
		return err
	}

	out, err := cli.ImagePull(app.Config.Context, app.Config.Docker.ImagePath, types.ImagePullOptions{RegistryAuth: authStr})
	if err != nil {
		return err
	}
	defer out.Close()

	err = ParseOutput(app, out)
	if err != nil {
		return err
	}
	return nil
}

func Unpack(app config.Application) error {
	var outb, errb bytes.Buffer
	dockerBin, err := exec.LookPath("docker")
	if err == nil {
		dockerBin, _ = filepath.Abs(dockerBin)
	}
	outputFile := "output.tar"
	outputFilePath := filepath.Join(app.Config.TmpDir, outputFile)
	saveArgs := []string{"save", app.Config.Docker.ImagePath, "--output", outputFilePath}
	saveCmd := exec.Command(dockerBin, saveArgs...)
	saveCmd.Stdout = &outb
	saveCmd.Stderr = &errb
	err = saveCmd.Run()
	if err != nil {
		app.ErrorLog.Fatal(errb.String())
	}

	// unpack layer
	manifestFile := "manifest.json"
	err = helpers.Untar(outputFilePath, manifestFile, app.Config.TmpDir)
	if err != nil {
		app.ErrorLog.Fatalf("Couldn't unpack %s", manifestFile)
	}

	// read manifest.json and extract layer with backup in it
	file, err := os.ReadFile(filepath.Join(app.Config.TmpDir, manifestFile))
	if err != nil {
		return err
	}
	var manifest []DockerImage
	err = json.Unmarshal(file, &manifest)
	if err != nil {
		return err
	}

	backupLayerTar := manifest[0].Layers[len(manifest[0].Layers)-1]

	err = helpers.Untar(filepath.Join(app.Config.TmpDir, outputFile), backupLayerTar, app.Config.TmpDir)
	if err != nil {
		return err
	}

	app.Config.DB.BackupFileName = fmt.Sprintf("%s_%s_backup.psql", app.Config.DB.SourceName, app.Config.Docker.Tag)
	err = helpers.Untar(filepath.Join(app.Config.TmpDir, backupLayerTar), app.Config.DB.BackupFileName, app.Config.TmpDir)
	if err != nil {
		return err
	}
	return nil
}
