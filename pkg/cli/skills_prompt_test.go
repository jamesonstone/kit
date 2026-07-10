package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestSkillPromptSuffix_ListsCopilotOnce(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	for _, relativePath := range []string{
		"AGENTS.md",
		"CLAUDE.md",
		".github/copilot-instructions.md",
	} {
		fullPath := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(fullPath, []byte("# stub\n"), 0644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	restore := chdirForTest(t, projectRoot)
	defer restore()

	got := skillPromptSuffix()

	copilotPath := filepath.Join(projectRoot, ".github", "copilot-instructions.md")
	if count := strings.Count(got, copilotPath); count != 1 {
		t.Fatalf("expected Copilot path once in skillPromptSuffix(), got %d\n%s", count, got)
	}
	if !strings.Contains(got, copilotPath) {
		t.Fatalf("expected Copilot path in skillPromptSuffix(), got:\n%s", got)
	}
}
