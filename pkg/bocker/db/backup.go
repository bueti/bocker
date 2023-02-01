package db

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"

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
