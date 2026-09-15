package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/instructions"
)

func TestNamingPolicyInstallation(t *testing.T) {
	for _, version := range []int{2, 3} {
		count := 0
		for _, f := range InstructionSupportFiles(version) {
			if f.RelativePath == "docs/references/thread-naming.md" {
				count++
				if f.Content != instructions.ThreadNamingPolicy {
					t.Fatal("installed policy diverges")
				}
			}
		}
		if count != 1 {
			t.Fatalf("version %d policy count %d", version, count)
		}
	}
	for _, version := range []int{1, 2, 3} {
		for _, path := range []string{"AGENTS.md", "CLAUDE.md"} {
			text := InstructionFileForVersion(path, version)
			if !strings.Contains(text, "docs/references/thread-naming.md") || !strings.Contains(text, "kit instructions naming") {
				t.Fatalf("%s v%d lacks shared policy route", path, version)
			}
		}
	}
	if !strings.Contains(MemoryAgentsMD, "only when the active host is Codex; Cursor and other hosts skip it") {
		t.Fatal("Cursor must skip Codex gate")
	}
	for _, path := range []string{"AGENTS.md", "CLAUDE.md"} {
		got, err := os.ReadFile(filepath.Join("..", "..", path))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != InstructionFileForVersion(path, 3) {
			t.Fatalf("%s not synchronized", path)
		}
	}
	got, err := os.ReadFile("../../docs/references/thread-naming.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != instructions.ThreadNamingPolicy {
		t.Fatal("checked-in policy projection differs")
	}
}
