package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const GlobalConfigDirName = "kit"

type Prompt struct {
	Content     string `yaml:"content"`
	Description string `yaml:"description,omitempty"`
}

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

func GlobalConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve home directory: %w", err)
	}
	if home == "" {
		return "", fmt.Errorf("failed to resolve home directory: HOME is empty")
	}

	return filepath.Join(home, ".config", GlobalConfigDirName), nil
}

func GlobalConfigPath() (string, error) {
	configDir, err := GlobalConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, ConfigFileName), nil
}
