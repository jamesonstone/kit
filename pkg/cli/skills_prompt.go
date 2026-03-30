package cli

import (
	"fmt"
	"os"
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
	paths := repoInstructionPromptPaths()

	var sb strings.Builder
	sb.WriteString("## Skills\n")
	sb.WriteString("- consult the repository instruction files for the active skills workflow before acting\n")
	for _, path := range paths {
		sb.WriteString(fmt.Sprintf("- repository instruction file: %s\n", path))
	}
	sb.WriteString("- if the work is feature-scoped, read that feature's SPEC.md and the `## SKILLS` table first\n")
	sb.WriteString("- open each referenced `SKILL.md` and use those skills during execution\n")
	sb.WriteString("- if the prompt provides explicit skill paths directly, open and use them\n")
	return sb.String()
}

func repoInstructionPromptPaths() []string {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return []string{
			"AGENTS.md",
			"CLAUDE.md",
			".github/copilot-instructions.md",
		}
	}

	cfg := config.LoadOrDefault(projectRoot)
	relativePaths := repoInstructionPaths(projectRoot, cfg)

	var existing []string
	for _, relativePath := range relativePaths {
		if document.Exists(relativePath) {
			existing = append(existing, relativePath)
		}
	}
	if len(existing) > 0 {
		return existing
	}

	return relativePaths
}

func codexHome() string {
	if home := os.Getenv("CODEX_HOME"); strings.TrimSpace(home) != "" {
		return home
	}

	userHome, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("~", ".codex")
	}

	return filepath.Join(userHome, ".codex")
}

func repoInstructionPaths(projectRoot string, cfg *config.Config) []string {
	relativePaths := instructionFiles(cfg)
	paths := make([]string, 0, len(relativePaths))
	for _, relativePath := range relativePaths {
		paths = append(paths, filepath.Join(projectRoot, relativePath))
	}
	return paths
}

func globalSkillDiscoveryInputs() []string {
	userHome, err := os.UserHomeDir()
	if err != nil {
		userHome = "~"
	}

	return []string{
		filepath.Join(userHome, ".claude", "CLAUDE.md"),
		filepath.Join(codexHome(), "AGENTS.md"),
		filepath.Join(codexHome(), "instructions.md"),
		filepath.Join(codexHome(), "skills", "*", "SKILL.md"),
	}
}
