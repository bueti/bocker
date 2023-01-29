package helpers

import (
	"os/exec"
	"path/filepath"
)

func Untar(tarFile, extractFile string) error {
	tarBin, err := exec.LookPath("tar")
	if err == nil {
		tarBin, _ = filepath.Abs(tarBin)
	} else {
		return err
	}
	unpackArgs := []string{"-xf", tarFile, extractFile}
	unpackCmd := exec.Command(tarBin, unpackArgs...)
	err = unpackCmd.Run()
	if err != nil {
		return err
	}
	return nil
}
