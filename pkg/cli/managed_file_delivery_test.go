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

// One batched check must classify a mix of paths exactly as per-path checks
// would: ignored and secret-like paths excluded, others included.
func TestManagedFileDeliveryEligiblePathsBatchesMixedPaths(t *testing.T) {
	root := t.TempDir()
	runGitForSourceAuditTest(t, root, "init", "-q", "-b", "main")
	writeFile(t, filepath.Join(root, ".gitignore"), "ignored.md\nbuild/\n")
	paths := []string{"AGENTS.md", "ignored.md", "build/out.md", ".env", "docs/x.md", "../outside.md", "./docs/y.md"}
	got := managedFileDeliveryEligiblePaths(root, paths)
	want := map[string]bool{"AGENTS.md": true, "docs/x.md": true, "docs/y.md": true}
	if len(got) != len(want) {
		t.Fatalf("eligible = %v, want %v", got, want)
	}
	for path := range want {
		if !got[path] || managedFileDeliveryPathEligible(root, path) != got[path] {
			t.Fatalf("eligible = %v, want %v", got, want)
		}
	}
}
