package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/templates"
)

func planRefreshInitScaffoldFiles(projectRoot string, opts initRefreshOptions, targets map[string]bool) ([]initRefreshFileChange, error) {
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

func planRefreshInitRulesets(
	ctx context.Context,
	projectRoot string,
	opts initRefreshOptions,
	cfg *config.Config,
	targets map[string]bool,
	registry []registryRuleset,
) ([]initRefreshFileChange, []string, bool, error) {
	var changes []initRefreshFileChange
	var notes []string
	registryChanged := false
	for _, item := range registry {
		relativePath := rulesetTarget(item.Slug)
		if !initRefreshTargetMatches(targets, relativePath) {
			continue
		}
		path := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
		exists := document.Exists(path)
		before := ""
		if exists {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, nil, false, fmt.Errorf("failed to read %s: %w", relativePath, err)
			}
			before = string(data)
		}
		if !exists {
			recordRulesetRegistryState(cfg, item, registryArtifactStateManaged, item.NormalizedHash, item.Content)
			registryChanged = true
			changes = append(changes, *newInitRefreshFileChange(projectRoot, relativePath, before, item.Content, instructionFileCreated))
			continue
		}

		state, tracked := rulesetRegistryState(cfg, item.Slug)
		if !tracked {
			localHash, err := normalizedRulesetContentHash(before, item.Metadata.Status)
			if err == nil && localHash == item.NormalizedHash {
				recordRulesetRegistryState(cfg, item, registryArtifactStateManaged, item.NormalizedHash, before)
				registryChanged = true
				changes = append(changes, *newInitRefreshFileChange(projectRoot, relativePath, before, before, instructionFileSkipped))
				continue
			}
			if !opts.force {
				hash := localHash
				if err != nil {
					hash = ""
				}
				recordRulesetRegistryState(cfg, item, registryArtifactStateLocalCustom, hash, before)
				registryChanged = true
				notes = append(notes, fmt.Sprintf("%s has local custom content; use --force to accept registry content", relativePath))
				changes = append(changes, *newInitRefreshFileChange(projectRoot, relativePath, before, before, instructionFileSkipped))
				continue
			}
		}

		syncResult, err := syncRulesetRegistryContent(ctx, item, state, before, opts.force)
		if err != nil {
			return nil, nil, false, fmt.Errorf("failed to sync %s: %w", relativePath, err)
		}
		recordRulesetRegistryState(cfg, item, syncResult.state, syncResult.hash, syncResult.content)
		registryChanged = true
		if len(syncResult.conflicts) > 0 {
			notes = append(notes, syncResult.conflicts...)
			changes = append(changes, *newInitRefreshFileChange(projectRoot, relativePath, before, before, instructionFileSkipped))
			continue
		}
		if before == syncResult.content {
			changes = append(changes, *newInitRefreshFileChange(projectRoot, relativePath, before, before, instructionFileSkipped))
			continue
		}
		changes = append(changes, *newInitRefreshFileChange(projectRoot, relativePath, before, syncResult.content, instructionFileUpdated))
	}
	return changes, notes, registryChanged, nil
}

func exactLegacyInstructionArtifact(projectRoot, relativePath string) bool {
	path := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(content)) == strings.TrimSpace(templates.InstructionFileForVersion(relativePath, config.InstructionScaffoldVersionVerbose))
}

func upsertConstitutionBaseline(content string) (string, bool) {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "### "+constitutionBaselineHeading {
			start = i
			break
		}
	}

	baselineLines := strings.Split(constitutionBaselineSection, "\n")
	if start >= 0 {
		end := len(lines)
		for i := start + 1; i < len(lines); i++ {
			trimmed := strings.TrimSpace(lines[i])
			if strings.HasPrefix(trimmed, "### ") || strings.HasPrefix(trimmed, "## ") {
				end = i
				break
			}
		}
		updatedLines := append([]string{}, lines[:start]...)
		updatedLines = append(updatedLines, baselineLines...)
		updatedLines = append(updatedLines, lines[end:]...)
		updated := strings.Join(updatedLines, "\n") + "\n"
		return updated, updated != content
	}

	constraints := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "## CONSTRAINTS" {
			constraints = i
			break
		}
	}
	if constraints == -1 {
		updated := strings.TrimRight(content, "\n") + "\n\n## CONSTRAINTS\n\n" + constitutionBaselineSection + "\n"
		return updated, true
	}

	insertAt := constraints + 1
	for insertAt < len(lines) && strings.TrimSpace(lines[insertAt]) == "" {
		insertAt++
	}

	updatedLines := append([]string{}, lines[:insertAt]...)
	updatedLines = append(updatedLines, "")
	updatedLines = append(updatedLines, baselineLines...)
	updatedLines = append(updatedLines, "")
	updatedLines = append(updatedLines, lines[insertAt:]...)
	return strings.Join(updatedLines, "\n") + "\n", true
}
