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

func TestResolveReconcileRefreshDecisionDefersPrimaryWrites(t *testing.T) {
	stubReconcileWorktreeInspection(t, worktreeprep.Location{
		InsideGit: true,
		IsPrimary: true,
	})

	decision, err := resolveReconcileRefreshDecision("/repo", true, false)
	if err != nil {
		t.Fatalf("resolveReconcileRefreshDecision() error = %v", err)
	}
	if decision.Apply || !decision.Deferred {
		t.Fatalf("decision = %#v, want deferred without apply", decision)
	}
}

func TestResolveReconcileRefreshDecisionPreservesWritableAndReadOnlyPaths(t *testing.T) {
	for _, test := range []struct {
		name      string
		location  worktreeprep.Location
		requested bool
		dryRun    bool
		wantApply bool
	}{
		{name: "linked worktree", location: worktreeprep.Location{InsideGit: true}, requested: true, wantApply: true},
		{name: "non-Git project", requested: true, wantApply: true},
		{name: "primary dry-run", location: worktreeprep.Location{InsideGit: true, IsPrimary: true}, requested: true, dryRun: true, wantApply: true},
		{name: "not requested", location: worktreeprep.Location{InsideGit: true, IsPrimary: true}},
	} {
		t.Run(test.name, func(t *testing.T) {
			stubReconcileWorktreeInspection(t, test.location)
			decision, err := resolveReconcileRefreshDecision("/repo", test.requested, test.dryRun)
			if err != nil {
				t.Fatalf("resolveReconcileRefreshDecision() error = %v", err)
			}
			if decision.Apply != test.wantApply || decision.Deferred {
				t.Fatalf("decision = %#v", decision)
			}
		})
	}
}

func TestResolveReconcileRefreshDecisionFailsClosed(t *testing.T) {
	previous := inspectReconcileWorktree
	inspectReconcileWorktree = func(string) (worktreeprep.Location, error) {
		return worktreeprep.Location{}, errors.New("inspect failed")
	}
	t.Cleanup(func() { inspectReconcileWorktree = previous })

	_, err := resolveReconcileRefreshDecision("/repo", true, false)
	if err == nil {
		t.Fatal("resolveReconcileRefreshDecision() unexpectedly succeeded")
	}
}

func TestBuildDeferredReconcileCommandPreservesWriteIntent(t *testing.T) {
	t.Run("default interactive refresh", func(t *testing.T) {
		resetReconcileFlags(t)
		if got := buildDeferredReconcileCommand(nil); got != "kit reconcile --include-files --output-only" {
			t.Fatalf("buildDeferredReconcileCommand() = %q", got)
		}
	})

	t.Run("whole project", func(t *testing.T) {
		resetReconcileFlags(t)
		reconcileAll = true
		if got := buildDeferredReconcileCommand(nil); got != "kit reconcile --all --include-files --output-only" {
			t.Fatalf("buildDeferredReconcileCommand() = %q", got)
		}
	})

	t.Run("filtered forced feature refresh", func(t *testing.T) {
		resetReconcileFlags(t)
		restorePromptProfileState(t, promptProfileFrontend, true)
		previousSingleAgent := singleAgent
		singleAgent = true
		t.Cleanup(func() { singleAgent = previousSingleAgent })
		reconcileForce = true
		reconcileRefreshFiles = []string{"docs/rules/owner's-rule.md", "AGENTS.md"}
		reconcileMigrateReferences = true
		reconcileMigrateVerification = true

		got := buildDeferredReconcileCommand([]string{"sample feature"})
		want := "kit reconcile 'sample feature' --include-files --force " +
			"--file 'docs/rules/owner'\"'\"'s-rule.md' --file 'AGENTS.md' " +
			"--migrate-references --migrate-verification --profile='frontend' --single-agent --output-only"
		if got != want {
			t.Fatalf("buildDeferredReconcileCommand() = %q, want %q", got, want)
		}
		instructions := strings.Join(
			managedFileDeliveryInstructionsForCommand("/repo", got),
			"\n",
		)
		if !strings.Contains(instructions, got) {
			t.Fatalf("delivery instructions omit exact rerun command:\n%s", instructions)
		}
	})
}

func TestRunReconcileDefersPrimaryRefreshWithoutChangingOutputContract(t *testing.T) {
	projectRoot := setupManagedSafetyGuidanceProject(t)
	initializeReconcileGitFixture(t, projectRoot)
	stubManagedSafetyRulesetRegistry(t)
	setWorkingDirectory(t, projectRoot)

	output := runManagedReconcileForWorktreeTest(t)
	if status := reconcileGitOutput(t, projectRoot, "status", "--porcelain"); status != "" {
		t.Fatalf("primary checkout changed:\n%s", status)
	}
	if _, err := os.Stat(filepath.Join(projectRoot, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatal("primary checkout unexpectedly received managed AGENTS.md")
	}
	for _, expected := range []string{
		"Delivery of command-created files:",
		"No exact command-owned path snapshot is present",
		"canonical non-primary writable worktree",
		"rerun the write-capable Kit command with this exact shell-safe invocation: kit reconcile --include-files --output-only",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("deferred reconcile output missing %q:\n%s", expected, output)
		}
	}
}

func TestRunReconcileAppliesManagedRefreshInLinkedWorktree(t *testing.T) {
	projectRoot := setupManagedSafetyGuidanceProject(t)
	initializeReconcileGitFixture(t, projectRoot)
	linkedRoot := filepath.Join(t.TempDir(), "GH-160")
	runGitForSourceAuditTest(t, projectRoot, "worktree", "add", "-b", "GH-160", linkedRoot, "main")
	stubManagedSafetyRulesetRegistry(t)
	setWorkingDirectory(t, linkedRoot)

	output := runManagedReconcileForWorktreeTest(t)
	if info, err := os.Stat(filepath.Join(linkedRoot, "AGENTS.md")); err != nil || !info.Mode().IsRegular() {
		t.Fatal("linked worktree did not receive managed AGENTS.md")
	}
	if status := reconcileGitOutput(t, projectRoot, "status", "--porcelain"); status != "" {
		t.Fatalf("primary checkout changed:\n%s", status)
	}
	for _, expected := range []string{
		"Delivery of command-created files:",
		"Treat only this exact snapshot as command-owned evidence",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("linked reconcile output missing %q:\n%s", expected, output)
		}
	}
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
	reconcileIncludeFiles = true
	reconcileOutputOnly = true

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().Bool("output-only", true, "")
	addPromptOnlyFlag(cmd)
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
