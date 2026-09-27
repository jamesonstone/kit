package cli

import (
	"fmt"
	"os"
	"path/filepath"
	stdreflect "reflect"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
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

	// --force resets .kit.yaml only when it is named explicitly; a whole-project
	// forced reconcile must never discard project configuration.
	if configSelected && opts.force && len(targets) > 0 {
		aws := cfg.AWS
		instructionVersion := cfg.InstructionScaffoldVersion
		registry := cfg.Registry
		cfg = defaultInitConfig()
		cfg.SchemaVersion = config.CurrentSchemaVersion
		if exists && config.IsKnownInstructionScaffoldVersion(instructionVersion) {
			cfg.InstructionScaffoldVersion = instructionVersion
		}
		cfg.AWS = aws
		cfg.Registry = registry
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
	if !exists {
		cfg.InstructionScaffoldVersion = config.CurrentInstructionScaffoldVersion
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
	// Rewrite only for a change in content, which includes dropping keys Kit
	// no longer reads; formatting and comments alone never trigger a rewrite.
	if before == after || (planned == nil && sameYAMLContent(before, after)) {
		return nil, nil
	}
	return newInitRefreshFileChange(projectRoot, config.ConfigFileName, before, after, result), nil
}

func sameYAMLContent(left, right string) bool {
	var a, b any
	if yaml.Unmarshal([]byte(left), &a) != nil || yaml.Unmarshal([]byte(right), &b) != nil {
		return false
	}
	return stdreflect.DeepEqual(a, b)
}

func marshalInitRefreshConfig(cfg *config.Config) (string, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}
	return string(data), nil
}

// droppedConfigNote names the top-level keys and comments a .kit.yaml rewrite
// removes, so nothing disappears silently.
func droppedConfigNote(before, after string) string {
	var old, updated map[string]any
	if before == "" || yaml.Unmarshal([]byte(before), &old) != nil || yaml.Unmarshal([]byte(after), &updated) != nil {
		return ""
	}
	var dropped []string
	for key := range old {
		if _, ok := updated[key]; !ok {
			dropped = append(dropped, key)
		}
	}
	sort.Strings(dropped)
	var parts []string
	if len(dropped) > 0 {
		parts = append(parts, "removed keys Kit no longer reads: "+strings.Join(dropped, ", "))
	}
	for _, line := range strings.Split(before, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			parts = append(parts, "comments are not preserved when Kit rewrites the file")
			break
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return config.ConfigFileName + ": " + strings.Join(parts, "; ")
}
