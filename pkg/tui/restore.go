package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"bocker.software-services.dev/pkg/config"
	"bocker.software-services.dev/pkg/db"
	"bocker.software-services.dev/pkg/docker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

func InitRestoreTui(opts config.Options) error {
	App = config.Application{}
	app := App.Setup()
	App = *app
	App.Config.DB.Owner = opts.Owner
	App.Config.DB.Host = opts.Host
	App.Config.DB.SourceName = opts.Source
	App.Config.DB.TargetName = opts.Target
	App.Config.DB.ExportRoles = opts.ExportRoles
	App.Config.Docker.ContainerID = opts.Container
	App.Config.Docker.Tag = opts.Tag
	App.Config.Docker.ImagePath = fmt.Sprintf("%s/%s:%s", opts.Namespace, opts.Repository, App.Config.Docker.Tag)

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		log.Error(err)
	}
	defer os.RemoveAll(tmpDir)
	App.Config.TmpDir = tmpDir

	var stages = []Stage{
		{
			Name: "Pull Backup Image",
			Action: func() error {
				err := docker.Pull(App)
				if err != nil {
					log.Error("docker pull failed", "err", err)
					return err
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
			IsComplete:     false,
		},
		{
			Name: "Extracting backup from image",
			Action: func() error {
				err := docker.Unpack(App)
				if err != nil {
					log.Error("failed to extract backup", "err", err)
					return err
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
			IsComplete:     false,
		},
		{
			Name: "Creating Database",
			Action: func() error {
				err := db.CreateDB(App)
				if err != nil {
					log.Error("failed to create database", "err", err)
					return err
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
			IsComplete:     false,
		},
		{
			Name: "Copy backup to container",
			Action: func() error {
				backupFile := filepath.Join(app.Config.TmpDir, fmt.Sprintf("%s_%s_backup.psql", App.Config.DB.SourceName, App.Config.Docker.Tag))
				if App.Config.Docker.ContainerID != "" {
					err = docker.CopyTo(App.Config.Docker.ContainerID, backupFile)
					if err != nil {
						log.Error("failed to copy backup to container", "err", err)
						return err
					}
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
			IsComplete:     false,
		},
		{
			Name: "Import Roles",
			Action: func() error {
				if App.Config.DB.ImportRoles {
					rolesFile := filepath.Join(App.Config.TmpDir, fmt.Sprintf("%s_%s_roles_backup.sql", App.Config.DB.SourceName, App.Config.DB.DateTime))
					if App.Config.Docker.ContainerID != "" {
						err = docker.CopyTo(App.Config.Docker.ContainerID, rolesFile)
						if err != nil {
							log.Error("failed to import roles", "err", err)
							return err
						}
					}
				}
				return nil
			},
			IsCompleteFunc: func() bool { return false },
			IsComplete:     false,
		},
		{
			Name: "Restoring Database",
			Action: func() error {
				err := db.Restore(App)
				if err != nil {
					log.Error("failed to restore database", "err", err)
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
