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

// GlobalConfigPath returns the canonical user settings file.
func GlobalConfigPath() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); filepath.IsAbs(dir) {
		return filepath.Join(dir, "kit", ConfigFileName), nil
	}
	dir, err := LocalDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName), nil
}

// globalConfigPaths lists user config files from highest to lowest precedence.
// Legacy files are read as fallbacks and never selected as write targets.
func globalConfigPaths() ([]string, error) {
	canonical, err := GlobalConfigPath()
	if err != nil {
		return nil, err
	}
	paths := []string{canonical}
	if home, err := homeDirectory(); err == nil {
		for _, path := range []string{
			filepath.Join(home, ".kit", ConfigFileName),
			filepath.Join(home, ".config", "kit", ConfigFileName),
		} {
			if path != canonical {
				paths = append(paths, path)
			}
		}
	}
	return paths, nil
}

// GlobalConfigLocationWarnings reports existing compatibility files outside
// the canonical user config path. It reads no settings and changes no files.
func GlobalConfigLocationWarnings() []Finding {
	paths, err := globalConfigPaths()
	if err != nil {
		return nil
	}
	canonical := paths[0]
	canonicalInfo, _ := os.Stat(canonical)
	var warnings []Finding
	for _, path := range paths[1:] {
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		message := fmt.Sprintf("Legacy Kit config found at %s. Merge its settings into %s, keeping existing canonical values, then move the legacy file aside.", path, canonical)
		if canonicalInfo != nil && os.SameFile(canonicalInfo, info) {
			message = fmt.Sprintf("Canonical Kit config %s aliases legacy config %s. Create a separate canonical file with those settings before moving the legacy file aside.", canonical, path)
		}
		warnings = append(warnings, Finding{
			Field: "global_config", Severity: FindingWarning, Message: message,
		})
	}
	return warnings
}

// GlobalConfigWritePath returns the canonical path after checking it is not
// an alias for a fallback config file.
func GlobalConfigWritePath() (string, error) {
	paths, err := globalConfigPaths()
	if err != nil {
		return "", err
	}
	return paths[0], checkGlobalConfigWrite(paths)
}

func checkGlobalConfigWrite(paths []string) error {
	canonical := paths[0]
	target, err := configDestination(canonical)
	if err != nil {
		return fmt.Errorf("cannot resolve global config destination %s: %w", canonical, err)
	}
	info, err := os.Stat(canonical)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, fallback := range paths[1:] {
		other, err := configDestination(fallback)
		if os.IsNotExist(err) {
			continue // A dangling fallback is not a settings file.
		}
		if err != nil {
			return err
		}
		old, err := os.Stat(fallback)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if target == other || (info != nil && old != nil && os.SameFile(info, old)) {
			return fmt.Errorf("global config path %s aliases fallback %s; use a separate canonical file before writing", canonical, fallback)
		}
	}
	return nil
}

// configDestination resolves aliases in the whole path, including ancestors
// of a file that has not been created. Reject dangling symlinks.
func configDestination(path string) (string, error) {
	target, err := filepath.EvalSymlinks(path)
	if !os.IsNotExist(err) {
		return target, err
	}
	if _, statErr := os.Lstat(path); !os.IsNotExist(statErr) {
		if statErr == nil {
			return "", err
		}
		return "", statErr
	}
	parent := filepath.Dir(path)
	if parent == path {
		return "", err
	}
	target, err = configDestination(parent)
	if err != nil {
		return "", err
	}
	return filepath.Join(target, filepath.Base(path)), nil
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
