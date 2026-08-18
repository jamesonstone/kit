package usage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func Directory() (string, error) {
	dir, err := config.GlobalConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "usage"), nil
}

func EffectiveSettings(projectRoot string) (Settings, error) {
	settings := Settings{
		Enabled:       true,
		GlobalEnabled: true,
		ProjectState:  "default-enabled",
		ProjectRoot:   projectRoot,
	}
	globalPath, err := config.GlobalConfigPath()
	if err != nil {
		return settings, err
	}
	if global, loadErr := config.LoadConfigFile(globalPath); loadErr == nil {
		settings.GlobalEnabled = global.IsUsageEnabled()
	} else if !errors.Is(loadErr, os.ErrNotExist) {
		return settings, loadErr
	}
	if !settings.GlobalEnabled {
		settings.Enabled = false
		settings.ProjectState = "suppressed-by-global"
		return settings, nil
	}
	if strings.TrimSpace(projectRoot) == "" {
		return settings, nil
	}
	project, loadErr := config.Load(projectRoot)
	if loadErr != nil {
		return settings, loadErr
	}
	if project.Usage != nil && project.Usage.Enabled != nil {
		if *project.Usage.Enabled {
			settings.ProjectState = "enabled"
		} else {
			settings.ProjectState = "disabled"
			settings.Enabled = false
		}
	}
	return settings, nil
}

func SetEnabled(scope, projectRoot string, enabled bool) (string, error) {
	switch scope {
	case "global":
		path, err := config.GlobalConfigPath()
		if err != nil {
			return "", err
		}
		return path, config.UpdateUsageEnabled(path, enabled)
	case "project":
		if strings.TrimSpace(projectRoot) == "" {
			return "", fmt.Errorf("project scope requires a Kit project")
		}
		path := filepath.Join(projectRoot, config.ConfigFileName)
		return path, config.UpdateUsageEnabled(path, enabled)
	default:
		return "", fmt.Errorf("usage scope must be global or project")
	}
}
