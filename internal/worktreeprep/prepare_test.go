package worktreeprep

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareBranchCreatesAndReusesCanonicalWorktree(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.createLocalBranch(t, "GH-139")
	for _, name := range environmentFileNames {
		if err := os.WriteFile(filepath.Join(fixture.primary, name), []byte(name+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	prepared, err := fixture.preparer.PrepareBranch(
		context.Background(),
		fixture.primary,
		"GH-139",
		true,
	)
	if err != nil {
		t.Fatalf("PrepareBranch() error = %v", err)
	}
	wantPath := filepath.Join(fixture.home, "worktrees", "owner", "project", "GH-139")
	if prepared.Path != wantPath || prepared.Branch != "GH-139" || !prepared.Created {
		t.Fatalf("prepared = %#v, want created %s", prepared, wantPath)
	}
	for _, name := range environmentFileNames {
		assertSymlink(t, filepath.Join(wantPath, name), filepath.Join(fixture.primary, name))
	}
	reused, err := fixture.preparer.PrepareBranch(
		context.Background(),
		fixture.primary,
		"GH-139",
		true,
	)
	if err != nil {
		t.Fatalf("PrepareBranch(reuse) error = %v", err)
	}
	if !samePath(reused.Path, wantPath) || reused.Branch != "GH-139" || reused.Created {
		t.Fatalf("reused = %#v", reused)
	}
}

func TestPrepareBranchTracksRemoteBranch(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.createRemoteBranch(t, "GH-140")
	prepared, err := fixture.preparer.PrepareBranch(
		context.Background(),
		fixture.primary,
		"GH-140",
		false,
	)
	if err != nil {
		t.Fatalf("PrepareBranch() error = %v", err)
	}
	if !prepared.Created || gitCommand(t, prepared.Path, "branch", "--show-current") != "GH-140" {
		t.Fatalf("prepared = %#v", prepared)
	}
	if upstream := gitCommand(t, prepared.Path, "rev-parse", "--abbrev-ref", "@{upstream}"); upstream != "origin/GH-140" {
		t.Fatalf("upstream = %q", upstream)
	}
}

func TestPrepareBranchRejectsProtectedPrimaryWorktree(t *testing.T) {
	fixture := newRepositoryFixture(t)
	_, err := fixture.preparer.PrepareBranch(
		context.Background(),
		fixture.primary,
		"main",
		true,
	)
	if err == nil || !strings.Contains(err.Error(), "protected primary worktree") {
		t.Fatalf("PrepareBranch(main) error = %v", err)
	}
}

func TestPrepareBranchRejectsMissingAndUnsafeBranches(t *testing.T) {
	fixture := newRepositoryFixture(t)
	for _, branch := range []string{"GH-404", "../escape", "/absolute"} {
		_, err := fixture.preparer.PrepareBranch(
			context.Background(),
			fixture.primary,
			branch,
			false,
		)
		if err == nil {
			t.Fatalf("PrepareBranch(%q) unexpectedly succeeded", branch)
		}
		if strings.Contains(err.Error(), "git wt") {
			t.Fatalf("error retains removed command guidance: %v", err)
		}
	}
}

func TestPreparePullRequestUsesOpenSameRepositoryHead(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.createRemoteBranch(t, "GH-141")
	fixture.preparer.resolvePullRequest = func(
		_ context.Context,
		_ string,
		repository string,
		number int,
	) (pullRequest, error) {
		if repository != fixture.repository || number != 141 {
			t.Fatalf("pull request target = %s#%d", repository, number)
		}
		return pullRequest{
			HeadRefName: "GH-141",
			HeadRefOID:  gitCommand(t, fixture.primary, "rev-parse", "origin/GH-141"),
			State:       "OPEN",
			URL:         "https://github.com/owner/project/pull/141",
		}, nil
	}
	prepared, err := fixture.preparer.PreparePullRequest(
		context.Background(),
		fixture.primary,
		141,
		false,
	)
	if err != nil {
		t.Fatalf("PreparePullRequest() error = %v", err)
	}
	if prepared.Repository != fixture.repository || prepared.Number != 141 ||
		prepared.Branch != "GH-141" || !prepared.Created || prepared.HeadRefOID == "" {
		t.Fatalf("prepared = %#v", prepared)
	}
}

func TestPreparePullRequestFailsClosed(t *testing.T) {
	fixture := newRepositoryFixture(t)
	for _, test := range []struct {
		name string
		pr   pullRequest
		want string
	}{
		{name: "fork", pr: pullRequest{IsCrossRepository: true}, want: "from a fork"},
		{name: "closed", pr: pullRequest{State: "CLOSED"}, want: "is closed, not open"},
		{name: "missing branch", pr: pullRequest{State: "OPEN"}, want: "has no head branch"},
		{name: "detached lane", pr: pullRequest{State: "OPEN", HeadRefName: "PR-9"}, want: "is not a durable branch"},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture.preparer.resolvePullRequest = func(context.Context, string, string, int) (pullRequest, error) {
				return test.pr, nil
			}
			_, err := fixture.preparer.PreparePullRequest(context.Background(), fixture.primary, 9, false)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestPrepareBranchRollsBackNewWorktreeAfterEnvironmentCollision(t *testing.T) {
	fixture := newRepositoryFixture(t)
	gitCommand(t, fixture.primary, "checkout", "-b", "GH-142")
	if err := os.WriteFile(filepath.Join(fixture.primary, ".env"), []byte("tracked\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitCommand(t, fixture.primary, "add", "-f", ".env")
	gitCommand(t, fixture.primary, "commit", "-m", "add environment collision")
	gitCommand(t, fixture.primary, "checkout", "main")
	destination := filepath.Join(fixture.home, "worktrees", "owner", "project", "GH-142")
	_, err := fixture.preparer.PrepareBranch(context.Background(), fixture.primary, "GH-142", true)
	if err == nil || !strings.Contains(err.Error(), "destination environment file already exists") {
		t.Fatalf("PrepareBranch() error = %v", err)
	}
	if _, statErr := os.Stat(destination); !os.IsNotExist(statErr) {
		t.Fatalf("rolled-back destination still exists: %v", statErr)
	}
	if strings.Contains(gitCommand(t, fixture.primary, "worktree", "list", "--porcelain"), destination) {
		t.Fatalf("rolled-back destination remains registered")
	}
}

func TestPrepareBranchPreservesRegularEnvironmentRC(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.createLocalBranch(t, "GH-143")
	prepared, err := fixture.preparer.PrepareBranch(context.Background(), fixture.primary, "GH-143", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture.primary, ".env"), []byte("shared\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	localRC := filepath.Join(prepared.Path, ".envrc")
	if err := os.WriteFile(localRC, []byte("local\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.preparer.PrepareBranch(context.Background(), fixture.primary, "GH-143", true); err != nil {
		t.Fatalf("PrepareBranch(reuse) error = %v", err)
	}
	assertSymlink(t, filepath.Join(prepared.Path, ".env"), filepath.Join(fixture.primary, ".env"))
	content, err := os.ReadFile(localRC)
	if err != nil || string(content) != "local\n" {
		t.Fatalf("preserved .envrc = %q, %v", content, err)
	}
}

func TestPrepareBranchRejectsMatchingLinkToNonRegularSource(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.createLocalBranch(t, "GH-144")
	prepared, err := fixture.preparer.PrepareBranch(context.Background(), fixture.primary, "GH-144", false)
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(fixture.primary, ".env")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(source, filepath.Join(prepared.Path, ".env")); err != nil {
		t.Fatal(err)
	}
	_, err = fixture.preparer.PrepareBranch(context.Background(), fixture.primary, "GH-144", true)
	if err == nil || !strings.Contains(err.Error(), "source environment file must be a regular file") {
		t.Fatalf("PrepareBranch(reuse) error = %v", err)
	}
}
