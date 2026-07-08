package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/templates"
)

func planRefreshInitScaffoldFiles(
	projectRoot string,
	opts initRefreshOptions,
	cfg *config.Config,
	targets map[string]bool,
) ([]initRefreshFileChange, error) {
	files := []struct {
		relativePath string
		content      string
		merge        bool
	}{
		{relativePath: gitignorePath, content: templates.Gitignore, merge: true},
		{relativePath: envPath, content: ""},
		{relativePath: envrcPath, content: templates.Envrc},
		{relativePath: codeRabbitConfigPath, content: templates.CodeRabbitConfig},
		{relativePath: pullRequestTemplatePath, content: templates.PullRequestTemplate},
	}

	var changes []initRefreshFileChange
	for _, file := range files {
		if !initRefreshTargetMatches(targets, file.relativePath) {
			continue
		}
		change, err := planRefreshInitScaffoldFile(projectRoot, opts, targets, file.relativePath, file.content, file.merge)
		if err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	if initRefreshTargetMatches(targets, autoAssignWorkflowPath) {
		change, err := planRefreshAutoAssignWorkflowFile(projectRoot, opts, cfg, targets)
		if err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	return changes, nil
}

func planRefreshInitScaffoldFile(
	projectRoot string,
	opts initRefreshOptions,
	targets map[string]bool,
	relativePath string,
	content string,
	merge bool,
) (initRefreshFileChange, error) {
	path := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
	exists := document.Exists(path)
	explicit := len(targets) > 0

	var before string
	if exists {
		data, err := os.ReadFile(path)
		if err != nil {
			return initRefreshFileChange{}, fmt.Errorf("failed to read %s: %w", relativePath, err)
		}
		before = string(data)
	}

	if exists && opts.force && explicit {
		if before == content {
			return *newInitRefreshFileChange(projectRoot, relativePath, before, before, instructionFileSkipped), nil
		}
		return *newInitRefreshFileChange(projectRoot, relativePath, before, content, instructionFileUpdated), nil
	}
	if exists && merge {
		missing := missingGitignorePatterns(before)
		if len(missing) == 0 {
			return *newInitRefreshFileChange(projectRoot, relativePath, before, before, instructionFileSkipped), nil
		}
		return *newInitRefreshFileChange(
			projectRoot,
			relativePath,
			before,
			appendGitignorePatterns(before, missing),
			instructionFileMerged,
		), nil
	}
	if exists {
		return *newInitRefreshFileChange(projectRoot, relativePath, before, before, instructionFileSkipped), nil
	}
	return *newInitRefreshFileChange(projectRoot, relativePath, before, content, instructionFileCreated), nil
}

func planRefreshInitConstitution(projectRoot string, cfg *config.Config, targets map[string]bool) (*initRefreshFileChange, error) {
	relativePath := filepath.ToSlash(cfg.ConstitutionPath)
	if !initRefreshTargetMatches(targets, relativePath) {
		return nil, nil
	}

	path := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
	exists := document.Exists(path)
	before := ""
	after := templates.Constitution
	result := instructionFileCreated
	if exists {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", relativePath, err)
		}
		before = string(data)
		after = before
		result = instructionFileSkipped
	}
	after = mergeDocumentContent(relativePath, after, templates.Constitution, document.TypeConstitution)
	updated, changed := upsertConstitutionBaseline(after)
	if changed {
		after = updated
	}

	if exists && before != after {
		result = instructionFileMerged
	}
	if exists && before == after {
		result = instructionFileSkipped
	}
	return newInitRefreshFileChange(projectRoot, relativePath, before, after, result), nil
}

func mergeDocumentContent(path string, content string, templateContent string, docType document.DocumentType) string {
	existing := document.Parse(content, path, docType)
	template := document.Parse(templateContent, "", docType)

	var missingSections []document.Section
	for _, section := range template.Sections {
		if !existing.HasSection(section.Name) {
			missingSections = append(missingSections, section)
		}
	}
	if len(missingSections) == 0 {
		return content
	}

	merged := content
	for _, section := range missingSections {
		merged += fmt.Sprintf("\n\n## %s\n\n%s", section.Name, section.Content)
	}
	return merged
}

func planRefreshInitInstructionArtifacts(
	projectRoot string,
	opts initRefreshOptions,
	cfg *config.Config,
	targets map[string]bool,
) ([]initRefreshFileChange, error) {
	var changes []initRefreshFileChange
	for _, relativePath := range instructionArtifactPaths(
		cfg,
		instructionFileSelection{},
		config.InstructionScaffoldVersionTOC,
		true,
	) {
		relativePath = filepath.ToSlash(relativePath)
		if !initRefreshTargetMatches(targets, relativePath) {
			continue
		}
		mode := instructionFileWriteModeAppendOnly
		if opts.force || exactLegacyInstructionArtifact(projectRoot, relativePath) {
			mode = instructionFileWriteModeOverwrite
		}
		plan, err := planInstructionArtifactWrite(projectRoot, relativePath, mode, config.InstructionScaffoldVersionTOC)
		if err != nil {
			return nil, err
		}
		change, err := initRefreshChangeFromInstructionPlan(projectRoot, plan)
		if err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	return changes, nil
}

func initRefreshChangeFromInstructionPlan(projectRoot string, plan instructionFileWritePlan) (initRefreshFileChange, error) {
	before := ""
	after := plan.content
	if document.Exists(plan.absolutePath) {
		data, err := os.ReadFile(plan.absolutePath)
		if err != nil {
			return initRefreshFileChange{}, fmt.Errorf("failed to read %s: %w", plan.relativePath, err)
		}
		before = string(data)
		if plan.result == instructionFileSkipped {
			after = before
		}
	}
	return *newInitRefreshFileChange(projectRoot, plan.relativePath, before, after, plan.result), nil
}
