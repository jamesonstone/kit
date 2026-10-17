package cli

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const mintPublishGuard = "${{ github.event_name == 'push' && github.ref == format('refs/heads/{0}', github.event.repository.default_branch) && github.event.deleted == false }}"

func mintSelectionRequested(s mintPipelineSelection) bool {
	return s.workflow != "" || s.job != "" || s.beforeStep != "" || s.releaseOwner != ""
}

func validateMintSelection(s mintPipelineSelection) error {
	if s.workflow == "" && s.job == "" && s.beforeStep == "" && s.releaseOwner == "" {
		return nil
	}
	if s.workflow == "" || s.job == "" || s.beforeStep == "" || (s.releaseOwner != "existing" && s.releaseOwner != "mint") {
		return fmt.Errorf("incomplete Mint selection: specify --mint-workflow, --mint-job, --mint-before-step and --mint-release-owner existing|mint")
	}
	return nil
}
func mintPreparedJob(job *yaml.Node) bool {
	steps := mintField(job, "steps")
	if steps == nil {
		return false
	}
	resolve, changelog := false, false
	for _, step := range steps.Content {
		id := mintField(step, "id")
		with := mintField(step, "with")
		command := mintField(with, "command")
		if id != nil && id.Value == "kit_mint_release" && command != nil && command.Value == "release-resolve" {
			resolve = true
		}
		if command != nil && command.Value == "changelog" && mintNodeContains(with, "steps.kit_mint_release.outputs.version_tag") {
			changelog = true
		}
	}
	return resolve && changelog
}
func mintFunctionalStages(node *yaml.Node) bool {
	jobs := mintField(node, "jobs")
	if jobs == nil {
		return false
	}
	for i := 1; i < len(jobs.Content); i += 2 {
		if mintPreparedJob(jobs.Content[i]) {
			return true
		}
	}
	return false
}
func mintNodeContains(node *yaml.Node, value string) bool {
	if node == nil {
		return false
	}
	if strings.Contains(node.Value, value) {
		return true
	}
	for _, child := range node.Content {
		if mintNodeContains(child, value) {
			return true
		}
	}
	return false
}
func mintHandoffReferences(step *yaml.Node) bool {
	return mintNodeContains(step, "steps.kit_mint_release.outputs.version_tag") && mintNodeContains(step, "runner.temp") && mintNodeContains(step, "kit-mint-CHANGELOG.md")
}
func mintInsertionPoint(steps *yaml.Node, before string) (int, error) {
	index := -1
	checkout := -1
	ids := map[string]bool{}
	for i, step := range steps.Content {
		if id := mintField(step, "id"); id != nil {
			if ids[id.Value] || id.Value == "kit_mint_release" || id.Value == "kit_mint_changelog" {
				return 0, fmt.Errorf("duplicate/reserved Mint step id %s", id.Value)
			}
			ids[id.Value] = true
			if id.Value == before {
				index = i
			}
		}
		uses := mintField(step, "uses")
		if uses != nil && strings.HasPrefix(uses.Value, "actions/checkout@") {
			with := mintField(step, "with")
			depth := mintField(with, "fetch-depth")
			if depth == nil || depth.Value != "0" || mintField(with, "repository") != nil || mintField(with, "path") != nil || mintField(with, "sparse-checkout") != nil || mintField(step, "if") != nil {
				return 0, fmt.Errorf("mint needs an unconditional same-repository full-history checkout before the insertion point")
			}
			if ref := mintField(with, "ref"); ref != nil && ref.Value != "${{ github.sha }}" {
				return 0, fmt.Errorf("mint checkout must use github.sha or the default triggering ref")
			}
			checkout = i
		}
	}
	if index < 0 || checkout < 0 || checkout >= index {
		return 0, fmt.Errorf("--mint-before-step must select a step after a full-history checkout")
	}
	return index, nil
}
func validateMintPublisher(root string, workflow mintWorkflow, job, steps *yaml.Node) error {
	permission := mintField(job, "permissions")
	if permission == nil {
		permission = mintField(&workflow.node, "permissions")
	}
	contents := mintField(permission, "contents")
	if contents == nil || contents.Value != "write" {
		return fmt.Errorf("mint publication incomplete: selected job must already grant explicit contents: write; Kit will not expand permissions")
	}
	concurrency := mintField(job, "concurrency")
	if concurrency == nil {
		concurrency = mintField(&workflow.node, "concurrency")
	}
	group := mintField(concurrency, "group")
	cancel := mintField(concurrency, "cancel-in-progress")
	if group == nil || group.Kind != yaml.ScalarNode || strings.Contains(group.Value, "${{") || group.Value == "" || cancel == nil || cancel.Value != "false" {
		return fmt.Errorf("mint publication incomplete: existing concurrency must use a fixed group and cancel-in-progress: false")
	}
	if err := validateMintPushTrigger(root, &workflow.node); err != nil {
		return err
	}
	for _, step := range steps.Content {
		if uses := mintField(step, "uses"); uses != nil && strings.HasPrefix(uses.Value, "actions/checkout@") {
			with := mintField(step, "with")
			if persist := mintField(with, "persist-credentials"); persist != nil && persist.Value != "true" {
				return fmt.Errorf("mint publication incomplete: checkout must already persist Git push credentials")
			}
			if mintField(with, "ssh-key") != nil {
				return fmt.Errorf("mint publication incomplete: custom SSH credential authority needs a project adapter")
			}
			if token := mintField(with, "token"); token != nil && token.Value != "${{ github.token }}" && token.Value != "${{ secrets.GITHUB_TOKEN }}" {
				return fmt.Errorf("mint publication incomplete: custom checkout token authority needs a project adapter")
			}
		}
	}
	workflows, err := mintWorkflows(root)
	if err != nil {
		return err
	}
	for _, existing := range workflows {
		if mintCompetingPublisher(&existing.node) {
			return fmt.Errorf("mint publication incomplete: existing publisher detected in %s; select existing ownership or migrate it explicitly", existing.path)
		}
	}
	return nil
}
func mintCompetingPublisher(node *yaml.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == yaml.ScalarNode {
		value := strings.ToLower(node.Value)
		for _, marker := range []string{"action-gh-release", "ncipollo/release-action", "create-release", "release-please", "semantic-release", "gh release create", "git push --tags", "git push origin refs/tags", "command: release-publish"} {
			if strings.Contains(value, marker) {
				return true
			}
		}
	}
	if uses := mintField(node, "uses"); uses != nil && strings.HasPrefix(strings.ToLower(uses.Value), "jamesonstone/mint@") {
		command := mintField(mintField(node, "with"), "command")
		if command != nil && (command.Value == "release-publish" || command.Value == "release-tag" || command.Value == "github-release" || command.Value == "release-version-main" || command.Value == "deployment-control") {
			return true
		}
	}
	for _, child := range node.Content {
		if mintCompetingPublisher(child) {
			return true
		}
	}
	return false
}
