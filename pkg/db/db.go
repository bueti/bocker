package db

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"bocker.software-services.dev/pkg/config"
	"bocker.software-services.dev/pkg/logger"
)

// identRE matches unquoted PostgreSQL identifiers: leading letter/underscore,
// then letters, digits, or underscores, up to 63 bytes (PG's NAMEDATALEN - 1).
var identRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,62}$`)

// validateIdent rejects anything that isn't a plain PostgreSQL identifier.
// We stay within the safe subset rather than supporting double-quoted names
// so operator mistakes (--db-target 'foo;drop...') can't reach psql.
func validateIdent(field, v string) error {
	if !identRE.MatchString(v) {
		return fmt.Errorf("invalid %s %q: must match %s", field, v, identRE.String())
	}
	return nil
}

// buildCmd resolves the binary and prepends `docker exec -- <container>` when
// a container ID is configured. When PGPASSWORD is set in the caller's env,
// it is forwarded into the container via `docker exec -e PGPASSWORD` (value
// not on argv); on the host path, children inherit the env automatically.
func buildCmd(containerID, tool string, args []string) (*exec.Cmd, error) {
	if containerID == "" {
		bin, err := exec.LookPath(tool)
		if err != nil {
			return nil, fmt.Errorf("%s not found: %w", tool, err)
		}
		bin, _ = filepath.Abs(bin)
		return exec.Command(bin, args...), nil
	}

	dockerBin, err := exec.LookPath("docker")
	if err != nil {
		return nil, fmt.Errorf("docker not found: %w", err)
	}
	dockerBin, _ = filepath.Abs(dockerBin)

	dockerArgs := []string{"exec"}
	if _, ok := os.LookupEnv("PGPASSWORD"); ok {
		dockerArgs = append(dockerArgs, "-e", "PGPASSWORD")
	}
	dockerArgs = append(dockerArgs, "--", containerID, tool)
	dockerArgs = append(dockerArgs, args...)
	return exec.Command(dockerBin, dockerArgs...), nil
}

// runCmd captures stderr, runs cmd, and wraps any non-zero exit with the
// underlying *exec.ExitError plus trimmed stderr.
func runCmd(cmd *exec.Cmd, tool string) (string, error) {
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	logger.LogCommand(cmd.Path + " " + strings.Join(cmd.Args[1:], " "))
	if err := cmd.Run(); err != nil {
		stderr := strings.TrimSpace(errb.String())
		if stderr == "" {
			return outb.String(), fmt.Errorf("%s failed: %w", tool, err)
		}
		return outb.String(), fmt.Errorf("%s failed: %w: %s", tool, err, stderr)
	}
	return outb.String(), nil
}

func Dump(app config.Application) error {
	if err := validateIdent("db-source", app.Config.DB.SourceName); err != nil {
		return err
	}
	if err := validateIdent("db-user", app.Config.DB.User); err != nil {
		return err
	}

	app.Config.DB.BackupFileName = fmt.Sprintf("%s_%s_backup.psql", app.Config.DB.SourceName, app.Config.DB.DateTime)
	var backupFilePath string
	if app.Config.Docker.ContainerID != "" {
		backupFilePath = filepath.Join("/var/tmp", app.Config.DB.BackupFileName)
	} else {
		backupFilePath = filepath.Join(app.Config.TmpDir, app.Config.DB.BackupFileName)
	}

	args := []string{"-F", "c", "-U", app.Config.DB.User, "-h", app.Config.DB.Host, app.Config.DB.SourceName, "-f", backupFilePath}
	cmd, err := buildCmd(app.Config.Docker.ContainerID, "pg_dump", args)
	if err != nil {
		return err
	}
	_, err = runCmd(cmd, "pg_dump")
	return err
}

func ExportRoles(app config.Application) error {
	if err := validateIdent("db-user", app.Config.DB.User); err != nil {
		return err
	}

	app.Config.DB.RolesFileName = fmt.Sprintf("%s_%s_roles_backup.sql", app.Config.DB.SourceName, app.Config.DB.DateTime)

	var rolesFilePath string
	if app.Config.Docker.ContainerID != "" {
		rolesFilePath = filepath.Join("/var/tmp/", app.Config.DB.RolesFileName)
	} else {
		rolesFilePath = filepath.Join(app.Config.TmpDir, app.Config.DB.RolesFileName)
	}

	args := []string{"-U", app.Config.DB.User, "--clean", "--if-exists", "--no-comments", "--globals-only", fmt.Sprintf("--file=%s", rolesFilePath)}
	cmd, err := buildCmd(app.Config.Docker.ContainerID, "pg_dumpall", args)
	if err != nil {
		return err
	}
	_, err = runCmd(cmd, "pg_dumpall")
	return err
}

func CreateDB(app config.Application) error {
	if err := validateIdent("db-target", app.Config.DB.TargetName); err != nil {
		return err
	}
	if err := validateIdent("db-owner", app.Config.DB.Owner); err != nil {
		return err
	}

	// Identifiers are validated above, but we still double-quote to defeat any
	// future validator regression and to match PG's own escaping conventions.
	stmt := fmt.Sprintf(`CREATE DATABASE "%s" OWNER "%s" ENCODING UTF8`,
		app.Config.DB.TargetName, app.Config.DB.Owner)
	args := []string{"-U", app.Config.DB.Owner, "-d", "postgres", "-c", stmt}

	cmd, err := buildCmd(app.Config.Docker.ContainerID, "psql", args)
	if err != nil {
		return err
	}
	if _, err := runCmd(cmd, "psql"); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			logger.LogCommand("Database already exists, skipping creation...")
			return nil
		}
		return err
	}
	return nil
}

func Restore(app config.Application) error {
	if err := validateIdent("db-source", app.Config.DB.SourceName); err != nil {
		return err
	}
	if err := validateIdent("db-target", app.Config.DB.TargetName); err != nil {
		return err
	}
	if err := validateIdent("db-owner", app.Config.DB.Owner); err != nil {
		return err
	}

	app.Config.DB.BackupFileName = fmt.Sprintf("%s_%s_backup.psql", app.Config.DB.SourceName, app.Config.Docker.Tag)
	var backupFile string
	if app.Config.Docker.ContainerID != "" {
		backupFile = filepath.Join("/var/tmp", app.Config.DB.BackupFileName)
	} else {
		backupFile = filepath.Join(app.Config.TmpDir, app.Config.DB.BackupFileName)
	}

	args := []string{
		"-U", app.Config.DB.Owner, "-F", "c", "-c", "-v",
		fmt.Sprintf("--dbname=%s", app.Config.DB.TargetName),
		"-h", app.Config.DB.Host,
		backupFile,
	}

	cmd, err := buildCmd(app.Config.Docker.ContainerID, "pg_restore", args)
	if err != nil {
		return err
	}
	if _, err := runCmd(cmd, "pg_restore"); err != nil {
		if strings.Contains(err.Error(), "errors ignored on restore") {
			logger.LogCommand("Some errors during restore where ignored.")
			logger.LogCommand(err.Error())
			return nil
		}
		return err
	}
	return nil
}
