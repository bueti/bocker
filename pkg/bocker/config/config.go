package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

type config struct {
	Docker struct {
		Namespace   string
		Repository  string
		Tag         string
		Username    string
		Password    string
		Host        string
		ImagePath   string
		ContainerID string
	}
	DB struct {
		SourceName     string
		TargetName     string
		User           string
		Host           string
		Owner          string
		DateTime       string
		BackupFileName string
		RolesFileName  string
		ExportRoles    bool
		ImportRoles    bool
	}
	TmpDir  string
	Context context.Context
}

type Application struct {
	Config   config
	ErrorLog *log.Logger
	InfoLog  *log.Logger
}

func (app Application) Setup() (*Application, error) {
	username, ok := os.LookupEnv("DOCKER_USERNAME")
	if !ok {
		return &Application{}, fmt.Errorf("DOCKER_USERNAME not set")
	}
	app.Config.Docker.Username = username

	password, ok := os.LookupEnv("DOCKER_PAT")
	if !ok {
		return &Application{}, fmt.Errorf("DOCKER_PAT not set")
	}
	app.Config.Docker.Password = password

	host, ok := os.LookupEnv("DOCKER_HOST")
	if !ok {
		app.Config.Docker.Host = "https://hub.docker.com"
	} else {
		app.Config.Docker.Host = host
	}

	dt := time.Now()
	app.Config.DB.DateTime = dt.Format("2006-01-02_15-04-05")
	app.Config.Context = context.Background()

	return &app, nil
}
