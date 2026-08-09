package registry

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitValidatesCompleteRegistryBeforeWriting(t *testing.T) {
	registryRoot, sourceCfg := writeTestRegistry(t, "Purpose.", "Rule.")
	writeTestFile(t, filepath.Join(registryRoot, "rules/example.md"), testRuleDocument("Tampered.", "Rule."))
	project := t.TempDir()
	agentsPath := filepath.Join(project, "AGENTS.md")
	writeTestFile(t, agentsPath, "# Project agents\n")

	_, err := BuildInitPlan(context.Background(), project, LocalSource{Root: registryRoot}, sourceCfg)
	if err == nil || !strings.Contains(err.Error(), "digest") {
		t.Fatalf("init error = %v", err)
	}
	if content := readTestFile(t, agentsPath); content != "# Project agents\n" {
		t.Fatalf("init changed routing before validation: %q", content)
	}
	if _, err := os.Stat(filepath.Join(project, ProjectFile)); !os.IsNotExist(err) {
		t.Fatalf("init left partial configuration: %v", err)
	}
}

func TestInitPreservesExistingArtifactAsLocalCustom(t *testing.T) {
	registryRoot, sourceCfg := writeTestRegistry(t, "Registry purpose.", "Registry rule.")
	project := t.TempDir()
	path := filepath.Join(project, "docs/references/rules/example.md")
	custom := testRuleDocument("Local purpose.", "Local rule.")
	writeTestFile(t, path, custom)
	plan, err := BuildInitPlan(context.Background(), project, LocalSource{Root: registryRoot}, sourceCfg)
	if err != nil {
		t.Fatal(err)
	}
	if got := findArtifactPlan(t, plan, KindRuleset, "example"); got.State != StateLocalCustom || got.Action != "none" {
		t.Fatalf("existing artifact plan = %#v", got)
	}
	if err := ApplyPlan(project, plan); err != nil {
		t.Fatal(err)
	}
	if readTestFile(t, path) != custom {
		t.Fatal("init overwrote project-owned artifact content")
	}
}
