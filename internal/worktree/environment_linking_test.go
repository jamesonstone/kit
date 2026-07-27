package worktree

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWritableLaneCreatedFromLinkedLaneUsesPrimaryEnvironment(t *testing.T) {
	fixture := newGitFixture(t)
	source := writeEnvironmentSource(t, fixture, "TOKEN=primary\n")
	rcSource := writeEnvironmentRCSource(t, fixture, "dotenv\n")
	runGit(t, fixture.primary, "branch", "--track", "topic/source-lane", "origin/main")
	runGit(t, fixture.primary, "branch", "--track", "topic/target-lane", "origin/main")
	runWT(t, fixture.app, fixture.primary, "add", "topic/source-lane")

	sourceLane := filepath.Join(
		fixture.worktreeRoot,
		"example",
		"project",
		"topic",
		"source-lane",
	)
	runWT(t, fixture.app, sourceLane, "add", "topic/target-lane")
	targetEnvironment := filepath.Join(
		fixture.worktreeRoot,
		"example",
		"project",
		"topic",
		"target-lane",
		environmentFileName,
	)
	targetEnvironmentRC := filepath.Join(filepath.Dir(targetEnvironment), environmentRCFileName)

	target, err := os.Readlink(targetEnvironment)
	if err != nil {
		t.Fatal(err)
	}
	expectedSource, err := resolvedPath(source)
	if err != nil {
		t.Fatal(err)
	}
	if target != expectedSource {
		t.Fatalf("target environment link = %q, want primary source %q", target, expectedSource)
	}
	assertEnvironmentSymlink(t, targetEnvironmentRC, rcSource)

	runWT(t, fixture.app, fixture.primary, "remove", "topic/source-lane")
	data, err := os.ReadFile(targetEnvironment)
	if err != nil {
		t.Fatalf("target environment broke after removing invoking lane: %v", err)
	}
	if string(data) != "TOKEN=primary\n" {
		t.Fatalf("target environment contents = %q", data)
	}
	if data, err := os.ReadFile(targetEnvironmentRC); err != nil || string(data) != "dotenv\n" {
		t.Fatalf("target environment configuration broke: data=%q err=%v", data, err)
	}
}

func TestHelpDocumentsWritableEnvironmentOptOut(t *testing.T) {
	fixture := newGitFixture(t)
	runWT(t, fixture.app, fixture.primary, "help")

	for _, want := range []string{
		"issue <number> [--no-link-env]",
		"add <branch> [--no-link-env]",
		"repair <number> [--no-link-env]",
		"list [flags]",
		"--plain",
		"path <lane>",
		"primary checkout's .env and .envrc",
		"omit both links",
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

func writeEnvironmentRCSource(t *testing.T, fixture gitFixture, contents string) string {
	t.Helper()
	path := filepath.Join(fixture.primary, environmentRCFileName)
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
