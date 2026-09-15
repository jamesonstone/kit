package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/templates"
)

func TestNamingAppendOnlyMigrationOverridesLegacyTitleRule(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	// Reconstruct the prior gate with its old format and preserve-title contract,
	// retaining current unrelated sections so the test isolates naming migration.
	sections := parseInstructionFileContent(templates.MemoryAgentsMD)
	var old strings.Builder
	old.WriteString(sections.preamble)
	for _, section := range sections.sections {
		if section.name == "Conversation Naming" {
			continue
		}
		text := section.raw
		if section.name == "Codex Thread Initialization Hard Gate" {
			text = strings.Replace(text, "[scope] domain / objective", "[<project>] <description>", 1)
			start := strings.Index(text, "- For a continued Codex task,")
			if start < 0 {
				t.Fatal("missing initialization continuation contract")
			}
			text = text[:start] + "- For a continued Codex task, preserve its current title and pin state unless either is missing or the user explicitly requests a change.\n\n"
		}
		old.WriteString(text)
	}
	path := filepath.Join(root, "AGENTS.md")
	writeFile(t, path, old.String())
	opts := initRefreshOptions{files: []string{"AGENTS.md"}, outputOnly: true}
	if err := runInitRefresh(root, opts); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, path)
	if !strings.Contains(got, "preserve its current title and pin state unless") {
		t.Fatal("append-only migration overwrote existing section")
	}
	if !strings.Contains(got, "This policy supersedes legacy initial-title formats and unconditional preserve-title wording") {
		t.Fatal("migration left old policy authoritative")
	}
	for _, snippet := range v3GuidanceExpectations()["AGENTS.md"] {
		if !strings.Contains(got, snippet) {
			t.Fatalf("migrated AGENTS fails reconcile expectation %q", snippet)
		}
	}
	if err := runInitRefresh(root, opts); err != nil {
		t.Fatal(err)
	}
	if readFile(t, path) != got {
		t.Fatal("second migration is not idempotent")
	}
}
