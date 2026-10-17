package cli

import (
	"os"
	"path/filepath"
	reflectpkg "reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func writePipelineFixture(t *testing.T, root, path, content string) {
	t.Helper()
	full := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

const pipelineDeploy = `name: Deploy
on:
  push:
    branches: [trunk]
permissions:
  contents: read
jobs:
  deploy:
    runs-on: ubuntu-latest
    needs: []
    steps:
      - name: Existing deploy
        env:
          LABEL: original
        run: echo original
`

func TestMintPipelineEmptyAndRepeat(t *testing.T) {
	root := t.TempDir()
	changes, _, err := planMintPipeline(root, mintPipelineSelection{})
	if err != nil || len(changes) != 1 || changes[0].relativePath != ".github/workflows/mint-release.yaml" {
		t.Fatalf("%+v %v", changes, err)
	}
	if _, err := os.Stat(filepath.Join(root, changes[0].relativePath)); !os.IsNotExist(err) {
		t.Fatal("planning wrote a workflow")
	}
	if err := applyInitRefreshFileChangesAtomically(changes); err != nil {
		t.Fatal(err)
	}
	changes, _, err = planMintPipeline(root, mintPipelineSelection{})
	if err != nil || len(changes) != 0 {
		t.Fatalf("repeat: %+v %v", changes, err)
	}
}
func TestMintPipelineExistingDeployPreserved(t *testing.T) {
	root := t.TempDir()
	path := ".github/workflows/deploy.yaml"
	writePipelineFixture(t, root, path, pipelineDeploy)
	changes, _, err := planMintPipeline(root, mintPipelineSelection{})
	if err != nil || len(changes) != 1 || changes[0].relativePath != path {
		t.Fatalf("%+v %v", changes, err)
	}
	before := map[string]any{}
	after := map[string]any{}
	if err := yaml.Unmarshal([]byte(pipelineDeploy), &before); err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal([]byte(changes[0].after), &after); err != nil {
		t.Fatal(err)
	}
	jobs := after["jobs"].(map[string]any)
	job := jobs["deploy"].(map[string]any)
	steps := job["steps"].([]any)
	if len(steps) != 2 || !strings.Contains(changes[0].after, "command: version") || strings.Contains(changes[0].after, "release-publish") {
		t.Fatal(changes[0].after)
	}
	job["steps"] = steps[:1]
	if !reflectpkg.DeepEqual(before, after) {
		t.Fatalf("original semantics changed: %#v", after)
	}
	if !strings.HasPrefix(changes[0].after, pipelineDeploy) {
		t.Fatal("original bytes changed")
	}
	if err := applyInitRefreshFileChangesAtomically(changes); err != nil {
		t.Fatal(err)
	}
	again, _, err := planMintPipeline(root, mintPipelineSelection{})
	if err != nil || len(again) != 0 {
		t.Fatalf("repeat: %+v %v", again, err)
	}
}
func TestMintPipelineAmbiguousSafeAdditive(t *testing.T) {
	for _, kind := range []string{"ci", "ambiguous", "reusable", "reusable callee", "dynamic", "multi-job"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			source := pipelineDeploy
			switch kind {
			case "ci":
				source = strings.Replace(source, "name: Deploy", "name: CI", 1)
			case "ambiguous":
				writePipelineFixture(t, root, ".github/workflows/release.yaml", pipelineDeploy)
			case "reusable":
				source = "name: Deploy\non: push\njobs:\n  deploy:\n    uses: acme/actions/.github/workflows/deploy.yaml@main\n    secrets: inherit\n"
			case "reusable callee":
				source = strings.Replace(source, "push:\n    branches: [trunk]", "workflow_call:", 1)
			case "dynamic":
				source = strings.Replace(source, "ubuntu-latest", "${{ matrix.runner }}", 1)
			case "multi-job":
				source += "  test:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo test\n"
			}
			path := ".github/workflows/workflow.yaml"
			writePipelineFixture(t, root, path, source)
			changes, notes, err := planMintPipeline(root, mintPipelineSelection{})
			if err != nil || len(changes) != 1 || changes[0].relativePath != ".github/workflows/mint-check.yaml" || len(notes) == 0 {
				t.Fatalf("%+v %v %v", changes, notes, err)
			}
			if strings.Contains(changes[0].after, "write") || strings.Contains(changes[0].after, "secrets.") || strings.Contains(changes[0].after, "release-publish") {
				t.Fatal(changes[0].after)
			}
			got, _ := os.ReadFile(filepath.Join(root, path))
			if string(got) != source {
				t.Fatal("existing workflow changed")
			}
		})
	}
}
func TestMintPipelineSelectionAndErrors(t *testing.T) {
	root := t.TempDir()
	path := ".github/workflows/ci.yaml"
	writePipelineFixture(t, root, path, strings.Replace(pipelineDeploy, "name: Deploy", "name: CI", 1))
	changes, _, err := planMintPipeline(root, mintPipelineSelection{path, "deploy"})
	if err != nil || len(changes) != 1 || changes[0].relativePath != path {
		t.Fatalf("%+v %v", changes, err)
	}
	for _, selection := range []mintPipelineSelection{{"", "deploy"}, {path, "missing"}, {".github/workflows/missing.yaml", ""}} {
		if _, _, err := planMintPipeline(root, selection); err == nil {
			t.Fatal(selection)
		}
	}
	writePipelineFixture(t, root, path, "jobs: [invalid\n")
	if _, _, err := planMintPipeline(root, mintPipelineSelection{}); err == nil {
		t.Fatal("invalid YAML accepted")
	}
}
func TestMintPipelineSymlinkRejected(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, ".github")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := planMintPipeline(root, mintPipelineSelection{}); err == nil {
		t.Fatal("symlink accepted")
	}
}

