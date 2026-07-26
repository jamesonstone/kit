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

func TestPreparePullRequestRepairCreatesThenReusesExactHeadWorktree(t *testing.T) {
	fixture := newGitFixture(t)
	headOID := commitOnRemoteBranch(t, fixture, "review-head")
	fixture.app.resolvePR = func(_ context.Context, _ string, repository string, number int) (PR, error) {
		if repository != "example/project" || number != 77 {
			t.Fatalf("resolve PR target = %s#%d", repository, number)
		}
		return PR{
			HeadRefName: "review-head",
			HeadRefOID:  headOID,
			State:       "OPEN",
			URL:         "https://github.com/example/project/pull/77",
		}, nil
	}

	prepared, err := fixture.app.PreparePullRequestRepair(
		context.Background(),
		fixture.primary,
		77,
		false,
	)
	if err != nil {
		t.Fatalf("PreparePullRequestRepair() error = %v", err)
	}
	wantPath := filepath.Join(fixture.worktreeRoot, "example", "project", "review-head")
	if prepared.Path != wantPath ||
		prepared.Branch != "review-head" ||
		!prepared.Created ||
		prepared.Repository != "example/project" ||
		prepared.Number != 77 ||
		prepared.URL != "https://github.com/example/project/pull/77" ||
		prepared.HeadRefOID != headOID {
		t.Fatalf("unexpected prepared PR repair: %#v", prepared)
	}
	assertBranch(t, wantPath, "review-head")

	reused, err := fixture.app.PreparePullRequestRepair(
		context.Background(),
		fixture.primary,
		77,
		false,
	)
	if err != nil {
		t.Fatalf("reused PreparePullRequestRepair() error = %v", err)
	}
	wantInfo, err := os.Stat(wantPath)
	if err != nil {
		t.Fatal(err)
	}
	reusedInfo, err := os.Stat(reused.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(wantInfo, reusedInfo) || reused.Created {
		t.Fatalf("expected exact worktree reuse, got %#v", reused)
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
	fmt.Println("git wt cd GH-76")
	fmt.Println(`cd "$(git wt path GH-76)"`)
	fmt.Println("git wt pr 77")
	fmt.Println("git wt repair 77")
	// Output:
	// git wt issue 76
	// git wt cd GH-76
	// cd "$(git wt path GH-76)"
	// git wt pr 77
	// git wt repair 77
}
