package worktree

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathPrintsOnlyExactRegisteredWorktree(t *testing.T) {
	fixture := newGitFixture(t)
	runWT(t, fixture.app, fixture.primary, "issue", "76", "--no-link-env")
	issuePath := filepath.Join(fixture.worktreeRoot, "example", "project", "GH-76")

	fixture.out.Reset()
	runWT(t, fixture.app, fixture.primary, "path", "GH-76")
	output := fixture.out.String()
	got := strings.TrimSuffix(output, "\n")
	if output != got+"\n" || !samePath(got, issuePath) {
		t.Fatalf("path output = %q, want filesystem-equivalent path %q and one newline", output, issuePath)
	}

	runGit(t, fixture.primary, "branch", "--track", "topic/navigate", "origin/main")
	runWT(t, fixture.app, fixture.primary, "add", "topic/navigate", "--no-link-env")
	nestedPath := filepath.Join(fixture.worktreeRoot, "example", "project", "topic", "navigate")

	fixture.out.Reset()
	runWT(t, fixture.app, fixture.primary, "path", "topic/navigate")
	output = fixture.out.String()
	got = strings.TrimSuffix(output, "\n")
	if output != got+"\n" || !samePath(got, nestedPath) {
		t.Fatalf("nested path output = %q, want filesystem-equivalent path %q and one newline", output, nestedPath)
	}

	err := fixture.app.Run(context.Background(), fixture.primary, []string{"path", "GH-999"})
	if err == nil || !strings.Contains(err.Error(), "not an exact registered worktree") {
		t.Fatalf("unregistered lane error = %v", err)
	}
	err = fixture.app.Run(context.Background(), fixture.primary, []string{"path"})
	if err == nil || err.Error() != "usage: git wt path <lane>" {
		t.Fatalf("path usage error = %v", err)
	}
}

func TestCDOpensShellInExactRegisteredWorktree(t *testing.T) {
	fixture := newGitFixture(t)
	runWT(t, fixture.app, fixture.primary, "issue", "76", "--no-link-env")
	want := filepath.Join(fixture.worktreeRoot, "example", "project", "GH-76")
	var got string
	fixture.app.runShell = func(_ context.Context, path string) error {
		got = path
		return nil
	}

	runWT(t, fixture.app, fixture.primary, "cd", "GH-76")
	if !samePath(got, want) {
		t.Fatalf("cd shell path = %q, want filesystem-equivalent path %q", got, want)
	}

	got = ""
	runWT(t, fixture.app, fixture.primary, "enter", "GH-76")
	if !samePath(got, want) {
		t.Fatalf("enter shell path = %q, want filesystem-equivalent path %q", got, want)
	}
}

func TestHomeOpensShellInPrimaryWorktreeFromLinkedLane(t *testing.T) {
	fixture := newGitFixture(t)
	runGit(t, fixture.primary, "branch", "--track", "topic/navigate", "origin/main")
	runWT(t, fixture.app, fixture.primary, "add", "topic/navigate", "--no-link-env")
	lanePath := filepath.Join(fixture.worktreeRoot, "example", "project", "topic", "navigate")
	var got string
	fixture.app.runShell = func(_ context.Context, path string) error {
		got = path
		return nil
	}

	runWT(t, fixture.app, lanePath, "home")
	if !samePath(got, fixture.primary) {
		t.Fatalf("home shell path = %q, want filesystem-equivalent path %q", got, fixture.primary)
	}

	err := fixture.app.Run(context.Background(), lanePath, []string{"home", "extra"})
	if err == nil || err.Error() != "home accepts no arguments" {
		t.Fatalf("home usage error = %v", err)
	}
}

func TestCDRejectsUnregisteredLane(t *testing.T) {
	fixture := newGitFixture(t)
	fixture.app.runShell = func(_ context.Context, _ string) error {
		t.Fatal("shell should not start for an unregistered lane")
		return nil
	}
	err := fixture.app.Run(context.Background(), fixture.primary, []string{"cd", "GH-999"})
	if err == nil || !strings.Contains(err.Error(), "not an exact registered worktree") {
		t.Fatalf("unregistered lane error = %v", err)
	}
}

func TestBranchArgumentOpensExistingWorktree(t *testing.T) {
	fixture := newGitFixture(t)
	runGit(t, fixture.primary, "branch", "GH-93", "origin/main")
	runWT(t, fixture.app, fixture.primary, "add", "GH-93", "--no-link-env")
	want := filepath.Join(fixture.worktreeRoot, "example", "project", "GH-93")
	var got string
	fixture.app.runShell = func(_ context.Context, path string) error {
		got = path
		return nil
	}

	runWT(t, fixture.app, fixture.primary, "GH-93")
	if !samePath(got, want) {
		t.Fatalf("branch navigation path = %q, want %q", got, want)
	}
}

func TestBranchArgumentPromptsBeforeCreatingMissingWorktree(t *testing.T) {
	fixture := newGitFixture(t)
	fixture.app.stdin = bytes.NewBufferString("y\n")
	var got string
	fixture.app.runShell = func(_ context.Context, path string) error {
		got = path
		return nil
	}

	runWT(t, fixture.app, fixture.primary, "GH-93")
	want := filepath.Join(fixture.worktreeRoot, "example", "project", "GH-93")
	if !samePath(got, want) {
		t.Fatalf("created branch navigation path = %q, want %q", got, want)
	}
	assertBranch(t, want, "GH-93")
	if !strings.Contains(fixture.out.String(), "do you want to create this worktree? (y/n)") {
		t.Fatalf("creation prompt missing from output: %q", fixture.out.String())
	}
}

func TestBranchArgumentDeclinesMissingWorktree(t *testing.T) {
	fixture := newGitFixture(t)
	fixture.app.stdin = bytes.NewBufferString("n\n")
	fixture.app.runShell = func(_ context.Context, _ string) error {
		t.Fatal("shell should not start when creation is declined")
		return nil
	}

	runWT(t, fixture.app, fixture.primary, "GH-93")
	want := filepath.Join(fixture.worktreeRoot, "example", "project", "GH-93")
	if _, err := os.Stat(want); !os.IsNotExist(err) {
		t.Fatalf("declined worktree exists or stat failed: %v", err)
	}
}
