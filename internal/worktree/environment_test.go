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
