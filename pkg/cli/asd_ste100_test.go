package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

func ste100RuleForTest(t *testing.T) registryRuleset {
	t.Helper()
	registry, err := embeddedRulesetRegistry(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range registry {
		if item.Slug == "asd-ste100" {
			return item
		}
	}
	t.Fatal("ASD-STE100 rule is absent from embedded registry")
	return registryRuleset{}
}

func TestSTE100DefaultDistribution(t *testing.T) {
	rule := ste100RuleForTest(t)
	if rule.Metadata.RegistryScope != rulesetRegistryScopeDownstream || rule.Metadata.ReadPolicyDefault != "must" || rule.Metadata.Status != "active" {
		t.Fatalf("rule is not an active required downstream default: %#v", rule.Metadata)
	}
	for _, refresh := range []bool{false, true} {
		t.Run(map[bool]string{false: "fresh init", true: "existing project"}[refresh], func(t *testing.T) {
			setupInitHome(t)
			root := t.TempDir()
			setWorkingDirectory(t, root)
			stubRulesetRegistry(t, rule)
			if refresh {
				if err := config.Save(root, config.Default()); err != nil {
					t.Fatal(err)
				}
			}
			withInitFlags(t, func() {
				initRefresh = refresh
				initOutputOnly = true
				_ = captureStdout(t, func() {
					if err := runInitForTest(initCmd, nil); err != nil {
						t.Fatal(err)
					}
				})
			})
			got, err := os.ReadFile(filepath.Join(root, rulesetTarget(rule.Slug)))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != rule.Content {
				t.Fatal("installed rule differs from canonical rule")
			}
			cfg, err := config.Load(root)
			if err != nil {
				t.Fatal(err)
			}
			artifact, ok := cfg.RegistryArtifact(rulesetKind, rule.Slug)
			if !ok || artifact.State != registryArtifactStateManaged || artifact.InstalledHash != rule.NormalizedHash {
				t.Fatalf("default rule not tracked as managed: %#v", artifact)
			}
			for _, path := range []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md"} {
				entry, err := os.ReadFile(filepath.Join(root, path))
				if err != nil {
					t.Fatal(err)
				}
				if strings.Count(string(entry), "`asd-ste100.md`") != 1 {
					t.Fatalf("%s lacks one default reference", path)
				}
				if strings.Contains(string(entry), "## Verified dictionary examples") {
					t.Fatalf("%s duplicates the canonical rule", path)
				}
			}
		})
	}
}

func TestSTE100ReconcilePreservesLocalRule(t *testing.T) {
	rule := ste100RuleForTest(t)
	root := setupRulesProject(t)
	local := rule.Content + "\nProject terminology: widget-controller.\n"
	writeFile(t, filepath.Join(root, rulesetTarget(rule.Slug)), local)
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	changes, notes, _, err := planRefreshInitRulesets(t.Context(), root, initRefreshOptions{}, cfg, nil, []registryRuleset{rule})
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || changes[0].after != local || len(notes) == 0 {
		t.Fatal("reconcile overwrites local custom content")
	}
}

func TestSTE100ContractKeepsOverrideAndVerificationBoundaries(t *testing.T) {
	rule := ste100RuleForTest(t)
	for _, marker := range []string{
		"The override changes prose style only", "permissions, approvals, and task-execution requirements still apply",
		"Do not rewrite code, commands, API names, identifiers, paths, URLs, quoted source material, or exact error messages",
		"Preserve required machine-readable formats", "Do not call a response ASD-STE100 compliant",
		"Full Issue 9 vocabulary verification was unavailable", "A fixture test does not establish",
	} {
		if !strings.Contains(rule.Content, marker) {
			t.Errorf("rule lost boundary %q", marker)
		}
	}
	contract := templates.UniversalContractBlock()
	if strings.Count(contract, "`asd-ste100.md`") != 1 || !strings.Contains(contract, "apply `asd-ste100.md` by default") || !strings.Contains(contract, "override this writing default only") {
		t.Fatal("shared contract must carry one default reference and limited user override")
	}
}
