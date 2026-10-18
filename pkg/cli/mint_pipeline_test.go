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
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - id: deploy
        name: Existing deploy
        env:
          LABEL: original
        run: echo original
`

func TestMintPipelineExistingDeployPreserved(t *testing.T) {
	root := t.TempDir()
	path := ".github/workflows/deploy.yaml"
	writePipelineFixture(t, root, path, pipelineDeploy)
	diagnostic, notes, err := planMintPipeline(root, mintPipelineSelection{})
	if err != nil || len(diagnostic) != 0 || !strings.Contains(strings.Join(notes, " "), "INCOMPLETE") {
		t.Fatalf("%+v %v %v", diagnostic, notes, err)
	}
	changes, notes, err := planMintPipeline(root, mintPipelineSelection{workflow: path, job: "deploy", beforeStep: "deploy"})
	if err != nil || len(changes) != 1 || !strings.Contains(strings.Join(notes, " "), "Added Mint preparation") {
		t.Fatalf("%+v %v %v", changes, notes, err)
	}
	before := map[string]any{}
	after := map[string]any{}
	if err := yaml.Unmarshal([]byte(pipelineDeploy), &before); err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal([]byte(changes[0].after), &after); err != nil {
		t.Fatal(err)
	}
	job := after["jobs"].(map[string]any)["deploy"].(map[string]any)
	steps := job["steps"].([]any)
	if len(steps) != 5 || strings.Contains(changes[0].after, "release-publish") {
		t.Fatal(changes[0].after)
	}
	job["steps"] = []any{steps[0], steps[4]}
	if !reflectpkg.DeepEqual(before, after) {
		t.Fatalf("original semantics changed: %#v", after)
	}
	if !strings.Contains(changes[0].after, "        name: Existing deploy\n        env:\n          LABEL: original\n        run: echo original") {
		t.Fatal("original step bytes changed")
	}
	if err := applyInitRefreshFileChangesAtomically(changes); err != nil {
		t.Fatal(err)
	}
	again, _, err := planMintPipeline(root, mintPipelineSelection{})
	if err != nil || len(again) != 0 {
		t.Fatalf("repeat: %+v %v", again, err)
	}
}
func TestMintPipelineAmbiguousPlacementPreservesFiles(t *testing.T) {
	root := t.TempDir()
	path := ".github/workflows/deploy.yaml"
	writePipelineFixture(t, root, path, pipelineDeploy)
	writePipelineFixture(t, root, ".github/workflows/ci.yaml", pipelineDeploy)
	changes, notes, err := planMintPipeline(root, mintPipelineSelection{})
	if err != nil || len(changes) != 0 || !strings.Contains(strings.Join(notes, " "), "--mint-before-step") {
		t.Fatalf("%+v %v %v", changes, notes, err)
	}
	if readFile(t, filepath.Join(root, path)) != pipelineDeploy {
		t.Fatal("existing workflow changed")
	}
}

func TestMintPipelineExistingMintPreserved(t *testing.T) {
	root := t.TempDir()
	path := ".github/workflows/custom.yaml"
	source := strings.Replace(pipelineDeploy, "run: echo original", "uses: jamesonstone/mint@custom", 1)
	writePipelineFixture(t, root, path, source)
	// A custom Mint use counts as inclusion, without Kit certifying its lifecycle.
	for _, selection := range []mintPipelineSelection{{}, {workflow: path, job: "deploy", beforeStep: "deploy"}} {
		changes, _, err := planMintPipeline(root, selection)
		if err != nil || len(changes) != 0 {
			t.Fatalf("%+v %v", changes, err)
		}
	}
	if readFile(t, filepath.Join(root, path)) != source {
		t.Fatal("existing Mint workflow changed")
	}
}

func TestMintPipelineSelectionAndErrors(t *testing.T) {
	root := t.TempDir()
	path := ".github/workflows/ci.yaml"
	writePipelineFixture(t, root, path, strings.Replace(pipelineDeploy, "name: Deploy", "name: CI", 1))
	changes, _, err := planMintPipeline(root, mintPipelineSelection{workflow: path, job: "deploy", beforeStep: "deploy"})
	if err != nil || len(changes) != 1 || changes[0].relativePath != path {
		t.Fatalf("%+v %v", changes, err)
	}
	for _, selection := range []mintPipelineSelection{{job: "deploy"}, {workflow: path, job: "missing", beforeStep: "deploy"}, {workflow: ".github/workflows/missing.yaml"}} {
		if _, _, err := planMintPipeline(root, selection); err == nil {
			t.Fatal(selection)
		}
	}
	for _, source := range []string{
		strings.Replace(pipelineDeploy, "id: deploy", "id: different", 1),
		strings.Replace(pipelineDeploy, "id: deploy", "id: kit_mint_release", 1),
		pipelineDeploy + "      - id: deploy\n        run: echo duplicate\n",
	} {
		writePipelineFixture(t, root, path, source)
		if _, _, err := planMintPipeline(root, mintPipelineSelection{workflow: path, job: "deploy", beforeStep: "deploy"}); err == nil {
			t.Fatal("missing anchor or conflicting step ID accepted")
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
	changes, _, err := planMintPipeline(root, mintPipelineSelection{workflow: path, job: "deploy", beforeStep: "deploy"})
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
	changes, _, err := planMintPipeline(root, mintPipelineSelection{workflow: path, job: "deploy", beforeStep: "deploy"})
	if err != nil || len(changes) != 1 {
		t.Fatalf("%+v %v", changes, err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(changes[0].after), &doc); err != nil {
		t.Fatal(err)
	}
	jobs := mintField(&doc, "jobs")
	if len(mintField(mintField(jobs, "deploy"), "steps").Content) != 5 || len(mintField(mintField(jobs, "observe"), "steps").Content) != 1 {
		t.Fatal(changes[0].after)
	}
}

func TestMintPipelinePreservesProjectWorkflowChoices(t *testing.T) {
	root := t.TempDir()
	path := ".github/workflows/deploy.yaml"
	source := strings.Replace(pipelineDeploy, "ubuntu-latest", "${{ matrix.runner }}", 1)
	source = strings.Replace(source, "push:\n    branches: [trunk]", "workflow_dispatch:", 1)
	source = strings.Replace(source, "          fetch-depth: 0\n", "          fetch-depth: 1\n", 1)
	writePipelineFixture(t, root, path, source)
	writeFile(t, filepath.Join(root, mintPolicyPath), "# project-owned policy\n")
	changes, _, err := planMintPipeline(root, mintPipelineSelection{workflow: path, job: "deploy", beforeStep: "deploy"})
	if err != nil || len(changes) != 1 {
		t.Fatalf("%+v %v", changes, err)
	}
	for _, original := range []string{"workflow_dispatch:", "runs-on: ${{ matrix.runner }}", "fetch-depth: 1", "contents: read", "run: echo original"} {
		if !strings.Contains(changes[0].after, original) {
			t.Fatalf("project choice changed: %s", original)
		}
	}
	if strings.Contains(changes[0].after, "release-publish") || strings.Contains(changes[0].after, "concurrency:") {
		t.Fatal("insertion imposed release policy")
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
