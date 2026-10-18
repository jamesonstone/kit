package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/spf13/cobra"
)

// Every package fixture stays offline; release transport is tested separately.
func TestMain(m *testing.M) {
	lookupMintActionRelease = func() (mintActionRelease, error) {
		return mintActionRelease{Version: "v0.9.0", SHA: mintPipelineActionRef}, nil
	}
	os.Exit(m.Run())
}

func TestMintActionRefsPreserveWorkflowChoices(t *testing.T) {
	release := mintActionRelease{Version: "v1.0.0", SHA: strings.Repeat("b", 40)}
	oldSHA := strings.Repeat("a", 40)
	source := "# keep\non: workflow_dispatch\njobs:\n  deploy:\n    steps:\n      - uses: 'jamesonstone/mint@v0.8.0' # retain tag format\n      - uses: jamesonstone/mint@" + oldSHA + " # v0.8.0 pinned\n      - uses: jamesonstone/mint@main\n      - uses: acme/action@v2\n        with: {uses: 'jamesonstone/mint@v0.1.0'}\n"
	got, notes, err := refreshMintActionRefs(source, release)
	want := strings.Replace(source, "'jamesonstone/mint@v0.8.0'", "'jamesonstone/mint@v1.0.0'", 1)
	want = strings.Replace(want, oldSHA, release.SHA, 1)
	want = strings.Replace(want, "# v0.8.0 pinned", "# v1.0.0 pinned", 1)
	if err != nil || got != want || len(notes) != 1 {
		t.Fatalf("%s\nnotes=%v err=%v", got, notes, err)
	}
	again, _, err := refreshMintActionRefs(got, release)
	if err != nil || again != got {
		t.Fatal("repeat changed workflow")
	}
}

func TestMintActionUpdatePlanDryRunAndWrite(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	if err := config.Save(root, defaultInitConfig()); err != nil {
		t.Fatal(err)
	}
	path := ".github/workflows/custom.yaml"
	source := strings.Replace(pipelineDeploy, "run: echo original", "uses: jamesonstone/mint@v0.8.0", 1)
	writePipelineFixture(t, root, path, source)
	release := mintActionRelease{Version: "v1.0.0", SHA: strings.Repeat("b", 40)}
	opts := initRefreshOptions{mintRelease: &release, dryRun: true, outputOnly: true}
	captureStdout(t, func() {
		if _, err := runInitRefreshWithSnapshot(root, opts); err != nil {
			t.Fatal(err)
		}
	})
	if readFile(t, filepath.Join(root, path)) != source {
		t.Fatal("preview wrote workflow")
	}
	opts.dryRun = false
	captureStdout(t, func() {
		if _, err := runInitRefreshWithSnapshot(root, opts); err != nil {
			t.Fatal(err)
		}
	})
	if got := readFile(t, filepath.Join(root, path)); got != strings.Replace(source, "mint@v0.8.0", "mint@v1.0.0", 1) {
		t.Fatal(got)
	}
}

func TestMintReleaseFailureStopsInitBeforeWrites(t *testing.T) {
	root := t.TempDir()
	setWorkingDirectory(t, root)
	setupInitHome(t)
	previous := lookupMintActionRelease
	t.Cleanup(func() { lookupMintActionRelease = previous })
	lookupMintActionRelease = func() (mintActionRelease, error) { return mintActionRelease{}, errors.New("offline fixture") }
	withInitFlags(t, func() {
		if err := runInitForTest(initCmd, nil); err == nil || !strings.Contains(err.Error(), "no files changed") {
			t.Fatal(err)
		}
	})
	if config.Exists(root) {
		t.Fatal("release lookup failure wrote configuration")
	}
}

func TestInitScaffoldUsesDiscoveredMintRelease(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, root)
	previous := lookupMintActionRelease
	t.Cleanup(func() { lookupMintActionRelease = previous })
	release := mintActionRelease{Version: "v1.0.0", SHA: strings.Repeat("c", 40)}
	lookupMintActionRelease = func() (mintActionRelease, error) { return release, nil }
	withInitFlags(t, func() {
		initOutputOnly = true
		captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatal(err)
			}
		})
	})
	got := readFile(t, filepath.Join(root, ".github/workflows/mint-release.yaml"))
	if strings.Contains(got, mintPipelineActionRef) || strings.Count(got, "jamesonstone/mint@"+release.SHA) != 3 {
		t.Fatal(got)
	}
}

func TestAliasedMintRefIsPreservedAndReported(t *testing.T) {
	for _, source := range []string{
		"env: {MINT_REF: &mint_ref jamesonstone/mint@v0.8.0}\njobs:\n  check:\n    steps:\n      - uses: *mint_ref\n",
		"x-steps: &mint_steps [{uses: jamesonstone/mint@v0.8.0}]\njobs: {check: {steps: *mint_steps}}\n",
	} {
		got, notes, err := refreshMintActionRefs(source, mintActionRelease{Version: "v1.0.0", SHA: strings.Repeat("a", 40)})
		if err != nil || got != source || len(notes) != 1 || !strings.Contains(notes[0], "unknown") {
			t.Fatalf("%s %v %v", got, notes, err)
		}
	}
}

func TestFeatureMintUpdatePreflightsCanonicalTarget(t *testing.T) {
	setupInitHome(t)
	root, cfg := setupLifecycleTestProject(t)
	path := ".github/workflows/mint.yaml"
	source := "on: push\njobs:\n  check:\n    steps:\n      - uses: jamesonstone/mint@v0.8.0\n"
	writePipelineFixture(t, root, path, source)
	initializeReconcileGitFixture(t, root)
	worktrees := t.TempDir()
	stubReconcileWorktreeRoot(t, worktrees)
	setWorkingDirectory(t, root)
	resetReconcileFlags(t)
	writeFile(t, filepath.Join(cfg.SpecsPath(root), "0001-sample", "SPEC.md"), withFeatureFrontMatter(validSpecWithRelationships("none\n"), "spec", "0001-sample"))
	var err error
	captureStdout(t, func() { err = runReconcile(&cobra.Command{}, []string{"sample"}) })
	if err == nil || !strings.Contains(err.Error(), "no workflow updates applied") {
		t.Fatal(err)
	}
	linked := filepath.Join(worktrees, filepath.Base(root), reconcileBranch)
	if readFile(t, filepath.Join(linked, path)) != source || readFile(t, filepath.Join(root, path)) != source {
		t.Fatal("failed target preflight changed workflow")
	}
}
