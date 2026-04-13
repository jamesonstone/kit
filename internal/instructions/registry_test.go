package instructions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestInstructionRelativePathsIncludesCopilotOnce(t *testing.T) {
	cfg := config.Default()
	cfg.Agents = []string{AgentsMDPath, ClaudeMDPath, CopilotInstructionsPath}

	got := InstructionRelativePaths(cfg)
	want := []string{AgentsMDPath, ClaudeMDPath, CopilotInstructionsPath}

	if len(got) != len(want) {
		t.Fatalf("InstructionRelativePaths() len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("InstructionRelativePaths()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDetectVersionPrefersConfig(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionVerbose

	if got := DetectVersion(projectRoot, cfg); got != config.InstructionScaffoldVersionVerbose {
		t.Fatalf("DetectVersion() = %d, want %d", got, config.InstructionScaffoldVersionVerbose)
	}
}

func TestDetectVersionFallsBackToSupportDocs(t *testing.T) {
	projectRoot := t.TempDir()
	path := filepath.Join(projectRoot, "docs", "agents", "README.md")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte("# Agents Docs\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if got := DetectVersion(projectRoot, nil); got != config.InstructionScaffoldVersionTOC {
		t.Fatalf("DetectVersion() = %d, want %d", got, config.InstructionScaffoldVersionTOC)
	}
}

func TestExistingSupportDocsFiltersMissingFiles(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC

	for _, relativePath := range []string{"docs/agents/README.md", "docs/references/README.md"} {
		fullPath := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(fullPath, []byte("# stub\n"), 0644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	docs := ExistingSupportDocs(projectRoot, cfg)
	if len(docs) != 2 {
		t.Fatalf("ExistingSupportDocs() len = %d, want 2", len(docs))
	}
	if docs[0].Label != "AGENTS DOCS" {
		t.Fatalf("ExistingSupportDocs()[0].Label = %q, want %q", docs[0].Label, "AGENTS DOCS")
	}
	if docs[1].Label != "REFERENCES" {
		t.Fatalf("ExistingSupportDocs()[1].Label = %q, want %q", docs[1].Label, "REFERENCES")
	}
}
