package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/document"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

func planRefreshAutoAssignWorkflowFile(
	projectRoot string,
	opts initRefreshOptions,
	cfg *config.Config,
	targets map[string]bool,
) (initRefreshFileChange, error) {
	assignees, err := resolveAutoAssignAssignees(projectRoot, cfg, opts.resetsProjectConfig(targets))
	if err != nil {
		return initRefreshFileChange{}, err
	}
	content := templates.BuildAutoAssignWorkflow(assignees)
	path := filepath.Join(projectRoot, filepath.FromSlash(autoAssignWorkflowPath))
	exists := document.Exists(path)
	explicit := len(targets) > 0

	var before string
	if exists {
		data, err := os.ReadFile(path)
		if err != nil {
			return initRefreshFileChange{}, fmt.Errorf("failed to read %s: %w", autoAssignWorkflowPath, err)
		}
		before = string(data)
		content = preserveNewerWorkflowActionPins(before, content)
	}
	if !exists {
		return *newInitRefreshFileChange(projectRoot, autoAssignWorkflowPath, before, content, instructionFileCreated), nil
	}
	if before == content {
		return *newInitRefreshFileChange(projectRoot, autoAssignWorkflowPath, before, before, instructionFileSkipped), nil
	}
	if opts.force && explicit {
		return *newInitRefreshFileChange(projectRoot, autoAssignWorkflowPath, before, content, instructionFileUpdated), nil
	}
	if isKitManagedAutoAssignWorkflow(before) {
		return *newInitRefreshFileChange(projectRoot, autoAssignWorkflowPath, before, content, instructionFileUpdated), nil
	}
	return *newInitRefreshFileChange(projectRoot, autoAssignWorkflowPath, before, before, instructionFileSkipped), nil
}

func resolveAutoAssignAssignees(projectRoot string, cfg *config.Config, reset bool) ([]string, error) {
	if cfg != nil && cfg.GitHub.DefaultAssignees != nil {
		return normalizeAutoAssignAssignees(*cfg.GitHub.DefaultAssignees), nil
	}

	if !reset {
		explicit, err := config.RepositoryDeclaresDefaultAssignees(projectRoot)
		if explicit || err != nil {
			return nil, err
		}
	}
	assignees, err := config.LoadGlobalDefaultAssignees()
	if err != nil {
		return nil, fmt.Errorf("failed to load global Kit config for auto-assign assignees: %w", err)
	}
	if assignees != nil {
		return normalizeAutoAssignAssignees(*assignees), nil
	}
	return nil, nil
}

func normalizeAutoAssignAssignees(assignees []string) []string {
	seen := make(map[string]bool, len(assignees))
	var normalized []string
	for _, assignee := range assignees {
		assignee = strings.TrimSpace(strings.TrimPrefix(assignee, "@"))
		if assignee == "" || seen[assignee] {
			continue
		}
		seen[assignee] = true
		normalized = append(normalized, assignee)
	}
	return normalized
}

func isKitManagedAutoAssignWorkflow(content string) bool {
	return strings.Contains(content, "# Kit-managed auto-assignment workflow.")
}
