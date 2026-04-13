package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestRepoInstructionContextRows_IncludeCopilotOnceInDefaultOrder(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC

	rows := repoInstructionContextRows(projectRoot, cfg)

	want := [][]string{
		{"AGENTS.md", filepath.Join(projectRoot, "AGENTS.md")},
		{"CLAUDE.md", filepath.Join(projectRoot, "CLAUDE.md")},
		{"COPILOT", filepath.Join(projectRoot, ".github", "copilot-instructions.md")},
	}

	for i, row := range want {
		if rows[i][0] != row[0] || rows[i][1] != row[1] {
			t.Fatalf("repoInstructionContextRows()[%d] = %v, want %v", i, rows[i], row)
		}
	}

	var copilotCount int
	for _, row := range rows {
		if row[1] == filepath.Join(projectRoot, ".github", "copilot-instructions.md") {
			copilotCount++
		}
	}
	if copilotCount != 1 {
		t.Fatalf("expected one Copilot entry, got %d (%v)", copilotCount, rows)
	}
}

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
