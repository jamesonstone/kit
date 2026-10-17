package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/spf13/cobra"
)

func TestInitAutomaticallyPlansMintBeforeAutoAssign(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, root)
	withInitFlags(t, func() {
		initOutputOnly = true
		output := captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatal(err)
			}
		})
		pipeline := readFile(t, filepath.Join(root, ".github/workflows/mint-release.yaml"))
		if !strings.Contains(pipeline, "command: release-publish") || !strings.Contains(output, ".github/workflows/mint-release.yaml") {
			t.Fatal("pipeline absent from init/delivery")
		}
		if _, err := os.Stat(filepath.Join(root, ".github/workflows/mint-check.yaml")); !os.IsNotExist(err) {
			t.Fatal("auto-assignment changed empty-workflow detection")
		}
	})
}
func TestMintPipelineReconcileDryRunAndForceIdempotence(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	if err := config.Save(root, defaultInitConfig()); err != nil {
		t.Fatal(err)
	}
	opts := initRefreshOptions{mintWorkflows: true, dryRun: true, diff: true}
	output := captureStdout(t, func() {
		if _, err := runInitRefreshWithSnapshot(root, opts); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(output, "mint-release.yaml") || !strings.Contains(output, "release-publish") {
		t.Fatal(output)
	}
	if _, err := os.Stat(filepath.Join(root, ".github/workflows")); !os.IsNotExist(err) {
		t.Fatal("dry-run wrote workflows")
	}
	opts.dryRun = false
	opts.diff = false
	opts.force = true
	opts.outputOnly = true
	if _, err := runInitRefreshWithSnapshot(root, opts); err != nil {
		t.Fatal(err)
	}
	before := readFile(t, filepath.Join(root, ".github/workflows/mint-release.yaml"))
	if _, err := runInitRefreshWithSnapshot(root, opts); err != nil {
		t.Fatal(err)
	}
	if after := readFile(t, filepath.Join(root, ".github/workflows/mint-release.yaml")); after != before {
		t.Fatal("repeat/force rewrote pipeline")
	}
}
func TestMintPipelineInvalidInitHasNoPartialConfig(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, root)
	writePipelineFixture(t, root, ".github/workflows/deploy.yaml", "jobs: [invalid\n")
	withInitFlags(t, func() {
		initOutputOnly = true
		if err := runInitForTest(initCmd, nil); err == nil {
			t.Fatal("invalid workflow accepted")
		}
		if config.Exists(root) {
			t.Fatal("init wrote config before rejecting invalid workflow")
		}
	})
}
func TestReconcileMintWorkflowsRejectsPartialScope(t *testing.T) {
	resetReconcileFlags(t)
	reconcileMintWorkflows = true
	if err := runReconcile(&cobra.Command{}, []string{"feature"}); err == nil {
		t.Fatal("feature-only integration accepted")
	}
	reconcileRefreshFiles = []string{"AGENTS.md"}
	if err := runReconcile(&cobra.Command{}, nil); err == nil {
		t.Fatal("file-only integration accepted")
	}
}
func TestMintPipelineReadsRemoteDefaultBranch(t *testing.T) {
	root := t.TempDir()
	if err := config.Save(root, defaultInitConfig()); err != nil {
		t.Fatal(err)
	}
	initializeReconcileGitFixture(t, root)
	reconcileGitOutput(t, root, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/trunk")
	changes, _, err := planMintPipeline(root, mintPipelineSelection{})
	if err != nil || len(changes) != 1 || !strings.Contains(changes[0].after, "- trunk") {
		t.Fatalf("%+v %v", changes, err)
	}
}
