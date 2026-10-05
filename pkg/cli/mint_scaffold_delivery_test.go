package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/spf13/cobra"
)

func TestReconcileMintFromPrimaryKeepsPolicyAndCheckoutUntouched(t *testing.T) {
	root := mintFixture(t)
	setupInitHome(t)
	if err := config.Save(root, defaultInitConfig()); err != nil {
		t.Fatal(err)
	}
	initializeReconcileGitFixture(t, root)
	worktrees := t.TempDir()
	stubReconcileWorktreeRoot(t, worktrees)
	setWorkingDirectory(t, root)
	resetReconcileFlags(t)
	reconcileMint = true
	captureStdout(t, func() {
		if err := runReconcile(&cobra.Command{}, nil); err != nil {
			t.Fatal(err)
		}
	})
	if status := reconcileGitOutput(t, root, "status", "--porcelain"); status != "" {
		t.Fatalf("primary changed: %s", status)
	}
	path := ".github/workflows/control.yaml"
	if _, err := os.Stat(filepath.Join(root, path)); !os.IsNotExist(err) {
		t.Fatal("controller was written into primary")
	}
	linked := filepath.Join(worktrees, filepath.Base(root), reconcileBranch)
	if !strings.Contains(readFile(t, filepath.Join(linked, path)), mintActionRef) {
		t.Fatal("linked worktree lacks Mint controller")
	}
	if readFile(t, filepath.Join(linked, mintPolicyPath)) != mintFixturePolicy {
		t.Fatal("reconcile changed project-owned policy")
	}
}

func TestReconcileMintRejectsPartialScope(t *testing.T) {
	resetReconcileFlags(t)
	reconcileMint = true
	if err := runReconcile(&cobra.Command{}, []string{"feature"}); err == nil {
		t.Fatal("feature-only Mint generation accepted")
	}
	reconcileRefreshFiles = []string{"AGENTS.md"}
	if err := runReconcile(&cobra.Command{}, nil); err == nil {
		t.Fatal("file-only Mint generation accepted")
	}
}
