package prfix

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/registry"
)

func LoadContract(root string) (registry.PRFeedbackContract, error) {
	config, migration, err := registry.LoadProject(root)
	if err != nil {
		return registry.PRFeedbackContract{}, err
	}
	if migration {
		return registry.PRFeedbackContract{}, fmt.Errorf("schema-v1 project requires `kit reconcile --json --diff`")
	}
	record, found := registry.RecordByKey(config.Registry.Artifacts, registry.KindWorkflow, "pr-feedback-repair")
	if !found {
		return registry.PRFeedbackContract{}, fmt.Errorf("workflow/pr-feedback-repair is not installed; run `kit reconcile --json --diff`")
	}
	if record.State == registry.StateConflict || record.State == registry.StateMissing {
		return registry.PRFeedbackContract{}, fmt.Errorf("workflow/pr-feedback-repair is %s; run `kit reconcile --json --diff`", record.State)
	}
	path, err := confinedWorkflowPath(root, record.Path)
	if err != nil {
		return registry.PRFeedbackContract{}, err
	}
	// #nosec G304 -- confinedWorkflowPath verified this registry path is inside root.
	content, err := os.ReadFile(path)
	if err != nil {
		return registry.PRFeedbackContract{}, fmt.Errorf("read workflow/pr-feedback-repair: %w", err)
	}
	document, err := registry.ParseMarkdown(string(content))
	if err != nil {
		return registry.PRFeedbackContract{}, err
	}
	if document.Metadata.PRFeedback == nil {
		return registry.PRFeedbackContract{}, fmt.Errorf("workflow/pr-feedback-repair has no structured feedback contract")
	}
	contract := *document.Metadata.PRFeedback
	if err := registry.ValidatePRFeedbackContract(contract); err != nil {
		return registry.PRFeedbackContract{}, fmt.Errorf("validate workflow/pr-feedback-repair: %w", err)
	}
	return contract, nil
}

func confinedWorkflowPath(root, relative string) (string, error) {
	if relative == "" || filepath.IsAbs(relative) {
		return "", fmt.Errorf("workflow path %q is not project-relative", relative)
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	path := filepath.Join(root, filepath.FromSlash(relative))
	within, err := filepath.Rel(root, path)
	if err != nil || within == ".." || strings.HasPrefix(within, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("workflow path %q escapes the project", relative)
	}
	return path, nil
}
