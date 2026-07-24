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
			if _, err := os.Lstat(destination); !os.IsNotExist(err) {
				t.Fatalf("opt-out destination exists or lstat failed: %v", err)
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
	assertBranch(t, filepath.Dir(destination), "GH-103")
}

func TestDetachedPRDoesNotLinkEnvironment(t *testing.T) {
	fixture := newGitFixture(t)
	writeEnvironmentSource(t, fixture, "TOKEN=source\n")
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
	if _, err := os.Lstat(destination); !os.IsNotExist(err) {
		t.Fatalf("detached PR environment exists or lstat failed: %v", err)
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
}

func TestExistingLaneReuseEnsuresEnvironmentLink(t *testing.T) {
	fixture := newGitFixture(t)
	source := writeEnvironmentSource(t, fixture, "TOKEN=reuse\n")
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
	if !strings.Contains(fixture.out.String(), "Reusing") {
		t.Fatalf("reuse output:\n%s", fixture.out.String())
	}
}

func TestHelpDocumentsWritableEnvironmentOptOut(t *testing.T) {
	fixture := newGitFixture(t)
	runWT(t, fixture.app, fixture.primary, "help")

	for _, want := range []string{
		"issue <number> [--no-link-env]",
		"add <branch> [--no-link-env]",
		"repair <number> [--no-link-env]",
		".envrc is never linked automatically",
		"No command starts applications or manages databases, ports, or runtime services",
	} {
		if !strings.Contains(fixture.out.String(), want) {
			t.Fatalf("help does not contain %q:\n%s", want, fixture.out.String())
		}
	}
	for _, command := range [][]string{
		{"issue", "106", "--unknown"},
		{"add", "topic/example", "--unknown"},
		{"repair", "81", "--unknown"},
	} {
		err := fixture.app.Run(context.Background(), fixture.primary, command)
		if err == nil || !strings.Contains(err.Error(), "[--no-link-env]") {
			t.Fatalf("invalid writable command %v error = %v", command, err)
		}
	}
}

func writeEnvironmentSource(t *testing.T, fixture gitFixture, contents string) string {
	t.Helper()
	path := filepath.Join(fixture.primary, environmentFileName)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertEnvironmentSymlink(t *testing.T, path, expectedSource string) {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s mode = %s, want symlink", path, info.Mode())
	}
	matches, _, err := environmentSymlinkMatches(path, expectedSource)
	if err != nil {
		t.Fatal(err)
	}
	if !matches {
		target, _ := os.Readlink(path)
		t.Fatalf("%s target = %q, want %q", path, target, expectedSource)
	}
}
