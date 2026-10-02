package cli

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/document"
)

// contextEvidenceRule returns the optional embedded evidence rule.
func contextEvidenceRule(t *testing.T) registryRuleset {
	t.Helper()
	registry, err := embeddedRulesetRegistry(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, rule := range registry {
		if rule.Slug == "context-evidence" {
			return rule
		}
	}
	t.Fatal("context-evidence is not shipped")
	return registryRuleset{}
}

// TestContextEvidenceIsOptionalAndPreservesLocalEdits verifies opt-in installation and retained project edits.
func TestContextEvidenceIsOptionalAndPreservesLocalEdits(t *testing.T) {
	rule := contextEvidenceRule(t)
	if rule.Metadata.RegistryScope != rulesetRegistryScopeOptional || rule.Metadata.ReadPolicyDefault != document.ReferenceReadPolicyConditional {
		t.Fatalf("unexpected installation/read policy: %#v", rule.Metadata)
	}
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)
	stubRulesetRegistry(t, rule)
	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	changes, _, changed, err := planRefreshInitRulesets(t.Context(), projectRoot, initRefreshOptions{}, cfg, nil, []registryRuleset{rule})
	if err != nil || changed || len(changes) != 0 {
		t.Fatalf("default refresh installs optional guidance: changes=%v changed=%v err=%v", changes, changed, err)
	}
	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)
	if err := runRulesAdd(cmd, []string{rule.Slug}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(projectRoot, rulesetTarget(rule.Slug))
	if got := readFile(t, path); got != rule.Content {
		t.Fatal("rules add did not install the exact embedded guidance")
	}
	cfg, err = config.Load(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	installed, ok := cfg.RegistryArtifact(rulesetKind, rule.Slug)
	if !ok || installed.InstalledHash != rule.NormalizedHash {
		t.Fatalf("installed artifact lacks embedded identity: %#v", installed)
	}
	edited := rule.Content + "\nProject note: retain only safe, material sources.\n"
	if err := os.WriteFile(path, []byte(edited), 0600); err != nil {
		t.Fatal(err)
	}
	result, err := syncRulesetRegistryContent(rule, installed, edited, false)
	if err != nil || result.content != edited || result.state != registryArtifactStateLocalCustom {
		t.Fatalf("local customization was not preserved: %#v, %v", result, err)
	}
}

// TestContextEvidenceCoversGeneralRuleUseAndMissingInputs checks source and missing-input safeguards.
func TestContextEvidenceCoversGeneralRuleUseAndMissingInputs(t *testing.T) {
	content := contextEvidenceRule(t).Content
	for _, invariant := range []string{
		"This rule never exempts a required rule",
		"Read the rule once per session unless its source changes",
		"Expand when dependencies, exceptions, contradictions, or uncertainty require it",
		"Never reconstruct missing requirements from a summary as though they were observed",
		"Do not capture credentials or secrets",
		"An unchanged file does not prove a finding still applies",
	} {
		if !strings.Contains(content, invariant) {
			t.Errorf("lost context/evidence invariant: %s", invariant)
		}
	}
	// Check the outline exposes source gaps instead of inviting silent recovery.
	outline := contextEvidenceFence(t, content, "markdown", 0)
	for _, field := range []string{"Original input:", "UNCAPTURED", "Observation:", "Inference:", "Open questions:", "Next:"} {
		if !strings.Contains(outline, field) {
			t.Errorf("record outline omits %q", field)
		}
	}
}

// TestContextEvidenceFreshInitKeepsDefaultFootprint verifies no default installation or instruction overhead.
func TestContextEvidenceFreshInitKeepsDefaultFootprint(t *testing.T) {
	rule := contextEvidenceRule(t)
	root := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, root)
	stubRulesetRegistry(t, rule)
	withInitFlags(t, func() {
		initOutputOnly = true
		_ = captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatal(err)
			}
		})
	})
	if document.Exists(filepath.Join(root, rulesetTarget(rule.Slug))) {
		t.Fatal("fresh init installed optional guidance")
	}
	for _, entry := range []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md"} {
		if strings.Contains(readFile(t, filepath.Join(root, entry)), "context-evidence") {
			t.Fatalf("%s added optional guidance to default context", entry)
		}
	}
	if words := len(strings.Fields(rule.Content)); words > 900 {
		t.Fatalf("optional guidance grew to %d words; keep its one-time context cost bounded", words)
	}
}
