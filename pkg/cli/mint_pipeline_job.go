package cli

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func integrateMintJob(root string, workflow mintWorkflow, selection mintPipelineSelection) ([]initRefreshFileChange, []string, error) {
	if mintComplexYAML(&workflow.node) {
		return nil, nil, fmt.Errorf("selected workflow uses YAML anchors/aliases/merges; choose a block-style steps list for insertion")
	}
	job := mintField(mintField(&workflow.node, "jobs"), selection.job)
	steps := mintField(job, "steps")
	if steps == nil || steps.Kind != yaml.SequenceNode || len(steps.Content) == 0 || mintField(job, "uses") != nil {
		return nil, nil, fmt.Errorf("select a job with an ordinary nonempty steps list")
	}
	index, err := mintInsertionPoint(steps, selection.beforeStep)
	if err != nil {
		return nil, nil, err
	}
	content, err := insertMintStages(workflow.data, steps, index)
	if err != nil {
		return nil, nil, err
	}
	note := "Added Mint preparation before step " + selection.beforeStep + ". Existing workflow settings and publication steps are unchanged; review checkout history and Mint inputs for your project."
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
