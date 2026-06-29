package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/templates"
)

func planRefreshAutoAssignWorkflowFile(
	projectRoot string,
	opts initRefreshOptions,
	cfg *config.Config,
	targets map[string]bool,
) (initRefreshFileChange, error) {
	assignees, err := resolveAutoAssignAssignees(cfg)
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

func resolveAutoAssignAssignees(cfg *config.Config) ([]string, error) {
	if cfg != nil && cfg.GitHub.DefaultAssignees != nil {
		return normalizeAutoAssignAssignees(*cfg.GitHub.DefaultAssignees), nil
	}

	global, found, err := config.LoadGlobal()
	if err != nil {
		return nil, fmt.Errorf("failed to load global Kit config for auto-assign assignees: %w", err)
	}
	if found && global.GitHub.DefaultAssignees != nil {
		return normalizeAutoAssignAssignees(*global.GitHub.DefaultAssignees), nil
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
