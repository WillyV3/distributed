package sync

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/WillyV3/distributed/internal/ui"
)

// defaultExcludes are patterns to exclude from sync
var defaultExcludes = []string{
	".git",
	"node_modules",
	".DS_Store",
	"*.pyc",
	"__pycache__",
	".venv",
	"venv",
	"dist",
	"build",
	".next",
	"target",
	".terraform",
}

// Push syncs a local directory to remote host(s)
func Push(localPath string, hosts []string, dryRun bool) error {
	// Resolve to absolute path
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Verify path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", absPath)
	}

	// Convert to relative path from home for remote
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	remotePath := strings.Replace(absPath, home, "~", 1)

	// Build rsync args
	args := []string{
		"-avz",
		"--progress",
	}

	if dryRun {
		args = append(args, "--dry-run")
	}

	// Add excludes
	for _, exclude := range defaultExcludes {
		args = append(args, "--exclude", exclude)
	}

	// Sync to each host
	for _, host := range hosts {
		hostArgs := append(args, absPath+"/", host+":"+remotePath+"/")

		title := fmt.Sprintf("Syncing to %s:%s", host, remotePath)

		if err := ui.SpinCommand(title, "rsync", hostArgs...); err != nil {
			return fmt.Errorf("rsync to %s failed: %w", host, err)
		}
	}

	return nil
}

// Pull syncs from a remote host to local
func Pull(host, remotePath, localPath string) error {
	// Ensure local directory exists
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	args := []string{
		"-avz",
		"--progress",
		host + ":" + remotePath + "/",
		localPath + "/",
	}

	cmd := exec.Command("rsync", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("← Pulling from %s:%s\n", host, remotePath)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rsync from %s failed: %w", host, err)
	}

	return nil
}
