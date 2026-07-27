package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpsertPromptFilePreservesUnrelatedAndUnknownFields(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), ConfigFileName)
	initial := []byte(`
goal_percentage: 90
custom_root: keep me
prompts:
  coding-agent:
    short:
      content: old
      description: old description
      tags:
        - keep
`)
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := UpsertPromptFile(configPath, "coding-agent", "short", Prompt{
		Content: "new",
	}, false)
	if err != nil {
		t.Fatalf("UpsertPromptFile() error = %v", err)
	}

	updated, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(updated)
	for _, want := range []string{
		"custom_root: keep me",
		"content: new",
		"tags:",
		"- keep",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("updated config missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "description: old description") {
		t.Fatalf("description was not removed:\n%s", text)
	}
}

func TestUpsertPromptFileRejectsEmptyContent(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), ConfigFileName)
	if err := os.WriteFile(configPath, []byte("{}\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := UpsertPromptFile(configPath, "coding-agent", "short", Prompt{}, false)
	if err == nil {
		t.Fatal("expected empty prompt content to fail")
	}
}

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
