package tui

import (
	"fmt"
	"os"

	"bocker.software-services.dev/pkg/config"
	"bocker.software-services.dev/pkg/db"
	"bocker.software-services.dev/pkg/docker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

func InitBackupTui(opts config.Options) error {
	App = config.Application{}
	app := App.Setup()
	App = *app
	App.Config.DB.User = opts.Username
	App.Config.DB.Host = opts.Host
	App.Config.DB.SourceName = opts.Source
	App.Config.DB.ExportRoles = opts.ExportRoles
	App.Config.Docker.ContainerID = opts.Container
	App.Config.Docker.Tag = App.Config.DB.DateTime
	App.Config.Docker.ImagePath = fmt.Sprintf("%s/%s:%s", opts.Namespace, opts.Repository, App.Config.Docker.Tag)

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		App.ErroLog.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	App.Config.TmpDir = tmpDir

	var stages = []Stage{
		{
			Name: "Creating Backup",
			Action: func() error {
				err := db.Dump(App)
				if err != nil {
					log.Error("dump failed", "err", err)
					return err
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
			IsComplete:     false,
		},
		{
			Name: "Exporting Rules",
			Action: func() error {
				if App.Config.DB.ExportRoles {
					App.InfoLog.Info("Exporting roles...")
					err := db.ExportRoles(App)
					if err != nil {
						log.Error("failed to export roles", "err", err)
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
				if App.Config.Docker.ContainerID != "" {
					err := docker.CopyFrom(App)
					if err != nil {
						log.Error("failed to copy backup from container", "err", err)
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
				err := docker.Build(App)
				if err != nil {
					log.Error("failed to building image", "err", err)
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
				err := docker.Push(App)
				if err != nil {
					log.Error("failed to push image", "err", err)
					return err
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
			IsComplete:     false,
		},
	}

	m := newModel(stages)
	_, err = tea.NewProgram(&m).Run()
	if err != nil {
		return err
	}

	return nil
}
