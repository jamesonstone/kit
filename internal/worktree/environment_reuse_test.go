package worktree

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigratePreservesExistingEnvironmentSymlink(t *testing.T) {
	fixture := newGitFixture(t)
	source := writeEnvironmentSource(t, fixture, "TOKEN=preserve-link\n")
	legacy := filepath.Join(fixture.worktreeRoot, "project-topic-linked")
	runGit(t, fixture.primary, "branch", "topic/linked", "origin/main")
	runGit(t, fixture.primary, "worktree", "add", legacy, "topic/linked")
	if err := os.Symlink(source, filepath.Join(legacy, environmentFileName)); err != nil {
		t.Fatal(err)
	}

	runWT(t, fixture.app, fixture.primary, "migrate", "--apply")
	destination := filepath.Join(fixture.worktreeRoot, "example", "project", "topic", "linked", environmentFileName)
	assertEnvironmentSymlink(t, destination, source)
}

func TestExistingDestinationEnvironmentCollisionIsPreserved(t *testing.T) {
	fixture := newGitFixture(t)
	writeEnvironmentSource(t, fixture, "TOKEN=source\n")
	runWT(t, fixture.app, fixture.primary, "issue", "104", "--no-link-env")
	destination := filepath.Join(fixture.worktreeRoot, "example", "project", "GH-104", environmentFileName)
	if err := os.WriteFile(destination, []byte("TOKEN=local\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := fixture.app.Run(context.Background(), fixture.primary, []string{"issue", "104"})
	if err == nil || !strings.Contains(err.Error(), "already exists and is not a symlink") {
		t.Fatalf("environment collision error = %v", err)
	}
	data, readErr := os.ReadFile(destination)
	if readErr != nil || string(data) != "TOKEN=local\n" {
		t.Fatalf("destination collision was modified: data=%q err=%v", data, readErr)
	}
	assertBranch(t, filepath.Dir(destination), "GH-104")
}

func TestExistingLaneReuseEnsuresEnvironmentLink(t *testing.T) {
	fixture := newGitFixture(t)
	source := writeEnvironmentSource(t, fixture, "TOKEN=reuse\n")
	rcSource := writeEnvironmentRCSource(t, fixture, "dotenv\n")
	runWT(t, fixture.app, fixture.primary, "issue", "105", "--no-link-env")

	fixture.out.Reset()
	runWT(t, fixture.app, fixture.primary, "issue", "105")
	destination := filepath.Join(fixture.worktreeRoot, "example", "project", "GH-105", environmentFileName)
	assertEnvironmentSymlink(t, destination, source)
	assertEnvironmentSymlink(t, filepath.Join(filepath.Dir(destination), environmentRCFileName), rcSource)
	if !strings.Contains(fixture.out.String(), "Reusing") {
		t.Fatalf("reuse output:\n%s", fixture.out.String())
	}
}

func TestExistingEnvironmentRCFileIsPreserved(t *testing.T) {
	fixture := newGitFixture(t)
	writeEnvironmentRCSource(t, fixture, "dotenv\n")
	runGit(t, fixture.primary, "add", "-f", environmentRCFileName)
	runGit(t, fixture.primary, "commit", "-m", "track environment configuration")
	runGit(t, fixture.primary, "push", "origin", "main")

	runWT(t, fixture.app, fixture.primary, "issue", "106")
	destination := filepath.Join(fixture.worktreeRoot, "example", "project", "GH-106", environmentRCFileName)
	info, err := os.Lstat(destination)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("tracked %s was replaced with a symlink", destination)
	}
	data, err := os.ReadFile(destination)
	if err != nil || string(data) != "dotenv\n" {
		t.Fatalf("tracked environment configuration changed: data=%q err=%v", data, err)
	}
}

func TestExistingEnvironmentRCSymlinkCollisionIsPreserved(t *testing.T) {
	fixture := newGitFixture(t)
	writeEnvironmentSource(t, fixture, "TOKEN=source\n")
	writeEnvironmentRCSource(t, fixture, "dotenv\n")
	runWT(t, fixture.app, fixture.primary, "issue", "107", "--no-link-env")
	destination := filepath.Join(fixture.worktreeRoot, "example", "project", "GH-107", environmentRCFileName)
	unexpected := filepath.Join(fixture.primary, ".other-envrc")
	if err := os.WriteFile(unexpected, []byte("export OTHER=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(unexpected, destination); err != nil {
		t.Fatal(err)
	}

	err := fixture.app.Run(context.Background(), fixture.primary, []string{"issue", "107"})
	if err == nil || !strings.Contains(err.Error(), "points somewhere unexpected") {
		t.Fatalf("environment configuration collision error = %v", err)
	}
	target, readErr := os.Readlink(destination)
	if readErr != nil || target != unexpected {
		t.Fatalf("destination environment configuration link was modified: target=%q err=%v", target, readErr)
	}
	environmentDestination := filepath.Join(filepath.Dir(destination), environmentFileName)
	if _, statErr := os.Lstat(environmentDestination); !os.IsNotExist(statErr) {
		t.Fatalf("transaction left a partial environment link: %v", statErr)
	}
	assertBranch(t, filepath.Dir(destination), "GH-107")
}

func TestBrokenEnvironmentRCSymlinkIsPreserved(t *testing.T) {
	fixture := newGitFixture(t)
	runWT(t, fixture.app, fixture.primary, "issue", "108", "--no-link-env")
	destination := filepath.Join(fixture.worktreeRoot, "example", "project", "GH-108", environmentRCFileName)
	source := filepath.Join(fixture.primary, environmentRCFileName)
	if err := os.Symlink(source, destination); err != nil {
		t.Fatal(err)
	}

	err := fixture.app.Run(context.Background(), fixture.primary, []string{"issue", "108"})
	if err == nil || !strings.Contains(err.Error(), "symlink is broken") {
		t.Fatalf("broken environment configuration error = %v", err)
	}
	target, readErr := os.Readlink(destination)
	if readErr != nil || target != source {
		t.Fatalf("broken environment configuration link was modified: target=%q err=%v", target, readErr)
	}
}
