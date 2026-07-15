package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

func appendSkillPromptSuffix(prompt string) string {
	trimmed := strings.TrimRight(prompt, "\n")
	suffix := skillPromptSuffix()
	if trimmed == "" {
		return suffix
	}
	return trimmed + "\n\n" + suffix
}

func skillPromptSuffix() string {
	paths, version := repoInstructionPromptPaths()

	var sb strings.Builder
	sb.WriteString("## Skills\n")
	if config.UsesInstructionSupportDocs(version) {
		sb.WriteString("- Treat repository instruction entrypoints as routing maps; start at `docs/agents/README.md` when present.\n")
	} else {
		sb.WriteString("- Read the repository instruction entrypoints before acting.\n")
	}
	if len(paths) > 0 {
		sb.WriteString(fmt.Sprintf("- Entrypoints: %s.\n", strings.Join(paths, ", ")))
	}
	sb.WriteString("- Load only the repo docs and skills needed for the current decision; repo-local guidance precedes secondary global inputs.\n")
	sb.WriteString("- For feature work, use canonical front matter `skills` (legacy `## SKILLS` only as fallback), and open every selected or explicitly provided `SKILL.md`.\n")
	return sb.String()
}

func repoInstructionPromptPaths() ([]string, int) {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return []string{
			"AGENTS.md",
			"CLAUDE.md",
			".github/copilot-instructions.md",
		}, config.DefaultInstructionScaffoldVersion
	}

	cfg := config.LoadOrDefault(projectRoot)
	version := detectInstructionScaffoldVersion(projectRoot, cfg)
	relativePaths := repoInstructionPaths(projectRoot, cfg)

	var existing []string
	for _, relativePath := range relativePaths {
		if document.Exists(relativePath) {
			existing = append(existing, relativePath)
		}
	}
	if len(existing) > 0 {
		return existing, version
	}

	return relativePaths, version
}

func repoInstructionPaths(projectRoot string, cfg *config.Config) []string {
	relativePaths := instructionFiles(cfg)
	paths := make([]string, 0, len(relativePaths))
	for _, relativePath := range relativePaths {
		paths = append(paths, filepath.Join(projectRoot, relativePath))
	}
	return paths
}
