package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/plugin"
)

func applyPluginInstructionPlan(projectRoot string, cfg *config.Config, plan instructionFileWritePlan) (instructionFileWritePlan, error) {
	if !isAgentEntryPath(cfg, plan.relativePath) || plan.note != "" {
		return plan, nil
	}
	base := plan.content
	if plan.result != instructionFileCreated && plan.result != instructionFileUpdated {
		data, err := os.ReadFile(plan.absolutePath)
		if err != nil {
			return instructionFileWritePlan{}, fmt.Errorf("failed to read %s: %w", plan.relativePath, err)
		}
		base = string(data)
	}
	var ids []string
	if cfg != nil {
		ids = cfg.Plugins.Enabled
	}
	blocks, err := plugin.LoadEnabled(projectRoot, ids)
	if err != nil {
		return instructionFileWritePlan{}, err
	}
	next, err := plugin.Apply(base, blocks)
	if err != nil {
		return instructionFileWritePlan{}, fmt.Errorf("%s: %w", plan.relativePath, err)
	}
	if next == base {
		return plan, nil
	}
	pluginOnly := plan.result == instructionFileSkipped
	plan.content = next
	if pluginOnly {
		plan.result = instructionFileUpdated
		plan.note = "enabled plugin instruction blocks differ from plugins.enabled"
	}
	return plan, nil
}

func isAgentEntryPath(cfg *config.Config, relativePath string) bool {
	slash := filepath.ToSlash(relativePath)
	for _, path := range instructionFiles(cfg) {
		if filepath.ToSlash(path) == slash {
			return true
		}
	}
	return false
}
