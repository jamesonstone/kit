package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

const teamNotes = "## Team Notes\n\n- Deploy only from the release branch.\n"

func TestProjectGuidanceOutsideTheBlockSurvivesMigration(t *testing.T) {
	setupMigrationEnvironment(t)
	t.Run("current block", func(t *testing.T) {
		root := copyMigrationFixture(t, "phase2", func(root string) {
			path := filepath.Join(root, "AGENTS.md")
			stale := strings.Replace(readFile(t, path), "## ", "## Stale ", 1)
			writeFile(t, path, "Intro kept above.\n\n"+stale+"\n"+teamNotes)
		})
		setWorkingDirectory(t, root)
		migrateProject(t, root, false)
		got := readFile(t, filepath.Join(root, "AGENTS.md"))
		if !strings.Contains(got, templates.UniversalContractBlock()) || !strings.HasPrefix(got, "Intro kept above.") || !strings.HasSuffix(got, teamNotes) {
			t.Fatalf("block not replaced in place:\n%s", got)
		}
	})
	t.Run("pre-contract file", func(t *testing.T) {
		root := copyMigrationFixture(t, "v2", func(root string) {
			path := filepath.Join(root, "AGENTS.md")
			writeFile(t, path, readFile(t, path)+"\n"+teamNotes)
		})
		setWorkingDirectory(t, root)
		migrateProject(t, root, false)
		got := readFile(t, filepath.Join(root, "AGENTS.md"))
		want := "# AGENTS\n\n" + templates.UniversalContractBlock() + "\n" + teamNotes
		if got != want {
			t.Fatalf("migrated AGENTS.md =\n%s\nwant\n%s", got, want)
		}
	})
}

func TestEditedKitSectionsBlockMigrationUnlessForced(t *testing.T) {
	setupMigrationEnvironment(t)
	root := copyMigrationFixture(t, "v2", func(root string) {
		path := filepath.Join(root, "AGENTS.md")
		edited := strings.Replace(readFile(t, path), "- Repo-local markdown under `docs/` is the system of record", "- Repo-local markdown under `docs/` is the system of record\n- Our local edit", 1)
		writeFile(t, path, edited+"\n"+teamNotes)
	})
	setWorkingDirectory(t, root)
	before := readFile(t, filepath.Join(root, "AGENTS.md"))
	plan := migrateProject(t, root, false)
	if readFile(t, filepath.Join(root, "AGENTS.md")) != before {
		t.Fatal("edited pre-contract entry file changed without --force")
	}
	if !strings.Contains(strings.Join(plan.notes, "\n"), "AGENTS.md predates the Kit-managed contract and has edited Kit sections (PURPOSE)") {
		t.Fatalf("missing diagnostic: %v", plan.notes)
	}
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.InstructionScaffoldVersion == config.CurrentInstructionScaffoldVersion {
		t.Fatal("scaffold version recorded as current while an entry file is unmigrated")
	}
	if strings.Contains(readFile(t, filepath.Join(root, "docs", "CONSTITUTION.md")), templates.ConstitutionBaselineSection) {
		t.Fatal("Constitution baseline points at a contract the entry files do not carry yet")
	}

	migrateProject(t, root, true)
	got := readFile(t, filepath.Join(root, "AGENTS.md"))
	if got != "# AGENTS\n\n"+templates.UniversalContractBlock()+"\n"+teamNotes {
		t.Fatalf("forced migration did not keep only project sections:\n%s", got)
	}
	if changes := plannedChanges(t, root); len(changes) != 0 {
		t.Fatalf("reconcile after forced migration is not a no-op: %v", changes)
	}
}

func TestMarkerProblemsAreReportedNotGuessed(t *testing.T) {
	setupMigrationEnvironment(t)
	root := copyMigrationFixture(t, "phase2", func(root string) {
		agents := filepath.Join(root, "AGENTS.md")
		writeFile(t, agents, strings.Replace(readFile(t, agents), templates.UniversalContractEndMarker, "", 1)+teamNotes)
		claude := filepath.Join(root, "CLAUDE.md")
		writeFile(t, claude, readFile(t, claude)+"\n"+templates.UniversalContractBlock())
	})
	setWorkingDirectory(t, root)
	incomplete := readFile(t, filepath.Join(root, "AGENTS.md"))
	plan := migrateProject(t, root, false)
	if readFile(t, filepath.Join(root, "AGENTS.md")) != incomplete {
		t.Fatal("file with incomplete markers was rewritten")
	}
	if !strings.Contains(strings.Join(plan.notes, "\n"), "AGENTS.md has incomplete Kit-managed contract markers") {
		t.Fatalf("missing marker diagnostic: %v", plan.notes)
	}
	if got := strings.Count(readFile(t, filepath.Join(root, "CLAUDE.md")), templates.UniversalContractBeginMarker); got != 1 {
		t.Fatalf("duplicate blocks not collapsed: %d", got)
	}
}

func TestRetiredRuleEditsAndProjectRulesArePreserved(t *testing.T) {
	setupMigrationEnvironment(t)
	custom := templates.BuildRuleset("team-style", []string{"team"})
	root := copyMigrationFixture(t, "v3-precontract", func(root string) {
		path := filepath.Join(root, rulesetTarget("work-lane-gating"))
		writeFile(t, path, readFile(t, path)+"\n- Team-specific lane rule.\n")
		writeFile(t, filepath.Join(root, rulesetTarget("team-style")), custom)
	})
	setWorkingDirectory(t, root)
	plan := migrateProject(t, root, false)
	if !strings.Contains(readFile(t, filepath.Join(root, rulesetTarget("work-lane-gating"))), "Team-specific lane rule") {
		t.Fatal("edited retired rule was removed")
	}
	if readFile(t, filepath.Join(root, rulesetTarget("team-style"))) != custom {
		t.Fatal("project rule changed")
	}
	cfg, _ := config.Load(root)
	if _, ok := cfg.RegistryArtifact(rulesetKind, "work-lane-gating"); ok {
		t.Fatal("edited retired rule is still registered")
	}
	if !strings.Contains(strings.Join(plan.notes, "\n"), "docs/references/rules/work-lane-gating.md (retired rule with local edits") {
		t.Fatalf("missing kept-file note: %v", plan.notes)
	}
}

func TestMissingGeneratedFilesAreRestored(t *testing.T) {
	setupMigrationEnvironment(t)
	root := copyMigrationFixture(t, "phase2", nil)
	runGitForSourceAuditTest(t, root, "rm", "-q", "CLAUDE.md", rulesetTarget("deletion-safety"))
	commitMigrationFixture(t, root)
	setWorkingDirectory(t, root)
	migrateProject(t, root, false)
	if readFile(t, filepath.Join(root, "CLAUDE.md")) != templates.InstructionEntryFile("CLAUDE.md") {
		t.Fatal("CLAUDE.md not restored")
	}
	if !strings.Contains(readFile(t, filepath.Join(root, rulesetTarget("deletion-safety"))), "slug: deletion-safety") {
		t.Fatal("default rule not restored")
	}
}
