package registry

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestInitAndReconcileAreCompleteAndIdempotent(t *testing.T) {
	registryRoot, sourceCfg := writeTestRegistry(t, "Registry purpose.", "Registry rule.")
	project := t.TempDir()
	source := LocalSource{Root: registryRoot}

	plan, err := BuildInitPlan(context.Background(), project, source, sourceCfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Artifacts) != 2 {
		t.Fatalf("artifacts = %d, want 2", len(plan.Artifacts))
	}
	if err := ApplyPlan(project, plan); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		"docs/references/rules/example.md", "docs/references/workflows/delivery.md",
		"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md", "docs/agents/README.md", ProjectFile,
	} {
		if _, err := os.Stat(filepath.Join(project, filepath.FromSlash(path))); err != nil {
			t.Errorf("expected %s: %v", path, err)
		}
	}

	reconcile, err := BuildReconcilePlan(context.Background(), project, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	if reconcile.State != "current" || len(reconcile.Changes) != 0 {
		t.Fatalf("idempotent state = %s, changes = %d", reconcile.State, len(reconcile.Changes))
	}
}

func TestReconcileMergesDisjointSectionsAndPreservesCustomization(t *testing.T) {
	registryRoot, sourceCfg := writeTestRegistry(t, "Registry purpose.", "Registry rule.")
	project := initializedProject(t, registryRoot, sourceCfg)
	rulePath := filepath.Join(project, "docs/references/rules/example.md")
	local := readTestFile(t, rulePath)
	writeTestFile(t, rulePath, strings.Replace(local, "Registry purpose.", "Local purpose.", 1))
	localOnly, err := BuildReconcilePlan(context.Background(), project, LocalSource{Root: registryRoot}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := findArtifactPlan(t, localOnly, KindRuleset, "example"); got.State != StateLocalCustom || got.Action != "none" {
		t.Fatalf("local-only artifact = %#v", got)
	}
	if err := ApplyPlan(project, localOnly); err != nil {
		t.Fatal(err)
	}
	updateTestRegistry(t, registryRoot, "Registry purpose.", "Updated registry rule.")

	plan, err := BuildReconcilePlan(context.Background(), project, LocalSource{Root: registryRoot}, nil)
	if err != nil {
		t.Fatal(err)
	}
	artifact := findArtifactPlan(t, plan, KindRuleset, "example")
	if artifact.State != StateLocalCustom || artifact.Action != "update" {
		t.Fatalf("artifact = %#v", artifact)
	}
	if err := ApplyPlan(project, plan); err != nil {
		t.Fatal(err)
	}
	merged := readTestFile(t, rulePath)
	if !strings.Contains(merged, "Local purpose.") || !strings.Contains(merged, "Updated registry rule.") {
		t.Fatalf("disjoint changes were not preserved:\n%s", merged)
	}
}

func TestReconcileBlocksSameSectionConflictUntilAccepted(t *testing.T) {
	registryRoot, sourceCfg := writeTestRegistry(t, "Registry purpose.", "Registry rule.")
	project := initializedProject(t, registryRoot, sourceCfg)
	rulePath := filepath.Join(project, "docs/references/rules/example.md")
	local := readTestFile(t, rulePath)
	writeTestFile(t, rulePath, strings.Replace(local, "Registry purpose.", "Local purpose.", 1))
	updateTestRegistry(t, registryRoot, "Updated registry purpose.", "Registry rule.")

	plan, err := BuildReconcilePlan(context.Background(), project, LocalSource{Root: registryRoot}, nil)
	if err != nil {
		t.Fatal(err)
	}
	artifact := findArtifactPlan(t, plan, KindRuleset, "example")
	if plan.State != "attention-needed" || artifact.State != StateConflict || artifact.Action != "blocked" {
		t.Fatalf("plan = %#v, artifact = %#v", plan, artifact)
	}
	if err := ApplyPlan(project, plan); err != nil {
		t.Fatal(err)
	}
	repeated, err := BuildReconcilePlan(context.Background(), project, LocalSource{Root: registryRoot}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := findArtifactPlan(t, repeated, KindRuleset, "example"); got.State != StateConflict {
		t.Fatalf("conflict did not persist: %#v", got)
	}
	accepted, err := BuildReconcilePlan(context.Background(), project, LocalSource{Root: registryRoot}, map[string]bool{"ruleset/example": true})
	if err != nil {
		t.Fatal(err)
	}
	if got := findArtifactPlan(t, accepted, KindRuleset, "example"); got.Action != "update" || got.State != StateManaged {
		t.Fatalf("accepted artifact = %#v", got)
	}
}

func TestSchemaV1MigrationPreservesRulesAndRoutingText(t *testing.T) {
	registryRoot, sourceCfg := writeTestRegistry(t, "Registry purpose.", "Registry rule.")
	project := t.TempDir()
	writeTestFile(t, filepath.Join(project, ProjectFile), "schema_version: 1\nregistry:\n  source:\n    path: "+registryRoot+"\n    catalog_path: registry/catalog.yaml\nlegacy_option: preserved\n")
	custom := testRuleDocument("Project purpose.", "Project rule.")
	writeTestFile(t, filepath.Join(project, "docs/references/rules/example.md"), custom)
	writeTestFile(t, filepath.Join(project, "AGENTS.md"), "# Project agents\n\nKeep this text.\n")

	plan, err := BuildReconcilePlan(context.Background(), project, LocalSource{Root: registryRoot}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Migration || findArtifactPlan(t, plan, KindRuleset, "example").State != StateLocalCustom {
		t.Fatalf("migration plan = %#v", plan)
	}
	if err := ApplyPlan(project, plan); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(readTestFile(t, filepath.Join(project, "AGENTS.md")), "Keep this text.") {
		t.Fatal("routing migration erased project text")
	}
	content := readTestFile(t, filepath.Join(project, ProjectFile))
	if !strings.Contains(content, "schema_version: 2") || !strings.Contains(content, "legacy_option: preserved") {
		t.Fatalf("config migration lost data:\n%s", content)
	}
	_ = sourceCfg
}

func TestApplyPlanRollsBackEarlierChanges(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "block"), "not a directory")
	plan := Plan{Changes: []Change{
		{Path: "first.md", Action: "create", After: "created\n"},
		{Path: "block/second.md", Action: "create", After: "blocked\n"},
	}}
	if err := ApplyPlan(root, plan); err == nil {
		t.Fatal("expected apply failure")
	}
	if _, err := os.Stat(filepath.Join(root, "first.md")); !os.IsNotExist(err) {
		t.Fatalf("first change was not rolled back: %v", err)
	}
}

func TestReconcileMovesManagedArtifactsAndPreservesCustomizedOldPaths(t *testing.T) {
	registryRoot, sourceCfg := writeTestRegistry(t, "Registry purpose.", "Registry rule.")
	managedProject := initializedProject(t, registryRoot, sourceCfg)
	customProject := initializedProject(t, registryRoot, sourceCfg)
	oldCustomPath := filepath.Join(customProject, "docs/references/rules/example.md")
	writeTestFile(t, oldCustomPath, strings.Replace(readTestFile(t, oldCustomPath), "Registry purpose.", "Local purpose.", 1))
	moveTestRuleTarget(t, registryRoot, "docs/references/rules/moved-example.md")

	managedPlan, err := BuildReconcilePlan(context.Background(), managedProject, LocalSource{Root: registryRoot}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := ApplyPlan(managedProject, managedPlan); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(managedProject, "docs/references/rules/example.md")); !os.IsNotExist(err) {
		t.Fatalf("managed old path was not removed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(managedProject, "docs/references/rules/moved-example.md")); err != nil {
		t.Fatalf("new managed path missing: %v", err)
	}

	customPlan, err := BuildReconcilePlan(context.Background(), customProject, LocalSource{Root: registryRoot}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := ApplyPlan(customProject, customPlan); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(readTestFile(t, oldCustomPath), "Local purpose.") {
		t.Fatal("customized old path was not preserved")
	}
}

func initializedProject(t *testing.T, registryRoot string, sourceCfg SourceConfig) string {
	t.Helper()
	project := t.TempDir()
	plan, err := BuildInitPlan(context.Background(), project, LocalSource{Root: registryRoot}, sourceCfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := ApplyPlan(project, plan); err != nil {
		t.Fatal(err)
	}
	return project
}

func writeTestRegistry(t *testing.T, purpose, rule string) (string, SourceConfig) {
	t.Helper()
	root := t.TempDir()
	updateTestRegistry(t, root, purpose, rule)
	return root, SourceConfig{Path: root, CatalogPath: "registry/catalog.yaml"}
}

func updateTestRegistry(t *testing.T, root, purpose, rule string) {
	t.Helper()
	ruleContent := testRuleDocument(purpose, rule)
	workflowContent := testWorkflowDocument()
	writeTestFile(t, filepath.Join(root, "rules/example.md"), ruleContent)
	writeTestFile(t, filepath.Join(root, "workflows/delivery.md"), workflowContent)
	catalog := Catalog{SchemaVersion: CatalogSchemaVersion, Artifacts: []CatalogArtifact{
		{Kind: KindRuleset, Slug: "example", Description: "Example rules", Visibility: "downstream", SourcePath: "rules/example.md", TargetPath: "docs/references/rules/example.md", Version: 1, Digest: HashContent(ruleContent), ReadPolicy: "must", AppliesTo: []string{"example"}},
		{Kind: KindWorkflow, Slug: "delivery", Description: "Delivery workflow", Visibility: "downstream", SourcePath: "workflows/delivery.md", TargetPath: "docs/references/workflows/delivery.md", Version: 1, Digest: HashContent(workflowContent), ReadPolicy: "must", Dependencies: []string{"ruleset/example"}},
	}}
	content, err := yaml.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(root, "registry/catalog.yaml"), string(content))
}

func moveTestRuleTarget(t *testing.T, root, target string) {
	t.Helper()
	path := filepath.Join(root, "registry/catalog.yaml")
	var catalog Catalog
	if err := yaml.Unmarshal([]byte(readTestFile(t, path)), &catalog); err != nil {
		t.Fatal(err)
	}
	for index := range catalog.Artifacts {
		if catalog.Artifacts[index].Kind == KindRuleset {
			catalog.Artifacts[index].TargetPath = target
		}
	}
	content, err := yaml.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, path, string(content))
}

func testWorkflowDocument() string {
	return "---\nkind: workflow\nslug: delivery\ndescription: Delivery workflow\nstatus: active\nregistry_scope: downstream\nread_policy_default: must\ndependencies:\n  - ruleset/example\n---\n\n# Workflow: Delivery\n\n## Phases\n\n1. Implement.\n"
}

func findArtifactPlan(t *testing.T, plan Plan, kind, slug string) ArtifactPlan {
	t.Helper()
	for _, artifact := range plan.Artifacts {
		if artifact.Kind == kind && artifact.Slug == slug {
			return artifact
		}
	}
	t.Fatalf("artifact %s/%s not found", kind, slug)
	return ArtifactPlan{}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
