// Package logger captures the "commands" issued by the TUI so they can be
// dumped to disk if a stage fails. Adapted from
// https://github.com/zackproser/bubbletea-stages.
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	mu         sync.Mutex
	commandLog []string
)

// LogCommand appends an entry to the in-memory command log. Safe to call from
// multiple goroutines (bubbletea runs Cmds concurrently with Update).
func LogCommand(s string) {
	mu.Lock()
	defer mu.Unlock()
	commandLog = append(commandLog, s)
}

// WriteCommandLogFile flushes the command log plus the given error to a file
// in the OS temp dir and returns the path. Intended for post-mortem
// inspection when a stage fails.
func WriteCommandLogFile(failure error) (string, error) {
	path := filepath.Join(os.TempDir(), "bocker-debug.log")

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create debug log: %w", err)
	}
	defer f.Close()

	mu.Lock()
	defer mu.Unlock()

	header := "Ran at: " + time.Now().UTC().String() + "\n" +
		"******************************************************************************\n" +
		"Human legible log of steps taken and commands run up to the point of failure:\n" +
		"******************************************************************************\n"
	if _, err := f.WriteString(header); err != nil {
		return "", fmt.Errorf("write debug log: %w", err)
	}
	for _, cmd := range commandLog {
		if _, err := f.WriteString(cmd + "\n"); err != nil {
			return "", fmt.Errorf("write debug log: %w", err)
		}
	}
	footer := "^ The above command is likely the one that caused the error!\n\n\n" +
		"******************************************************************************\n" +
		"Complete log of the error that halted the deployment:\n" +
		"******************************************************************************\n\n\n" +
		failure.Error() + "\n"
	if _, err := f.WriteString(footer); err != nil {
		return "", fmt.Errorf("write debug log: %w", err)
	}
	return path, nil
}
