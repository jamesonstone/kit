package usage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const staleLockAge = 5 * time.Minute

func withStoreLock(dir string, action func() error) error {
	if err := ensurePrivateDirectory(dir); err != nil {
		return err
	}
	lockPath := filepath.Join(dir, ".lock")
	if err := os.Mkdir(lockPath, 0o700); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
		info, statErr := os.Stat(lockPath)
		if statErr != nil || time.Since(info.ModTime()) <= staleLockAge {
			return fmt.Errorf("usage store is busy")
		}
		if err := os.Remove(lockPath); err != nil {
			return fmt.Errorf("usage store has a stale lock: %w", err)
		}
		if err := os.Mkdir(lockPath, 0o700); err != nil {
			return fmt.Errorf("usage store is busy")
		}
	}
	defer func() { _ = os.Remove(lockPath) }()
	return action()
}
