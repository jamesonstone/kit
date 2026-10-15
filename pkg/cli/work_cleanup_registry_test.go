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
		for _, obsolete := range []string{"fast-forward-only update", "command-owned leftovers", "remove only what clear evidence"} {
			if strings.Contains(content, obsolete) {
				t.Errorf("%s restores obsolete cleanup authority %q", path, obsolete)
			}
		}
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
	migrateProject(t, root, false)
	if got := readFile(t, filepath.Join(root, rulesetTarget("work-cleanup"))); got != content {
		t.Fatal("second reconcile overwrote custom cleanup")
	}
	if changes := plannedChanges(t, root); len(changes) != 0 {
		t.Fatalf("custom cleanup does not converge: %v", changes)
	}
}

// TestWorkCleanupContractRunsQuietly checks the always-loaded contract makes cleanup automatic and prompt-free.
func TestWorkCleanupContractRunsQuietly(t *testing.T) {
	block := templates.UniversalContractBlock()
	for _, want := range []string{
		"apply `work-cleanup` quietly",
		"inventory and retain persistent local state as `KEEP_LOCAL_STATE`",
		"never ask for routine cleanup approval",
		"Removal requires a separate human request naming the targets and complete recovery evidence",
		workCleanupRoute,
	} {
		if !strings.Contains(block, want) {
			t.Errorf("universal contract lacks %q", want)
		}
	}
	if strings.Contains(block, "hard-delete confirmation") {
		t.Error("universal contract still asks for cleanup confirmation")
	}
}

// TestWorkCleanupNeverAsksAndStaysNarrow checks the rule never prompts and only runs low-risk commands.
func TestWorkCleanupNeverAsksAndStaysNarrow(t *testing.T) {
	content := workCleanupRule(t).Content
	for _, invariant := range []string{
		"Keep every", "worktree and branch as `KEEP_LOCAL_STATE`",
		"Do not automatically remove worktrees, delete branches, prune refs or caches",
		"remove scratch data, or update the primary checkout",
		"do not move or copy local files as cleanup",
		"no outstanding work", "Do not close unrelated issues or tasks",
		"A separate human request must identify",
		"recovery of the complete relevant local state",
		"A remote SHA is not a substitute for that proof",
		"If any state or recovery evidence is unknown, retain the target",
		"Never force removal, reset or clean a checkout",
		"This rule grants no destructive authority",
	} {
		if !strings.Contains(content, invariant) {
			t.Errorf("lost cleanup invariant: %s", invariant)
		}
	}
	for _, removed := range []string{"git branch -D", "go clean -modcache", "npm cache clean", "confirmation list", "Ask once for confirmation"} {
		if strings.Contains(content, removed) {
			t.Errorf("rule still includes alarming or prompting action %q", removed)
		}
	}
	for _, state := range workCleanupStates {
		if !strings.Contains(content, "`"+state+"`") {
			t.Errorf("rule does not explain classifier state %s", state)
		}
	}
	// Keep the inventory policy short.
	if words := len(strings.Fields(content)); words > 1400 {
		t.Fatalf("work-cleanup grew to %d words; keep it short", words)
	}
}

// TestWorkCleanupUpgradesReleasedPolicy replaces a managed v3.0.44 install and stale contract edits.
func TestWorkCleanupUpgradesReleasedPolicy(t *testing.T) {
	setupMigrationEnvironment(t)
	old := readFile(t, filepath.Join("testdata", "work-cleanup", "v3.0.44.md"))
	root := freshInitProject(t)
	rule := workCleanupRule(t)
	writeFile(t, filepath.Join(root, rulesetTarget(rule.Slug)), old)
	hash, err := normalizedRulesetContentHash(old, rule.Metadata.Status)
	if err != nil {
		t.Fatal(err)
	}
	cfg := loadMigratedConfig(t, root)
	recordRulesetRegistryState(cfg, rule, registryArtifactStateManaged, hash)
	if err := config.Save(root, cfg); err != nil {
		t.Fatal(err)
	}
	project := "\n## Project retention\n\nKeep local results for the human.\n"
	for _, path := range instructionFiles(cfg) {
		content := readFile(t, filepath.Join(root, path))
		content = strings.Replace(content, "inventory and retain persistent local state", "remove only what clear evidence shows a remote can restore", 1)
		writeFile(t, filepath.Join(root, path), content+project)
	}
	constitution := readFile(t, filepath.Join(root, "docs", "CONSTITUTION.md"))
	writeFile(t, filepath.Join(root, "docs", "CONSTITUTION.md"), constitution+project)
	commitMigrationFixture(t, root)
	migrateProject(t, root, false)
	assertWorkCleanupInstalled(t, root, rule)
	for _, path := range instructionFiles(cfg) {
		if !strings.HasSuffix(readFile(t, filepath.Join(root, path)), project) {
			t.Errorf("%s lost project retention guidance", path)
		}
	}
	if !strings.HasSuffix(readFile(t, filepath.Join(root, "docs", "CONSTITUTION.md")), project) {
		t.Fatal("reconcile lost constitutional retention guidance")
	}
	migrateProject(t, root, false)
	assertWorkCleanupInstalled(t, root, rule)
	if changes := plannedChanges(t, root); len(changes) != 0 {
		t.Fatalf("released cleanup upgrade does not converge: %v", changes)
	}
}
