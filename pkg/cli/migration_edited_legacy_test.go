package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

func editedLegacyFixture(t *testing.T) string {
	t.Helper()
	setupMigrationEnvironment(t)
	return copyMigrationFixture(t, "v2", func(root string) {
		path := filepath.Join(root, "AGENTS.md")
		edited := strings.Replace(readFile(t, path), "- Repo-local markdown under `docs/` is the system of record", "- Repo-local markdown under `docs/` is the system of record\n- Our local edit", 1)
		writeFile(t, path, edited+"\n"+teamNotes)
	})
}

// Without --force an edited legacy project keeps its file, reports why, never
// claims version 4, and reaches a stable state with identical diagnostics.
func TestEditedLegacyProjectIsStableWithoutForce(t *testing.T) {
	root := editedLegacyFixture(t)
	setWorkingDirectory(t, root)
	before := readFile(t, filepath.Join(root, "AGENTS.md"))
	first := migrateProject(t, root, false)
	commitMigrationFixture(t, root)
	second, err := buildInitRefreshPlan(t.Context(), root, initRefreshOptions{outputOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range second.changes {
		if change.result != instructionFileSkipped {
			t.Fatalf("second pass changes %s %s", change.result, change.relativePath)
		}
	}
	// The entry-file diagnostic repeats unchanged; the one-time .kit.yaml
	// dropped-keys note appears only on the pass that rewrote the file.
	entryNote := func(notes []string) string {
		for _, note := range notes {
			if strings.HasPrefix(note, "AGENTS.md predates") {
				return note
			}
		}
		return ""
	}
	if entryNote(first.notes) == "" || entryNote(first.notes) != entryNote(second.notes) {
		t.Fatalf("entry diagnostics changed between passes:\n%v\n%v", first.notes, second.notes)
	}
	if readFile(t, filepath.Join(root, "AGENTS.md")) != before {
		t.Fatal("edited entry file changed")
	}
	if v := loadMigratedConfig(t, root).InstructionScaffoldVersion; v == config.CurrentInstructionScaffoldVersion {
		t.Fatal("incomplete migration recorded as current")
	}
	for _, path := range []string{"CLAUDE.md", ".github/copilot-instructions.md"} {
		if !strings.Contains(readFile(t, filepath.Join(root, path)), templates.UniversalContractBeginMarker) {
			t.Errorf("unedited %s was not migrated", path)
		}
	}
}

// The note's own remedy, `--force --file <entry>`, followed by a normal
// reconcile completes the migration and converges.
func TestEditedLegacyProjectCompletesWithTargetedForce(t *testing.T) {
	root := editedLegacyFixture(t)
	setWorkingDirectory(t, root)
	migrateProject(t, root, false)
	plan, err := buildInitRefreshPlan(t.Context(), root, initRefreshOptions{force: true, files: []string{"AGENTS.md"}, outputOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := applyInitRefreshFileChangesAtomically(plan.changes); err != nil {
		t.Fatal(err)
	}
	for _, change := range plan.changes {
		if change.relativePath != "AGENTS.md" && change.result != instructionFileSkipped {
			t.Fatalf("targeted force touched %s", change.relativePath)
		}
	}
	if got := readFile(t, filepath.Join(root, "AGENTS.md")); got != "# AGENTS\n\n"+templates.UniversalContractBlock()+"\n"+teamNotes {
		t.Fatalf("forced AGENTS.md =\n%s", got)
	}
	migrateProject(t, root, false)
	if v := loadMigratedConfig(t, root).InstructionScaffoldVersion; v != config.CurrentInstructionScaffoldVersion {
		t.Fatalf("scaffold version = %d after completing migration", v)
	}
	if changes := plannedChanges(t, root); len(changes) != 0 {
		t.Fatalf("not converged: %v", changes)
	}
}

func TestBrokenMarkersNeverClaimCurrentVersion(t *testing.T) {
	setupMigrationEnvironment(t)
	root := copyMigrationFixture(t, "v3-precontract", func(root string) {
		path := filepath.Join(root, "CLAUDE.md")
		writeFile(t, path, readFile(t, path)+"\n"+templates.UniversalContractBeginMarker+"\n")
	})
	setWorkingDirectory(t, root)
	migrateProject(t, root, false)
	if v := loadMigratedConfig(t, root).InstructionScaffoldVersion; v == config.CurrentInstructionScaffoldVersion {
		t.Fatal("broken markers recorded as current")
	}
	if changes := plannedChanges(t, root); len(changes) != 0 {
		t.Fatalf("second pass not stable: %v", changes)
	}
}
