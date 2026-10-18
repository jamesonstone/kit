package cli

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func integrateMintJob(root string, workflow mintWorkflow, selection mintPipelineSelection) ([]initRefreshFileChange, []string, error) {
	if mintComplexYAML(&workflow.node) {
		return nil, nil, fmt.Errorf("selected workflow uses YAML anchors/aliases/merges; prepare a project-owned adapter")
	}
	if _, err := readMintProjectFile(root, mintPolicyPath); err == nil {
		return nil, nil, fmt.Errorf("existing .mint.yaml owns release policy; use reviewed policy adapters instead of adding another pipeline")
	} else if !os.IsNotExist(err) {
		return nil, nil, err
	}
	job := mintField(mintField(&workflow.node, "jobs"), selection.job)
	steps := mintField(job, "steps")
	runner := mintField(job, "runs-on")
	if steps == nil || steps.Kind != yaml.SequenceNode || len(steps.Content) == 0 || mintField(job, "uses") != nil || mintField(job, "container") != nil || mintField(job, "strategy") != nil || runner == nil || runner.Value != "ubuntu-latest" {
		return nil, nil, fmt.Errorf("select an ordinary ubuntu-latest steps job without a container; reusable/dynamic jobs require an adapter")
	}
	if mintPreparedJob(job) {
		return nil, []string{"INCOMPLETE adoption review: existing Mint preparation preserved; verify publication ownership, handoff and project-owned pins manually."}, nil
	}
	index, err := mintInsertionPoint(steps, selection.beforeStep)
	if err != nil {
		return nil, nil, err
	}
	if selection.releaseOwner == "mint" {
		if err := validateMintPublisher(root, workflow, job, steps); err != nil {
			return nil, nil, err
		}
	}
	content, err := insertMintStages(workflow.data, steps, index, selection.releaseOwner == "mint")
	if err != nil {
		return nil, nil, err
	}
	note := "CONFIGURED Mint version resolution and enriched changelog artifact before step " + selection.beforeStep + "."
	if selection.releaseOwner == "mint" {
		note += " Mint owns publication after existing job steps succeed; no deployment trigger, permission or environment changed."
	} else {
		note += " INCOMPLETE publication handoff review: verify or bind the existing consumer to steps.kit_mint_release.outputs.version_tag and runner.temp/kit-mint-CHANGELOG.md; existing publisher was preserved."
	}
	change := newMintPipelineFileChange(root, workflow.path, string(workflow.data), content, instructionFileUpdated)
	return []initRefreshFileChange{*change}, []string{note}, nil
}

func mintComplexYAML(node *yaml.Node) bool {
	if node.Anchor != "" || node.Kind == yaml.AliasNode || node.Tag == "!!merge" {
		return true
	}
	for _, child := range node.Content {
		if mintComplexYAML(child) {
			return true
		}
	}
	return false
}
func renderMintValidationPipeline() []byte {
	return []byte(`# Scaffolded by Kit. Read-only Mint tooling check; no application code runs.
name: Mint tooling
on:
  push:
  pull_request:
permissions:
  contents: read
jobs:
  mint:
    runs-on: ubuntu-latest
    steps:
      - name: Verify Mint tooling
        uses: jamesonstone/mint@` + mintPipelineActionRef + ` # v0.9.0
        with:
          command: version
`)
}
