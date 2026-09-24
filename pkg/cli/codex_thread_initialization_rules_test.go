package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodexThreadInitializationRulesetIsRetired(t *testing.T) {
	const slug = "codex-thread-initialization"
	path := filepath.Join("..", "..", "docs", "references", "rules", slug+".md")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("retired %s ruleset still exists: %v", slug, err)
	}

	index, err := os.ReadFile(filepath.Join("..", "..", "docs", "references", "README.md"))
	if err != nil {
		t.Fatalf("read references index: %v", err)
	}
	if strings.Contains(string(index), "| `"+slug+"` |") {
		t.Fatal("references index still lists codex-thread-initialization")
	}

	kitYAML, err := os.ReadFile(filepath.Join("..", "..", ".kit.yaml"))
	if err != nil {
		t.Fatalf("read .kit.yaml: %v", err)
	}
	if strings.Contains(string(kitYAML), "slug: "+slug) {
		t.Fatal(".kit.yaml still tracks codex-thread-initialization")
	}
}
