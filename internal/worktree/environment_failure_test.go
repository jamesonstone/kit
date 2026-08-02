package worktree

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnvironmentLinkCreationRollsBackAfterApplyFailure(t *testing.T) {
	sourceRoot := t.TempDir()
	destinationRoot := t.TempDir()
	for _, name := range environmentFileNames {
		if err := os.WriteFile(filepath.Join(sourceRoot, name), []byte("source\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	environmentRCDestination := filepath.Join(destinationRoot, environmentRCFileName)
	writer := &environmentLinkRaceWriter{destination: environmentRCDestination}
	app := NewApp(writer, writer)
	err := app.ensureEnvironmentLinks(sourceRoot, destinationRoot, true)
	if err == nil || !strings.Contains(err.Error(), "link environment file") {
		t.Fatalf("environment apply failure = %v", err)
	}
	if _, statErr := os.Lstat(filepath.Join(destinationRoot, environmentFileName)); !os.IsNotExist(statErr) {
		t.Fatalf("transaction left a newly created environment link: %v", statErr)
	}
	data, readErr := os.ReadFile(environmentRCDestination)
	if readErr != nil || string(data) != "race\n" {
		t.Fatalf("concurrent destination was modified: data=%q err=%v", data, readErr)
	}
}

func TestEnvironmentLinkCreationRollsBackAfterOutputFailure(t *testing.T) {
	sourceRoot := t.TempDir()
	destinationRoot := t.TempDir()
	for _, name := range environmentFileNames {
		if err := os.WriteFile(filepath.Join(sourceRoot, name), []byte("source\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	app := NewApp(failingWriter{}, failingWriter{})
	err := app.ensureEnvironmentLinks(sourceRoot, destinationRoot, true)
	if err == nil || !strings.Contains(err.Error(), "write output") {
		t.Fatalf("environment output failure = %v", err)
	}
	for _, name := range environmentFileNames {
		if _, statErr := os.Lstat(filepath.Join(destinationRoot, name)); !os.IsNotExist(statErr) {
			t.Fatalf("output failure left environment material %s: %v", name, statErr)
		}
	}
}

func TestNewWorktreeSetupFailureRemovesOnlyFreshRegistration(t *testing.T) {
	tests := []struct {
		name        string
		branch      string
		prepare     func(*testing.T, gitFixture)
		command     []string
		destination string
	}{
		{name: "issue", branch: "GH-109", command: []string{"issue", "109"}, destination: "GH-109"},
		{
			name:   "add existing branch",
			branch: "topic/setup-failure",
			prepare: func(t *testing.T, fixture gitFixture) {
				runGit(t, fixture.primary, "branch", "topic/setup-failure", "origin/main")
			},
			command: []string{"add", "topic/setup-failure"}, destination: filepath.Join("topic", "setup-failure"),
		},
		{
			name: "interactive branch creation", branch: "GH-110", command: []string{"GH-110"}, destination: "GH-110",
			prepare: func(t *testing.T, fixture gitFixture) { fixture.app.stdin = strings.NewReader("y\n") },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newGitFixture(t)
			writeEnvironmentSource(t, fixture, "TOKEN=tracked\n")
			runGit(t, fixture.primary, "add", "-f", environmentFileName)
			runGit(t, fixture.primary, "commit", "-m", "track environment")
			runGit(t, fixture.primary, "push", "origin", "main")
			if test.prepare != nil {
				test.prepare(t, fixture)
			}

			err := fixture.app.Run(context.Background(), fixture.primary, test.command)
			if err == nil || !strings.Contains(err.Error(), "already exists and is not a symlink") {
				t.Fatalf("worktree setup failure = %v", err)
			}
			destination := filepath.Join(fixture.worktreeRoot, "example", "project", test.destination)
			if _, statErr := os.Lstat(destination); !os.IsNotExist(statErr) {
				t.Fatalf("failed worktree registration remains at %s: %v", destination, statErr)
			}
			if output := gitText(t, fixture.primary, "worktree", "list", "--porcelain"); strings.Contains(output, destination) {
				t.Fatalf("failed worktree remains registered:\n%s", output)
			}
			if output := gitText(t, fixture.primary, "show-ref", "--verify", "refs/heads/"+test.branch); output == "" {
				t.Fatalf("setup rollback deleted branch %s", test.branch)
			}
		})
	}
}

type environmentLinkRaceWriter struct {
	destination string
	wrote       bool
}

func (w *environmentLinkRaceWriter) Write(data []byte) (int, error) {
	if !w.wrote {
		w.wrote = true
		if err := os.WriteFile(w.destination, []byte("race\n"), 0o600); err != nil {
			return 0, err
		}
	}
	return len(data), nil
}
