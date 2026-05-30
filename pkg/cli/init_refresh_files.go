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

func refreshInitScaffoldFiles(projectRoot string, opts initRefreshOptions, targets map[string]bool, stats *initRefreshStats) error {
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

	for _, file := range files {
		if !initRefreshTargetMatches(targets, file.relativePath) {
			continue
		}
		err := refreshInitScaffoldFile(projectRoot, opts, targets, file.relativePath, file.content, file.merge, stats)
		if err != nil {
			return err
		}
	}
	return nil
}

func refreshInitScaffoldFile(
	projectRoot string,
	opts initRefreshOptions,
	targets map[string]bool,
	relativePath string,
	content string,
	merge bool,
	stats *initRefreshStats,
) error {
	path := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
	exists := document.Exists(path)
	explicit := len(targets) > 0

	if exists && opts.force && explicit {
		if err := document.Write(path, content); err != nil {
			return fmt.Errorf("failed to overwrite %s: %w", relativePath, err)
		}
		stats.updated++
		return nil
	}
	if exists && merge {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", relativePath, err)
		}
		missing := missingGitignorePatterns(string(data))
		if len(missing) == 0 {
			stats.skipped++
			return nil
		}
		if err := document.Write(path, appendGitignorePatterns(string(data), missing)); err != nil {
			return fmt.Errorf("failed to update %s: %w", relativePath, err)
		}
		stats.merged++
		return nil
	}
	if exists {
		stats.skipped++
		return nil
	}
	if err := document.Write(path, content); err != nil {
		return fmt.Errorf("failed to create %s: %w", relativePath, err)
	}
	stats.created++
	return nil
}

func refreshInitConstitution(projectRoot string, cfg *config.Config, targets map[string]bool, stats *initRefreshStats) error {
	relativePath := filepath.ToSlash(cfg.ConstitutionPath)
	if !initRefreshTargetMatches(targets, relativePath) {
		return nil
	}

	path := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
	created := false
	if !document.Exists(path) {
		if err := document.Write(path, templates.Constitution); err != nil {
			return fmt.Errorf("failed to create %s: %w", relativePath, err)
		}
		stats.created++
		created = true
	}

	before, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", relativePath, err)
	}
	if err := document.MergeDocument(path, templates.Constitution, document.TypeConstitution); err != nil {
		return fmt.Errorf("failed to merge %s: %w", relativePath, err)
	}
	merged, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", relativePath, err)
	}
	updated, changed := upsertConstitutionBaseline(string(merged))
	if changed {
		if err := document.Write(path, updated); err != nil {
			return fmt.Errorf("failed to update %s: %w", relativePath, err)
		}
		stats.merged++
		return nil
	}
	if string(before) != string(merged) {
		stats.merged++
		return nil
	}
	if !created {
		stats.skipped++
	}
	return nil
}

func refreshInitInstructionArtifacts(
	projectRoot string,
	opts initRefreshOptions,
	cfg *config.Config,
	targets map[string]bool,
	stats *initRefreshStats,
) error {
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
			return err
		}
		result, err := applyInstructionFileWritePlan(plan)
		if err != nil {
			return err
		}
		stats.recordInstructionResult(result)
	}
	return nil
}

func refreshInitRulesets(
	projectRoot string,
	opts initRefreshOptions,
	targets map[string]bool,
	registry []registryRuleset,
	stats *initRefreshStats,
) error {
	for _, item := range registry {
		relativePath := rulesetTarget(item.Slug)
		if !initRefreshTargetMatches(targets, relativePath) {
			continue
		}
		path := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
		exists := document.Exists(path)
		if exists && !opts.force {
			stats.skipped++
			continue
		}
		if err := document.Write(path, item.Content); err != nil {
			return fmt.Errorf("failed to write %s: %w", relativePath, err)
		}
		if exists {
			stats.updated++
		} else {
			stats.created++
		}
	}
	return nil
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
