package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const mintPolicyPath = ".mint.yaml"

func validateMintControllerOptions(controller, workflows bool, selection mintPipelineSelection) error {
	if controller && (workflows || mintSelectionRequested(selection)) {
		return fmt.Errorf("--mint cannot be combined with --mint-workflows or Mint workflow selection flags; select one scaffold at a time")
	}
	return nil
}

func incompleteMintNotes(notes []string) []string {
	var incomplete []string
	for _, note := range notes {
		if strings.Contains(note, "INCOMPLETE") {
			incomplete = append(incomplete, note)
		}
	}
	return incomplete
}

func appendMintIncompletePrompt(prompt string, notes []string) string {
	if notes = incompleteMintNotes(notes); len(notes) > 0 {
		prompt += "\n\n## Incomplete Mint integration\n\n- " + strings.Join(notes, "\n- ") + "\n"
	}
	return prompt
}

func printInitRefreshNotes(notes []string, opts initRefreshOptions) {
	if opts.outputOnly {
		notes = incompleteMintNotes(notes)
	}
	if len(notes) == 0 {
		return
	}
	fmt.Println("\nNotes:")
	for _, note := range notes {
		fmt.Printf("   - %s\n", note)
	}
}

func printCurrentMintIncompleteNotes(root string, opts initRefreshOptions) error {
	if opts.mintRelease != nil {
		_, notes, err := planMintActionUpdates(root, nil, *opts.mintRelease)
		if err != nil {
			return err
		}
		printInitRefreshNotes(notes, initRefreshOptions{outputOnly: true})
	}
	if !opts.mintWorkflows && !mintSelectionRequested(opts.mintSelection) {
		return nil
	}
	_, notes, err := planMintPipeline(root, opts.mintSelection)
	if err != nil {
		return err
	}
	printInitRefreshNotes(notes, initRefreshOptions{outputOnly: true})
	return nil
}

// Plan only a missing v1 controller. Mint validates and renders the policy;
// Kit checks ownership, link safety and callback collisions before invoking it.
func planMintWorkflow(root string, release *mintActionRelease) (*initRefreshFileChange, error) {
	if release != nil && !release.hasController() {
		return nil, fmt.Errorf("deployment controller unavailable in Mint %s; use project-owned deployment workflows", release.Version)
	}
	data, err := readMintProjectFile(root, mintPolicyPath)
	if err != nil {
		return nil, fmt.Errorf("--mint requires an existing .mint.yaml: %w", err)
	}
	var policy struct {
		Schema          int    `yaml:"schema_version"`
		Mode            string `yaml:"mode"`
		Authorization   string `yaml:"authorization"`
		ControlWorkflow string `yaml:"control_workflow"`
		BuildWorkflow   string `yaml:"build_workflow"`
		Environments    map[string]struct {
			Scope               string `yaml:"scope"`
			PromotionWorkflow   string `yaml:"promotion_workflow"`
			ObservationWorkflow string `yaml:"observation_workflow"`
			Configuration       string `yaml:"configuration_sha256"`
		} `yaml:"environments"`
	}
	if err := yaml.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("read Mint policy: %w", err)
	}
	if policy.Schema != 2 || policy.Mode != "deployment" || policy.Authorization != "repository-write" {
		return nil, fmt.Errorf("--mint requires a schema 2 deployment policy with authorization: repository-write; package/artifact projects retain their publication workflows")
	}
	path := policy.ControlWorkflow
	if path == "" {
		path = ".github/workflows/mint-recovery.yaml"
	}
	if !validMintWorkflowPath(path) {
		return nil, fmt.Errorf("invalid Mint control_workflow %q", path)
	}
	before, err := readMintProjectFile(root, path)
	if err == nil {
		return newInitRefreshFileChange(root, path, string(before), string(before), instructionFileSkipped), nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	paths := map[string]bool{policy.BuildWorkflow: true}
	var environment string
	for name, env := range policy.Environments {
		if environment == "" || name < environment {
			environment = name
		}
		if env.Scope == "shared" && (env.PromotionWorkflow == "" || env.ObservationWorkflow == "" || env.Configuration == "") {
			return nil, fmt.Errorf("environment %s needs reviewed promotion/observation adapters and configuration identity before scaffolding", name)
		}
		for _, callback := range []string{env.PromotionWorkflow, env.ObservationWorkflow} {
			if callback != "" {
				paths[callback] = true
			}
		}
	}
	if paths[path] {
		return nil, fmt.Errorf("control_workflow must be separate from build, promotion and observation adapters")
	}
	if err := checkMintCallbackNames(root, paths); err != nil {
		return nil, err
	}
	if release == nil {
		return nil, fmt.Errorf("latest Mint release is required to render the controller")
	}
	content, err := renderMintController(root, *release, environment)
	if err != nil {
		return nil, err
	}
	content = "# Scaffolded by Kit. This workflow is project owned.\n" + content
	return newMintPipelineFileChange(root, path, "", content, instructionFileCreated), nil
}

func checkMintCallbackNames(root string, paths map[string]bool) error {
	seenNames := map[string]bool{"Mint environment control": true}
	for callback := range paths {
		if !validMintWorkflowPath(callback) {
			return fmt.Errorf("invalid Mint callback workflow path %q", callback)
		}
		data, err := readMintProjectFile(root, callback)
		if err != nil {
			return fmt.Errorf("read Mint adapter %s: %w", callback, err)
		}
		var workflow struct {
			Name string `yaml:"name"`
		}
		if err := yaml.Unmarshal(data, &workflow); err != nil {
			return fmt.Errorf("parse Mint adapter %s: %w", callback, err)
		}
		if workflow.Name == "" {
			workflow.Name = callback
		}
		if strings.Contains(workflow.Name, "${{") {
			return fmt.Errorf("mint callback name must not contain a GitHub Actions expression")
		}
		if seenNames[workflow.Name] {
			return fmt.Errorf("mint callback names must be unique and separate from the environment controller")
		}
		seenNames[workflow.Name] = true
	}
	return nil
}

var mintWorkflowPath = regexp.MustCompile(`^\.github/workflows/[A-Za-z0-9_.-]+\.ya?ml$`)

func validMintWorkflowPath(path string) bool {
	return mintWorkflowPath.MatchString(path)
}

// Reject links before reads and writes, including linked parent directories.
func readMintProjectFile(root, relative string) ([]byte, error) {
	if relative != mintPolicyPath && !validMintWorkflowPath(relative) {
		return nil, fmt.Errorf("unsafe Mint project path %q", relative)
	}
	path := root
	for _, component := range strings.Split(relative, "/") {
		path = filepath.Join(path, component)
		info, err := os.Lstat(path)
		if err != nil {
			return nil, err
		}
		if path == filepath.Join(root, filepath.FromSlash(relative)) && !info.Mode().IsRegular() {
			return nil, fmt.Errorf("mint project file %s must be a regular file", relative)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("mint project path %s is a symbolic link", relative)
		}
	}
	return os.ReadFile(path)
}
