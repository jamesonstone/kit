package worktree

import (
	"context"
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
