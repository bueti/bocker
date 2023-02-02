package db

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"bocker.software-services.dev/pkg/bocker/config"
)

func Dump(app config.Application) error {
	var outb, errb bytes.Buffer
	app.Config.DB.BackupFileName = fmt.Sprintf("%s_%s_backup.psql", app.Config.DB.Name, app.Config.DB.DateTime)
	backupFilePath := filepath.Join(app.Config.TmpDir, app.Config.DB.BackupFileName)

	pgDumpBin, err := exec.LookPath("pg_dump")
	if err == nil {
		pgDumpBin, _ = filepath.Abs(pgDumpBin)
	} else {
		return fmt.Errorf("pg_dump not found")
	}

	bkpCmd := exec.Command(pgDumpBin, "-F", "c", "-U", "postgres", "-h", app.Config.DB.Host, app.Config.DB.Name, "-f", backupFilePath)
	bkpCmd.Stdout = &outb
	bkpCmd.Stderr = &errb
	err = bkpCmd.Run()
	if err != nil {
		return fmt.Errorf(errb.String())
	}
	return nil
}

func ExportRoles(app config.Application) error {
	var outb, errb bytes.Buffer
	app.Config.DB.RolesFileName = fmt.Sprintf("%s_%s_roles_backup.sql", app.Config.DB.Name, app.Config.DB.DateTime)
	rolesFilePath := filepath.Join(app.Config.TmpDir, app.Config.DB.RolesFileName)

	pgDumallBin, err := exec.LookPath("pg_dumpall")
	if err == nil {
		pgDumallBin, _ = filepath.Abs(pgDumallBin)
	} else {
		return fmt.Errorf("pg_dumpall not found")
	}

	bkpCmd := exec.Command(pgDumallBin, "--clean", "--if-exists", "--no-comments", "--globals-only", fmt.Sprintf("--file=%s", rolesFilePath))
	bkpCmd.Stdout = &outb
	bkpCmd.Stderr = &errb
	err = bkpCmd.Run()
	if err != nil {
		return fmt.Errorf(errb.String())
	}
	return nil
}

func CreateDB(app config.Application) error {
	var outb, errb bytes.Buffer

	pgsqlBin, err := exec.LookPath("psql")
	if err == nil {
		pgsqlBin, _ = filepath.Abs(pgsqlBin)
	} else {
		return fmt.Errorf("psql not found")
	}

	stmt := fmt.Sprintf("CREATE DATABASE %s OWNER %s ENCODING UTF8", app.Config.DB.Name, app.Config.DB.Owner)
	psqlArgs := []string{"-U", app.Config.DB.Owner, "-d", "postgres", "-c", stmt}

	psqlCmd := exec.Command(pgsqlBin, psqlArgs...)
	psqlCmd.Stdout = &outb
	psqlCmd.Stderr = &errb
	err = psqlCmd.Run()
	if err != nil {
		if strings.Contains(errb.String(), "already exists") {
			app.InfoLog.Println("Database already exists, skipping creation...")
		} else {
			return fmt.Errorf(errb.String())
		}
	}
	return nil
}

func Restore(app config.Application) error {
	var outb, errb bytes.Buffer

	pgRestoreBin, err := exec.LookPath("pg_restore")
	if err == nil {
		pgRestoreBin, _ = filepath.Abs(pgRestoreBin)
	} else {
		return fmt.Errorf("psql not found")
	}

	app.Config.DB.BackupFileName = fmt.Sprintf("%s_%s_backup.psql", app.Config.DB.Name, app.Config.Docker.Tag)
	pgRestoreArgs := []string{
		"-U", app.Config.DB.Owner, "-F", "c", "-c", "-v",
		fmt.Sprintf("--dbname=%s", app.Config.DB.Name),
		"-h", app.Config.DB.Host,
		filepath.Join(app.Config.TmpDir, app.Config.DB.BackupFileName),
	}
	pgRestoreCmd := exec.Command(pgRestoreBin, pgRestoreArgs...)
	pgRestoreCmd.Stdout = &outb
	pgRestoreCmd.Stderr = &errb
	err = pgRestoreCmd.Run()
	if err != nil {
		return fmt.Errorf(errb.String())
	}
	return nil
}
