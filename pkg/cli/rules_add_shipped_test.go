package cli

import (
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestRunRulesAddInstallsShippedRuleBySlug(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)
	if err := runRulesAdd(cmd, []string{"llms-txt"}); err != nil {
		t.Fatalf("runRulesAdd() error = %v", err)
	}
	registry, err := embeddedRulesetRegistry(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	var shipped registryRuleset
	for _, item := range registry {
		if item.Slug == "llms-txt" {
			shipped = item
		}
	}
	content := readFile(t, filepath.Join(projectRoot, "docs", "references", "rules", "llms-txt.md"))
	if !strings.Contains(content, "# Ruleset: llms-txt") || strings.Contains(content, "Capture durable project guidance") {
		t.Fatalf("expected the shipped llms-txt rule, got:\n%s", content)
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	artifact, ok := cfg.RegistryArtifact(rulesetKind, "llms-txt")
	if !ok || artifact.State != registryArtifactStateManaged || artifact.InstalledHash != shipped.NormalizedHash {
		t.Fatalf("registry artifact = %#v, ok = %v", artifact, ok)
	}
	if err := runRulesAdd(cmd, []string{"work-lane-gating"}); err == nil || !strings.Contains(err.Error(), "retired") {
		t.Fatalf("expected retired rule error, got %v", err)
	}
}

func TestRunRulesAddForceReplacesEditedShippedRule(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)
	if err := runRulesAdd(cmd, []string{"llms-txt"}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(projectRoot, "docs", "references", "rules", "llms-txt.md")
	writeFile(t, path, readFile(t, path)+"\n- local edit\n")
	if err := runRulesAdd(cmd, []string{"llms-txt"}); err == nil {
		t.Fatal("expected existing rule to be preserved without --force")
	}
	rulesAddForce = true
	if err := runRulesAdd(cmd, []string{"llms-txt"}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(readFile(t, path), "local edit") {
		t.Fatal("--force did not replace the edited rule")
	}
}
