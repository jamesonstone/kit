package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamesonstone/mint/pkg/promotion"
	"gopkg.in/yaml.v3"
)

const mintActionRef = "b97969136d5a43d0982c46c6f185868db16d14bf" // v0.5.0
const mintPolicyPath = ".mint.yaml"

// planMintWorkflow uses the published renderer, without installing or running
// project code. Policy and adapters remain project owned; only a missing
// controller is scaffolded. Existing controllers are preserved even with force.
func planMintWorkflow(root string) (*initRefreshFileChange, error) {
	data, err := readMintProjectFile(root, mintPolicyPath)
	if err != nil {
		return nil, fmt.Errorf("--mint requires an existing .mint.yaml: %w", err)
	}
	policy, err := promotion.ParsePolicy(data)
	if err != nil {
		return nil, err
	}
	if policy.Schema != 2 || policy.Mode != "deployment" || policy.Authorization != "repository-write" {
		return nil, fmt.Errorf("--mint requires a schema 2 deployment policy with authorization: repository-write; package/artifact projects retain their publication workflows")
	}
	path := (promotion.Config{ControlWorkflow: policy.ControlWorkflow}).ControlWorkflowPath()
	before, err := readMintProjectFile(root, path)
	if err == nil {
		return newInitRefreshFileChange(root, path, string(before), string(before), instructionFileSkipped), nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	paths := map[string]bool{policy.BuildWorkflow: true}
	for name, env := range policy.Environments {
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
	ordered := make([]string, 0, len(paths))
	for callback := range paths {
		ordered = append(ordered, callback)
	}
	sort.Strings(ordered)
	names := make([]string, 0, len(ordered))
	seenNames := map[string]bool{}
	for _, callback := range ordered {
		data, err := readMintProjectFile(root, callback)
		if err != nil {
			return nil, fmt.Errorf("read Mint adapter %s: %w", callback, err)
		}
		var workflow struct {
			Name string `yaml:"name"`
		}
		if err := yaml.Unmarshal(data, &workflow); err != nil {
			return nil, fmt.Errorf("parse Mint adapter %s: %w", callback, err)
		}
		if workflow.Name == "" {
			workflow.Name = callback
		}
		if workflow.Name == "Mint environment control" || seenNames[workflow.Name] {
			return nil, fmt.Errorf("mint callback names must be unique and separate from the environment controller")
		}
		seenNames[workflow.Name] = true
		names = append(names, workflow.Name)
	}
	content, err := promotion.RenderEnvironmentWorkflow(policy, mintActionRef, mintPolicyPath, names...)
	if err != nil {
		return nil, err
	}
	content = "# Scaffolded by Kit with Mint v0.5.0. This workflow is project owned.\n" + content
	return newInitRefreshFileChange(root, path, "", content, instructionFileCreated), nil
}

// Reject links before reads and writes, including linked parent directories.
func readMintProjectFile(root, relative string) ([]byte, error) {
	if !promotion.SafeRepositoryPath(relative) {
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
