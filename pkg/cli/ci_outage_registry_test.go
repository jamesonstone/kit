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

// ciOutageRoute is the contract line that routes a hosted CI outage to its rule.
const ciOutageRoute = "- Hosted CI cannot start or run because of an outage or exhausted minutes or billing: `ci-outage.md`.\n"

// ciOutageRule returns the embedded CI-outage rule.
func ciOutageRule(t *testing.T) registryRuleset {
	t.Helper()
	registry, err := embeddedRulesetRegistry(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, rule := range registry {
		if rule.Slug == "ci-outage" {
			return rule
		}
	}
	t.Fatal("ci-outage is not shipped")
	return registryRuleset{}
}

// assertCIOutageInstalled verifies the managed rule and its routing in every instruction file.
func assertCIOutageInstalled(t *testing.T, root string, rule registryRuleset) {
	t.Helper()
	if got := readFile(t, filepath.Join(root, rulesetTarget(rule.Slug))); got != rule.Content {
		t.Fatal("installed ci-outage differs from the embedded rule")
	}
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	artifact, ok := cfg.RegistryArtifact(rulesetKind, rule.Slug)
	if !ok || artifact.State != registryArtifactStateManaged || artifact.InstalledHash != rule.NormalizedHash {
		t.Fatalf("ci-outage registry state = %#v", artifact)
	}
	for _, path := range instructionFiles(cfg) {
		content := readFile(t, filepath.Join(root, path))
		if !strings.Contains(content, templates.UniversalContractBlock()) || !strings.Contains(content, ciOutageRoute) {
			t.Errorf("%s lacks current ci-outage routing", path)
		}
	}
}

// TestCIOutageInstallsByDefaultAcrossGenerations verifies kit init installs the rule and
// kit reconcile adds it to existing projects of every generation, idempotently.
func TestCIOutageInstallsByDefaultAcrossGenerations(t *testing.T) {
	setupMigrationEnvironment(t)
	rule := ciOutageRule(t)
	if rule.Metadata.RegistryScope != rulesetRegistryScopeDownstream {
		t.Fatal("ci-outage must be a default downstream rule")
	}
	t.Run("init", func(t *testing.T) {
		assertCIOutageInstalled(t, freshInitProject(t), rule)
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
			writeFile(t, file, strings.ReplaceAll(readFile(t, file), ciOutageRoute, ""))
		}
		commitMigrationFixture(t, root)
		migrateProject(t, root, false)
		assertCIOutageInstalled(t, root, rule)
		if changes := plannedChanges(t, root); len(changes) != 0 {
			t.Fatalf("second reconcile plans changes: %v", changes)
		}
	})
	for _, generation := range migrationFixtures {
		t.Run("reconcile "+generation, func(t *testing.T) {
			root := copyMigrationFixture(t, generation, nil)
			migrateProject(t, root, false)
			assertCIOutageInstalled(t, root, rule)
			if changes := plannedChanges(t, root); len(changes) != 0 {
				t.Fatalf("second reconcile plans changes: %v", changes)
			}
		})
	}
}

// TestCIOutageResolvesMergeGateConflict checks the contract and github-pr-merge name the
// exception, so the merge gates and the bypass do not read as contradictory.
func TestCIOutageResolvesMergeGateConflict(t *testing.T) {
	block := templates.UniversalContractBlock()
	for _, want := range []string{
		"The only exception is the human-approved `ci-outage` bypass, which replaces only hosted required checks that could not run",
		"under the `ci-outage` bypass, full local validation",
		ciOutageRoute,
	} {
		if !strings.Contains(block, want) {
			t.Errorf("universal contract lacks %q", want)
		}
	}
	registry, err := embeddedRulesetRegistry(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, rule := range registry {
		if rule.Slug == "github-pr-merge" && strings.Count(rule.Content, "`ci-outage`") < 2 {
			t.Error("github-pr-merge must name ci-outage as the exception to both its check gates")
		}
	}
}
