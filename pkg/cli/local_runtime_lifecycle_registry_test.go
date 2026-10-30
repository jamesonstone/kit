package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

// localRuntimeRoute is the contract line that routes local environment work to its rule.
const localRuntimeRoute = "- Starting, restarting, or stopping a local environment, including for validation or post-merge checks, or opening a terminal for one: `local-runtime-lifecycle.md`.\n"

// localRuntimeRule returns the embedded local runtime rule.
func localRuntimeRule(t *testing.T) registryRuleset {
	t.Helper()
	registry, err := embeddedRulesetRegistry(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, rule := range registry {
		if rule.Slug == "local-runtime-lifecycle" {
			return rule
		}
	}
	t.Fatal("local-runtime-lifecycle is not shipped")
	return registryRuleset{}
}

// assertLocalRuntimeInstalled verifies the managed rule and its routing in every instruction file.
func assertLocalRuntimeInstalled(t *testing.T, root string, rule registryRuleset) {
	t.Helper()
	if got := readFile(t, filepath.Join(root, rulesetTarget(rule.Slug))); got != rule.Content {
		t.Fatal("installed local-runtime-lifecycle differs from the embedded rule")
	}
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	artifact, ok := cfg.RegistryArtifact(rulesetKind, rule.Slug)
	if !ok || artifact.State != registryArtifactStateManaged || artifact.InstalledHash != rule.NormalizedHash {
		t.Fatalf("local-runtime-lifecycle registry state = %#v", artifact)
	}
	for _, path := range instructionFiles(cfg) {
		content := readFile(t, filepath.Join(root, path))
		if !strings.Contains(content, templates.UniversalContractBlock()) || !strings.Contains(content, localRuntimeRoute) {
			t.Errorf("%s lacks current local-runtime-lifecycle routing", path)
		}
	}
}

// TestLocalRuntimeLifecycleInstallsByDefaultAcrossGenerations verifies kit init installs the rule and
// kit reconcile adds it to existing projects of every generation, idempotently.
func TestLocalRuntimeLifecycleInstallsByDefaultAcrossGenerations(t *testing.T) {
	setupMigrationEnvironment(t)
	rule := localRuntimeRule(t)
	if rule.Metadata.RegistryScope != rulesetRegistryScopeDownstream {
		t.Fatal("local-runtime-lifecycle must be a default downstream rule")
	}
	t.Run("init", func(t *testing.T) {
		assertLocalRuntimeInstalled(t, freshInitProject(t), rule)
	})
	t.Run("reconcile current without newly shipped rule", func(t *testing.T) {
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
			writeFile(t, file, strings.ReplaceAll(readFile(t, file), localRuntimeRoute, ""))
		}
		commitMigrationFixture(t, root)
		migrateProject(t, root, false)
		assertLocalRuntimeInstalled(t, root, rule)
		if changes := plannedChanges(t, root); len(changes) != 0 {
			t.Fatalf("second reconcile plans changes: %v", changes)
		}
	})
	for _, generation := range migrationFixtures {
		t.Run("reconcile "+generation, func(t *testing.T) {
			root := copyMigrationFixture(t, generation, nil)
			migrateProject(t, root, false)
			assertLocalRuntimeInstalled(t, root, rule)
			if changes := plannedChanges(t, root); len(changes) != 0 {
				t.Fatalf("second reconcile plans changes: %v", changes)
			}
		})
	}
}
