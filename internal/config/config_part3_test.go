package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadGlobalReturnsDefaultWhenAbsent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, found, err := LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error = %v", err)
	}
	if found {
		t.Fatal("LoadGlobal() found = true, want false")
	}
	if cfg == nil {
		t.Fatal("LoadGlobal() cfg = nil")
	}
	if cfg.Prompts != nil {
		t.Fatalf("Prompts = %v, want nil", cfg.Prompts)
	}
}

func TestPopulateGlobalConfigCreatesDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	defaults := Default()
	defaults.InstructionScaffoldVersion = DefaultInstructionScaffoldVersion

	path, changed, err := PopulateGlobalConfig(defaults)
	if err != nil {
		t.Fatalf("PopulateGlobalConfig() error = %v", err)
	}
	if !changed {
		t.Fatal("PopulateGlobalConfig() changed = false, want true")
	}
	wantPath := filepath.Join(home, ".config", "kit", ConfigFileName)
	if path != wantPath {
		t.Fatalf("path = %q, want %q", path, wantPath)
	}

	cfg, found, err := LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error = %v", err)
	}
	if !found {
		t.Fatal("LoadGlobal() found = false, want true")
	}
	if cfg.GoalPercentage != defaults.GoalPercentage {
		t.Fatalf("GoalPercentage = %d, want %d", cfg.GoalPercentage, defaults.GoalPercentage)
	}
	if cfg.InstructionScaffoldVersion != DefaultInstructionScaffoldVersion {
		t.Fatalf("InstructionScaffoldVersion = %d, want %d", cfg.InstructionScaffoldVersion, DefaultInstructionScaffoldVersion)
	}
}

func TestPopulateGlobalConfigPreservesPromptsAndAddsMissingDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := filepath.Join(home, ".config", "kit", ConfigFileName)
	initial := []byte(`prompts:
  custom:
    review:
      content: keep existing prompt
      description: existing description
`)
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	defaults := Default()
	defaults.InstructionScaffoldVersion = DefaultInstructionScaffoldVersion

	_, changed, err := PopulateGlobalConfig(defaults)
	if err != nil {
		t.Fatalf("PopulateGlobalConfig() error = %v", err)
	}
	if !changed {
		t.Fatal("PopulateGlobalConfig() changed = false, want true")
	}

	cfg, found, err := LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error = %v", err)
	}
	if !found {
		t.Fatal("LoadGlobal() found = false, want true")
	}
	got := cfg.Prompts["custom"]["review"]
	if got.Content != "keep existing prompt" {
		t.Fatalf("Content = %q, want keep existing prompt", got.Content)
	}
	if got.Description != "existing description" {
		t.Fatalf("Description = %q, want existing description", got.Description)
	}
	if cfg.GoalPercentage != defaults.GoalPercentage {
		t.Fatalf("GoalPercentage = %d, want %d", cfg.GoalPercentage, defaults.GoalPercentage)
	}
	if cfg.InstructionScaffoldVersion != DefaultInstructionScaffoldVersion {
		t.Fatalf("InstructionScaffoldVersion = %d, want %d", cfg.InstructionScaffoldVersion, DefaultInstructionScaffoldVersion)
	}
}

func TestUpsertGlobalPromptCreatesConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	err := UpsertGlobalPrompt("coding-agent", "short", Prompt{
		Content:     "clarify first",
		Description: "short prompt",
	})
	if err != nil {
		t.Fatalf("UpsertGlobalPrompt() error = %v", err)
	}

	cfg, found, err := LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error = %v", err)
	}
	if !found {
		t.Fatal("LoadGlobal() found = false, want true")
	}
	got := cfg.Prompts["coding-agent"]["short"]
	if got.Content != "clarify first" {
		t.Fatalf("Content = %q, want clarify first", got.Content)
	}
	if got.Description != "short prompt" {
		t.Fatalf("Description = %q, want short prompt", got.Description)
	}
}

func TestUpsertLocalPromptWritesProjectConfig(t *testing.T) {
	projectRoot := t.TempDir()
	configPath := filepath.Join(projectRoot, ConfigFileName)
	if err := os.WriteFile(configPath, []byte("goal_percentage: 90\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := UpsertLocalPrompt(projectRoot, "coding-agent", "short", Prompt{
		Content:     "clarify first",
		Description: "short prompt",
	})
	if err != nil {
		t.Fatalf("UpsertLocalPrompt() error = %v", err)
	}

	cfg, err := Load(projectRoot)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	got := cfg.Prompts["coding-agent"]["short"]
	if got.Content != "clarify first" {
		t.Fatalf("Content = %q, want clarify first", got.Content)
	}
	if got.Description != "short prompt" {
		t.Fatalf("Description = %q, want short prompt", got.Description)
	}
}

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
