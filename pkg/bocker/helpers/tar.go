package helpers

import (
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
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
	if output, err := unpackCmd.CombinedOutput(); err != nil {
		return errors.Wrap(err, string(output))
	}
	return nil
}
