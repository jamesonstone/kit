package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRetiredThreadLifecycleIsAbsentFromScaffolds(t *testing.T) {
	stale := []string{
		"docs/references/thread-naming.md",
		"kit instructions naming",
		"## Conversation Naming",
		"## Codex Thread Initialization Hard Gate",
		"set_thread_title",
		"set_thread_pinned",
	}
	for _, version := range []int{1, 2, 3} {
		for _, file := range InstructionSupportFiles(version) {
			if file.RelativePath == "docs/references/thread-naming.md" {
				t.Fatalf("version %d still installs thread-naming.md", version)
			}
			for _, snippet := range stale {
				if strings.Contains(file.Content, snippet) {
					t.Fatalf("version %d %s still contains %q", version, file.RelativePath, snippet)
				}
			}
		}
		for _, path := range []string{"AGENTS.md", "CLAUDE.md"} {
			text := InstructionFileForVersion(path, version)
			for _, snippet := range stale {
				if strings.Contains(text, snippet) {
					t.Fatalf("%s v%d still contains %q", path, version, snippet)
				}
			}
		}
	}
	if _, err := os.Stat(filepath.Join("..", "..", "docs", "references", "thread-naming.md")); !os.IsNotExist(err) {
		t.Fatalf("checked-in thread-naming.md still exists: %v", err)
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
}
