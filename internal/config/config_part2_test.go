package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

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

func TestLoadParsesProjectRefreshConfig(t *testing.T) {
	projectRoot := t.TempDir()
	configPath := filepath.Join(projectRoot, ConfigFileName)
	if err := os.WriteFile(configPath, []byte(`
project_refresh:
  constitution:
    feature_interval: 3
    max_age_days: 14
    last_reviewed_at: "2026-06-01T12:00:00Z"
    last_completed_feature_count: 8
`), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(projectRoot)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	refresh := cfg.ProjectRefresh.Constitution
	if refresh.FeatureInterval != 3 || refresh.MaxAgeDays != 14 {
		t.Fatalf("ProjectRefresh.Constitution thresholds = %#v", refresh)
	}
	if refresh.LastReviewedAt != "2026-06-01T12:00:00Z" {
		t.Fatalf("LastReviewedAt = %q", refresh.LastReviewedAt)
	}
	if refresh.LastCompletedFeatureCount != 8 {
		t.Fatalf("LastCompletedFeatureCount = %d, want 8", refresh.LastCompletedFeatureCount)
	}
}

func TestLoadParsesRegistryConfig(t *testing.T) {
	projectRoot := t.TempDir()
	configPath := filepath.Join(projectRoot, ConfigFileName)
	if err := os.WriteFile(configPath, []byte(`
registry:
  schema_version: 1
  source:
    repo: jamesonstone/kit
    branch: main
  artifacts:
    - kind: ruleset
      slug: github-pr-delivery
      path: docs/references/rules/github-pr-delivery.md
      source_repo: jamesonstone/kit
      source_branch: main
      source_commit: abc123
      source_path: docs/references/rules/github-pr-delivery.md
      installed_hash: sha256:deadbeef
      state: managed
`), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(projectRoot)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Registry.SchemaVersion != 1 {
		t.Fatalf("Registry.SchemaVersion = %d, want 1", cfg.Registry.SchemaVersion)
	}
	if cfg.Registry.Source.Repo != "jamesonstone/kit" || cfg.Registry.Source.Branch != "main" {
		t.Fatalf("Registry.Source = %#v", cfg.Registry.Source)
	}
	artifact, ok := cfg.RegistryArtifact("ruleset", "github-pr-delivery")
	if !ok {
		t.Fatal("expected registry artifact")
	}
	if artifact.InstalledHash != "sha256:deadbeef" || artifact.State != "managed" {
		t.Fatalf("artifact = %#v", artifact)
	}
	cfg.UpsertRegistryArtifact(RegistryArtifact{
		Kind:          "ruleset",
		Slug:          "github-pr-delivery",
		Path:          "docs/references/rules/github-pr-delivery.md",
		InstalledHash: "sha256:feedface",
		State:         "local-custom",
	})
	artifact, ok = cfg.RegistryArtifact("ruleset", "github-pr-delivery")
	if !ok || artifact.InstalledHash != "sha256:feedface" || artifact.State != "local-custom" {
		t.Fatalf("updated artifact = %#v", artifact)
	}
}

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
