package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

// TestMintLifecycleInstallsAcrossGenerations verifies default routing and idempotent adoption.
func TestMintLifecycleInstallsAcrossGenerations(t *testing.T) {
	setupMigrationEnvironment(t)
	registry, err := embeddedRulesetRegistry(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	var rule registryRuleset
	for _, item := range registry {
		if item.Slug == "mint-deployment-lifecycle" {
			rule = item
		}
	}
	if rule.Metadata.RegistryScope != rulesetRegistryScopeDownstream {
		t.Fatal("Mint lifecycle must be a core downstream rule")
	}
	assertInstalled := func(t *testing.T, root string) {
		t.Helper()
		if got := readFile(t, filepath.Join(root, rulesetTarget(rule.Slug))); got != rule.Content {
			t.Fatal("installed Mint lifecycle differs from the embedded rule")
		}
		cfg, err := config.Load(root)
		if err != nil {
			t.Fatal(err)
		}
		artifact, ok := cfg.RegistryArtifact(rulesetKind, rule.Slug)
		if !ok || artifact.State != registryArtifactStateManaged || artifact.InstalledHash != rule.NormalizedHash {
			t.Fatalf("Mint lifecycle registry state = %#v", artifact)
		}
		for _, path := range instructionFiles(cfg) {
			content := readFile(t, filepath.Join(root, path))
			if !strings.Contains(content, templates.UniversalContractBlock()) || !strings.Contains(content, "`mint-deployment-lifecycle.md`") {
				t.Errorf("%s lacks current Mint lifecycle routing", path)
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
			content = strings.ReplaceAll(content, "- Mint adoption, releases, deployments, environment state, hotfix, roll-forward or rollback: `mint-deployment-lifecycle.md` when the project uses Mint.\n", "")
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

// TestMintLifecyclePreservesProjectOwnedRule verifies local guidance survives migration.
func TestMintLifecyclePreservesProjectOwnedRule(t *testing.T) {
	setupMigrationEnvironment(t)
	content := readFile(t, filepath.Join("..", "..", rulesetTarget("mint-deployment-lifecycle"))) + "\nProject-specific recovery guidance.\n"
	root := copyMigrationFixture(t, "v3-precontract", func(root string) {
		writeFile(t, filepath.Join(root, rulesetTarget("mint-deployment-lifecycle")), content)
	})
	plan := migrateProject(t, root, false)
	if got := readFile(t, filepath.Join(root, rulesetTarget("mint-deployment-lifecycle"))); got != content {
		t.Fatal("reconcile overwrote project-owned Mint lifecycle guidance")
	}
	artifact, ok := loadMigratedConfig(t, root).RegistryArtifact(rulesetKind, "mint-deployment-lifecycle")
	if !ok || artifact.State != registryArtifactStateLocalCustom {
		t.Fatalf("project-owned Mint lifecycle registry state = %#v", artifact)
	}
	if !strings.Contains(strings.Join(plan.notes, "\n"), rulesetTarget("mint-deployment-lifecycle")) {
		t.Fatal("reconcile did not report preserved project-owned content")
	}
}
