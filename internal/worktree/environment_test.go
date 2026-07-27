package worktree

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWritableCommandsLinkEnvironmentByDefault(t *testing.T) {
	tests := []struct {
		name        string
		prepare     func(*testing.T, gitFixture)
		command     []string
		destination string
	}{
		{
			name:        "issue",
			command:     []string{"issue", "101"},
			destination: "GH-101",
		},
		{
			name: "add",
			prepare: func(t *testing.T, fixture gitFixture) {
				runGit(t, fixture.primary, "branch", "--track", "topic/env", "origin/main")
			},
			command:     []string{"add", "topic/env"},
			destination: filepath.Join("topic", "env"),
		},
		{
			name: "repair",
			prepare: func(t *testing.T, fixture gitFixture) {
				commitOnRemoteBranch(t, fixture, "repair-env")
				fixture.app.resolvePR = func(context.Context, string, string, int) (PR, error) {
					return PR{HeadRefName: "repair-env", State: "OPEN"}, nil
				}
			},
			command:     []string{"repair", "79"},
			destination: "repair-env",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newGitFixture(t)
			source := writeEnvironmentSource(t, fixture, "TOKEN=original\n")
			rcSource := writeEnvironmentRCSource(t, fixture, "dotenv\n")
			if test.prepare != nil {
				test.prepare(t, fixture)
			}

			runWT(t, fixture.app, fixture.primary, test.command...)
			destination := filepath.Join(
				fixture.worktreeRoot,
				"example",
				"project",
				test.destination,
				environmentFileName,
			)
			assertEnvironmentSymlink(t, destination, source)
			assertEnvironmentSymlink(
				t,
				filepath.Join(filepath.Dir(destination), environmentRCFileName),
				rcSource,
			)

			if err := os.WriteFile(source, []byte("TOKEN=updated\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(destination)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != "TOKEN=updated\n" {
				t.Fatalf("environment link copied stale contents: %q", data)
			}
		})
	}
}

func TestWritableCommandsCanDisableEnvironmentLinking(t *testing.T) {
	tests := []struct {
		name        string
		prepare     func(*testing.T, gitFixture)
		command     []string
		destination string
	}{
		{
			name:        "issue",
			command:     []string{"issue", "102", "--no-link-env"},
			destination: "GH-102",
		},
		{
			name: "add",
			prepare: func(t *testing.T, fixture gitFixture) {
				runGit(t, fixture.primary, "branch", "--track", "topic/isolated", "origin/main")
			},
			command:     []string{"add", "topic/isolated", "--no-link-env"},
			destination: filepath.Join("topic", "isolated"),
		},
		{
			name: "repair",
			prepare: func(t *testing.T, fixture gitFixture) {
				commitOnRemoteBranch(t, fixture, "repair-isolated")
				fixture.app.resolvePR = func(context.Context, string, string, int) (PR, error) {
					return PR{HeadRefName: "repair-isolated", State: "OPEN"}, nil
				}
			},
			command:     []string{"repair", "80", "--no-link-env"},
			destination: "repair-isolated",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newGitFixture(t)
			writeEnvironmentSource(t, fixture, "TOKEN=isolated\n")
			writeEnvironmentRCSource(t, fixture, "dotenv\n")
			if test.prepare != nil {
				test.prepare(t, fixture)
			}

			runWT(t, fixture.app, fixture.primary, test.command...)
			destination := filepath.Join(
				fixture.worktreeRoot,
				"example",
				"project",
				test.destination,
				environmentFileName,
			)
			for _, name := range environmentFileNames {
				path := filepath.Join(filepath.Dir(destination), name)
				if _, err := os.Lstat(path); !os.IsNotExist(err) {
					t.Fatalf("opt-out destination %s exists or lstat failed: %v", path, err)
				}
			}
		})
	}
}

