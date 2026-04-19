package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tui "bocker.software-services.dev/pkg/config/tui/setup"
	tea "charm.land/bubbletea/v2"
	"github.com/adrg/xdg"
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
	TmpDir     string
	DaemonMode bool
}

type Application struct {
	Config config
}

type Username struct {
	Username string `yaml:"username,omitempty"`
}

// Setup populates runtime fields (credentials, Docker host, timestamp) on the
// Application. It mutates the receiver; call on a *Application shared with the
// rest of the program.
func (app *Application) Setup() error {
	cfg, err := GetUsername()
	if err != nil {
		return fmt.Errorf("read config: %w (try running `bocker config` to fix)", err)
	}
	if cfg.Username == "" {
		return errors.New("username not set; run `bocker config` first")
	}
	app.Config.Docker.Username = cfg.Username

	app.Config.Docker.Password, err = GetKey(AppName)
	if err != nil {
		return fmt.Errorf("read keyring: %w", err)
	}

	if host, ok := os.LookupEnv("DOCKER_HOST"); ok {
		app.Config.Docker.Host = host
	} else {
		app.Config.Docker.Host = "https://hub.docker.com"
	}

	app.Config.DB.DateTime = time.Now().Format("2006-01-02_15-04-05")
	return nil
}

// SetKey creates a new entry in the OS keyring
func SetKey(service, secret string) error {
	if err := keyring.Set(service, AppName, secret); err != nil {
		return fmt.Errorf("keyring set: %w", err)
	}
	return nil
}

// GetKey retrieves a key from the OS keyring
func GetKey(service string) (string, error) {
	if pw := os.Getenv("DOCKER_PASSWORD"); pw != "" {
		return pw, nil
	}

	secret, err := keyring.Get(service, AppName)
	if err != nil {
		return "", fmt.Errorf("keyring get: %w", err)
	}
	return secret, nil
}

// GetUsername from configuration stored on the disk
func GetUsername() (*Username, error) {
	var username Username

	if os.Getenv("DOCKER_USERNAME") != "" {
		return &Username{Username: os.Getenv("DOCKER_USERNAME")}, nil
	}

	dir, err := xdg.ConfigFile(AppName)
	if err != nil {
		return nil, err
	}
	fullPath := filepath.Join(dir, cfgFile)

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

	fullPath, err := xdg.ConfigFile(AppName)
	if err != nil {
		return err
	}
	err = os.MkdirAll(fullPath, 0700)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(fullPath, cfgFile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
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

	err = SetKey(AppName, ans.Password)
	if err != nil {
		return err
	}

	err = SetUsername(ans.Username)
	if err != nil {
		return err
	}

	return nil
}
