package config

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const cfgFile = "config.yaml"

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

type Credentials struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

func (app Application) Setup() *Application {
	creds, err := Read()
	if err != nil {
		log.Fatalf("Can't read configuration: %s\nTry running `bocker config` to fix the issue.", err)
	}

	if creds.Username == "" {
		log.Fatal("Username not set. Run `bocker config` first.")
	}
	app.Config.Docker.Username = creds.Username
	if creds.Password == "" {
		log.Fatal("Password not set. Run `bocker config` first.")
	}
	app.Config.Docker.Password = creds.Password

	host, ok := os.LookupEnv("DOCKER_HOST")
	if !ok {
		app.Config.Docker.Host = "https://hub.docker.com"
	} else {
		app.Config.Docker.Host = host
	}

	dt := time.Now()
	app.Config.DB.DateTime = dt.Format("2006-01-02_15-04-05")
	app.Config.Context = context.Background()

	return &app
}

// Read the Docker username and password configuration stored on the disk
func Read() (*Credentials, error) {
	var creds Credentials

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	fullPath := filepath.Join(home, ".config", "bocker", cfgFile)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Credentials{}, nil
		}
		return nil, err
	}

	err = yaml.Unmarshal(data, &creds)
	if err != nil {
		return nil, err
	}

	return &creds, nil
}

// Write the docker username and password to the disk
func Write(username, password string) error {
	creds, err := Read()
	if err != nil {
		return err
	}
	if username != "" {
		creds.Username = username
	}
	if password != "" {
		creds.Password = password
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	fullPath := filepath.Join(home, ".config", "bocker")
	err = os.MkdirAll(fullPath, os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(fullPath, cfgFile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(&creds)
	if err != nil {
		return err
	}

	if _, err := f.Write(data); err != nil {
		f.Close() // ignore error; Write error takes precedence
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	return nil
}
