package bootstrap

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/contract"
	"github.com/jamesonstone/kit/internal/registry"
)

var exactBootstrapPaths = []string{
	".coderabbit.yaml", ".env", ".envrc", ".github/copilot-instructions.md",
	".github/pull_request_template.md", ".github/workflows/auto-assign.yml",
	".gitignore", ".kit.yaml", "AGENTS.md", "CLAUDE.md", "Makefile", "README.md",
	"docs/CONSTITUTION.md", "docs/PROJECT_PROGRESS_SUMMARY.md",
	"docs/agents/GUARDRAILS.md", "docs/agents/README.md", "docs/agents/RLM.md",
	"docs/agents/TOOLING.md", "docs/agents/WORKFLOWS.md", "docs/references/README.md",
	"docs/references/external-systems.md", "docs/references/testing.md",
	"docs/references/tooling.md", "docs/references/worktrees.md",
}

func TestFreshPlanCreatesExactBootstrapAndResolvableWorkflow(t *testing.T) {
	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	sourceConfig := registry.SourceConfig{Path: repositoryRoot, CatalogPath: "registry/catalog.yaml"}
	user := testUserConfigPlan(t)
	plan, err := BuildPlan(context.Background(), project, registry.LocalSource{Root: repositoryRoot}, sourceConfig, UserConfig{}, user)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, file := range plan.Files {
		paths = append(paths, file.Path)
	}
	if !slices.Equal(paths, exactBootstrapPaths) {
		t.Fatalf("bootstrap paths = %#v", paths)
	}
	if len(plan.Directories) != 1 || plan.Directories[0].Path != "docs/specs" {
		t.Fatalf("directories = %#v", plan.Directories)
	}
	if !strings.Contains(plan.Prompt, "kit contract resolve --workflow repository-bootstrap --json") {
		t.Fatal("bootstrap prompt does not route through the repository-bootstrap contract")
	}
	if err := Apply(plan); err != nil {
		t.Fatal(err)
	}
	for _, path := range exactBootstrapPaths {
		if _, err := os.Lstat(filepath.Join(project, filepath.FromSlash(path))); err != nil {
			t.Errorf("bootstrap path %s: %v", path, err)
		}
	}
	if info, err := os.Stat(filepath.Join(project, "docs/specs")); err != nil || !info.IsDir() {
		t.Fatalf("docs/specs directory: %v", err)
	}
	env, err := os.ReadFile(filepath.Join(project, ".env"))
	if err != nil || len(env) != 0 {
		t.Fatalf(".env = %q, err %v", env, err)
	}
	for _, excluded := range []string{"BRAINSTORM.md", "PLAN.md", "TASKS.md", "docs/specs/0000_INIT_PROJECT.md", "docs/notes", ".kit/loops"} {
		if _, err := os.Stat(filepath.Join(project, excluded)); !os.IsNotExist(err) {
			t.Errorf("excluded bootstrap path %s exists: %v", excluded, err)
		}
	}
	resolved, err := contract.Resolve(project, contract.Hints{Workflows: []string{"repository-bootstrap"}})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.State != "ready" || !resolvedWorkflow(resolved, "repository-bootstrap") {
		t.Fatalf("resolved repository bootstrap = %#v", resolved)
	}
	assertResolvedRule(t, resolved, "constitution-curation")
	assertResolvedRule(t, resolved, "feature-specification")
	assertResolvedRule(t, resolved, "testing-and-environment-validation")
	assertResolvedRule(t, resolved, "readme-header-tagline")
	agentWorkflow := readFile(t, project, "docs/agents/WORKFLOWS.md")
	if !strings.Contains(agentWorkflow, "complete living V3 spec") ||
		!strings.Contains(agentWorkflow, "--work-type feature --feature <feature>") ||
		!strings.Contains(agentWorkflow, "--work-type maintenance") {
		t.Fatalf("generated implementation routing is incomplete: %s", agentWorkflow)
	}
}

