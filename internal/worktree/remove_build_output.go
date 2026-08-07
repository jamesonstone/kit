package worktree

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	ignoredBuildOutputDirName  = "bin"
	preserveIgnoredBuildOutput = false
	discardIgnoredBuildOutput  = true
)

func inspectIgnoredRootBuildOutput(
	worktreePath string,
	status string,
) (string, string, error) {
	lines := strings.Split(status, "\n")
	kept := lines[:0]
	found := false
	for _, line := range lines {
		if isIgnoredRootBuildOutputStatus(line) {
			found = true
			continue
		}
		if line != "" {
			kept = append(kept, line)
		}
	}
	if !found {
		return strings.Join(kept, "\n"), "", nil
	}

	path := filepath.Join(worktreePath, ignoredBuildOutputDirName)
	if err := requireIgnoredRootBuildOutputDirectory(path); err != nil {
		return "", "", err
	}
	return strings.Join(kept, "\n"), path, nil
}

func isIgnoredRootBuildOutputStatus(line string) bool {
	return line == "!! "+ignoredBuildOutputDirName+"/"
}

func requireIgnoredRootBuildOutputDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("inspect ignored build output %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf(
			"%s is not a regular root build-output directory; refusing removal",
			path,
		)
	}
	return nil
}

func (a *App) removeIgnoredRootBuildOutput(path string) error {
	if err := requireIgnoredRootBuildOutputDirectory(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if err := a.removeAll(path); err != nil {
		return fmt.Errorf("remove ignored build output %s: %w", path, err)
	}
	return nil
}
