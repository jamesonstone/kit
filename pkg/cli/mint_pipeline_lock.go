package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

// Serialize cooperating Kit Mint writers across validation, staging, replacement
// and rollback. External editors do not honor this protocol: byte/identity checks
// detect their edits at validation boundaries, not atomically with replacement.
func withMintWorkflowLock(path string, action func() error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	lockPath := path + ".kit-mint.lock"
	lock, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("cannot acquire Mint workflow lock %s; another writer or interrupted command may own it; inspect before retrying: %w", lockPath, err)
	}
	defer func() { _ = lock.Close(); _ = os.Remove(lockPath) }()
	if _, err := fmt.Fprintf(lock, "pid=%d\n", os.Getpid()); err != nil {
		return err
	}
	return action()
}
