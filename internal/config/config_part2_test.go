package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadParsesGitHubConfig(t *testing.T) {
	projectRoot := t.TempDir()
	configPath := filepath.Join(projectRoot, ConfigFileName)
	if err := os.WriteFile(configPath, []byte(`
github:
  repository: jamesonstone/kit
  default_branch: main
  default_assignees:
    - jamesonstone
    - octocat
`), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(projectRoot)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.GitHub.Repository != "jamesonstone/kit" {
		t.Fatalf("GitHub.Repository = %q, want jamesonstone/kit", cfg.GitHub.Repository)
	}
	if cfg.GitHub.DefaultBranch != "main" {
		t.Fatalf("GitHub.DefaultBranch = %q, want main", cfg.GitHub.DefaultBranch)
	}
	if cfg.GitHub.DefaultAssignees == nil {
		t.Fatalf("GitHub.DefaultAssignees = nil, want configured assignees")
	}
	if !reflect.DeepEqual(*cfg.GitHub.DefaultAssignees, []string{"jamesonstone", "octocat"}) {
		t.Fatalf("GitHub.DefaultAssignees = %v, want jamesonstone and octocat", *cfg.GitHub.DefaultAssignees)
	}

	if err := Save(projectRoot, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	for _, check := range []string{"github:", "repository: jamesonstone/kit", "default_branch: main", "default_assignees:", "- jamesonstone", "- octocat"} {
		if !strings.Contains(string(data), check) {
			t.Fatalf("saved config missing %q, got:\n%s", check, data)
		}
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