func TestRepositoryBootstrapPromptGolden(t *testing.T) {
	want, err := os.ReadFile("testdata/repository-bootstrap.golden.txt")
	if err != nil {
		t.Fatal(err)
	}
	if got := repositoryBootstrapPrompt(); got != string(want) {
		t.Fatalf("repository bootstrap prompt changed\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
	for _, duty := range []string{
		"docs/PROJECT_PROGRESS_SUMMARY.md from current specifications",
		"retain its valid empty state", "actual safe, verified test", "Makefile help-only starter",
		"files changed and preserved", "Repository Memory decision",
	} {
		if !strings.Contains(string(want), duty) {
			t.Errorf("prompt is missing legacy semantic duty %q", duty)
		}
	}
}

func TestProgressSummarySupportsEmptyAndExistingSpecPopulation(t *testing.T) {
	for _, heading := range []string{
		"## FEATURE PROGRESS TABLE", "## PROJECT INTENT", "## GLOBAL CONSTRAINTS",
		"## FEATURE SUMMARIES", "## LAST UPDATED",
	} {
		if !strings.Contains(progressSummaryStarter, heading) {
			t.Errorf("empty progress starter is missing %q", heading)
		}
	}
	if strings.Contains(progressSummaryStarter, "0000_INIT_PROJECT") {
		t.Fatal("empty progress starter invents a legacy init feature")
	}
	workflow, err := os.ReadFile("../../docs/references/workflows/repository-bootstrap.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"from current specifications", "repository history", "never invent features"} {
		if !strings.Contains(string(workflow), expected) {
			t.Errorf("existing-spec population contract is missing %q", expected)
		}
	}
}

func resolvedWorkflow(resolved contract.Resolved, slug string) bool {
	for _, workflow := range resolved.Workflows {
		if workflow.Slug == slug {
			return true
		}
	}
	return false
}

func TestExistingBootstrapPreservesOwnedFilesAndIsIdempotent(t *testing.T) {
	project := t.TempDir()
	writeProjectConfig(t, project)
	owned := map[string]string{}
	for _, starter := range createIfMissingFiles {
		owned[starter.path] = "project owned: " + starter.path + "\n"
	}
	for path := range routingStarters {
		owned[path] = "project owned: " + path + "\n"
	}
	owned[".env"] = "TOP_SECRET=do-not-read\n"
	owned[".envrc"] = "project-owned envrc\n"
	owned["README.md"] = "# Existing\n\nProject prose.\n"
	owned["docs/CONSTITUTION.md"] = "# Existing Constitution\n\nProject truth.\n"
	owned[".gitignore"] = "project.cache\n"
	for path, content := range owned {
		writeFile(t, project, path, content)
	}
	plan, err := BuildPlan(context.Background(), project, nil, registry.SourceConfig{}, UserConfig{}, UserConfigDisposition{})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "TOP_SECRET") || strings.Contains(plan.Prompt, "TOP_SECRET") {
		t.Fatal("bootstrap output exposed .env content")
	}
	if err := Apply(plan); err != nil {
		t.Fatal(err)
	}
	for path, content := range owned {
		actual := readFile(t, project, path)
		switch path {
		case "README.md":
			if !strings.Contains(actual, content) || !strings.Contains(actual, readmeBadgesStart) || !strings.Contains(actual, readmeMaintainersStart) {
				t.Errorf("README bounded merge = %q", actual)
			}
		case "docs/CONSTITUTION.md":
			if !strings.Contains(actual, content) || !strings.Contains(actual, constitutionStart) {
				t.Errorf("Constitution bounded merge = %q", actual)
			}
		case ".gitignore":
			if !strings.Contains(actual, content) || !strings.Contains(actual, ".env\n") {
				t.Errorf("gitignore append = %q", actual)
			}
		default:
			if actual != content {
				t.Errorf("%s was overwritten: %q", path, actual)
			}
		}
	}
	second, err := BuildPlan(context.Background(), project, nil, registry.SourceConfig{}, UserConfig{}, UserConfigDisposition{})
	if err != nil {
		t.Fatal(err)
	}
	if second.State != "current" || len(second.Registry.Changes) != 0 {
		t.Fatalf("repeat init = state %s, changes %#v", second.State, second.Registry.Changes)
	}
	customReadme := strings.Replace(readFile(t, project, "README.md"),
		"repository-bootstrap agent may add only badges", "project customized badges", 1)
	writeFile(t, project, "README.md", customReadme)
	customPlan, err := BuildPlan(context.Background(), project, nil, registry.SourceConfig{}, UserConfig{}, UserConfigDisposition{})
	if err != nil {
		t.Fatal(err)
	}
	if len(customPlan.Registry.Changes) != 0 || fileDisposition(customPlan, "README.md").State != registry.StateLocalCustom {
		t.Fatalf("custom managed README was not preserved: %#v", fileDisposition(customPlan, "README.md"))
	}
	if got := readFile(t, project, "README.md"); got != customReadme {
		t.Fatal("planning overwrote customized README")
	}
}

func fileDisposition(plan Plan, path string) FileDisposition {
	for _, file := range plan.Files {
		if file.Path == path {
			return file
		}
	}
	return FileDisposition{}
}

func assertResolvedRule(t *testing.T, resolved contract.Resolved, slug string) {
	t.Helper()
	for _, rules := range resolved.Rules {
		for _, rule := range rules {
			if rule.Slug == slug {
				return
			}
		}
	}
	t.Errorf("resolved contract is missing ruleset/%s", slug)
}

func testUserConfigPlan(t *testing.T) UserConfigDisposition {
	t.Helper()
	path := filepath.Join(t.TempDir(), "kit", ".kit.yaml")
	return UserConfigDisposition{Path: path, State: "planned", Action: "create", after: "schema_version: 1\n", exists: false}
}

func writeProjectConfig(t *testing.T, project string) {
	t.Helper()
	content, err := registry.MarshalProject(registry.NewProjectConfig(registry.SourceConfig{Path: ".", CatalogPath: "registry/catalog.yaml"}))
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, project, registry.ProjectFile, string(content))
}

func writeFile(t *testing.T, root, path, content string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, root, path string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
