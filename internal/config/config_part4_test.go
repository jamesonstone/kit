package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpsertPromptFileRejectsNonMappingRoot(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), ConfigFileName)
	if err := os.WriteFile(configPath, []byte("- not-a-mapping\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := UpsertPromptFile(configPath, "coding-agent", "short", Prompt{Content: "prompt"}, false)
	if err == nil {
		t.Fatal("expected non-mapping config root to fail")
	}
	if !strings.Contains(err.Error(), "config root must be a YAML mapping") {
		t.Fatalf("error = %q, want YAML mapping guidance", err)
	}
}
