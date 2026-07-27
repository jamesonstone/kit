package worktree

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
