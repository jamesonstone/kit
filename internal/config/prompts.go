package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func FindProjectRootOptional() (string, bool, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false, fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		if Exists(dir) {
			return dir, true, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false, nil
		}
		dir = parent
	}
}

// GlobalConfigPath returns the default settings file shared by Kit projects.
func GlobalConfigPath() (string, error) {
	home, err := homeDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".kit", ConfigFileName), nil
}

// LocalDataDir remains separate from configuration so existing usage records
// stay in their original location.
func LocalDataDir() (string, error) {
	home, err := homeDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "kit"), nil
}

func homeDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve home directory: %w", err)
	}
	if home == "" {
		return "", fmt.Errorf("failed to resolve home directory: HOME is empty")
	}
	return home, nil
}
