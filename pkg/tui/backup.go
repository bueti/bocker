package tui

import (
	"fmt"
	"os"

	"bocker.software-services.dev/pkg/config"
	"bocker.software-services.dev/pkg/db"
	"bocker.software-services.dev/pkg/docker"
	"bocker.software-services.dev/pkg/logger"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
)

func InitBackupTui(app *config.Application) error {
	app = app.Setup()
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
				err := db.Dump(*app)
				if err != nil {
					logger.LogCommand("pg_dump failed")
					logger.LogCommand(err.Error())
					return err
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
			IsComplete:     false,
		},
		{
			Name: "Exporting Roles",
			Action: func() error {
				if app.Config.DB.ExportRoles {
					err := db.ExportRoles(*app)
					if err != nil {
						logger.LogCommand("failed to export roles")
						logger.LogCommand(err.Error())
						return err
					}
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
			IsComplete:     false,
		},
		{
			Name: "Copy from Container",
			Action: func() error {
				if app.Config.Docker.ContainerID != "" {
					err := docker.CopyFrom(*app)
					if err != nil {
						logger.LogCommand("failed to copy backup from container")
						logger.LogCommand(err.Error())
						return err
					}
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
			IsComplete:     false,
		},
		{
			Name: "Building Image",
			Action: func() error {
				err := docker.Build(*app)
				if err != nil {
					logger.LogCommand("failed to building image")
					logger.LogCommand(err.Error())
					return err
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
			IsComplete:     false,
		},
		{
			Name: "Pushing Image",
			Action: func() error {
				err := docker.Push(*app)
				if err != nil {
					logger.LogCommand("failed to push image")
					logger.LogCommand(err.Error())
					return err
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
			IsComplete:     false,
		},
	}

	m := newModel(stages)

	var opts []tea.ProgramOption
	if app.Config.DaemonMode || !isatty.IsTerminal(os.Stdout.Fd()) {
		opts = []tea.ProgramOption{tea.WithoutRenderer()}
	}
	_, err = tea.NewProgram(&m, opts...).Run()
	if err != nil {
		return err
	}

	return nil
}
