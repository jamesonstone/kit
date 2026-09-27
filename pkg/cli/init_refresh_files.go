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

func planRefreshInitScaffoldFiles(
	projectRoot string,
	opts initRefreshOptions,
	cfg *config.Config,
	targets map[string]bool,
) ([]initRefreshFileChange, error) {
	files := []struct {
		relativePath     string
		content          string
		merge            bool
		preserveExisting bool
	}{
		{relativePath: gitignorePath, content: templates.Gitignore, merge: true},
		{relativePath: envPath, content: ""},
		{relativePath: envrcPath, content: templates.Envrc},
		{relativePath: makefilePath, content: templates.Makefile, preserveExisting: true},
		{relativePath: codeRabbitConfigPath, content: templates.CodeRabbitConfig},
		{relativePath: pullRequestTemplatePath, content: templates.PullRequestTemplate},
	}

	var changes []initRefreshFileChange
	for _, file := range files {
		if !initRefreshTargetMatches(targets, file.relativePath) {
			continue
		}
		change, err := planRefreshInitScaffoldFile(
			projectRoot,
			opts,
			targets,
			file.relativePath,
			file.content,
			file.merge,
			file.preserveExisting,
		)
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
	preserveExisting bool,
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

	if exists && preserveExisting {
		return *newInitRefreshFileChange(projectRoot, relativePath, before, before, instructionFileSkipped), nil
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

func planRefreshInitConstitution(projectRoot string, cfg *config.Config, targets map[string]bool, entriesConverged bool) (*initRefreshFileChange, error) {
	relativePath := filepath.ToSlash(cfg.ConstitutionPath)
	if !initRefreshTargetMatches(targets, relativePath) {
		return nil, nil
	}

	path := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
	if !document.Exists(path) {
		after, _ := upsertConstitutionBaseline(templates.Constitution, templates.ConstitutionBaselineSection)
		return newInitRefreshFileChange(projectRoot, relativePath, "", after, instructionFileCreated), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", relativePath, err)
	}
	before := string(data)
	after := migrateConstitutionSections(before, entriesConverged)
	after = mergeDocumentContent(relativePath, after, templates.Constitution, document.TypeConstitution)
	// The baseline points at the contract, so it moves only once every entry
	// file carries the managed block; otherwise the existing baseline stays.
	if entriesConverged {
		after, _ = upsertConstitutionBaseline(after, templates.ConstitutionBaselineSection)
	}
	result := instructionFileSkipped
	if before != after {
		result = instructionFileMerged
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

	merged := strings.TrimRight(content, "\n")
	for _, section := range missingSections {
		merged += fmt.Sprintf("\n\n## %s\n\n%s", section.Name, strings.TrimSpace(section.Content))
	}
	return merged + "\n"
}

// planRefreshInitInstructionArtifacts creates missing instruction artifacts and
// migrates existing entry files onto the managed contract block. It reports
// whether every entry file carries exactly one current block afterwards.
func planRefreshInitInstructionArtifacts(
	projectRoot string,
	opts initRefreshOptions,
	cfg *config.Config,
	targets map[string]bool,
) ([]initRefreshFileChange, []string, bool, error) {
	var changes []initRefreshFileChange
	var notes []string
	converged := true
	seen := map[string]bool{}
	for _, relativePath := range instructionArtifactPaths(cfg) {
		relativePath = filepath.ToSlash(relativePath)
		// Entry files symlinked to one another are one file; plan it once.
		if resolved, err := filepath.EvalSymlinks(filepath.Join(projectRoot, filepath.FromSlash(relativePath))); err == nil {
			if seen[resolved] {
				continue
			}
			seen[resolved] = true
		}
		plan, err := planInstructionArtifactWrite(projectRoot, relativePath, instructionFileWriteModeConverge, opts.force)
		if err != nil {
			return nil, nil, false, err
		}
		converged = converged && plan.converged
		if !initRefreshTargetMatches(targets, relativePath) {
			continue
		}
		if plan.note != "" {
			notes = append(notes, plan.note)
		}
		change, err := initRefreshChangeFromInstructionPlan(projectRoot, plan)
		if err != nil {
			return nil, nil, false, err
		}
		changes = append(changes, change)
	}
	return changes, notes, converged, nil
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
