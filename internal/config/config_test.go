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
	if cfg.ProjectRefresh.Constitution.FeatureInterval != 5 {
		t.Fatalf("ProjectRefresh.Constitution.FeatureInterval = %d, want 5", cfg.ProjectRefresh.Constitution.FeatureInterval)
	}
	if cfg.ProjectRefresh.Constitution.MaxAgeDays != 30 {
		t.Fatalf("ProjectRefresh.Constitution.MaxAgeDays = %d, want 30", cfg.ProjectRefresh.Constitution.MaxAgeDays)
	}
}

func TestSaveOmitsDefaultLoopConfigAndKeepsCustomLoopConfig(t *testing.T) {
	projectRoot := t.TempDir()
	defaults := Default()
	if err := Save(projectRoot, defaults); err != nil {
		t.Fatalf("Save(defaults) error = %v", err)
	}
	data, err := os.ReadFile(filepath.Join(projectRoot, ConfigFileName))
	if err != nil {
		t.Fatalf("ReadFile(default config) error = %v", err)
	}
	if strings.Contains(string(data), "loop:") {
		t.Fatalf("default loop config should be omitted, got:\n%s", data)
	}
	if strings.Contains(string(data), "github:") {
		t.Fatalf("empty github config should be omitted, got:\n%s", data)
	}

	defaults.Loop.MaxIterations = 7
	defaults.Loop.Agent.Command = "codex"
	if err := Save(projectRoot, defaults); err != nil {
		t.Fatalf("Save(custom loop) error = %v", err)
	}
	data, err = os.ReadFile(filepath.Join(projectRoot, ConfigFileName))
	if err != nil {
		t.Fatalf("ReadFile(custom config) error = %v", err)
	}
	for _, check := range []string{"loop:", "max_iterations: 7", "command: codex"} {
		if !strings.Contains(string(data), check) {
			t.Fatalf("custom loop config missing %q, got:\n%s", check, data)
		}
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
