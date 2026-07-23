package worktree

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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
	if err := os.WriteFile(filepath.Join(issuePath, "issue.txt"), []byte("local\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, issuePath, "add", "issue.txt")
	runGit(t, issuePath, "commit", "-m", "local issue work")

	runGit(t, fixture.primary, "branch", "--track", "topic/existing", "origin/main")
	runWT(t, fixture.app, fixture.primary, "add", "topic/existing")
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
	assertBranch(t, destination, "topic/legacy")
}

func TestListDoesNotPruneAndPruneIsExplicit(t *testing.T) {
	fixture := newGitFixture(t)
	runWT(t, fixture.app, fixture.primary, "list")
	if !strings.Contains(fixture.out.String(), "STATE\tHEAD\tPATH") {
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

type gitFixture struct {
	app          *App
	out          *bytes.Buffer
	remote       string
	primary      string
	worktreeRoot string
}

func newGitFixture(t *testing.T) gitFixture {
	t.Helper()
	root := t.TempDir()
	remote := filepath.Join(root, "remotes", "example", "project.git")
	seed := filepath.Join(root, "seed")
	primary := filepath.Join(root, "primary")
	worktreeRoot := filepath.Join(root, "worktrees")

	if err := os.MkdirAll(filepath.Dir(remote), 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "init", "--bare", "--initial-branch=main", remote)
	runGit(t, root, "init", "--initial-branch=main", seed)
	runGit(t, seed, "config", "user.name", "Test User")
	runGit(t, seed, "config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(seed, "README.md"), []byte("fixture\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, seed, "add", "README.md")
	runGit(t, seed, "commit", "-m", "initial")
	runGit(t, seed, "remote", "add", "origin", remote)
	runGit(t, seed, "push", "-u", "origin", "main")
	runGit(t, root, "clone", remote, primary)
	runGit(t, primary, "config", "user.name", "Test User")
	runGit(t, primary, "config", "user.email", "test@example.com")

	out := &bytes.Buffer{}
	app := NewApp(out, &bytes.Buffer{})
	app.getenv = func(key string) string {
		if key == "GIT_WT_ROOT" {
			return worktreeRoot
		}
		return ""
	}
	return gitFixture{app: app, out: out, remote: remote, primary: primary, worktreeRoot: worktreeRoot}
}

func commitOnRemoteBranch(t *testing.T, fixture gitFixture, branch string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "branch")
	runGit(t, filepath.Dir(path), "clone", fixture.remote, path)
	runGit(t, path, "config", "user.name", "Test User")
	runGit(t, path, "config", "user.email", "test@example.com")
	runGit(t, path, "switch", "-c", branch)
	if err := os.WriteFile(filepath.Join(path, branch+".txt"), []byte("review\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, path, "add", branch+".txt")
	runGit(t, path, "commit", "-m", "review")
	runGit(t, path, "push", "-u", "origin", branch)
	return gitText(t, path, "rev-parse", "HEAD")
}

func runWT(t *testing.T, app *App, cwd string, args ...string) {
	t.Helper()
	if err := app.Run(context.Background(), cwd, args); err != nil {
		t.Fatalf("git wt %s: %v", strings.Join(args, " "), err)
	}
}

func runGit(t *testing.T, cwd string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = cwd
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
}

func gitText(t *testing.T, cwd string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = cwd
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func assertBranch(t *testing.T, path, want string) {
	t.Helper()
	if got := gitText(t, path, "symbolic-ref", "--quiet", "--short", "HEAD"); got != want {
		t.Fatalf("branch at %s = %q, want %q", path, got, want)
	}
}

func TestRepairRefusesFork(t *testing.T) {
	fixture := newGitFixture(t)
	fixture.app.resolvePR = func(context.Context, string, string, int) (PR, error) {
		return PR{HeadRefName: "fork-branch", IsCrossRepository: true, State: "OPEN"}, nil
	}
	err := fixture.app.Run(context.Background(), fixture.primary, []string{"repair", "9"})
	if err == nil || !strings.Contains(err.Error(), "from a fork") {
		t.Fatalf("repair fork error = %v", err)
	}
}

func TestUnknownCommandShowsHelp(t *testing.T) {
	fixture := newGitFixture(t)
	err := fixture.app.Run(context.Background(), fixture.primary, []string{"nope"})
	if err == nil || !strings.Contains(err.Error(), "Usage: git wt") {
		t.Fatalf("unknown command error = %v", err)
	}
}

func TestOutputFailureIsReturned(t *testing.T) {
	app := NewApp(failingWriter{}, io.Discard)
	err := app.Run(context.Background(), t.TempDir(), []string{"help"})
	if err == nil || !strings.Contains(err.Error(), "write output") {
		t.Fatalf("help output error = %v", err)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("write failed")
}

func Example() {
	fmt.Println("git wt issue 76")
	fmt.Println("git wt pr 77")
	fmt.Println("git wt repair 77")
	// Output:
	// git wt issue 76
	// git wt pr 77
	// git wt repair 77
}
