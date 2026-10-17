package cli

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func integrateMintJob(root string, workflow mintWorkflow, selected string) ([]initRefreshFileChange, []string, error) {
	if mintUses(&workflow.node) {
		return nil, []string{"Selected workflow already uses Mint; existing pins and behavior preserved."}, nil
	}
	if mintComplexYAML(&workflow.node) {
		return nil, nil, fmt.Errorf("select a workflow without YAML anchors, aliases or merge keys for Mint integration")
	}
	jobs := mintField(&workflow.node, "jobs")
	if jobs == nil || jobs.Kind != yaml.MappingNode {
		return nil, nil, fmt.Errorf("selected workflow needs a jobs mapping")
	}
	var job *yaml.Node
	if selected != "" {
		job = mintField(jobs, selected)
	} else if len(jobs.Content) == 2 {
		job = jobs.Content[1]
	}
	if job == nil {
		return nil, nil, fmt.Errorf("mint workflow %s needs an explicit --mint-job selection", workflow.path)
	}
	steps := mintField(job, "steps")
	runner := mintField(job, "runs-on")
	if steps == nil || steps.Kind != yaml.SequenceNode || len(steps.Content) == 0 || mintField(job, "uses") != nil || mintField(job, "container") != nil || runner == nil || runner.Kind != yaml.ScalarNode || runner.Value != "ubuntu-latest" {
		return nil, nil, fmt.Errorf("mint integration needs an ordinary ubuntu-latest steps job without a container; reusable/dynamic jobs require a project-owned adapter")
	}
	// Append after the last original step, without re-encoding any existing YAML.
	// Mint's action sets up Go and PATH, so placing it last avoids changing earlier
	// build/deploy tool selection. It does not acquire a token or publish releases.
	lines := strings.Split(string(workflow.data), "\n")
	start := steps.Content[0].Line - 1
	if start < 0 || start >= len(lines) {
		return nil, nil, fmt.Errorf("cannot locate Mint step insertion")
	}
	prefix := lines[start][:len(lines[start])-len(strings.TrimLeft(lines[start], " "))]
	if !strings.HasPrefix(strings.TrimSpace(lines[start]), "-") {
		return nil, nil, fmt.Errorf("mint integration requires block-style steps")
	}
	end := start + 1
	for end < len(lines) {
		line := lines[end]
		trim := strings.TrimSpace(line)
		if trim != "" && !strings.HasPrefix(trim, "#") && len(line)-len(strings.TrimLeft(line, " ")) < len(prefix) {
			break
		}
		end++
	}
	step := prefix + "- name: Verify Mint tooling (Kit)\n" + prefix + "  uses: jamesonstone/mint@" + mintPipelineActionRef + " # v0.9.0\n" + prefix + "  with:\n" + prefix + "    command: version\n"
	before := strings.Join(lines[:end], "\n")
	after := strings.Join(lines[end:], "\n")
	if !strings.HasSuffix(before, "\n") {
		before += "\n"
	}
	content := before + step + after
	var check yaml.Node
	if err := yaml.Unmarshal([]byte(content), &check); err != nil {
		return nil, nil, fmt.Errorf("validate Mint workflow insertion: %w", err)
	}
	change := newInitRefreshFileChange(root, workflow.path, string(workflow.data), content, instructionFileUpdated)
	return []initRefreshFileChange{*change}, []string{"Mint install/version step added after existing steps in " + workflow.path + ". Existing publishers, deploys, triggers and permissions preserved."}, nil
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
func renderMintValidationPipeline(_ string) []byte {
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
