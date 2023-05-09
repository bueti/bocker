package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	tui "bocker.software-services.dev/pkg/config/tui/setup"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

const AppName = "bocker"
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
	Config  config
	InfoLog log.Logger
	ErroLog log.Logger
}

type Username struct {
	Username string `yaml:"username,omitempty"`
}

func (app Application) Setup() *Application {
	cfg, err := GetUsername()
	if err != nil {
		log.Fatal("Can't read configuration. Try running `bocker config` to fix the issue.", "err", err)
	}

	if cfg.Username == "" {
		log.Fatal("Username not set. Run `bocker config` first.")
	}
	app.Config.Docker.Username = cfg.Username

	app.Config.Docker.Password, err = GetKey(AppName)
	if err != nil {
		os.Exit(1)
	}

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

// SetKey creates a new entry in the OS keyring
func SetKey(service, secret string) error {
	err := keyring.Set(service, AppName, secret)
	if err != nil {
		log.Error("failed to fetch secret", "err", err)
		return err
	}

	return nil
}

// GetKey retrieves a key from the OS keyring
func GetKey(service string) (string, error) {
	// get password
	secret, err := keyring.Get(service, AppName)
	if err != nil {
		log.Error("failed to fetch key", "err", err)
		return "", err
	}

	return secret, nil
}

// GetUsername from configuration stored on the disk
func GetUsername() (*Username, error) {
	var username Username

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	fullPath := filepath.Join(home, ".config", "bocker", cfgFile)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Username{}, nil
		}
		return nil, err
	}

	err = yaml.Unmarshal(data, &username)
	if err != nil {
		return nil, err
	}

	return &username, nil
}

// SetUsername writes the docker username to the disk
func SetUsername(username string) error {
	creds, err := GetUsername()
	if err != nil {
		return err
	}
	if username != "" {
		creds.Username = username
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
		f.Close() // ignore error; SetUsername error takes precedence
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

// ConfigTui starts the Bubbletea Configuration TUI
func ConfigTui() error {
	finalModel, err := tea.NewProgram(tui.InitialModel()).Run()
	if err != nil {
		return err
	}

	ans := finalModel.(tui.Model)

	if !ans.Done {
		return nil
	}

	SetKey(AppName, ans.Password)
	err = SetUsername(ans.Username)
	if err != nil {
		return err
	}

	return nil

}
