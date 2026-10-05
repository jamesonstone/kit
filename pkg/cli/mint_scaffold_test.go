package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"gopkg.in/yaml.v3"
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

func TestMintScaffoldUsesPublishedControllerAndProjectEnvironmentNames(t *testing.T) {
	root := mintFixture(t)
	change, err := planMintWorkflow(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"jamesonstone/mint@" + mintActionRef,
		"command: deployment-control", "production-config: '.mint.yaml'",
		"branches: ['trunk']", `options: ["preview","workstation"]`,
		"promote, hotfix, rollback, resume, observe, reconcile",
		"vars.MINT_RELEASE_ENABLED == 'true'", "persist-credentials: false",
		`ref: ${{ github.event.repository.default_branch }}`,
		`workflows: ["Build candidates","Deploy \\[preview\\]","Observe preview"]`,
	} {
		if !strings.Contains(change.after, want) {
			t.Errorf("generated controller lacks %q", want)
		}
	}
	for _, unsafe := range []string{"cache: true", "cache-from", "id-token: write", "hotfix_pr }}", "human_login"} {
		if strings.Contains(change.after, unsafe) {
			t.Errorf("controller contains %q", unsafe)
		}
	}
	var workflow map[string]any
	if err := yaml.Unmarshal([]byte(change.after), &workflow); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(change.absolutePath); !os.IsNotExist(err) {
		t.Fatal("planning wrote the controller")
	}
}

func TestMintScaffoldRejectsUnsupportedPoliciesAndMissingAdapters(t *testing.T) {
	for _, item := range []struct{ name, old, replacement string }{
		{"package", "mode: deployment", "mode: package"},
		{"personal", "authorization: repository-write", "human_login: operator"},
		{"missing observer", "    observation_workflow: .github/workflows/observe.yaml\n", ""},
		{"missing adapter", "build.yaml", "absent.yaml"},
		{"unsafe callback", "default_branch: trunk", "default_branch: ../outside"},
	} {
		t.Run(item.name, func(t *testing.T) {
			root := mintFixture(t)
			writeFile(t, filepath.Join(root, mintPolicyPath), strings.Replace(mintFixturePolicy, item.old, item.replacement, 1))
			if _, err := planMintWorkflow(root); err == nil {
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
			if _, err := planMintWorkflow(root); err == nil {
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
	if !strings.Contains(readFile(t, path), mintActionRef) {
		t.Fatal("init did not scaffold the published controller")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	captureStdout(t, func() {
		if _, err := runInitRefreshWithSnapshot(root, initRefreshOptions{mint: true, dryRun: true, diff: true}); err != nil {
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
		if _, err := runInitRefreshWithSnapshot(root, initRefreshOptions{mint: true}); err != nil {
			t.Fatal(err)
		}
	})
	before := readFile(t, path)
	change, err := planMintWorkflow(root)
	if err != nil || change.result != instructionFileSkipped || change.after != before {
		t.Fatalf("second generation did not preserve the controller: %v", err)
	}
}
