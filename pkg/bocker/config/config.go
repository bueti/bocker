package config

import (
	"log"
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
	TmpDir string
}

type Application struct {
	Config   config
	ErrorLog *log.Logger
	InfoLog  *log.Logger
}
