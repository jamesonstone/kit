package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

const globalLedgerForTest = `work_ledger:
  provider: plane
  owner_label: agent:global-example
  plane:
    workspace: personal-default
    project: GLOBAL
    states:
      active: My active state
`

func TestReconcilePreservesGlobalLedgerProvenance(t *testing.T) {
	for _, tc := range []struct {
		name, override              string
		force, empty, issueTemplate bool
	}{
		{name: "global only"},
		{name: "existing empty file", empty: true, issueTemplate: true},
		{name: "comments-only config", override: "# project note\n", empty: true, issueTemplate: true},
		{name: "nested project and explicit empty label", override: "work_ledger:\n  owner_label: ''\n  plane:\n    project: REPO\n"},
		{name: "partial project target under explicit force", override: "work_ledger:\n  provider: plane\n  plane:\n    project: REPO\n", force: true},
		{name: "provider disabled", override: "work_ledger:\n  provider: none\n"},
		{name: "null override", override: "work_ledger: null\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := setupInitHome(t)
			globalPath := filepath.Join(home, ".config", "kit", config.ConfigFileName)
			writeFile(t, globalPath, globalLedgerForTest)
			root := t.TempDir()
			projectPath := filepath.Join(root, config.ConfigFileName)
			before := tc.override
			if !tc.empty {
				before = "schema_version: 1\n" + before
			}
			writeFile(t, projectPath, before)
			initial, err := config.Load(root)
			if err != nil {
				t.Fatal(err)
			}

			opts := initRefreshOptions{
				dryRun: true, diff: true, force: tc.force, issueTemplate: tc.issueTemplate,
				files: []string{config.ConfigFileName, agentsMDPath},
			}
			stubRulesetRegistry(t)
			if tc.force {
				opts.files = []string{config.ConfigFileName}
			}
			preview, err := buildInitRefreshPlan(t.Context(), root, opts)
			if err != nil {
				t.Fatal(err)
			}
			assertPlanOmitsGlobalLedger(t, preview.changes)
			if got := readFile(t, projectPath); got != before {
				t.Fatalf("preview wrote config: %q", got)
			}

			opts.dryRun = false
			plan, err := buildInitRefreshPlan(t.Context(), root, opts)
			if err != nil {
				t.Fatal(err)
			}
			if err := applyInitRefreshFileChangesAtomically(plan.changes); err != nil {
				t.Fatal(err)
			}
			assertPlanOmitsGlobalLedger(t, plan.changes)
			final, err := config.Load(root)
			if err != nil || final.WorkLedger != initial.WorkLedger {
				t.Fatalf("effective ledger changed: before=%+v after=%+v err=%v", initial.WorkLedger, final.WorkLedger, err)
			}
			if tc.issueTemplate && !final.Issues.IssueTemplateEnabled() {
				t.Fatal("issue template opt-in was not written")
			}
			if final.SchemaVersion != config.CurrentSchemaVersion {
				t.Fatalf("schema version = %d", final.SchemaVersion)
			}
			if got := readFile(t, globalPath); got != globalLedgerForTest {
				t.Fatalf("global file changed: %q", got)
			}
		})
	}
}

func assertPlanOmitsGlobalLedger(t *testing.T, changes []initRefreshFileChange) {
	t.Helper()
	for _, change := range changes {
		for _, value := range []string{"personal-default", "agent:global-example", "GLOBAL", "My active state"} {
			if strings.Contains(change.after, value) {
				t.Errorf("planned %s includes global value %q", change.relativePath, value)
			}
		}
	}
}

func TestInstallRuleKeepsGlobalLedgerOutOfRegistryWrite(t *testing.T) {
	home := setupInitHome(t)
	writeFile(t, filepath.Join(home, ".config", "kit", config.ConfigFileName), globalLedgerForTest)
	root := t.TempDir()
	path := filepath.Join(root, config.ConfigFileName)
	writeFile(t, path, "schema_version: 2\nwork_ledger:\n  owner_label: ''\n  plane:\n    project: REPO\n")
	rule := registryRulesetForTest("test-rule", []string{"git"})
	if _, err := applyRegistryRulesetSelection(root, []registrySelectorEntry{{Registry: rule, DesiredActive: true}}); err != nil {
		t.Fatal(err)
	}
	assertPlanOmitsGlobalLedger(t, []initRefreshFileChange{{relativePath: config.ConfigFileName, after: readFile(t, path)}})
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.RegistryArtifact(rulesetKind, rule.Slug); !ok {
		t.Fatal("rule was not registered")
	}
	if !cfg.WorkLedger.Enabled() || cfg.WorkLedger.OwnerLabel != "" || cfg.WorkLedger.Plane.Project != "REPO" {
		t.Fatalf("effective ledger = %+v", cfg.WorkLedger)
	}
}

func TestReconcilePreservesAndPlansFromInheritedProjectSettings(t *testing.T) {
	home := setupInitHome(t)
	globalPath := filepath.Join(home, ".config", "kit", config.ConfigFileName)
	writeFile(t, globalPath, "specs_dir: docs/custom\nconstitution_path: policy/CONSTITUTION.md\nshell:\n  command: [example-shell]\n")
	root := t.TempDir()
	path := filepath.Join(root, config.ConfigFileName)
	writeFile(t, path, "schema_version: 1\nshell: null\n")
	plan, err := buildInitRefreshPlan(t.Context(), root, initRefreshOptions{
		files: []string{config.ConfigFileName, "policy/CONSTITUTION.md"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := applyInitRefreshFileChangesAtomically(plan.changes); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(root)
	if err != nil || cfg.SpecsDir != "docs/custom" || len(cfg.Shell.Command) != 0 {
		t.Fatalf("reconcile changed effective settings: cfg=%+v err=%v", cfg, err)
	}
	if body := readFile(t, path); strings.Contains(body, "specs_dir:") || strings.Contains(body, "constitution_path:") || !strings.Contains(body, "shell: null") {
		t.Fatalf("reconcile wrote inherited settings or removed null: %q", body)
	}
	if body := readFile(t, filepath.Join(root, "policy", "CONSTITUTION.md")); !strings.Contains(body, "CONSTITUTION") {
		t.Fatal("reconcile did not scaffold the configured Constitution path")
	}
}
