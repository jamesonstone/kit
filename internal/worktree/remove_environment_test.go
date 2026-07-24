package worktree

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemoveAllowsOnlyMatchingManagedEnvironmentLink(t *testing.T) {
	t.Run("matching link", func(t *testing.T) {
		fixture := newGitFixture(t)
		source := writeEnvironmentSource(t, fixture, "TOKEN=preserve\n")
		if err := os.WriteFile(
			filepath.Join(fixture.primary, ".gitignore"),
			[]byte(environmentFileName+"\n"),
			0o644,
		); err != nil {
			t.Fatal(err)
		}
		runGit(t, fixture.primary, "add", ".gitignore")
		runGit(t, fixture.primary, "commit", "-m", "ignore environment")
		runGit(t, fixture.primary, "push", "origin", "main")
		runGit(t, fixture.primary, "branch", "--track", "topic/remove-env", "origin/main")
		runWT(t, fixture.app, fixture.primary, "add", "topic/remove-env")
		worktreePath := filepath.Join(
			fixture.worktreeRoot,
			"example",
			"project",
			"topic",
			"remove-env",
		)

		runWT(t, fixture.app, fixture.primary, "remove", "topic/remove-env")
		if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
			t.Fatalf("removed worktree still exists or stat failed: %v", err)
		}
		data, err := os.ReadFile(source)
		if err != nil || string(data) != "TOKEN=preserve\n" {
			t.Fatalf("source environment was not preserved: data=%q err=%v", data, err)
		}
	})

	t.Run("regular file", func(t *testing.T) {
		fixture := newGitFixture(t)
		writeEnvironmentSource(t, fixture, "TOKEN=source\n")
		runGit(t, fixture.primary, "branch", "--track", "topic/regular-env", "origin/main")
		runWT(t, fixture.app, fixture.primary, "add", "topic/regular-env", "--no-link-env")
		worktreePath := filepath.Join(
			fixture.worktreeRoot,
			"example",
			"project",
			"topic",
			"regular-env",
		)
		destination := filepath.Join(worktreePath, environmentFileName)
		if err := os.WriteFile(destination, []byte("TOKEN=local\n"), 0o600); err != nil {
			t.Fatal(err)
		}

		err := fixture.app.Run(
			context.Background(),
			fixture.primary,
			[]string{"remove", "topic/regular-env"},
		)
		if err == nil || !strings.Contains(err.Error(), "not a GitWT-managed environment symlink") {
			t.Fatalf("regular environment removal error = %v", err)
		}
		if data, readErr := os.ReadFile(destination); readErr != nil ||
			string(data) != "TOKEN=local\n" {
			t.Fatalf("regular environment file was modified: data=%q err=%v", data, readErr)
		}
		assertBranch(t, worktreePath, "topic/regular-env")
	})

	t.Run("unexpected symlink", func(t *testing.T) {
		fixture := newGitFixture(t)
		writeEnvironmentSource(t, fixture, "TOKEN=source\n")
		unexpected := filepath.Join(fixture.primary, ".other-env")
		if err := os.WriteFile(unexpected, []byte("TOKEN=other\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		runGit(t, fixture.primary, "branch", "--track", "topic/unexpected-env", "origin/main")
		runWT(t, fixture.app, fixture.primary, "add", "topic/unexpected-env", "--no-link-env")
		worktreePath := filepath.Join(
			fixture.worktreeRoot,
			"example",
			"project",
			"topic",
			"unexpected-env",
		)
		destination := filepath.Join(worktreePath, environmentFileName)
		if err := os.Symlink(unexpected, destination); err != nil {
			t.Fatal(err)
		}

		err := fixture.app.Run(
			context.Background(),
			fixture.primary,
			[]string{"remove", "topic/unexpected-env"},
		)
		if err == nil || !strings.Contains(err.Error(), "points somewhere other than") {
			t.Fatalf("unexpected environment symlink removal error = %v", err)
		}
		target, readErr := os.Readlink(destination)
		if readErr != nil || target != unexpected {
			t.Fatalf(
				"unexpected environment symlink was modified: target=%q err=%v",
				target,
				readErr,
			)
		}
		assertBranch(t, worktreePath, "topic/unexpected-env")
	})
}

func TestRemoveRestoresManagedEnvironmentLinkWhenGitRemovalFails(t *testing.T) {
	fixture := newGitFixture(t)
	source := writeEnvironmentSource(t, fixture, "TOKEN=restore\n")
	runGit(t, fixture.primary, "branch", "--track", "topic/restore-env", "origin/main")
	runWT(t, fixture.app, fixture.primary, "add", "topic/restore-env")
	destination := filepath.Join(
		fixture.worktreeRoot,
		"example",
		"project",
		"topic",
		"restore-env",
		environmentFileName,
	)

	run := fixture.app.run
	fixture.app.run = func(
		ctx context.Context,
		cwd string,
		name string,
		args ...string,
	) ([]byte, error) {
		if name == "git" && len(args) >= 2 && args[0] == "worktree" && args[1] == "remove" {
			return []byte("simulated removal failure"), fmt.Errorf("simulated failure")
		}
		return run(ctx, cwd, name, args...)
	}

	err := fixture.app.Run(
		context.Background(),
		fixture.primary,
		[]string{"remove", "topic/restore-env"},
	)
	if err == nil || !strings.Contains(err.Error(), "restored environment symlink") {
		t.Fatalf("failed removal error = %v", err)
	}
	assertEnvironmentSymlink(t, destination, source)
}
