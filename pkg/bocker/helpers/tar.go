package helpers

import (
	"os/exec"
	"path/filepath"
)

// Extracts a single file from a tar file.
func Untar(tarFile, extractFile, dir string) error {
	tarBin, err := exec.LookPath("tar")
	if err == nil {
		tarBin, _ = filepath.Abs(tarBin)
	} else {
		return err
	}
	unpackArgs := []string{"-xf", tarFile, extractFile}
	unpackCmd := exec.Command(tarBin, unpackArgs...)
	unpackCmd.Dir = dir
	err = unpackCmd.Run()
	if err != nil {
		return err
	}
	return nil
}
