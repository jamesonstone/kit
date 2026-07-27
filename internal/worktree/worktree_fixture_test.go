package worktree

import (
	"bytes"
	"context"
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

type gitFixture struct {
	app          *App
	out          *bytes.Buffer
	remote       string
	primary      string
	worktreeRoot string
}
