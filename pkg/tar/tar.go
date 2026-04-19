package tar

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Untar extracts a single file from a tar file into dir.
func Untar(ctx context.Context, tarFile, extractFile, dir string) error {
	tarBin, err := exec.LookPath("tar")
	if err != nil {
		return fmt.Errorf("tar not found: %w", err)
	}
	tarBin, _ = filepath.Abs(tarBin)

	unpackCmd := exec.CommandContext(ctx, tarBin, "-xf", tarFile, extractFile)
	unpackCmd.Dir = dir
	output, err := unpackCmd.CombinedOutput()
	if err != nil {
		out := strings.TrimSpace(string(output))
		if out == "" {
			return fmt.Errorf("tar failed: %w", err)
		}
		return fmt.Errorf("tar failed: %w: %s", err, out)
	}
	return nil
}
