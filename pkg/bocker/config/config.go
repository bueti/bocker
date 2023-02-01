package config

import (
	"log"
)

type config struct {
	Docker struct {
		Namespace  string
		Repository string
		Tag        string
		Username   string
		Password   string
		Host       string
	}
	DB struct {
		Name           string
		User           string
		Host           string
		Owner          string
		ExportRoles    bool
		DateTime       string
		BackupFileName string
		RolesFileName  string
	}
	TmpDir string
}

type Application struct {
	Config   config
	ErrorLog *log.Logger
	InfoLog  *log.Logger
}
