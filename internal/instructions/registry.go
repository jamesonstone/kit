// Package instructions names the agent entry files Kit manages.
package instructions

import "github.com/jamesonstone/kit/v3/internal/config"

const (
	AgentsMDPath            = "AGENTS.md"
	ClaudeMDPath            = "CLAUDE.md"
	CopilotInstructionsPath = ".github/copilot-instructions.md"
)

// InstructionRelativePaths lists the configured agent entry files plus the
// Copilot instructions, which every project carries.
func InstructionRelativePaths(cfg *config.Config) []string {
	if cfg == nil {
		cfg = config.Default()
	}
	files := make([]string, 0, len(cfg.Agents)+1)
	for _, file := range cfg.Agents {
		files = appendUnique(files, file)
	}
	return appendUnique(files, CopilotInstructionsPath)
}

func appendUnique(items []string, value string) []string {
	for _, existing := range items {
		if existing == value {
			return items
		}
	}
	return append(items, value)
}
