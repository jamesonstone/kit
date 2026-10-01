package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

func TestDeploymentRecoveryInstallsAcrossGenerations(t *testing.T) {
	setupMigrationEnvironment(t)
	registry, err := embeddedRulesetRegistry(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	var rule registryRuleset
	for _, item := range registry {
		if item.Slug == "deployment-recovery" {
			rule = item
		}
	}
	if rule.Metadata.RegistryScope != rulesetRegistryScopeDownstream {
		t.Fatal("deployment recovery must be a core downstream rule")
	}
	assertInstalled := func(t *testing.T, root string) {
		t.Helper()
		if got := readFile(t, filepath.Join(root, rulesetTarget(rule.Slug))); got != rule.Content {
			t.Fatal("installed deployment recovery differs from the embedded rule")
		}
		cfg, err := config.Load(root)
		if err != nil {
			t.Fatal(err)
		}
		artifact, ok := cfg.RegistryArtifact(rulesetKind, rule.Slug)
		if !ok || artifact.State != registryArtifactStateManaged || artifact.InstalledHash != rule.NormalizedHash {
			t.Fatalf("deployment recovery registry state = %#v", artifact)
		}
		for _, path := range instructionFiles(cfg) {
			content := readFile(t, filepath.Join(root, path))
			if !strings.Contains(content, templates.UniversalContractBlock()) || !strings.Contains(content, "`deployment-recovery.md`") {
				t.Errorf("%s lacks current deployment recovery routing", path)
			}
		}
	}

	t.Run("fresh", func(t *testing.T) {
		assertInstalled(t, freshInitProject(t))
	})
	t.Run("current without newly shipped rule", func(t *testing.T) {
		root := freshInitProject(t)
		if err := os.Remove(filepath.Join(root, rulesetTarget(rule.Slug))); err != nil {
			t.Fatal(err)
		}
		cfg := loadMigratedConfig(t, root)
		var artifacts []config.RegistryArtifact
		for _, artifact := range cfg.Registry.Artifacts {
			if artifact.Slug != rule.Slug {
				artifacts = append(artifacts, artifact)
			}
		}
		cfg.Registry.Artifacts = artifacts
		if err := config.Save(root, cfg); err != nil {
			t.Fatal(err)
		}
		for _, path := range instructionFiles(cfg) {
			file := filepath.Join(root, path)
			content := readFile(t, file)
			content = strings.ReplaceAll(content, "- Following a release or deployment, including a missing or failed pipeline: `deployment-recovery.md`.\n", "")
			writeFile(t, file, content)
		}
		commitMigrationFixture(t, root)
		migrateProject(t, root, false)
		assertInstalled(t, root)
		if changes := plannedChanges(t, root); len(changes) != 0 {
			t.Fatalf("second reconcile plans changes: %v", changes)
		}
	})
	for _, generation := range migrationFixtures {
		t.Run(generation, func(t *testing.T) {
			root := copyMigrationFixture(t, generation, nil)
			migrateProject(t, root, false)
			assertInstalled(t, root)
			if changes := plannedChanges(t, root); len(changes) != 0 {
				t.Fatalf("second reconcile plans changes: %v", changes)
			}
		})
	}
}

func TestDeploymentRecoveryPreservesProjectOwnedRule(t *testing.T) {
	setupMigrationEnvironment(t)
	content := readFile(t, filepath.Join("..", "..", rulesetTarget("deployment-recovery"))) + "\nProject-specific recovery guidance.\n"
	root := copyMigrationFixture(t, "v3-precontract", func(root string) {
		writeFile(t, filepath.Join(root, rulesetTarget("deployment-recovery")), content)
	})
	plan := migrateProject(t, root, false)
	if got := readFile(t, filepath.Join(root, rulesetTarget("deployment-recovery"))); got != content {
		t.Fatal("reconcile overwrote project-owned deployment recovery guidance")
	}
	artifact, ok := loadMigratedConfig(t, root).RegistryArtifact(rulesetKind, "deployment-recovery")
	if !ok || artifact.State != registryArtifactStateLocalCustom {
		t.Fatalf("project-owned deployment recovery registry state = %#v", artifact)
	}
	if !strings.Contains(strings.Join(plan.notes, "\n"), rulesetTarget("deployment-recovery")) {
		t.Fatal("reconcile did not report preserved project-owned content")
	}
}
