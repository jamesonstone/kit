package bootstrap

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/registry"
)

func TestApplyRollsBackRepositoryAndUserConfig(t *testing.T) {
	registryRoot := t.TempDir()
	writeFile(t, registryRoot, "registry/catalog.yaml", "schema_version: 1\nartifacts: []\n")
	project := t.TempDir()
	userPath := filepath.Join(t.TempDir(), "kit", ".kit.yaml")
	user := UserConfigDisposition{
		Path: userPath, State: "planned", Action: "create",
		after: "schema_version: 1\n", exists: false,
	}
	sourceConfig := registry.SourceConfig{Path: registryRoot, CatalogPath: "registry/catalog.yaml"}
	plan, err := BuildPlan(context.Background(), project, registry.LocalSource{Root: registryRoot}, sourceConfig, UserConfig{}, user)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, project, ".gitignore", "changed after planning\n")
	if err := Apply(plan); err == nil || !strings.Contains(err.Error(), "changed after planning") {
		t.Fatalf("apply error = %v", err)
	}
	for _, path := range []string{".env", ".envrc", ".kit.yaml", ".coderabbit.yaml", "docs", ".github", "AGENTS.md"} {
		if _, err := os.Lstat(filepath.Join(project, path)); !os.IsNotExist(err) {
			t.Errorf("rollback left %s: %v", path, err)
		}
	}
	if _, err := os.Stat(userPath); !os.IsNotExist(err) {
		t.Fatalf("rollback left user config: %v", err)
	}
	if got := readFile(t, project, ".gitignore"); got != "changed after planning\n" {
		t.Fatalf("collision file changed: %q", got)
	}
}

func TestUserConfigPreservesContentAndProvidesConsumedDefaults(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("KIT_CONFIG_HOME", configHome)
	path := filepath.Join(configHome, "kit", ".kit.yaml")
	content := `# user-owned comment
schema_version: 1
registry:
  repo: example/catalog
  branch: stable
  catalog_path: custom/catalog.yaml
bootstrap:
  copy_prompt: false
github:
  default_assignees:
    - maintainer
custom:
  retained: true
`
	writeFile(t, configHome, "kit/.kit.yaml", content)
	config, plan, err := PlanUserConfig()
	if err != nil {
		t.Fatal(err)
	}
	if plan.Path != path || plan.Action != "none" || plan.before != content || plan.after != content {
		t.Fatalf("user plan = %#v", plan)
	}
	if config.Source().Repo != "example/catalog" || config.Source().Branch != "stable" || config.CopyPrompt() ||
		len(config.GitHub.DefaultAssignees) != 1 || config.GitHub.DefaultAssignees[0] != "maintainer" {
		t.Fatalf("consumed defaults = %#v", config)
	}
	if workflow := buildAutoAssign(config.GitHub.DefaultAssignees); !strings.Contains(workflow, `const assignees = ["maintainer"];`) {
		t.Fatalf("auto-assign did not consume user defaults: %s", workflow)
	}

	partialHome := t.TempDir()
	t.Setenv("KIT_CONFIG_HOME", partialHome)
	writeFile(t, partialHome, "kit/.kit.yaml", "schema_version: 1\ncustom: retained\n")
	_, merged, err := PlanUserConfig()
	if err != nil {
		t.Fatal(err)
	}
	if merged.Action != "merge" || !strings.Contains(merged.after, "custom: retained") || !strings.Contains(merged.after, "copy_prompt: true") {
		t.Fatalf("merged user config = %q", merged.after)
	}
}

func TestUserConfigRefusesSymlinkWithoutReplacingIt(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("KIT_CONFIG_HOME", configHome)
	target := filepath.Join(t.TempDir(), "owned.yaml")
	if err := os.WriteFile(target, []byte("schema_version: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(configHome, "kit", ".kit.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	if _, _, err := PlanUserConfig(); err == nil || !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("symlink user config error = %v", err)
	}
	if linked, err := os.Readlink(path); err != nil || linked != target {
		t.Fatalf("user config symlink changed: %q, %v", linked, err)
	}
}

func TestEnvironmentAndPromptNeverTrustOrExposeSecrets(t *testing.T) {
	project := t.TempDir()
	writeProjectConfig(t, project)
	writeFile(t, project, ".env", "SECRET_VALUE=hidden\n")
	writeFile(t, project, ".envrc", "project-owned envrc\n")
	plan, err := BuildPlan(context.Background(), project, nil, registry.SourceConfig{}, UserConfig{}, UserConfigDisposition{})
	if err != nil {
		t.Fatal(err)
	}
	if err := Apply(plan); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, project, ".env"); got != "SECRET_VALUE=hidden\n" {
		t.Fatalf(".env changed: %q", got)
	}
	if got := readFile(t, project, ".envrc"); got != "project-owned envrc\n" {
		t.Fatalf(".envrc changed: %q", got)
	}
	for _, content := range []string{plan.Prompt, envrcStarter} {
		if strings.Contains(content, "direnv allow") || strings.Contains(content, "SECRET_VALUE") {
			t.Fatalf("unsafe bootstrap content: %q", content)
		}
	}
}
