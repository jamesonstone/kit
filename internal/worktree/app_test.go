package worktree

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseRemoteIdentity(t *testing.T) {
	t.Parallel()
	for remote, want := range map[string]string{
		"git@github.com:LSMC-Bio/LabCore.git":         "LSMC-Bio/LabCore",
		"ssh://git@github.com/jamesonstone/kit.git":   "jamesonstone/kit",
		"https://github.com/patient-driven-care/mypa": "patient-driven-care/mypa",
		"/tmp/remotes/example/project.git":            "example/project",
	} {
		owner, repo, err := parseRemoteIdentity(remote)
		if err != nil {
			t.Fatalf("parse %q: %v", remote, err)
		}
		if got := owner + "/" + repo; got != want {
			t.Fatalf("parse %q = %q, want %q", remote, got, want)
		}
	}
}

func TestParseRemoteIdentityRejectsDotSegments(t *testing.T) {
	t.Parallel()
	for _, remote := range []string{
		"git@github.com:../project.git",
		"git@github.com:example/...git",
	} {
		if _, _, err := parseRemoteIdentity(remote); err == nil {
			t.Fatalf("parse %q expected an error", remote)
		}
	}
	for _, value := range []string{".", ".."} {
		if isSafeProjectPart(value) {
			t.Fatalf("project identity segment %q must be rejected", value)
		}
	}
}

func TestValidateLaneRejectsTraversal(t *testing.T) {
	t.Parallel()
	for _, lane := range []string{"", "/tmp/GH-1", "../GH-1", "topic/../GH-1", "topic//GH-1", `topic\GH-1`} {
		if _, err := validateLane(lane); err == nil {
			t.Fatalf("expected %q to be rejected", lane)
		}
	}
	for _, lane := range []string{"GH-76", "PR-77", "codex/consent-service-fix"} {
		if got, err := validateLane(lane); err != nil || got != lane {
			t.Fatalf("validate %q = %q, %v", lane, got, err)
		}
	}
}

