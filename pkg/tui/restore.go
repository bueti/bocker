package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"bocker.software-services.dev/pkg/config"
	"bocker.software-services.dev/pkg/db"
	"bocker.software-services.dev/pkg/docker"
	"bocker.software-services.dev/pkg/logger"
	tea "charm.land/bubbletea/v2"
)

func InitRestoreTui(ctx context.Context, app *config.Application) error {
	if err := app.Setup(); err != nil {
		return err
	}
	app.Config.Docker.ImagePath = fmt.Sprintf("%s/%s:%s", app.Config.Docker.Namespace, app.Config.Docker.Repository, app.Config.Docker.Tag)

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return fmt.Errorf("create tmp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	app.Config.TmpDir = tmpDir

	var stages = []Stage{
		{
			Name: "Pull Backup Image",
			Action: func() error {
				if err := docker.Pull(ctx, *app); err != nil {
					logger.LogCommand("docker pull failed")
					logger.LogCommand(err.Error())
					return err
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
		},
		{
			Name: "Extracting backup from image",
			Action: func() error {
				if err := docker.Unpack(ctx, *app); err != nil {
					logger.LogCommand("failed to extract backup")
					logger.LogCommand(err.Error())
					return err
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
		},
		{
			Name: "Creating Database",
			Action: func() error {
				if err := db.CreateDB(ctx, *app); err != nil {
					logger.LogCommand("failed to create database")
					logger.LogCommand(err.Error())
					return err
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
		},
		{
			Name: "Copy backup to container",
			Action: func() error {
				if app.Config.Docker.ContainerID == "" {
					return nil
				}
				backupFile := filepath.Join(app.Config.TmpDir, fmt.Sprintf("%s_%s_backup.psql", app.Config.DB.SourceName, app.Config.Docker.Tag))
				if err := docker.CopyTo(ctx, app.Config.Docker.ContainerID, backupFile); err != nil {
					logger.LogCommand("failed to copy backup to container")
					logger.LogCommand(err.Error())
					return err
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
		},
		{
			Name: "Import Roles",
			Action: func() error {
				if !app.Config.DB.ImportRoles || app.Config.Docker.ContainerID == "" {
					return nil
				}
				rolesFile := filepath.Join(app.Config.TmpDir, fmt.Sprintf("%s_%s_roles_backup.sql", app.Config.DB.SourceName, app.Config.DB.DateTime))
				if err := docker.CopyTo(ctx, app.Config.Docker.ContainerID, rolesFile); err != nil {
					logger.LogCommand("failed to import roles")
					logger.LogCommand(err.Error())
					return err
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
		},
		{
			Name: "Restoring Database",
			Action: func() error {
				if err := db.Restore(ctx, *app); err != nil {
					logger.LogCommand("failed to restore database")
					logger.LogCommand(err.Error())
					return err
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
		},
	}

	m := newModel(stages)
	if _, err := tea.NewProgram(&m).Run(); err != nil {
		return fmt.Errorf("failed to run restore tui: %w", err)
	}
	return nil
}
