package tui

import (
	"context"
	"fmt"
	"os"

	"bocker.software-services.dev/pkg/config"
	"bocker.software-services.dev/pkg/db"
	"bocker.software-services.dev/pkg/docker"
	"bocker.software-services.dev/pkg/logger"
	tea "charm.land/bubbletea/v2"
	"github.com/mattn/go-isatty"
)

func InitBackupTui(ctx context.Context, app *config.Application) error {
	if err := app.Setup(); err != nil {
		return err
	}
	app.Config.Docker.Tag = app.Config.DB.DateTime
	app.Config.Docker.ImagePath = fmt.Sprintf("%s/%s:%s", app.Config.Docker.Namespace, app.Config.Docker.Repository, app.Config.Docker.Tag)

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		logger.LogCommand(err.Error())
		return err
	}
	defer os.RemoveAll(tmpDir)
	app.Config.TmpDir = tmpDir

	var stages = []Stage{
		{
			Name: "Creating Backup",
			Action: func() error {
				if err := db.Dump(ctx, *app); err != nil {
					logger.LogCommand("pg_dump failed")
					logger.LogCommand(err.Error())
					return err
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
		},
		{
			Name: "Exporting Roles",
			Action: func() error {
				if !app.Config.DB.ExportRoles {
					return nil
				}
				if err := db.ExportRoles(ctx, *app); err != nil {
					logger.LogCommand("failed to export roles")
					logger.LogCommand(err.Error())
					return err
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
		},
		{
			Name: "Copy from Container",
			Action: func() error {
				if app.Config.Docker.ContainerID == "" {
					return nil
				}
				if err := docker.CopyFrom(ctx, *app); err != nil {
					logger.LogCommand("failed to copy backup from container")
					logger.LogCommand(err.Error())
					return err
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
		},
		{
			Name: "Building Image",
			Action: func() error {
				if err := docker.Build(ctx, *app); err != nil {
					logger.LogCommand("failed to building image")
					logger.LogCommand(err.Error())
					return err
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
		},
		{
			Name: "Pushing Image",
			Action: func() error {
				if err := docker.Push(ctx, *app); err != nil {
					logger.LogCommand("failed to push image")
					logger.LogCommand(err.Error())
					return err
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
		},
	}

	m := newModel(stages)

	var opts []tea.ProgramOption
	if app.Config.DaemonMode || !isatty.IsTerminal(os.Stdout.Fd()) {
		opts = []tea.ProgramOption{tea.WithoutRenderer(), tea.WithInput(nil)}
	}
	_, err = tea.NewProgram(&m, opts...).Run()
	if err != nil {
		return fmt.Errorf("failed to run backup tui: %w", err)
	}

	return nil
}
