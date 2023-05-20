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

func InitRestoreTui(app *config.Application) error {
	app = app.Setup()
	app.Config.Docker.ImagePath = fmt.Sprintf("%s/%s:%s", app.Config.Docker.Namespace, app.Config.Docker.Repository, app.Config.Docker.Tag)

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		log.Error(err)
	}
	defer os.RemoveAll(tmpDir)
	app.Config.TmpDir = tmpDir

	var stages = []Stage{
		{
			Name: "Pull Backup Image",
			Action: func() error {
				err := docker.Pull(*app)
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
				err := docker.Unpack(*app)
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
				err := db.CreateDB(*app)
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
				backupFile := filepath.Join(app.Config.TmpDir, fmt.Sprintf("%s_%s_backup.psql", app.Config.DB.SourceName, app.Config.Docker.Tag))
				if app.Config.Docker.ContainerID != "" {
					err = docker.CopyTo(app.Config.Docker.ContainerID, backupFile)
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
				if app.Config.DB.ImportRoles {
					rolesFile := filepath.Join(app.Config.TmpDir, fmt.Sprintf("%s_%s_roles_backup.sql", app.Config.DB.SourceName, app.Config.DB.DateTime))
					if app.Config.Docker.ContainerID != "" {
						err = docker.CopyTo(app.Config.Docker.ContainerID, rolesFile)
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
				err := db.Restore(*app)
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
