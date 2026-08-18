package worktreeprep

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectIdentifiesPrimaryAndLinkedWorktrees(t *testing.T) {
	fixture := newRepositoryFixture(t)

	primary, err := fixture.preparer.Inspect(context.Background(), fixture.primary)
	if err != nil {
		t.Fatalf("Inspect(primary) error = %v", err)
	}
	if !primary.InsideGit || !primary.IsPrimary || !samePath(primary.Path, fixture.primary) {
		t.Fatalf("primary location = %#v", primary)
	}

	fixture.createLocalBranch(t, "GH-160")
	prepared, err := fixture.preparer.PrepareBranch(context.Background(), fixture.primary, "GH-160", false)
	if err != nil {
		t.Fatalf("PrepareBranch() error = %v", err)
	}
	linked, err := fixture.preparer.Inspect(context.Background(), prepared.Path)
	if err != nil {
		t.Fatalf("Inspect(linked) error = %v", err)
	}
	if !linked.InsideGit || linked.IsPrimary || !samePath(linked.PrimaryPath, fixture.primary) {
		t.Fatalf("linked location = %#v", linked)
	}
}

func TestInspectPreservesNonGitProjects(t *testing.T) {
	location, err := New().Inspect(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("Inspect(non-Git) error = %v", err)
	}
	if location.InsideGit || location.IsPrimary {
		t.Fatalf("non-Git location = %#v", location)
	}
}

func TestInspectFailsClosedForBrokenGitMetadata(t *testing.T) {
	projectRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectRoot, ".git"), []byte("gitdir: missing\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := New().Inspect(context.Background(), projectRoot)
	if err == nil || !strings.Contains(err.Error(), "resolve current Git worktree") {
		t.Fatalf("Inspect(broken Git metadata) error = %v", err)
	}
}

func TestInspectFailsClosedForBrokenAncestorGitMetadata(t *testing.T) {
	repositoryRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(repositoryRoot, ".git"), []byte("gitdir: missing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	projectRoot := filepath.Join(repositoryRoot, "nested", "project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := New().Inspect(context.Background(), projectRoot)
	if err == nil || !strings.Contains(err.Error(), "resolve current Git worktree") {
		t.Fatalf("Inspect(nested broken Git metadata) error = %v", err)
	}
}
