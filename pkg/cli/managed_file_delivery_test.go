package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestManagedFileDeliverySnapshotFromInitRefreshCapturesExactBoundary(t *testing.T) {
	projectRoot := t.TempDir()
	if output, err := exec.Command("git", "-C", projectRoot, "init", "--quiet").CombinedOutput(); err != nil {
		t.Fatalf("git init error = %v\n%s", err, output)
	}
	writeFile(t, filepath.Join(projectRoot, ".gitignore"), "ignored.md\n")
	changes := []initRefreshFileChange{
		{
			relativePath: "AGENTS.md",
			absolutePath: filepath.Join(projectRoot, "AGENTS.md"),
			before:       "local baseline\n",
			after:        "refreshed guidance\n",
			result:       instructionFileUpdated,
		},
		{
			relativePath: ".env",
			absolutePath: filepath.Join(projectRoot, ".env"),
			after:        "MACHINE_LOCAL=value\n",
			result:       instructionFileCreated,
		},
		{
			relativePath: "ignored.md",
			absolutePath: filepath.Join(projectRoot, "ignored.md"),
			after:        "ignored generated state\n",
			result:       instructionFileCreated,
		},
	}

	snapshot := managedFileDeliverySnapshotFromInitRefresh(projectRoot, changes)
	if len(snapshot) != 1 {
		t.Fatalf("snapshot = %#v, want one version-control-eligible command-owned path", snapshot)
	}
	got := snapshot[0]
	if got.Path != "AGENTS.md" ||
		got.Action != "update" ||
		got.PreCommandState != managedFileContentState("local baseline\n") ||
		got.ResultState != managedFileContentState("refreshed guidance\n") {
		t.Fatalf("snapshot[0] = %#v, want exact AGENTS.md before/result states", got)
	}

	instructions := strings.Join(managedFileDeliveryInstructions(projectRoot, snapshot), "\n")
	for _, expected := range []string{
		"Treat only this exact snapshot as command-owned evidence",
		"`AGENTS.md` (update; pre-command sha256:",
		"never expand the command-owned boundary from post-command status",
		"default to a new worklane without asking",
		"explicitly directed continuation of an existing lane",
		"exact existing-PR review repair, CI repair, base refresh, and ordered merge coordination",
		"never create a coordinator or corrective pull request for scope-preserving work",
		"Pull-Request Landing Plan",
		"snapshot came from the primary checkout",
		"do not adopt, transfer, stage, commit, push, restore, discard",
		"already selected writable lane",
		"contain exactly the captured command-owned change",
	} {
		if !strings.Contains(instructions, expected) {
			t.Fatalf("expected delivery instructions to contain %q, got:\n%s", expected, instructions)
		}
	}
	if strings.Contains(instructions, "`.env` (") {
		t.Fatalf("delivery instructions included machine-local .env path:\n%s", instructions)
	}
}

func TestManagedFileDeliveryInstructionsCarryRemovalOnlyChange(t *testing.T) {
	projectRoot := t.TempDir()
	relativePath := "docs/agents/README.md"
	absolutePath := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
	writeFile(t, absolutePath, "obsolete managed guidance\n")

	snapshot, err := managedFileDeliverySnapshotFromScaffold(
		projectRoot,
		nil,
		[]instructionRemovalPlan{{
			relativePath: relativePath,
			absolutePath: absolutePath,
		}},
	)
	if err != nil {
		t.Fatalf("managedFileDeliverySnapshotFromScaffold() error = %v", err)
	}

	instructions := strings.Join(managedFileDeliveryInstructions(projectRoot, snapshot), "\n")
	for _, expected := range []string{
		"`docs/agents/README.md` (remove; pre-command sha256:",
		"expected absent",
		"trigger the work-lane tripwire",
		"explicitly stage only the captured paths (including deleted paths)",
	} {
		if !strings.Contains(instructions, expected) {
			t.Fatalf("expected removal-only delivery instructions to contain %q, got:\n%s", expected, instructions)
		}
	}
}

func TestManagedFileDeliveryInstructionsWithoutSnapshotRequiresFreshBoundary(t *testing.T) {
	projectRoot := t.TempDir()
	instructions := strings.Join(managedFileDeliveryInstructions(projectRoot), "\n")

	for _, expected := range []string{
		"No exact command-owned path snapshot is present",
		"apply only the listed manual findings until a fresh snapshot exists",
		"default to a new worklane without asking",
		"explicitly directed continuation of an existing lane",
		"exact existing-PR review repair, CI repair, base refresh, and ordered merge coordination",
		"never create a coordinator or corrective pull request for scope-preserving work",
		"complete Pull-Request Landing Plan",
		"canonical non-primary writable worktree",
		"rerun the write-capable Kit command",
		"require the rerun to emit a new exact command-owned snapshot",
		"if it cannot, do not adopt managed-file changes and report the blocker",
		"explicitly stage only those paths",
		"match the snapshot exactly",
	} {
		if !strings.Contains(instructions, expected) {
			t.Fatalf("expected no-snapshot instructions to contain %q, got:\n%s", expected, instructions)
		}
	}
	for _, forbidden := range []string{
		"Only when the snapshot was produced",
		"trigger the work-lane tripwire",
	} {
		if strings.Contains(instructions, forbidden) {
			t.Fatalf("no-snapshot instructions require unavailable state %q:\n%s", forbidden, instructions)
		}
	}
}

