package cli

import (
	"strings"

	"github.com/jamesonstone/kit/internal/config"
)

func defaultInitLoopAgentConfig() config.LoopAgentConfig {
	return config.LoopAgentConfig{
		Command: "codex",
		Args: []string{
			"--ask-for-approval",
			"never",
			"exec",
			"--model",
			"gpt-5.5",
			"--sandbox",
			"workspace-write",
			"--ignore-user-config",
			"--color",
			"never",
			"-",
		},
	}
}

func ensureInitLoopReviewConfig(cfg *config.Config) bool {
	if cfg == nil {
		return false
	}

	changed := false
	if cfg.Loop.MinConfidence <= 0 {
		cfg.Loop.MinConfidence = 95
		changed = true
	}
	if cfg.Loop.MaxIterations <= 0 {
		cfg.Loop.MaxIterations = 10
		changed = true
	}
	if shouldBackfillInitLoopAgent(cfg.Loop.Agent) {
		cfg.Loop.Agent = defaultInitLoopAgentConfig()
		changed = true
	}
	return changed
}

func isMissingInitLoopAgentCommand(command string) bool {
	switch strings.TrimSpace(command) {
	case "", "your-agent":
		return true
	default:
		return false
	}
}

func shouldBackfillInitLoopAgent(agent config.LoopAgentConfig) bool {
	if isMissingInitLoopAgentCommand(agent.Command) {
		return true
	}
	return isGeneratedInitLoopAgentConfig(agent) && !sameLoopAgentConfig(agent, defaultInitLoopAgentConfig())
}

func isGeneratedInitLoopAgentConfig(agent config.LoopAgentConfig) bool {
	if strings.TrimSpace(agent.Command) != "codex" {
		return false
	}
	generatedArgs := [][]string{
		{
			"--ask-for-approval",
			"never",
			"exec",
			"--sandbox",
			"workspace-write",
			"--ignore-user-config",
			"--color",
			"never",
			"-",
		},
		{
			"--ask-for-approval",
			"never",
			"exec",
			"--sandbox",
			"workspace-write",
			"--color",
			"never",
			"-",
		},
		{
			"exec",
			"--sandbox",
			"workspace-write",
			"--ask-for-approval",
			"never",
			"--color",
			"never",
			"-",
		},
	}
	for _, args := range generatedArgs {
		if sameStringSlice(agent.Args, args) {
			return true
		}
	}
	return false
}

func sameLoopAgentConfig(left, right config.LoopAgentConfig) bool {
	return strings.TrimSpace(left.Command) == strings.TrimSpace(right.Command) && sameStringSlice(left.Args, right.Args)
}

func sameStringSlice(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
