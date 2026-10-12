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

// workCleanupRoute is the contract line that routes completed work to the cleanup rule.
const workCleanupRoute = "- Work completes, stops, or is superseded; before a new lane or session end; or free disk space is low: `work-cleanup.md`.\n"

// workCleanupRule returns the embedded cleanup rule.
func workCleanupRule(t *testing.T) registryRuleset {
	t.Helper()
	registry, err := embeddedRulesetRegistry(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, rule := range registry {
		if rule.Slug == "work-cleanup" {
			return rule
		}
	}
	t.Fatal("work-cleanup is not shipped")
	return registryRuleset{}
}

// assertWorkCleanupInstalled verifies the managed rule and its automatic routing in every instruction file.
func assertWorkCleanupInstalled(t *testing.T, root string, rule registryRuleset) {
	t.Helper()
	if got := readFile(t, filepath.Join(root, rulesetTarget(rule.Slug))); got != rule.Content {
		t.Fatal("installed work-cleanup differs from the embedded rule")
	}
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	artifact, ok := cfg.RegistryArtifact(rulesetKind, rule.Slug)
	if !ok || artifact.State != registryArtifactStateManaged || artifact.InstalledHash != rule.NormalizedHash {
		t.Fatalf("work-cleanup registry state = %#v", artifact)
	}
	for _, path := range instructionFiles(cfg) {
		content := readFile(t, filepath.Join(root, path))
		if !strings.Contains(content, templates.UniversalContractBlock()) || !strings.Contains(content, workCleanupRoute) {
			t.Errorf("%s lacks current work-cleanup routing", path)
		}
	}
}

// TestWorkCleanupInstallsByDefaultAcrossGenerations verifies default installation and idempotent adoption.
func TestWorkCleanupInstallsByDefaultAcrossGenerations(t *testing.T) {
	setupMigrationEnvironment(t)
	rule := workCleanupRule(t)
	if rule.Metadata.RegistryScope != rulesetRegistryScopeDownstream {
		t.Fatal("work-cleanup must be a default downstream rule")
	}
	t.Run("fresh", func(t *testing.T) {
		assertWorkCleanupInstalled(t, freshInitProject(t), rule)
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
			writeFile(t, file, strings.ReplaceAll(readFile(t, file), workCleanupRoute, ""))
		}
		commitMigrationFixture(t, root)
		migrateProject(t, root, false)
		assertWorkCleanupInstalled(t, root, rule)
		if changes := plannedChanges(t, root); len(changes) != 0 {
			t.Fatalf("second reconcile plans changes: %v", changes)
		}
	})
	for _, generation := range migrationFixtures {
		t.Run(generation, func(t *testing.T) {
			root := copyMigrationFixture(t, generation, nil)
			migrateProject(t, root, false)
			assertWorkCleanupInstalled(t, root, rule)
			if changes := plannedChanges(t, root); len(changes) != 0 {
				t.Fatalf("second reconcile plans changes: %v", changes)
			}
		})
	}
}

// TestWorkCleanupPreservesProjectOwnedRule verifies a project can tune or narrow the rule.
func TestWorkCleanupPreservesProjectOwnedRule(t *testing.T) {
	setupMigrationEnvironment(t)
	content := readFile(t, filepath.Join("..", "..", rulesetTarget("work-cleanup"))) + "\nProject note: keep the Go module cache.\n"
	root := copyMigrationFixture(t, "v3-precontract", func(root string) {
		writeFile(t, filepath.Join(root, rulesetTarget("work-cleanup")), content)
	})
	migrateProject(t, root, false)
	if got := readFile(t, filepath.Join(root, rulesetTarget("work-cleanup"))); got != content {
		t.Fatal("reconcile overwrote project-owned work-cleanup guidance")
	}
	artifact, ok := loadMigratedConfig(t, root).RegistryArtifact(rulesetKind, "work-cleanup")
	if !ok || artifact.State != registryArtifactStateLocalCustom {
		t.Fatalf("project-owned work-cleanup registry state = %#v", artifact)
	}
}

// TestWorkCleanupContractRunsAutomatically checks the always-loaded contract makes cleanup automatic.
func TestWorkCleanupContractRunsAutomatically(t *testing.T) {
	block := templates.UniversalContractBlock()
	for _, want := range []string{
		"apply `work-cleanup` immediately without asking",
		"clear evidence shows a remote can restore",
		"batch anything else into one hard-delete confirmation",
		workCleanupRoute,
	} {
		if !strings.Contains(block, want) {
			t.Errorf("universal contract lacks %q", want)
		}
	}
}

// TestWorkCleanupKeepsSafetyBoundaries checks that automatic pruning never bypasses hard-delete rules.
func TestWorkCleanupKeepsSafetyBoundaries(t *testing.T) {
	content := workCleanupRule(t).Content
	for _, invariant := range []string{
		"never weakens `deletion-safety`, `delivery`, or `infrastructure-change-approval`",
		"you must observe both in the current session",
		"Do not ask, do not offer, and do not defer cleanup",
		"Never infer completion or restorability from names, dates alone, or memory of an earlier session",
		"checked it by digest, size, or a successful fetch",
		"Never delete a branch, stash, or worktree to unblock a new lane",
		"never use `--force` on any Git command",
		"Never touch the primary checkout",
		"Never run `git clean -X`",
		"Close, never delete. Never post to Slack.",
		"Never run `docker volume prune`",
		"Ask once for confirmation of the complete hard-delete list",
		"It does not mean the work is completed",
	} {
		if !strings.Contains(content, invariant) {
			t.Errorf("lost cleanup invariant: %s", invariant)
		}
	}
	for _, state := range workCleanupStates {
		if !strings.Contains(content, "`"+state+"`") {
			t.Errorf("rule does not explain classifier state %s", state)
		}
	}
	if words := len(strings.Fields(content)); words > 1500 {
		t.Fatalf("work-cleanup grew to %d words; keep its context cost bounded", words)
	}
}
