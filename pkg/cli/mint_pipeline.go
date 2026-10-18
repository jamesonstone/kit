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

type mintPipelineSelection struct{ workflow, job, beforeStep string }

var initMintSelection, reconcileMintSelection mintPipelineSelection
var reconcileMintWorkflows bool

func bindMintPipelineSelection(cmd *cobra.Command, selection *mintPipelineSelection) {
	cmd.Flags().StringVar(&selection.workflow, "mint-workflow", "", "existing workflow path for Mint preparation steps")
	cmd.Flags().StringVar(&selection.job, "mint-job", "", "steps job (required with --mint-workflow and --mint-before-step)")
	cmd.Flags().StringVar(&selection.beforeStep, "mint-before-step", "", "existing step id before which to resolve version and generate changelog")
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

// Only job/step uses fields are Action references; similarly named inputs are not.
func mintActionMappings(workflow *yaml.Node) []*yaml.Node {
	jobs := mintField(workflow, "jobs")
	if jobs == nil || jobs.Kind != yaml.MappingNode {
		return nil
	}
	var mappings []*yaml.Node
	for i := 1; i < len(jobs.Content); i += 2 {
		job := jobs.Content[i]
		mappings = append(mappings, job)
		if steps := mintField(job, "steps"); steps != nil && steps.Kind == yaml.SequenceNode {
			mappings = append(mappings, steps.Content...)
		}
	}
	return mappings
}

func mintUses(node *yaml.Node) bool {
	for _, mapping := range mintActionMappings(node) {
		if value := mintField(mapping, "uses"); value != nil {
			if value.Kind == yaml.AliasNode && value.Alias != nil {
				value = value.Alias
			}
			if strings.HasPrefix(strings.ToLower(value.Value), "jamesonstone/mint@") {
				return true
			}
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
	if err := validateMintSelection(selection); err != nil {
		return nil, nil, err
	}
	for _, workflow := range workflows {
		if mintUses(&workflow.node) {
			return nil, []string{"Existing Mint Actions integration preserved."}, nil
		}
	}
	if selection.workflow != "" {
		for _, workflow := range workflows {
			if workflow.path == selection.workflow {
				return integrateMintJob(root, workflow, selection)
			}
		}
		return nil, nil, fmt.Errorf("selected Mint workflow %q is not an existing workflow", selection.workflow)
	}

	if len(workflows) == 0 {
		content := renderMintReleasePipeline(mintDefaultBranch(root))
		if len(content) == 0 {
			return nil, nil, fmt.Errorf("cannot render Mint pipeline for the remote default branch")
		}
		change := newMintPipelineFileChange(root, ".github/workflows/mint-release.yaml", "", string(content), instructionFileCreated)
		return []initRefreshFileChange{*change}, []string{"Mint starter workflow added; review it for your project."}, nil
	}
	return nil, []string{"INCOMPLETE Mint placement: existing workflows were preserved. Select --mint-workflow <path> --mint-job <id> --mint-before-step <id>, then preview with --dry-run --diff."}, nil
}

func mintKnownDefaultBranch(root string) (string, bool) {
	cmd := exec.Command("git", "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD")
	cmd.Dir = root
	if out, err := cmd.Output(); err == nil {
		if branch := strings.TrimPrefix(strings.TrimSpace(string(out)), "origin/"); branch != "" {
			return branch, true
		}
	}
	return "", false
}
func mintDefaultBranch(root string) string {
	if branch, known := mintKnownDefaultBranch(root); known {
		return branch
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
		if !change.strictMintWorkflow {
			continue
		}
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

func newMintPipelineFileChange(root, path, before, after string, result instructionFileWriteResult) *initRefreshFileChange {
	change := newInitRefreshFileChange(root, path, before, after, result)
	change.strictMintWorkflow = true
	return change
}
