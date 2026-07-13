package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"gopkg.in/yaml.v3"
)

func initRefreshConfig(
	projectRoot string,
	opts initRefreshOptions,
	targets map[string]bool,
) (*config.Config, *initRefreshFileChange, error) {
	cfg := defaultInitConfig()
	configSelected := initRefreshTargetMatches(targets, config.ConfigFileName)
	shouldTouchConfig := len(targets) == 0 || configSelected
	path := filepath.Join(projectRoot, config.ConfigFileName)
	exists := config.Exists(projectRoot)
	var before string
	var inspection config.Inspection

	if exists {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read %s: %w", config.ConfigFileName, err)
		}
		before = string(data)
		existing, currentInspection, err := config.LoadWithInspection(projectRoot)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load %s: %w", config.ConfigFileName, err)
		}
		if currentInspection.SchemaState == config.SchemaStateNewer {
			return nil, nil, fmt.Errorf("%s", currentInspection.Findings[0].Message)
		}
		cfg = existing
		inspection = currentInspection
	}

	if configSelected && opts.force {
		aws := cfg.AWS
		cfg = defaultInitConfig()
		cfg.SchemaVersion = config.CurrentSchemaVersion
		cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
		cfg.AWS = aws
		after, err := marshalInitRefreshConfig(cfg)
		if err != nil {
			return nil, nil, err
		}
		result := instructionFileCreated
		if exists {
			result = instructionFileUpdated
		}
		return cfg, newInitRefreshFileChange(projectRoot, config.ConfigFileName, before, after, result), nil
	}

	configChanged := false
	if inspection.NeedsSchemaMigration() {
		cfg.SchemaVersion = config.CurrentSchemaVersion
		configChanged = true
	}
	if cfg.InstructionScaffoldVersion != config.InstructionScaffoldVersionTOC {
		cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
		configChanged = true
	}
	if ensureInitLoopReviewConfig(cfg) {
		configChanged = true
	}
	if configChanged && shouldTouchConfig {
		after, err := marshalInitRefreshConfig(cfg)
		if err != nil {
			return nil, nil, err
		}
		result := instructionFileUpdated
		if !exists {
			result = instructionFileCreated
		}
		return cfg, newInitRefreshFileChange(projectRoot, config.ConfigFileName, before, after, result), nil
	}

	if !exists && shouldTouchConfig {
		after, err := marshalInitRefreshConfig(cfg)
		if err != nil {
			return nil, nil, err
		}
		return cfg, newInitRefreshFileChange(projectRoot, config.ConfigFileName, before, after, instructionFileCreated), nil
	}

	return cfg, nil, nil
}

func finalizeInitRefreshConfigChange(projectRoot string, cfg *config.Config, planned *initRefreshFileChange) (*initRefreshFileChange, error) {
	before := ""
	result := instructionFileCreated
	if planned != nil {
		before = planned.before
		result = planned.result
	} else if config.Exists(projectRoot) {
		data, err := os.ReadFile(filepath.Join(projectRoot, config.ConfigFileName))
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", config.ConfigFileName, err)
		}
		before = string(data)
		result = instructionFileUpdated
	}

	after, err := marshalInitRefreshConfig(cfg)
	if err != nil {
		return nil, err
	}
	if before == after {
		return nil, nil
	}
	return newInitRefreshFileChange(projectRoot, config.ConfigFileName, before, after, result), nil
}

func marshalInitRefreshConfig(cfg *config.Config) (string, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}
	return string(data), nil
}
