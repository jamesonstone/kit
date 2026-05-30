package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDefaultIncludesAllRepositoryInstructionFiles(t *testing.T) {
	cfg := Default()

	want := []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md"}
	if !reflect.DeepEqual(cfg.Agents, want) {
		t.Fatalf("Default().Agents = %v, want %v", cfg.Agents, want)
	}
}

func TestDefaultIncludesLoopPolicy(t *testing.T) {
	cfg := Default()

	if cfg.Loop.MinConfidence != 95 {
		t.Fatalf("Loop.MinConfidence = %d, want 95", cfg.Loop.MinConfidence)
	}
	if cfg.Loop.MaxIterations != 20 {
		t.Fatalf("Loop.MaxIterations = %d, want 20", cfg.Loop.MaxIterations)
	}
}

func TestLoadParsesLoopConfig(t *testing.T) {
	projectRoot := t.TempDir()
	configPath := filepath.Join(projectRoot, ConfigFileName)
	if err := os.WriteFile(configPath, []byte(`
loop:
  min_confidence: 91
  max_iterations: 7
  agent:
    command: codex
    args:
      - run
      - --stdin
`), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(projectRoot)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Loop.MinConfidence != 91 {
		t.Fatalf("Loop.MinConfidence = %d, want 91", cfg.Loop.MinConfidence)
	}
	if cfg.Loop.MaxIterations != 7 {
		t.Fatalf("Loop.MaxIterations = %d, want 7", cfg.Loop.MaxIterations)
	}
	if cfg.Loop.Agent.Command != "codex" {
		t.Fatalf("Loop.Agent.Command = %q, want codex", cfg.Loop.Agent.Command)
	}
	wantArgs := []string{"run", "--stdin"}
	if !reflect.DeepEqual(cfg.Loop.Agent.Args, wantArgs) {
		t.Fatalf("Loop.Agent.Args = %v, want %v", cfg.Loop.Agent.Args, wantArgs)
	}
}

func TestLoadAllowsMissingPrompts(t *testing.T) {
	projectRoot := t.TempDir()
	configPath := filepath.Join(projectRoot, ConfigFileName)
	if err := os.WriteFile(configPath, []byte("goal_percentage: 90\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(projectRoot)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.GoalPercentage != 90 {
		t.Fatalf("GoalPercentage = %d, want 90", cfg.GoalPercentage)
	}
	if cfg.Prompts != nil {
		t.Fatalf("Prompts = %v, want nil", cfg.Prompts)
	}
}

func TestLoadParsesPromptSchema(t *testing.T) {
	projectRoot := t.TempDir()
	configPath := filepath.Join(projectRoot, ConfigFileName)
	configText := []byte(`
prompts:
  coding-agent:
    short:
      content: clarify first
      description: short planning prompt
`)
	if err := os.WriteFile(configPath, configText, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(projectRoot)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	got := cfg.Prompts["coding-agent"]["short"]
	if got.Content != "clarify first" {
		t.Fatalf("Content = %q, want %q", got.Content, "clarify first")
	}
	if got.Description != "short planning prompt" {
		t.Fatalf("Description = %q, want %q", got.Description, "short planning prompt")
	}
}

func TestLoadPromptSchemaIgnoresUnknownMetadata(t *testing.T) {
	projectRoot := t.TempDir()
	configPath := filepath.Join(projectRoot, ConfigFileName)
	configText := []byte(`
prompts:
  coding-agent:
    short:
      content: clarify first
      description: short planning prompt
      tags:
        - planning
      future:
        owner: agent
`)
	if err := os.WriteFile(configPath, configText, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(projectRoot)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	got := cfg.Prompts["coding-agent"]["short"]
	if got.Content != "clarify first" {
		t.Fatalf("Content = %q, want clarify first", got.Content)
	}
	if got.Description != "short planning prompt" {
		t.Fatalf("Description = %q, want short planning prompt", got.Description)
	}
}

func TestRecordRemovedFeatureClearsPausedStateAndReplacesTombstone(t *testing.T) {
	cfg := Default()
	cfg.SetFeaturePaused("0002-bravo", true)
	cfg.RecordRemovedFeature(RemovedFeature{
		Number:    2,
		Slug:      "bravo",
		DirName:   "0002-bravo",
		CreatedAt: "2026-04-05T00:00:00Z",
		RemovedAt: "2026-05-06T12:00:00Z",
	})

	if cfg.IsFeaturePaused("0002-bravo") {
		t.Fatalf("expected removed feature to clear paused state")
	}
	if len(cfg.RemovedFeatures) != 1 {
		t.Fatalf("RemovedFeatures length = %d, want 1", len(cfg.RemovedFeatures))
	}
	if cfg.RemovedFeatures[0].RemovedAt != "2026-05-06T12:00:00Z" {
		t.Fatalf("RemovedAt = %q, want first timestamp", cfg.RemovedFeatures[0].RemovedAt)
	}

	cfg.RecordRemovedFeature(RemovedFeature{
		Number:    2,
		Slug:      "bravo",
		DirName:   "0002-bravo",
		CreatedAt: "2026-04-05T00:00:00Z",
		RemovedAt: "2026-05-07T12:00:00Z",
	})

	if len(cfg.RemovedFeatures) != 1 {
		t.Fatalf("RemovedFeatures length after replace = %d, want 1", len(cfg.RemovedFeatures))
	}
	if cfg.RemovedFeatures[0].RemovedAt != "2026-05-07T12:00:00Z" {
		t.Fatalf("RemovedAt = %q, want replacement timestamp", cfg.RemovedFeatures[0].RemovedAt)
	}
}

func TestFindProjectRootOptionalReturnsRootWhenConfigExists(t *testing.T) {
	projectRoot := t.TempDir()
	nested := filepath.Join(projectRoot, "a", "b")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, ConfigFileName), []byte("goal_percentage: 95\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Chdir(nested)

	root, found, err := FindProjectRootOptional()
	if err != nil {
		t.Fatalf("FindProjectRootOptional() error = %v", err)
	}
	if !found {
		t.Fatal("FindProjectRootOptional() found = false, want true")
	}
	if root != projectRoot {
		t.Fatalf("FindProjectRootOptional() root = %q, want %q", root, projectRoot)
	}
}

func TestFindProjectRootOptionalReturnsNotFoundWithoutError(t *testing.T) {
	t.Chdir(t.TempDir())

	root, found, err := FindProjectRootOptional()
	if err != nil {
		t.Fatalf("FindProjectRootOptional() error = %v", err)
	}
	if found {
		t.Fatal("FindProjectRootOptional() found = true, want false")
	}
	if root != "" {
		t.Fatalf("FindProjectRootOptional() root = %q, want empty", root)
	}
}

func TestGlobalConfigPathUsesDotConfigKit(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := GlobalConfigPath()
	if err != nil {
		t.Fatalf("GlobalConfigPath() error = %v", err)
	}

	want := filepath.Join(home, ".config", "kit", ConfigFileName)
	if got != want {
		t.Fatalf("GlobalConfigPath() = %q, want %q", got, want)
	}
}

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