func TestMergeManagedFileDeliverySnapshotsPreservesWholeCommandBaseline(t *testing.T) {
	primary := []managedFileDeliverySnapshot{{
		Path:            config.ConfigFileName,
		Action:          "create",
		PreCommandState: managedFileAbsentState,
		ResultState:     managedFileContentState("final config\n"),
	}}
	secondary := []managedFileDeliverySnapshot{
		{
			Path:            "./" + config.ConfigFileName,
			Action:          "update",
			PreCommandState: managedFileContentState("intermediate config\n"),
			ResultState:     managedFileContentState("final config\n"),
		},
		{
			Path:            "docs/references/rules/example.md",
			Action:          "create",
			PreCommandState: managedFileAbsentState,
			ResultState:     managedFileContentState("registry rule\n"),
		},
	}

	merged := mergeManagedFileDeliverySnapshots(primary, secondary)
	if len(merged) != 2 {
		t.Fatalf("merged = %#v, want two unique paths", merged)
	}
	for _, change := range merged {
		if change.Path == config.ConfigFileName && change != primary[0] {
			t.Fatalf("config snapshot = %#v, want whole-command baseline %#v", change, primary[0])
		}
		if strings.HasPrefix(change.Path, "./") {
			t.Fatalf("merged snapshot retained aliased path identity: %#v", change)
		}
	}
}

func TestManagedFileDeliveryRejectsPathsOutsideProject(t *testing.T) {
	projectRoot := t.TempDir()
	outsidePath := filepath.Join(projectRoot, "..", "outside.md")
	writeFile(t, outsidePath, "outside project\n")

	baseline, err := captureManagedFileDeliveryBaseline(
		projectRoot,
		[]string{"../outside.md", outsidePath},
	)
	if err != nil {
		t.Fatalf("captureManagedFileDeliveryBaseline() error = %v", err)
	}
	if len(baseline) != 0 {
		t.Fatalf("baseline = %#v, want escaping and absolute paths excluded", baseline)
	}
	if managedFileDeliveryPathEligible(projectRoot, "../outside.md") {
		t.Fatal("escaping path was considered version-control eligible")
	}

	instructions := strings.Join(
		managedFileDeliveryInstructions(projectRoot, []managedFileDeliverySnapshot{{
			Path:            "../outside.md",
			Action:          "update",
			PreCommandState: managedFileContentState("before\n"),
			ResultState:     managedFileContentState("after\n"),
		}}),
		"\n",
	)
	if strings.Contains(instructions, "`../outside.md`") {
		t.Fatalf("delivery instructions included escaping path:\n%s", instructions)
	}

	if err := os.Mkdir(filepath.Join(projectRoot, ".git"), 0o755); err != nil {
		t.Fatalf("os.Mkdir(.git) error = %v", err)
	}
	if managedFileDeliveryPathEligible(projectRoot, "AGENTS.md") {
		t.Fatal("path was considered eligible after git check-ignore failed")
	}
}

func TestManagedFileDeliveryExcludesIgnoredPathFromNestedProject(t *testing.T) {
	repositoryRoot := t.TempDir()
	if output, err := exec.Command("git", "-C", repositoryRoot, "init", "--quiet").CombinedOutput(); err != nil {
		t.Fatalf("git init error = %v\n%s", err, output)
	}
	projectRoot := filepath.Join(repositoryRoot, "nested", "project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(projectRoot) error = %v", err)
	}
	writeFile(
		t,
		filepath.Join(repositoryRoot, ".gitignore"),
		"nested/project/ignored.md\n",
	)

	if managedFileDeliveryPathEligible(projectRoot, "ignored.md") {
		t.Fatal("ignored path in nested Kit project was considered version-control eligible")
	}
}

func TestManagedFileDeliveryInstructionsRequirePostMergePrimaryClean(t *testing.T) {
	projectRoot := t.TempDir()
	if output, err := exec.Command("git", "-C", projectRoot, "init", "--quiet").CombinedOutput(); err != nil {
		t.Fatalf("git init error = %v\n%s", err, output)
	}
	snapshot := []managedFileDeliverySnapshot{{
		Path:            "AGENTS.md",
		Action:          "create",
		PreCommandState: managedFileAbsentState,
		ResultState:     managedFileContentState("guidance\n"),
	}}
	cases := []struct {
		name         string
		instructions string
	}{
		{
			name:         "with snapshot",
			instructions: strings.Join(managedFileDeliveryInstructions(projectRoot, snapshot), "\n"),
		},
		{
			name:         "without snapshot",
			instructions: strings.Join(managedFileDeliveryInstructions(projectRoot), "\n"),
		},
	}
	expected := []string{
		"`git clean -fd`",
		"Remain in this coding-agent session",
		"Do not treat pull-request creation as session completion",
		"address remaining pull-request review feedback",
		"Merge the worktree pull request only after merge is authorized",
		"names the exact authorized pull request set",
		"This leftover cleanup does not create merge authority",
		"After remaining pull-request feedback is addressed",
		"Confirm the merge first",
		"enumerate or dry-run all untracked files",
		"verify every candidate is command-owned",
		"`git clean -fd` with only those verified paths",
		"restore those exact paths in both the index and the worktree",
		"only after revalidating",
		"still match the captured command-owned snapshot",
		"if any path mismatches or is ambiguous, stop",
		"Do not run `git clean` before merge",
		"Then pull the merged default branch",
	}
	for _, test := range cases {
		for _, snippet := range expected {
			if !strings.Contains(test.instructions, snippet) {
				t.Fatalf("%s missing %q:\n%s", test.name, snippet, test.instructions)
			}
		}
		assertScopedPostMergeCleanupOrder(t, test.instructions)
	}
}
