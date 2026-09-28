package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/worktreeprep"
	"github.com/spf13/cobra"
)

func TestResolveReconcileTargetWritesInPlaceOutsidePrimaryCheckout(t *testing.T) {
	for _, test := range []struct {
		name     string
		location worktreeprep.Location
		dryRun   bool
	}{
		{name: "linked worktree", location: worktreeprep.Location{InsideGit: true}},
		{name: "non-Git project"},
		{name: "primary dry run", location: worktreeprep.Location{InsideGit: true, IsPrimary: true}, dryRun: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			stubReconcileWorktreeInspection(t, test.location)
			target, err := resolveReconcileTarget(&bytes.Buffer{}, "/repo", test.dryRun, false)
			if err != nil || target.worktree || target.projectRoot != "/repo" {
				t.Fatalf("target = %#v, err = %v", target, err)
			}
		})
	}
}

func TestResolveReconcileTargetFailsClosed(t *testing.T) {
	previous := inspectReconcileWorktree
	inspectReconcileWorktree = func(string) (worktreeprep.Location, error) {
		return worktreeprep.Location{}, errors.New("broken git metadata")
	}
	t.Cleanup(func() { inspectReconcileWorktree = previous })
	if _, err := resolveReconcileTarget(&bytes.Buffer{}, "/repo", false, false); err == nil {
		t.Fatal("expected inspection failure to stop reconcile")
	}
}

func TestRunReconcileFromPrimaryMigratesLinkedWorktreeAndLeavesPrimaryUntouched(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	writeFile(t, filepath.Join(projectRoot, "AGENTS.md"), "# AGENTS\n\n## Team Notes\n\n- Keep me.\n")
	writeFile(t, filepath.Join(projectRoot, envPath), "SECRET=local\n")
	writeFile(t, filepath.Join(projectRoot, ".gitignore"), ".env\n")
	initializeReconcileGitFixture(t, projectRoot)
	worktreeRoot := t.TempDir()
	stubReconcileWorktreeRoot(t, worktreeRoot)
	setWorkingDirectory(t, projectRoot)

	output := runManagedReconcileForWorktreeTest(t)
	if status := reconcileGitOutput(t, projectRoot, "status", "--porcelain"); status != "" {
		t.Fatalf("primary checkout changed:\n%s", status)
	}
	linked := filepath.Join(worktreeRoot, filepath.Base(projectRoot), reconcileBranch)
	agents := readFile(t, filepath.Join(linked, "AGENTS.md"))
	if strings.Count(agents, "BEGIN KIT-MANAGED CONTRACT") != 1 || !strings.Contains(agents, "- Keep me.") {
		t.Fatalf("linked AGENTS.md was not migrated with project guidance kept:\n%s", agents)
	}
	if target, err := os.Readlink(filepath.Join(linked, envPath)); err != nil || filepath.Base(target) != envPath || readFile(t, target) != "SECRET=local\n" {
		t.Fatalf(".env link = %q, %v", target, err)
	}
	if branch := reconcileGitOutput(t, linked, "branch", "--show-current"); branch != reconcileBranch {
		t.Fatalf("linked branch = %q", branch)
	}
	if !strings.Contains(output, linked) || !strings.Contains(output, "git -C") {
		t.Fatalf("output does not point at the worktree:\n%s", output)
	}

	// A second run reuses the worktree and changes nothing.
	output = runManagedReconcileForWorktreeTest(t)
	if !strings.Contains(output, "reusing worktree") {
		t.Fatalf("second run did not reuse the worktree:\n%s", output)
	}
	if status := reconcileGitOutput(t, projectRoot, "status", "--porcelain"); status != "" {
		t.Fatalf("primary checkout changed on rerun:\n%s", status)
	}
}

func TestReconcileWorktreeRefusesUnrelatedExistingPath(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	initializeReconcileGitFixture(t, projectRoot)
	worktreeRoot := t.TempDir()
	stubReconcileWorktreeRoot(t, worktreeRoot)
	writeFile(t, filepath.Join(worktreeRoot, filepath.Base(projectRoot), reconcileBranch, "keep.txt"), "mine\n")
	location, err := worktreeprep.New().Inspect(context.Background(), projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := prepareReconcileWorktree(projectRoot, location); err == nil || !strings.Contains(err.Error(), "move it aside") {
		t.Fatalf("expected refusal, got %v", err)
	}
}

func stubReconcileWorktreeRoot(t *testing.T, root string) {
	t.Helper()
	previous := reconcileWorktreeRoot
	reconcileWorktreeRoot = func() (string, error) { return root, nil }
	t.Cleanup(func() { reconcileWorktreeRoot = previous })
}

func initializeReconcileGitFixture(t *testing.T, projectRoot string) {
	t.Helper()
	runGitForSourceAuditTest(t, projectRoot, "init", "-b", "main")
	runGitForSourceAuditTest(t, projectRoot, "config", "user.name", "Test User")
	runGitForSourceAuditTest(t, projectRoot, "config", "user.email", "test@example.com")
	runGitForSourceAuditTest(t, projectRoot, "add", "--all")
	runGitForSourceAuditTest(t, projectRoot, "commit", "-m", "fixture")
}

func runManagedReconcileForWorktreeTest(t *testing.T) string {
	t.Helper()
	resetReconcileFlags(t)
	reconcileOutputOnly = true

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().Bool("output-only", true, "")
	cmd.SetContext(context.Background())
	cmd.SetOut(&out)
	stdout := captureStdout(t, func() {
		if err := runReconcile(cmd, nil); err != nil {
			t.Fatalf("runReconcile() error = %v", err)
		}
	})
	return out.String() + stdout
}

func reconcileGitOutput(t *testing.T, projectRoot string, args ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", projectRoot}, args...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v error = %v\n%s", args, err, output)
	}
	return strings.TrimSpace(string(output))
}

func stubReconcileWorktreeInspection(t *testing.T, location worktreeprep.Location) {
	t.Helper()
	previous := inspectReconcileWorktree
	inspectReconcileWorktree = func(string) (worktreeprep.Location, error) {
		return location, nil
	}
	t.Cleanup(func() { inspectReconcileWorktree = previous })
}

// runManagedReconcileForWorktreeTestKeepingFlags runs reconcile with flags the
// caller already set.
func runManagedReconcileForWorktreeTestKeepingFlags(t *testing.T) string {
	t.Helper()
	reconcileOutputOnly = true
	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().Bool("output-only", true, "")
	cmd.SetContext(context.Background())
	cmd.SetOut(&out)
	stdout := captureStdout(t, func() {
		if err := runReconcile(cmd, nil); err != nil {
			t.Fatalf("runReconcile() error = %v", err)
		}
	})
	return out.String() + stdout
}
