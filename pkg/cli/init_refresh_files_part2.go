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
