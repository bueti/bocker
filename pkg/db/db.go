package db

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"bocker.software-services.dev/pkg/config"
	"bocker.software-services.dev/pkg/logger"
)

func Dump(app config.Application) error {
	var err error
	var outb, errb bytes.Buffer
	var pgDumpBin string
	var backupFilePath string
	var pgDumpArgs []string

	app.Config.DB.BackupFileName = fmt.Sprintf("%s_%s_backup.psql", app.Config.DB.SourceName, app.Config.DB.DateTime)
	if app.Config.Docker.ContainerID != "" {
		backupFilePath = filepath.Join("/var/tmp", app.Config.DB.BackupFileName)
	} else {
		backupFilePath = filepath.Join(app.Config.TmpDir, app.Config.DB.BackupFileName)
	}

	pgDumpArgs = []string{"-F", "c", "-U", app.Config.DB.User, "-h", app.Config.DB.Host, app.Config.DB.SourceName, "-f", backupFilePath}

	if app.Config.Docker.ContainerID != "" {
		pgDumpBin, err = exec.LookPath("docker")
		if err == nil {
			pgDumpBin, _ = filepath.Abs(pgDumpBin)
		} else {
			return errors.New("docker not found")
		}
		pgDumpArgs = append([]string{"exec", app.Config.Docker.ContainerID, "pg_dump"}, pgDumpArgs...)
	} else {
		pgDumpBin, err = exec.LookPath("pg_dump")
		if err == nil {
			pgDumpBin, _ = filepath.Abs(pgDumpBin)
		} else {
			return err
		}
	}

	logger.LogCommand(pgDumpBin + " " + strings.Join(pgDumpArgs, " "))
	bkpCmd := exec.Command(pgDumpBin, pgDumpArgs...)
	bkpCmd.Stdout = &outb
	bkpCmd.Stderr = &errb
	err = bkpCmd.Run()
	if err != nil {
		return errors.New(errb.String())
	}
	return nil
}

func ExportRoles(app config.Application) error {
	var err error
	var outb, errb bytes.Buffer
	var rolesFilePath string
	var pgDumpallBin string
	var pgDumpallArgs []string

	app.Config.DB.RolesFileName = fmt.Sprintf("%s_%s_roles_backup.sql", app.Config.DB.SourceName, app.Config.DB.DateTime)

	if app.Config.Docker.ContainerID != "" {
		rolesFilePath = filepath.Join("/var/tmp/", app.Config.DB.RolesFileName)
	} else {
		rolesFilePath = filepath.Join(app.Config.TmpDir, app.Config.DB.RolesFileName)
	}

	pgDumpallArgs = []string{"-U", app.Config.DB.User, "--clean", "--if-exists", "--no-comments", "--globals-only", fmt.Sprintf("--file=%s", rolesFilePath)}

	if app.Config.Docker.ContainerID != "" {
		pgDumpallBin, err = exec.LookPath("docker")
		if err == nil {
			pgDumpallBin, _ = filepath.Abs(pgDumpallBin)
		} else {
			return errors.New("docker not found")
		}
		pgDumpallArgs = append([]string{"exec", app.Config.Docker.ContainerID, "pg_dumpall"}, pgDumpallArgs...)
	} else {
		pgDumpallBin, err = exec.LookPath("pg_dumpall")
		if err == nil {
			pgDumpallBin, _ = filepath.Abs(pgDumpallBin)
		} else {
			return errors.New("pg_dumpall not found")
		}
	}

	logger.LogCommand(pgDumpallBin + " " + strings.Join(pgDumpallArgs, " "))
	bkpCmd := exec.Command(pgDumpallBin, pgDumpallArgs...)
	bkpCmd.Stdout = &outb
	bkpCmd.Stderr = &errb
	err = bkpCmd.Run()
	if err != nil {
		return errors.New(errb.String())
	}
	return nil
}

func CreateDB(app config.Application) error {
	var err error
	var outb, errb bytes.Buffer
	var pgsqlBin string

	stmt := fmt.Sprintf("CREATE DATABASE %s OWNER %s ENCODING UTF8", app.Config.DB.TargetName, app.Config.DB.Owner)
	psqlArgs := []string{"-U", app.Config.DB.Owner, "-d", "postgres", "-c", stmt}

	if app.Config.Docker.ContainerID != "" {
		pgsqlBin, err = exec.LookPath("docker")
		if err == nil {
			pgsqlBin, _ = filepath.Abs(pgsqlBin)
		} else {
			return errors.New("docker not found")
		}
		psqlArgs = append([]string{"exec", app.Config.Docker.ContainerID, "psql"}, psqlArgs...)
	} else {
		pgsqlBin, err = exec.LookPath("psql")
		if err == nil {
			pgsqlBin, _ = filepath.Abs(pgsqlBin)
		} else {
			return errors.New("psql not found")
		}
	}

	logger.LogCommand(pgsqlBin + " " + strings.Join(psqlArgs, " "))
	psqlCmd := exec.Command(pgsqlBin, psqlArgs...)
	psqlCmd.Stdout = &outb
	psqlCmd.Stderr = &errb
	err = psqlCmd.Run()
	if err != nil {
		if strings.Contains(errb.String(), "already exists") {
			logger.LogCommand("Database already exists, skipping creation...")
		} else {
			return errors.New(errb.String())
		}
	}
	return nil
}

func Restore(app config.Application) error {
	var err error
	var outb, errb bytes.Buffer
	var pgRestoreBin string
	var backupFile string

	app.Config.DB.BackupFileName = fmt.Sprintf("%s_%s_backup.psql", app.Config.DB.SourceName, app.Config.Docker.Tag)

	if app.Config.Docker.ContainerID != "" {
		backupFile = filepath.Join("/var/tmp", app.Config.DB.BackupFileName)
	} else {
		backupFile = filepath.Join(app.Config.TmpDir, app.Config.DB.BackupFileName)
	}

	pgRestoreArgs := []string{
		"-U", app.Config.DB.Owner, "-F", "c", "-c", "-v",
		fmt.Sprintf("--dbname=%s", app.Config.DB.TargetName),
		"-h", app.Config.DB.Host,
		backupFile,
	}

	if app.Config.Docker.ContainerID != "" {
		pgRestoreBin, err = exec.LookPath("docker")
		if err == nil {
			pgRestoreBin, _ = filepath.Abs(pgRestoreBin)
		} else {
			return errors.New("docker not found")
		}
		pgRestoreArgs = append([]string{"exec", app.Config.Docker.ContainerID, "pg_restore"}, pgRestoreArgs...)
	} else {
		pgRestoreBin, err = exec.LookPath("pg_restore")
		if err == nil {
			pgRestoreBin, _ = filepath.Abs(pgRestoreBin)
		} else {
			return errors.New("pg_restore not found")
		}
	}

	logger.LogCommand(pgRestoreBin + " " + strings.Join(pgRestoreArgs, " "))
	pgRestoreCmd := exec.Command(pgRestoreBin, pgRestoreArgs...)
	pgRestoreCmd.Stdout = &outb
	pgRestoreCmd.Stderr = &errb
	err = pgRestoreCmd.Run()
	if err != nil {
		if strings.Contains(errb.String(), "errors ignored on restore") {
			logger.LogCommand("Some errors during restore where ignored.")
			logger.LogCommand(errb.String())
		} else {
			return errors.New(errb.String())
		}
	}
	return nil
}