func TestIssueAddPRRepairAndSafeRemove(t *testing.T) {
	fixture := newGitFixture(t)
	ctx := context.Background()

	runWT(t, fixture.app, fixture.primary, "issue", "76")
	issuePath := filepath.Join(fixture.worktreeRoot, "example", "project", "GH-76")
	assertBranch(t, issuePath, "GH-76")
	runGit(t, fixture.primary, "remote", "set-url", "origin", filepath.Join(fixture.worktreeRoot, "missing.git"))
	runWT(t, fixture.app, fixture.primary, "issue", "76")
	runGit(t, fixture.primary, "remote", "set-url", "origin", fixture.remote)
	if err := os.WriteFile(filepath.Join(issuePath, "issue.txt"), []byte("local\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, issuePath, "add", "issue.txt")
	runGit(t, issuePath, "commit", "-m", "local issue work")

	runGit(t, fixture.primary, "branch", "--track", "topic/existing", "origin/main")
	offlineRemote := filepath.Join(fixture.worktreeRoot, "offline", "example", "project.git")
	runGit(t, fixture.primary, "remote", "set-url", "origin", offlineRemote)
	runWT(t, fixture.app, fixture.primary, "add", "topic/existing")
	runGit(t, fixture.primary, "remote", "set-url", "origin", fixture.remote)
	topicPath := filepath.Join(fixture.worktreeRoot, "example", "project", "topic", "existing")
	assertBranch(t, topicPath, "topic/existing")

	prCommit := commitOnRemoteBranch(t, fixture, "review-head")
	runGit(t, fixture.remote, "update-ref", "refs/pull/77/head", prCommit)
	runWT(t, fixture.app, fixture.primary, "pr", "77")
	prPath := filepath.Join(fixture.worktreeRoot, "example", "project", "PR-77")
	if branch := gitText(t, prPath, "symbolic-ref", "--quiet", "--short", "HEAD"); branch != "" {
		t.Fatalf("PR lane branch = %q, want detached", branch)
	}

	fixture.app.resolvePR = func(context.Context, string, string, int) (PR, error) {
		return PR{HeadRefName: "review-head", State: "OPEN"}, nil
	}
	runWT(t, fixture.app, fixture.primary, "repair", "77")
	repairPath := filepath.Join(fixture.worktreeRoot, "example", "project", "review-head")
	assertBranch(t, repairPath, "review-head")

	if err := fixture.app.Run(ctx, fixture.primary, []string{"remove", "GH-76"}); err == nil || !strings.Contains(err.Error(), "ahead of") {
		t.Fatalf("remove unpushed issue lane error = %v", err)
	}

	runWT(t, fixture.app, fixture.primary, "remove", "PR-77")
	if _, err := os.Stat(prPath); !os.IsNotExist(err) {
		t.Fatalf("detached PR path still exists or stat failed: %v", err)
	}
}

func TestRemoveRefusesDirtyAndIgnoredMaterial(t *testing.T) {
	fixture := newGitFixture(t)
	ctx := context.Background()
	runGit(t, fixture.primary, "branch", "--track", "topic/clean", "origin/main")
	runWT(t, fixture.app, fixture.primary, "add", "topic/clean")
	path := filepath.Join(fixture.worktreeRoot, "example", "project", "topic", "clean")

	if err := os.WriteFile(filepath.Join(path, "untracked.txt"), []byte("preserve\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := fixture.app.Run(ctx, fixture.primary, []string{"remove", "topic/clean"}); err == nil || !strings.Contains(err.Error(), "refusing removal") {
		t.Fatalf("remove dirty worktree error = %v", err)
	}
	if err := os.Remove(filepath.Join(path, "untracked.txt")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, ".gitignore"), []byte("ignored.txt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, path, "add", ".gitignore")
	runGit(t, path, "commit", "-m", "add ignore")
	runGit(t, path, "push", "-u", "origin", "topic/clean")
	if err := os.WriteFile(filepath.Join(path, "ignored.txt"), []byte("preserve\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := fixture.app.Run(ctx, fixture.primary, []string{"remove", "topic/clean"}); err == nil || !strings.Contains(err.Error(), "ignored material") {
		t.Fatalf("remove ignored worktree error = %v", err)
	}
}

func TestMigratePreviewsThenMovesDirtyLegacyWorktree(t *testing.T) {
	fixture := newGitFixture(t)
	writeEnvironmentSource(t, fixture, "TOKEN=do-not-link\n")
	legacy := filepath.Join(fixture.worktreeRoot, "project-topic-legacy")
	runGit(t, fixture.primary, "branch", "topic/legacy", "origin/main")
	runGit(t, fixture.primary, "worktree", "add", legacy, "topic/legacy")
	if err := os.WriteFile(filepath.Join(legacy, "dirty.txt"), []byte("preserve\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	fixture.out.Reset()
	runWT(t, fixture.app, fixture.primary, "migrate")
	destination := filepath.Join(fixture.worktreeRoot, "example", "project", "topic", "legacy")
	if !strings.Contains(fixture.out.String(), "WOULD MOVE") || !strings.Contains(fixture.out.String(), destination) {
		t.Fatalf("unexpected migration preview:\n%s", fixture.out.String())
	}
	if _, err := os.Stat(legacy); err != nil {
		t.Fatalf("preview moved legacy worktree: %v", err)
	}

	fixture.out.Reset()
	runWT(t, fixture.app, fixture.primary, "migrate", "--apply")
	if data, err := os.ReadFile(filepath.Join(destination, "dirty.txt")); err != nil || string(data) != "preserve\n" {
		t.Fatalf("dirty state not preserved: data=%q err=%v", data, err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatalf("legacy path still exists or stat failed: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(destination, environmentFileName)); !os.IsNotExist(err) {
		t.Fatalf("migration created an environment link or lstat failed: %v", err)
	}
	assertBranch(t, destination, "topic/legacy")
}

func TestListDoesNotPruneAndPruneIsExplicit(t *testing.T) {
	fixture := newGitFixture(t)
	runWT(t, fixture.app, fixture.primary, "list")
	if !strings.Contains(fixture.out.String(), "STATE\tHEAD\tLAST UPDATED\tPATH") {
		t.Fatalf("list output:\n%s", fixture.out.String())
	}
	fixture.out.Reset()
	runWT(t, fixture.app, fixture.primary, "prune", "--dry-run")
	if !strings.Contains(fixture.out.String(), "Dry run complete") {
		t.Fatalf("prune dry-run output:\n%s", fixture.out.String())
	}
	fixture.out.Reset()
	runWT(t, fixture.app, fixture.primary, "prune")
	if !strings.Contains(fixture.out.String(), "Pruned stale worktree metadata") {
		t.Fatalf("prune output:\n%s", fixture.out.String())
	}
}

func TestListSortsByLastUpdatedByDefaultAndSupportsOtherAttributes(t *testing.T) {
	fixture := newGitFixture(t)
	runGit(t, fixture.primary, "branch", "topic/old", "origin/main")
	runGit(t, fixture.primary, "branch", "topic/new", "origin/main")
	runWT(t, fixture.app, fixture.primary, "add", "topic/old")
	runWT(t, fixture.app, fixture.primary, "add", "topic/new")
	oldPath := filepath.Join(fixture.worktreeRoot, "example", "project", "topic", "old")
	newPath := filepath.Join(fixture.worktreeRoot, "example", "project", "topic", "new")
	time.Sleep(1100 * time.Millisecond)
	runGit(t, newPath, "commit", "--allow-empty", "-m", "newer")

	fixture.out.Reset()
	runWT(t, fixture.app, fixture.primary, "list")
	lines := strings.Split(strings.TrimSpace(fixture.out.String()), "\n")
	if len(lines) < 4 || !strings.Contains(lines[1], fixture.primary) || !strings.Contains(lines[2], newPath) {
		t.Fatalf("default list should pin the primary worktree before the newest lane:\n%s", fixture.out.String())
	}
	if !strings.Contains(fixture.out.String(), oldPath) || !strings.Contains(fixture.out.String(), "\nclean\tmain\t") {
		t.Fatalf("default list should retain the older worktrees:\n%s", fixture.out.String())
	}
	columns := strings.Split(lines[2], "\t")
	if len(columns) != 4 {
		t.Fatalf("list row should have four columns: %q", lines[2])
	}
	if _, err := time.Parse("Jan 02, 2006 15:04", columns[2]); err != nil {
		t.Fatalf("last updated value should show a local human-readable minute, got %q: %v", columns[2], err)
	}

	fixture.out.Reset()
	runWT(t, fixture.app, fixture.primary, "list", "--sort", "path")
	pathLines := strings.Split(strings.TrimSpace(fixture.out.String()), "\n")
	if len(pathLines) < 4 || !strings.Contains(pathLines[1], fixture.primary) {
		t.Fatalf("path sort should start with the primary path:\n%s", fixture.out.String())
	}

	fixture.out.Reset()
	runWT(t, fixture.app, fixture.primary, "list", "--sort=path", "--reverse")
	reverseLines := strings.Split(strings.TrimSpace(fixture.out.String()), "\n")
	if len(reverseLines) < 4 || !strings.Contains(reverseLines[1], fixture.primary) {
		t.Fatalf("primary pin should override reverse path sorting:\n%s", fixture.out.String())
	}

	fixture.out.Reset()
	runWT(t, fixture.app, fixture.primary, "list", "--sort=path", "--reverse", "--root-position=bottom")
	bottomLines := strings.Split(strings.TrimSpace(fixture.out.String()), "\n")
	if len(bottomLines) < 4 || !strings.Contains(bottomLines[len(bottomLines)-1], fixture.primary) {
		t.Fatalf("bottom root position should pin the primary path last:\n%s", fixture.out.String())
	}
}

func TestParseListOptionsRejectsUnknownSort(t *testing.T) {
	if _, err := parseListOptions([]string{"--sort", "branch"}); err == nil {
		t.Fatal("unknown sort attribute should fail")
	}
	if _, err := parseListOptions([]string{"--root-position", "middle"}); err == nil {
		t.Fatal("unknown root position should fail")
	}
}

type gitFixture struct {
	app          *App
	out          *bytes.Buffer
	remote       string
	primary      string
	worktreeRoot string
}