func TestMintPipelineRefusesChangedWorkflowBeforeBatchWrites(t *testing.T) {
	root := t.TempDir()
	path := ".github/workflows/deploy.yaml"
	writePipelineFixture(t, root, path, pipelineDeploy)
	changes, _, err := planMintPipeline(root, mintPipelineSelection{})
	if err != nil {
		t.Fatal(err)
	}
	newer := pipelineDeploy + "# concurrent human edit\n"
	writePipelineFixture(t, root, path, newer)
	changes = append([]initRefreshFileChange{*newInitRefreshFileChange(root, "unrelated.txt", "", "must not write", instructionFileCreated)}, changes...)
	if err := applyInitRefreshFileChangesAtomically(changes); err == nil {
		t.Fatal("concurrent edit overwritten")
	}
	if got := readFile(t, filepath.Join(root, path)); got != newer {
		t.Fatal("human edit changed")
	}
	if _, err := os.Stat(filepath.Join(root, "unrelated.txt")); !os.IsNotExist(err) {
		t.Fatal("batch wrote before validation")
	}
}
func TestMintPipelineExplicitJobPreservesFollowingJob(t *testing.T) {
	root := t.TempDir()
	path := ".github/workflows/deploy.yaml"
	source := pipelineDeploy + "  observe:\n    needs: deploy\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo observe\n"
	writePipelineFixture(t, root, path, source)
	changes, _, err := planMintPipeline(root, mintPipelineSelection{path, "deploy"})
	if err != nil || len(changes) != 1 {
		t.Fatalf("%+v %v", changes, err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(changes[0].after), &doc); err != nil {
		t.Fatal(err)
	}
	jobs := mintField(&doc, "jobs")
	if len(mintField(mintField(jobs, "deploy"), "steps").Content) != 2 || len(mintField(mintField(jobs, "observe"), "steps").Content) != 1 {
		t.Fatal(changes[0].after)
	}
}

func TestMintPipelinePrivilegedJobsRequireExplicitSelection(t *testing.T) {
	for _, source := range []string{
		strings.Replace(pipelineDeploy, "run: echo original", "uses: aws-actions/configure-aws-credentials@v4\n        with:\n          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}", 1),
		strings.Replace(pipelineDeploy, "run: echo original", "run: echo TOKEN=value >> $GITHUB_ENV", 1),
		strings.Replace(pipelineDeploy, "run: echo original", "uses: acme/action@v1\n        with:\n          token: ${{ secrets['TOKEN'] }}", 1),
		strings.Replace(pipelineDeploy, "  contents: read", "  contents: write", 1),
		strings.Replace(pipelineDeploy, "  contents: read", "  contents: read\n  id-token: write", 1),
		strings.Replace(pipelineDeploy, "    needs: []", "    needs: []\n    environment: production", 1),
		strings.Replace(pipelineDeploy, "    needs: []", "    needs: []\n    env:\n      TOKEN: ${{ secrets.TOKEN }}", 1),
		strings.Replace(pipelineDeploy, "permissions:\n  contents: read\n", "", 1),
		strings.Replace(pipelineDeploy, "permissions:\n  contents: read", "permissions: write-all", 1),
	} {
		root := t.TempDir()
		path := ".github/workflows/deploy.yaml"
		writePipelineFixture(t, root, path, source)
		changes, _, err := planMintPipeline(root, mintPipelineSelection{})
		if err != nil || len(changes) != 1 || changes[0].relativePath != ".github/workflows/mint-check.yaml" {
			t.Fatalf("automatic: %+v %v", changes, err)
		}
		changes, _, err = planMintPipeline(root, mintPipelineSelection{path, "deploy"})
		if err != nil || len(changes) != 1 || changes[0].relativePath != path {
			t.Fatalf("explicit: %+v %v", changes, err)
		}
	}
}
func TestMintSnapshotValidationPreservesGenericSymlinkHandling(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "owned.yaml")
	writePipelineFixture(t, root, "owned.yaml", "before")
	path := ".github/workflows/auto-assign.yml"
	if err := os.MkdirAll(filepath.Dir(filepath.Join(root, path)), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, path)); err != nil {
		t.Fatal(err)
	}
	change := newInitRefreshFileChange(root, path, "before", "after", instructionFileUpdated)
	if err := applyInitRefreshFileChangesAtomically([]initRefreshFileChange{*change}); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, target); got != "after" {
		t.Fatal(got)
	}
	strict := newMintPipelineFileChange(root, path, "after", "forbidden", instructionFileUpdated)
	if err := applyInitRefreshFileChangesAtomically([]initRefreshFileChange{*strict}); err == nil {
		t.Fatal("Mint symlink accepted")
	}
}
