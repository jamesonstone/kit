package cli

import (
	stdreflect "reflect"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

func TestFreshInitReconcilesAsNoOp(t *testing.T) {
	setupMigrationEnvironment(t)
	root := freshInitProject(t)
	if changes := plannedChanges(t, root); len(changes) != 0 {
		t.Fatalf("reconcile after a fresh init plans changes: %v", changes)
	}
	state := captureKitOwnedState(t, root)
	if state.ScaffoldVersion != config.CurrentInstructionScaffoldVersion || len(state.RetiredPresent) != 0 || len(state.RequiredMissing) != 0 {
		t.Fatalf("fresh init state = %#v", state)
	}
	for path, block := range state.EntryBlocks {
		if block+"\n" != templates.UniversalContractBlock() {
			t.Errorf("%s block differs from the universal contract", path)
		}
	}
}

// Every released generation converges on the structure a fresh init creates,
// and a second reconcile changes nothing.
func TestLegacyGenerationsConvergeOnFreshInitStructure(t *testing.T) {
	setupMigrationEnvironment(t)
	fresh := captureKitOwnedState(t, freshInitProject(t))
	for _, name := range migrationFixtures {
		t.Run(name, func(t *testing.T) {
			root := copyMigrationFixture(t, name, nil)
			setWorkingDirectory(t, root)
			migrateProject(t, root, false)
			if changes := plannedChanges(t, root); len(changes) != 0 {
				t.Fatalf("second reconcile is not a no-op: %v", changes)
			}
			migrated := captureKitOwnedState(t, root)
			if !stdreflect.DeepEqual(migrated, fresh) {
				t.Fatalf("migrated Kit-owned state differs from fresh init:\nmigrated: %#v\nfresh:    %#v", migrated, fresh)
			}
		})
	}
}

func TestLegacyMigrationRemovesRetiredStateAndKeepsOptionalRules(t *testing.T) {
	setupMigrationEnvironment(t)
	root := copyMigrationFixture(t, "v3-precontract", nil)
	setWorkingDirectory(t, root)
	plan := migrateProject(t, root, false)

	removed := map[string]bool{}
	for _, change := range plan.changes {
		if change.result == instructionFileRemoved {
			removed[change.relativePath] = true
		}
	}
	for _, path := range []string{
		"docs/agents/README.md",
		"docs/references/workflows/implementation-delivery.md",
		"docs/references/worktrees.md",
		"docs/references/rules/work-lane-gating.md",
		"docs/references/rules/source-file-size.md",
	} {
		if !removed[path] {
			t.Errorf("%s was not removed; removed = %v", path, removed)
		}
	}
	assertFileDoesNotExist(t, root+"/docs/agents")
	assertFileDoesNotExist(t, root+"/docs/references/workflows")
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, artifact := range cfg.Registry.Artifacts {
		if isRetiredRuleset(artifact.Slug) {
			t.Errorf("retired rule %s still registered", artifact.Slug)
		}
	}
	// Optional rules the project already had stay installed and managed.
	if artifact, ok := cfg.RegistryArtifact(rulesetKind, "llms-txt"); !ok || artifact.State != registryArtifactStateManaged {
		t.Errorf("optional llms-txt rule not kept as managed: %#v", artifact)
	}
	// Every removal is recoverable from Git.
	runGitForSourceAuditTest(t, root, "checkout", "HEAD", "--", "docs/references/rules/work-lane-gating.md")
	if !strings.Contains(readFile(t, root+"/docs/references/rules/work-lane-gating.md"), "work-lane-gating") {
		t.Fatal("removed retired rule could not be restored from Git")
	}
}

func TestLegacyConfigDropsObsoleteKeysOnly(t *testing.T) {
	setupMigrationEnvironment(t)
	root := copyMigrationFixture(t, "v1", nil)
	setWorkingDirectory(t, root)
	migrateProject(t, root, false)
	content := readFile(t, root+"/.kit.yaml")
	for _, obsolete := range []string{"goal_percentage", "allow_out_of_order", "branching"} {
		if strings.Contains(content, obsolete+":") {
			t.Errorf(".kit.yaml still has obsolete key %s:\n%s", obsolete, content)
		}
	}

	// Formatting and comments alone never trigger a rewrite.
	commented := "# project note\n" + content
	writeFile(t, root+"/.kit.yaml", commented)
	if changes := plannedChanges(t, root); len(changes) != 0 {
		t.Fatalf("comment-only .kit.yaml difference planned changes: %v", changes)
	}
}
