package registry

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestReconcileRemoteOnlyAndMissingArtifacts(t *testing.T) {
	registryRoot, sourceCfg := writeTestRegistry(t, "Registry purpose.", "Registry rule.")
	remoteProject := initializedProject(t, registryRoot, sourceCfg)
	missingProject := initializedProject(t, registryRoot, sourceCfg)
	updateTestRegistry(t, registryRoot, "Registry purpose.", "Remote-only update.")

	remotePlan, err := BuildReconcilePlan(context.Background(), remoteProject, LocalSource{Root: registryRoot}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := findArtifactPlan(t, remotePlan, KindRuleset, "example"); got.Action != "update" || got.State != StateManaged {
		t.Fatalf("remote-only artifact = %#v", got)
	}

	missingPath := filepath.Join(missingProject, "docs/references/rules/example.md")
	if err := os.Remove(missingPath); err != nil {
		t.Fatal(err)
	}
	missingPlan, err := BuildReconcilePlan(context.Background(), missingProject, LocalSource{Root: registryRoot}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := findArtifactPlan(t, missingPlan, KindRuleset, "example"); got.Action != "create" || got.State != StateManaged {
		t.Fatalf("missing artifact = %#v", got)
	}
	if err := ApplyPlan(missingProject, missingPlan); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(readTestFile(t, missingPath), "Remote-only update.") {
		t.Fatal("missing artifact was not restored from current registry content")
	}
}

func TestReconcileRetiresOnlyUnchangedArtifacts(t *testing.T) {
	registryRoot, sourceCfg := writeTestRegistry(t, "Registry purpose.", "Registry rule.")
	managedProject := initializedProject(t, registryRoot, sourceCfg)
	customProject := initializedProject(t, registryRoot, sourceCfg)
	customPath := filepath.Join(customProject, "docs/references/workflows/delivery.md")
	writeTestFile(t, customPath, readTestFile(t, customPath)+"\n## Project Gate\n\nPreserve this.\n")
	removeTestArtifact(t, registryRoot, KindWorkflow, "delivery")

	managedPlan, err := BuildReconcilePlan(context.Background(), managedProject, LocalSource{Root: registryRoot}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := findArtifactPlan(t, managedPlan, KindWorkflow, "delivery"); got.Action != "delete" || got.State != StateMissing {
		t.Fatalf("retired managed artifact = %#v", got)
	}
	if err := ApplyPlan(managedProject, managedPlan); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(managedProject, "docs/references/workflows/delivery.md")); !os.IsNotExist(err) {
		t.Fatalf("retired managed artifact remains: %v", err)
	}

	customPlan, err := BuildReconcilePlan(context.Background(), customProject, LocalSource{Root: registryRoot}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(customPlan.Diagnostics) == 0 || !strings.Contains(customPlan.Diagnostics[0], "preserved as local-custom") {
		t.Fatalf("custom retirement diagnostics = %v", customPlan.Diagnostics)
	}
	if err := ApplyPlan(customProject, customPlan); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(readTestFile(t, customPath), "Preserve this.") {
		t.Fatal("retired customized artifact was deleted")
	}
}

func removeTestArtifact(t *testing.T, root, kind, slug string) {
	t.Helper()
	path := filepath.Join(root, "registry/catalog.yaml")
	var catalog Catalog
	if err := yaml.Unmarshal([]byte(readTestFile(t, path)), &catalog); err != nil {
		t.Fatal(err)
	}
	filtered := catalog.Artifacts[:0]
	for _, artifact := range catalog.Artifacts {
		if artifact.Kind != kind || artifact.Slug != slug {
			filtered = append(filtered, artifact)
		}
	}
	catalog.Artifacts = filtered
	content, err := yaml.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, path, string(content))
}
