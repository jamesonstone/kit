package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

const sopRuleSlug = "standard-operating-procedures"

func assertSOPScaffold(t *testing.T, root string) {
	t.Helper()
	if got := readFile(t, filepath.Join(root, templates.SOPGuidePath)); got != templates.SOPGuide {
		t.Fatal("installed SOP guide differs from its template")
	}
	cfg := loadMigratedConfig(t, root)
	registry, err := embeddedRulesetRegistry(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, rule := range registry {
		if rule.Slug != sopRuleSlug {
			continue
		}
		found = true
		if rule.Metadata.RegistryScope != rulesetRegistryScopeDownstream {
			t.Fatal("SOP guidance must be a default downstream rule")
		}
		if readFile(t, filepath.Join(root, rulesetTarget(rule.Slug))) != rule.Content {
			t.Fatal("installed SOP rule differs from its embedded source")
		}
		artifact, ok := cfg.RegistryArtifact(rulesetKind, rule.Slug)
		if !ok || artifact.State != registryArtifactStateManaged || artifact.InstalledHash != rule.NormalizedHash {
			t.Fatalf("SOP rule registry state = %#v", artifact)
		}
	}
	if !found {
		t.Fatal("SOP rule is not embedded")
	}
	for _, path := range instructionFiles(cfg) {
		content := readFile(t, filepath.Join(root, path))
		if !strings.Contains(content, templates.UniversalContractBlock()) || !strings.Contains(content, sopRuleSlug+".md") {
			t.Errorf("%s does not route agents to SOP guidance", path)
		}
	}
}

func TestSOPScaffoldInstallsAcrossGenerations(t *testing.T) {
	setupMigrationEnvironment(t)
	t.Run("init", func(t *testing.T) {
		root := freshInitProject(t)
		assertSOPScaffold(t, root)
		if changes := plannedChanges(t, root); len(changes) != 0 {
			t.Fatalf("reconcile after init changes files: %v", changes)
		}
	})
	for _, generation := range migrationFixtures {
		t.Run(generation, func(t *testing.T) {
			root := copyMigrationFixture(t, generation, nil)
			migrateProject(t, root, false)
			assertSOPScaffold(t, root)
			if changes := plannedChanges(t, root); len(changes) != 0 {
				t.Fatalf("second reconcile changes files: %v", changes)
			}
		})
	}
}

func TestSOPScaffoldPreservesProjectContent(t *testing.T) {
	setupMigrationEnvironment(t)
	root := freshInitProject(t)
	guide := "# Our procedures\n\nFollow the project-maintained index.\n"
	procedurePath := filepath.Join(root, "docs/sops/workbook-sync/SOP.md")
	procedure := "# Workbook sync\n\nProject-owned operating steps.\n"
	writeFile(t, filepath.Join(root, templates.SOPGuidePath), guide)
	writeFile(t, procedurePath, procedure)
	for _, opts := range []initRefreshOptions{
		{}, {force: true}, {force: true, files: []string{templates.SOPGuidePath}},
	} {
		plan, err := buildInitRefreshPlan(t.Context(), root, opts)
		if err != nil {
			t.Fatal(err)
		}
		if err := applyInitRefreshFileChangesAtomically(plan.changes); err != nil {
			t.Fatal(err)
		}
		if got := readFile(t, filepath.Join(root, templates.SOPGuidePath)); got != guide {
			t.Fatal("reconcile overwrote the project SOP guide")
		}
		if got := readFile(t, procedurePath); got != procedure {
			t.Fatal("reconcile overwrote a project SOP")
		}
	}
}

func TestSOPScaffoldDryRunAndTargetedCreate(t *testing.T) {
	setupMigrationEnvironment(t)
	root := t.TempDir()
	cfg := config.Default()
	if err := config.Save(root, cfg); err != nil {
		t.Fatal(err)
	}
	opts := initRefreshOptions{dryRun: true, files: []string{templates.SOPGuidePath}}
	plan, err := buildInitRefreshPlan(t.Context(), root, opts)
	if err != nil {
		t.Fatal(err)
	}
	assertFileDoesNotExist(t, filepath.Join(root, "docs/sops"))
	if len(plan.changes) != 1 || plan.changes[0].relativePath != templates.SOPGuidePath || plan.changes[0].result != instructionFileCreated {
		t.Fatalf("unexpected targeted plan: %#v", plan.changes)
	}
	if _, err := runInitRefreshWithSnapshot(root, opts); err != nil {
		t.Fatal(err)
	}
	assertFileDoesNotExist(t, filepath.Join(root, "docs/sops"))
	opts.dryRun = false
	if _, err := runInitRefreshWithSnapshot(root, opts); err != nil {
		t.Fatal(err)
	}
	if readFile(t, filepath.Join(root, templates.SOPGuidePath)) != templates.SOPGuide {
		t.Fatal("targeted reconcile did not create the SOP guide")
	}
	assertFileDoesNotExist(t, filepath.Join(root, "AGENTS.md"))
}

func TestSOPScaffoldDoesNotWriteThroughLinks(t *testing.T) {
	setupMigrationEnvironment(t)
	for _, dangling := range []bool{false, true} {
		t.Run(map[bool]string{false: "directory", true: "dangling file"}[dangling], func(t *testing.T) {
			root, external := t.TempDir(), t.TempDir()
			if err := config.Save(root, config.Default()); err != nil {
				t.Fatal(err)
			}
			link, target := filepath.Join(root, "docs/sops"), external
			if dangling {
				link = filepath.Join(root, templates.SOPGuidePath)
				target = filepath.Join(external, "README.md")
			}
			if err := os.MkdirAll(filepath.Dir(link), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, link); err != nil {
				t.Fatal(err)
			}
			plan, err := buildInitRefreshPlan(t.Context(), root, initRefreshOptions{files: []string{templates.SOPGuidePath}, force: true})
			if err != nil {
				t.Fatal(err)
			}
			if err := applyInitRefreshFileChangesAtomically(plan.changes); err != nil {
				t.Fatal(err)
			}
			assertFileDoesNotExist(t, filepath.Join(external, "README.md"))
			if len(plan.notes) == 0 {
				t.Fatal("linked scaffold needs a preservation diagnostic")
			}
		})
	}
}

func TestSOPScaffoldMissingGuideIsDiagnosed(t *testing.T) {
	setupMigrationEnvironment(t)
	root := freshInitProject(t)
	if err := os.Remove(filepath.Join(root, templates.SOPGuidePath)); err != nil {
		t.Fatal(err)
	}
	for _, finding := range auditInitScaffoldArtifacts(root) {
		if finding.FilePath == filepath.Join(root, templates.SOPGuidePath) {
			return
		}
	}
	t.Fatal("missing SOP guide was not diagnosed")
}
