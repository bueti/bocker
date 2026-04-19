package docker

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"bocker.software-services.dev/pkg/config"
	"bocker.software-services.dev/pkg/logger"
	"bocker.software-services.dev/pkg/tar"
	"github.com/docker/docker/api/types/image"
)

// wrapExecErr produces an error that carries the underlying *exec.ExitError
// (so callers can inspect the exit code) plus trimmed stderr for context.
func wrapExecErr(tool string, err error, stderr string) error {
	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return fmt.Errorf("%s failed: %w", tool, err)
	}
	return fmt.Errorf("%s failed: %w: %s", tool, err, stderr)
}

//go:embed "Dockerfile"
var Dockerfile []byte

type DockerImage struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

// dockerBin resolves an absolute path to the docker binary.
func dockerBin() (string, error) {
	bin, err := exec.LookPath("docker")
	if err != nil {
		return "", fmt.Errorf("docker not found: %w", err)
	}
	bin, _ = filepath.Abs(bin)
	return bin, nil
}

// Copies a file from a running docker container to the app.Config.TmpDir
func CopyFrom(ctx context.Context, app config.Application) error {
	app.Config.DB.BackupFileName = fmt.Sprintf("%s_%s_backup.psql", app.Config.DB.SourceName, app.Config.DB.DateTime)

	bin, err := dockerBin()
	if err != nil {
		return err
	}

	// docker cp -- <container>:/var/tmp/<file> <dest>
	cpArgs := []string{"cp", "--", app.Config.Docker.ContainerID + ":/var/tmp/" + app.Config.DB.BackupFileName, app.Config.TmpDir}
	var outb, errb bytes.Buffer
	cpCmd := exec.CommandContext(ctx, bin, cpArgs...)
	cpCmd.Stdout = &outb
	cpCmd.Stderr = &errb
	if err := cpCmd.Run(); err != nil {
		return wrapExecErr("docker cp", err, errb.String())
	}
	return nil
}

// Copies a file to a running docker container to /var/tmp
func CopyTo(ctx context.Context, container, filename string) error {
	bin, err := dockerBin()
	if err != nil {
		return err
	}

	// docker cp -- <filename> <container>:/var/tmp/
	cpArgs := []string{"cp", "--", filename, container + ":/var/tmp/"}
	var outb, errb bytes.Buffer
	cpCmd := exec.CommandContext(ctx, bin, cpArgs...)
	cpCmd.Stdout = &outb
	cpCmd.Stderr = &errb
	if err := cpCmd.Run(); err != nil {
		return wrapExecErr("docker cp", err, errb.String())
	}
	return nil
}

func Build(ctx context.Context, app config.Application) error {
	dockerfilePath := filepath.Join(app.Config.TmpDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, Dockerfile, 0600); err != nil {
		return fmt.Errorf("unable to write Dockerfile: %w", err)
	}

	app.Config.DB.BackupFileName = fmt.Sprintf("%s_%s_backup.psql", app.Config.DB.SourceName, app.Config.DB.DateTime)

	bin, err := dockerBin()
	if err != nil {
		return err
	}

	buildArgs := []string{"build",
		"--build-arg", fmt.Sprintf("backup_file=%s", app.Config.DB.BackupFileName),
	}
	if app.Config.DB.ExportRoles {
		buildArgs = append(buildArgs, "--build-arg", fmt.Sprintf("roles_file=%s", app.Config.DB.RolesFileName))
	}
	buildArgs = append(buildArgs, "-t", app.Config.Docker.ImagePath, "-f", dockerfilePath, app.Config.TmpDir)

	var outb, errb bytes.Buffer
	buildCmd := exec.CommandContext(ctx, bin, buildArgs...)
	buildCmd.Stdout = &outb
	buildCmd.Stderr = &errb
	if err := buildCmd.Run(); err != nil {
		return wrapExecErr("docker build", err, errb.String())
	}
	return nil
}

func Push(ctx context.Context, app config.Application) error {
	c, err := NewClient()
	if err != nil {
		return err
	}
	defer c.docker.Close()

	authStr, err := c.Authentication(app)
	if err != nil {
		return err
	}

	out, err := c.docker.ImagePush(ctx, app.Config.Docker.ImagePath, image.PushOptions{RegistryAuth: authStr})
	if err != nil {
		return err
	}
	defer out.Close()

	return c.ParseOutput(out)
}

func Pull(ctx context.Context, app config.Application) error {
	c, err := NewClient()
	if err != nil {
		return err
	}
	defer c.docker.Close()

	authStr, err := c.Authentication(app)
	if err != nil {
		return err
	}

	out, err := c.docker.ImagePull(ctx, app.Config.Docker.ImagePath, image.PullOptions{RegistryAuth: authStr})
	if err != nil {
		return err
	}
	defer out.Close()

	return c.ParseOutput(out)
}

func Save(ctx context.Context, app config.Application, outputFile string) (string, error) {
	outputFilePath := filepath.Join(app.Config.TmpDir, outputFile)

	c, err := NewClient()
	if err != nil {
		return "", err
	}
	defer c.docker.Close()

	rc, err := c.docker.ImageSave(ctx, []string{app.Config.Docker.ImagePath})
	if err != nil {
		return "", err
	}
	defer rc.Close()

	f, err := os.Create(outputFilePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, rc); err != nil {
		return "", err
	}
	return outputFilePath, nil
}

func Unpack(ctx context.Context, app config.Application) error {
	outputFile := "output.tar"
	outputFilePath, err := Save(ctx, app, outputFile)
	if err != nil {
		return err
	}

	// unpack layer
	manifestFile := "manifest.json"
	if err := tar.Untar(ctx, outputFilePath, manifestFile, app.Config.TmpDir); err != nil {
		logger.LogCommand("Couldn't unpack file")
		logger.LogCommand(err.Error())
		return err
	}

	// read manifest.json and extract layer with backup in it
	file, err := os.ReadFile(filepath.Join(app.Config.TmpDir, manifestFile))
	if err != nil {
		return err
	}
	var manifest []DockerImage
	if err := json.Unmarshal(file, &manifest); err != nil {
		return err
	}
	if len(manifest) == 0 || len(manifest[0].Layers) == 0 {
		return fmt.Errorf("docker image manifest has no layers")
	}

	backupLayerTar := manifest[0].Layers[len(manifest[0].Layers)-1]
	if err := tar.Untar(ctx, filepath.Join(app.Config.TmpDir, outputFile), backupLayerTar, app.Config.TmpDir); err != nil {
		return err
	}

	app.Config.DB.BackupFileName = fmt.Sprintf("%s_%s_backup.psql", app.Config.DB.SourceName, app.Config.Docker.Tag)
	return tar.Untar(ctx, filepath.Join(app.Config.TmpDir, backupLayerTar), app.Config.DB.BackupFileName, app.Config.TmpDir)
}
