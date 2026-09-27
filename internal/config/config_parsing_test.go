package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

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

	if _, err := Load(projectRoot); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
}
