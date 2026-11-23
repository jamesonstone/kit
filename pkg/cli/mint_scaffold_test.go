package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

const mintFixturePolicy = `schema_version: 2
mode: deployment
repository: team/service
default_branch: trunk
authorization: repository-write
build_workflow: .github/workflows/build.yaml
control_workflow: .github/workflows/control.yaml
environments:
  preview:
    scope: shared
    deploy: manual
    follow: latest
    configuration_sha256: sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
    promotion_workflow: .github/workflows/deploy.yaml
    observation_workflow: .github/workflows/observe.yaml
  workstation:
    scope: local
    deploy: manual
    follow: latest
`

func mintFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, mintPolicyPath), mintFixturePolicy)
	for path, name := range map[string]string{"build.yaml": "Build candidates", "deploy.yaml": "Deploy [preview]", "observe.yaml": "Observe preview"} {
		writeFile(t, filepath.Join(root, ".github/workflows", path), "name: '"+name+"'\non: workflow_dispatch\njobs: {}\n")
	}
	return root
}

func TestMintScaffoldPassesVerifiedReleaseToRenderer(t *testing.T) {
	root := mintFixture(t)
	previous := renderMintController
	t.Cleanup(func() { renderMintController = previous })
	release := mintActionRelease{Version: "v1.3.0", SHA: strings.Repeat("c", 40)}
	renderMintController = func(dir string, got mintActionRelease, environment string) (string, error) {
		if dir != root || got != release || environment != "preview" {
			t.Fatalf("renderer got %q %+v %q", dir, got, environment)
		}
		return "name: Mint environment control\n", nil
	}
	change, err := planMintWorkflow(root, &release)
	if err != nil || !strings.HasSuffix(change.after, "name: Mint environment control\n") {
		t.Fatalf("plan=%+v err=%v", change, err)
	}
	if _, err := os.Stat(change.absolutePath); !os.IsNotExist(err) {
		t.Fatal("planning wrote the controller")
	}
}

func TestMintV2CannotScaffoldLegacyController(t *testing.T) {
	root := mintFixture(t)
	release := mintActionRelease{Version: "v2.0.0", SHA: strings.Repeat("b", 40)}
	if change, err := planMintWorkflow(root, &release); change != nil || err == nil || !strings.Contains(err.Error(), "project-owned deployment workflows") {
		t.Fatalf("plan=%+v err=%v", change, err)
	}
}

func TestMintScaffoldRejectsUnsupportedPoliciesAndMissingAdapters(t *testing.T) {
	for _, item := range []struct{ name, old, replacement string }{
		{"package", "mode: deployment", "mode: package"},
		{"personal", "authorization: repository-write", "human_login: operator"},
		{"missing observer", "    observation_workflow: .github/workflows/observe.yaml\n", ""},
		{"missing adapter", "build.yaml", "absent.yaml"},
	} {
		t.Run(item.name, func(t *testing.T) {
			root := mintFixture(t)
			writeFile(t, filepath.Join(root, mintPolicyPath), strings.Replace(mintFixturePolicy, item.old, item.replacement, 1))
			if _, err := planMintWorkflow(root, &testMintRelease); err == nil {
				t.Fatal("invalid or incomplete policy accepted")
			}
		})
	}
}

func TestMintScaffoldPreservesExistingControllerAndAdapters(t *testing.T) {
	root := mintFixture(t)
	path := filepath.Join(root, ".github/workflows/control.yaml")
	writeFile(t, path, "project-owned controller\n")
	if err := config.Save(root, defaultInitConfig()); err != nil {
		t.Fatal(err)
	}
	setupInitHome(t)
	plan, err := buildInitRefreshPlan(context.Background(), root, initRefreshOptions{mint: true, force: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range plan.changes {
		if change.absolutePath == path && (change.result != instructionFileSkipped || change.before != change.after) {
			t.Fatal("force changed the existing controller")
		}
	}
	if err := applyInitRefreshFileChangesAtomically(plan.changes); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(root, mintPolicyPath)); got != mintFixturePolicy {
		t.Fatal("policy was rewritten")
	}
	if got := readFile(t, path); got != "project-owned controller\n" {
		t.Fatal("controller was rewritten")
	}
}

func TestMintScaffoldNeverReadsOrWritesThroughLinks(t *testing.T) {
	for _, relative := range []string{mintPolicyPath, ".github/workflows/control.yaml", ".github/workflows/build.yaml", ".github/workflows"} {
		t.Run(relative, func(t *testing.T) {
			root := mintFixture(t)
			path := filepath.Join(root, relative)
			if err := os.RemoveAll(path); err != nil {
				t.Fatal(err)
			}
			outside := filepath.Join(t.TempDir(), "private")
			writeFile(t, outside, "unchanged")
			if err := os.Symlink(outside, path); err != nil {
				t.Fatal(err)
			}
			if _, err := planMintWorkflow(root, &testMintRelease); err == nil {
				t.Fatal("linked path accepted")
			}
			if readFile(t, outside) != "unchanged" {
				t.Fatal("external file changed")
			}
		})
	}
}

func TestInitMintPreflightsBeforeCreatingKitConfig(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, root)
	withInitFlags(t, func() {
		initMint, initOutputOnly = true, true
		if err := runInitForTest(initCmd, nil); err == nil {
			t.Fatal("missing policy accepted")
		}
		if config.Exists(root) {
			t.Fatal("invalid Mint request partially initialized Kit")
		}
	})
}

func TestInitMintAndReconcileDryRunAndConvergence(t *testing.T) {
	root := mintFixture(t)
	setupInitHome(t)
	setWorkingDirectory(t, root)
	withInitFlags(t, func() {
		initMint, initOutputOnly = true, true
		captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatal(err)
			}
		})
	})
	path := filepath.Join(root, ".github/workflows/control.yaml")
	if !strings.Contains(readFile(t, path), testMintRelease.SHA) {
		t.Fatal("init did not scaffold the published controller")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	captureStdout(t, func() {
		if _, err := runInitRefreshWithSnapshot(root, initRefreshOptions{mint: true, mintRelease: &testMintRelease, dryRun: true, diff: true}); err != nil {
			t.Fatal(err)
		}
	})
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("dry-run wrote the controller")
	}
	plan, err := buildInitRefreshPlan(context.Background(), root, initRefreshOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range plan.changes {
		if change.absolutePath == path {
			t.Fatal("Mint scaffolding was not opt-in")
		}
	}
	captureStdout(t, func() {
		if _, err := runInitRefreshWithSnapshot(root, initRefreshOptions{mint: true, mintRelease: &testMintRelease}); err != nil {
			t.Fatal(err)
		}
	})
	before := readFile(t, path)
	change, err := planMintWorkflow(root, &testMintRelease)
	if err != nil || change.result != instructionFileSkipped || change.after != before {
		t.Fatalf("second generation did not preserve the controller: %v", err)
	}
}

func TestMintScaffoldRejectsAmbiguousAndSelfReferencingCallbacks(t *testing.T) {
	for _, name := range []string{"Mint environment control", "Deploy [preview]", "Build ${{ github.token }}"} {
		t.Run(name, func(t *testing.T) {
			root := mintFixture(t)
			writeFile(t, filepath.Join(root, ".github/workflows/build.yaml"), "name: '"+name+"'\non: workflow_dispatch\njobs: {}\n")
			if _, err := planMintWorkflow(root, &testMintRelease); err == nil {
				t.Fatal("ambiguous, self-triggering or expression-bearing callback accepted")
			}
		})
	}
}
