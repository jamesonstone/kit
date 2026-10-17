package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type mintPipelineSelection struct{ workflow, job string }

var initMintSelection, reconcileMintSelection mintPipelineSelection
var reconcileMintWorkflows bool

func bindMintPipelineSelection(cmd *cobra.Command, selection *mintPipelineSelection) {
	cmd.Flags().StringVar(&selection.workflow, "mint-workflow", "", "existing workflow path for a Mint install/version step")
	cmd.Flags().StringVar(&selection.job, "mint-job", "", "steps job in --mint-workflow (required when multiple jobs exist)")
}

type mintWorkflow struct {
	path string
	data []byte
	node yaml.Node
}

func mintField(node *yaml.Node, key string) *yaml.Node {
	if node == nil {
		return nil
	}
	if node.Kind == yaml.DocumentNode && len(node.Content) == 1 {
		node = node.Content[0]
	}
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}
func mintUses(node *yaml.Node) bool {
	if node == nil {
		return false
	}
	if value := mintField(node, "uses"); value != nil && strings.HasPrefix(strings.ToLower(value.Value), "jamesonstone/mint@") {
		return true
	}
	for _, child := range node.Content {
		if mintUses(child) {
			return true
		}
	}
	return false
}
func mintWorkflows(root string) ([]mintWorkflow, error) {
	for _, part := range []string{".github", ".github/workflows"} {
		info, err := os.Lstat(filepath.Join(root, part))
		if os.IsNotExist(err) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("mint workflow directory must be a real directory: %s", part)
		}
	}
	entries, err := os.ReadDir(filepath.Join(root, ".github/workflows"))
	if err != nil {
		return nil, err
	}
	var result []mintWorkflow
	for _, entry := range entries {
		if entry.IsDir() || (filepath.Ext(entry.Name()) != ".yaml" && filepath.Ext(entry.Name()) != ".yml") {
			continue
		}
		path := ".github/workflows/" + entry.Name()
		data, err := readMintProjectFile(root, path)
		if err != nil {
			return nil, err
		}
		var node yaml.Node
		if err := yaml.Unmarshal(data, &node); err != nil {
			return nil, fmt.Errorf("parse workflow %s: %w", path, err)
		}
		if mintField(&node, "jobs") == nil {
			return nil, fmt.Errorf("workflow %s has no jobs mapping", path)
		}
		result = append(result, mintWorkflow{path, data, node})
	}
	return result, nil
}

// Plan before init creates auto-assignment, so a truly empty repository gets
// one release publisher. All existing files remain project owned.
func planMintPipeline(root string, selection mintPipelineSelection) ([]initRefreshFileChange, []string, error) {
	workflows, err := mintWorkflows(root)
	if err != nil {
		return nil, nil, err
	}
	if selection.job != "" && selection.workflow == "" {
		return nil, nil, fmt.Errorf("--mint-job requires --mint-workflow")
	}
	if selection.workflow != "" {
		for _, workflow := range workflows {
			if workflow.path == selection.workflow {
				return integrateMintJob(root, workflow, selection.job)
			}
		}
		return nil, nil, fmt.Errorf("selected Mint workflow %q is not an existing workflow", selection.workflow)
	}
	for _, workflow := range workflows {
		if mintUses(&workflow.node) {
			return nil, []string{"Existing Mint integration preserved; action pins and publisher behavior remain project owned."}, nil
		}
	}
	if len(workflows) == 0 {
		content := renderMintReleasePipeline(mintDefaultBranch(root))
		if len(content) == 0 {
			return nil, nil, fmt.Errorf("cannot render Mint pipeline for the remote default branch")
		}
		change := newInitRefreshFileChange(root, ".github/workflows/mint-release.yaml", "", string(content), instructionFileCreated)
		return []initRefreshFileChange{*change}, []string{"Mint release-state pipeline added. Runtime deployment requires reviewed project adapters; no deployment is inferred."}, nil
	}
	var candidates []mintWorkflow
	for _, workflow := range workflows {
		name := workflow.path
		if n := mintField(&workflow.node, "name"); n != nil {
			name += " " + n.Value
		}
		name = strings.ToLower(name)
		if strings.Contains(name, "deploy") || strings.Contains(name, "release") {
			candidates = append(candidates, workflow)
		}
	}
	if len(candidates) == 1 && !mintReusableWorkflow(&candidates[0].node) {
		changes, notes, err := integrateMintJob(root, candidates[0], "")
		if err == nil {
			return changes, notes, nil
		}
	}
	path := ".github/workflows/mint-check.yaml"
	for _, workflow := range workflows {
		if workflow.path == path {
			return nil, []string{"Mint target is ambiguous; existing mint-check.yaml preserved. Select --mint-workflow and --mint-job with --dry-run --diff."}, nil
		}
	}
	change := newInitRefreshFileChange(root, path, "", string(renderMintValidationPipeline(mintDefaultBranch(root))), instructionFileCreated)
	return []initRefreshFileChange{*change}, []string{"No single safe release/deploy job selected. Added read-only Mint validation; preview an explicit target with --mint-workflows --mint-workflow <path> --mint-job <id> --dry-run --diff. Existing publishers and deploys are unchanged."}, nil
}

func mintDefaultBranch(root string) string {
	cmd := exec.Command("git", "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD")
	cmd.Dir = root
	if out, err := cmd.Output(); err == nil {
		if branch := strings.TrimPrefix(strings.TrimSpace(string(out)), "origin/"); branch != "" {
			return branch
		}
	}
	// A new repository without remote metadata uses GitHub's conventional starter.
	return "main"
}

func appendMintPipelinePlan(root string, selection mintPipelineSelection, changes *[]initRefreshFileChange, notes *[]string) error {
	planned, messages, err := planMintPipeline(root, selection)
	if err != nil {
		return err
	}
	*changes = append(*changes, planned...)
	*notes = append(*notes, messages...)
	return nil
}

// A workflow can be edited while init gathers interactive configuration. Check
// the reviewed snapshot again before any files in the atomic batch are written.
func validateMintWorkflowChanges(changes []initRefreshFileChange) error {
	seen := map[string]bool{}
	for _, change := range changes {
		if !strings.HasPrefix(change.relativePath, ".github/workflows/") || change.result == instructionFileSkipped {
			continue
		}
		if seen[change.absolutePath] {
			return fmt.Errorf("conflicting workflow plans for %s", change.relativePath)
		}
		seen[change.absolutePath] = true
		root := strings.TrimSuffix(change.absolutePath, filepath.FromSlash(change.relativePath))
		data, err := readMintProjectFile(root, change.relativePath)
		if change.result == instructionFileCreated && os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("workflow changed since planning: %s: %w", change.relativePath, err)
		}
		if change.result == instructionFileCreated || string(data) != change.before {
			return fmt.Errorf("workflow changed since planning: %s", change.relativePath)
		}
	}
	return nil
}

func mintReusableWorkflow(node *yaml.Node) bool {
	trigger := mintField(node, "on")
	if trigger == nil {
		return false
	}
	if trigger.Kind == yaml.ScalarNode {
		return trigger.Value == "workflow_call"
	}
	if trigger.Kind == yaml.MappingNode {
		return mintField(trigger, "workflow_call") != nil
	}
	for _, event := range trigger.Content {
		if event.Value == "workflow_call" {
			return true
		}
	}
	return false
}