func TestMissingEnvironmentSourceDoesNotBlockIssueLane(t *testing.T) {
	fixture := newGitFixture(t)

	runWT(t, fixture.app, fixture.primary, "issue", "103")
	destination := filepath.Join(
		fixture.worktreeRoot,
		"example",
		"project",
		"GH-103",
		environmentFileName,
	)
	if _, err := os.Lstat(destination); !os.IsNotExist(err) {
		t.Fatalf("destination environment exists or lstat failed: %v", err)
	}
	if !strings.Contains(fixture.out.String(), "no .env link was created") {
		t.Fatalf("missing-source output:\n%s", fixture.out.String())
	}
	if !strings.Contains(fixture.out.String(), "no .envrc link was created") {
		t.Fatalf("missing-source output:\n%s", fixture.out.String())
	}
	assertBranch(t, filepath.Dir(destination), "GH-103")
}

func TestDetachedPRDoesNotLinkEnvironment(t *testing.T) {
	fixture := newGitFixture(t)
	writeEnvironmentSource(t, fixture, "TOKEN=source\n")
	writeEnvironmentRCSource(t, fixture, "dotenv\n")
	prCommit := commitOnRemoteBranch(t, fixture, "detached-env")
	runGit(t, fixture.remote, "update-ref", "refs/pull/82/head", prCommit)

	runWT(t, fixture.app, fixture.primary, "pr", "82")
	destination := filepath.Join(
		fixture.worktreeRoot,
		"example",
		"project",
		"PR-82",
		environmentFileName,
	)
	for _, name := range environmentFileNames {
		path := filepath.Join(filepath.Dir(destination), name)
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("detached PR environment %s exists or lstat failed: %v", path, err)
		}
	}
}

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
	destination := filepath.Join(
		fixture.worktreeRoot,
		"example",
		"project",
		"topic",
		"linked",
		environmentFileName,
	)
	assertEnvironmentSymlink(t, destination, source)
}

func TestExistingDestinationEnvironmentCollisionIsPreserved(t *testing.T) {
	fixture := newGitFixture(t)
	writeEnvironmentSource(t, fixture, "TOKEN=source\n")
	runWT(t, fixture.app, fixture.primary, "issue", "104", "--no-link-env")
	destination := filepath.Join(
		fixture.worktreeRoot,
		"example",
		"project",
		"GH-104",
		environmentFileName,
	)
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
	destination := filepath.Join(
		fixture.worktreeRoot,
		"example",
		"project",
		"GH-105",
		environmentFileName,
	)
	assertEnvironmentSymlink(t, destination, source)
	assertEnvironmentSymlink(
		t,
		filepath.Join(filepath.Dir(destination), environmentRCFileName),
		rcSource,
	)
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
	destination := filepath.Join(
		fixture.worktreeRoot,
		"example",
		"project",
		"GH-106",
		environmentRCFileName,
	)
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
	destination := filepath.Join(
		fixture.worktreeRoot,
		"example",
		"project",
		"GH-107",
		environmentRCFileName,
	)
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
		t.Fatalf(
			"destination environment configuration link was modified: target=%q err=%v",
			target,
			readErr,
		)
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
	destination := filepath.Join(
		fixture.worktreeRoot,
		"example",
		"project",
		"GH-108",
		environmentRCFileName,
	)
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
		t.Fatalf(
			"broken environment configuration link was modified: target=%q err=%v",
			target,
			readErr,
		)
	}
}

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
		{
			name:        "issue",
			branch:      "GH-109",
			command:     []string{"issue", "109"},
			destination: "GH-109",
		},
		{
			name:   "add existing branch",
			branch: "topic/setup-failure",
			prepare: func(t *testing.T, fixture gitFixture) {
				runGit(t, fixture.primary, "branch", "topic/setup-failure", "origin/main")
			},
			command:     []string{"add", "topic/setup-failure"},
			destination: filepath.Join("topic", "setup-failure"),
		},
		{
			name:        "interactive branch creation",
			branch:      "GH-110",
			command:     []string{"GH-110"},
			destination: "GH-110",
			prepare: func(t *testing.T, fixture gitFixture) {
				fixture.app.stdin = strings.NewReader("y\n")
			},
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
			destination := filepath.Join(
				fixture.worktreeRoot,
				"example",
				"project",
				test.destination,
			)
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
