package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func buildExecutablePath() (string, error) {
	execPath, err := executablePath()
	if err != nil {
		return "", fmt.Errorf("failed to resolve installed binary path: %w", err)
	}
	return execPath, nil
}

func replaceExecutable(execPath string, binary []byte) error {
	info, err := os.Stat(execPath)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", execPath, err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(execPath), ".kit-upgrade-*")
	if err != nil {
		return writePathError(execPath, err)
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if _, err := tmp.Write(binary); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("failed to write replacement binary: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to finalize replacement binary: %w", err)
	}
	mode := info.Mode().Perm()
	if mode == 0 {
		mode = 0755
	}
	if err := os.Chmod(tmpPath, mode); err != nil {
		return writePathError(execPath, err)
	}
	if runtime.GOOS == "windows" {
		return replaceExecutableWindows(execPath, tmpPath)
	}
	return writePathError(execPath, os.Rename(tmpPath, execPath))
}

func replaceExecutableWindows(execPath, tmpPath string) error {
	// windows cannot rename over a running executable, so move the old binary aside first
	backupPath := execPath + ".old"
	_ = os.Remove(backupPath)
	if err := os.Rename(execPath, backupPath); err != nil {
		return writePathError(execPath, err)
	}
	if err := os.Rename(tmpPath, execPath); err != nil {
		_ = os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to replace %s on windows: %w. install manually with `%s`", execPath, err, manualInstallHint)
	}
	_ = os.Remove(backupPath)
	return nil
}

func writePathError(execPath string, err error) error {
	if err == nil {
		return nil
	}
	if os.IsPermission(err) {
		return fmt.Errorf("install path %s is not writable: %w. try rerunning with appropriate permissions or install manually with `%s`", execPath, err, manualInstallHint)
	}
	return fmt.Errorf("failed to replace %s: %w", execPath, err)
}
