package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
)

func syncRulesetRegistryContentFromBase(
	item registryRuleset,
	state config.RegistryArtifact,
	baseContent string,
	localContent string,
	localStatus string,
) (rulesetRegistrySyncResult, error) {
	baseNormalized, err := normalizeRulesetContentForRegistry(baseContent, item.Metadata.Status)
	if err != nil {
		return rulesetRegistrySyncResult{}, err
	}
	localNormalized, err := normalizeRulesetContentForRegistry(localContent, item.Metadata.Status)
	if err != nil {
		return rulesetRegistrySyncResult{}, err
	}
	remoteNormalized, err := normalizeRulesetContentForRegistry(item.Content, item.Metadata.Status)
	if err != nil {
		return rulesetRegistrySyncResult{}, err
	}
	merged, conflicts := mergeRulesetSections(baseNormalized, localNormalized, remoteNormalized)
	if len(conflicts) > 0 {
		for i := range conflicts {
			conflicts[i] = fmt.Sprintf("%s section %s changed locally and in registry", rulesetTarget(item.Slug), conflicts[i])
		}
		return rulesetRegistrySyncResult{
			content:   localContent,
			state:     registryArtifactStateConflict,
			hash:      state.InstalledHash,
			conflicts: conflicts,
		}, nil
	}
	updated, err := setRulesetStatus(merged, localStatus)
	if err != nil {
		return rulesetRegistrySyncResult{}, err
	}
	mergedHash, err := normalizedRulesetContentHash(updated, item.Metadata.Status)
	if err != nil {
		return rulesetRegistrySyncResult{}, err
	}
	return rulesetRegistrySyncResult{
		content: updated,
		state:   registryArtifactStateManaged,
		hash:    mergedHash,
	}, nil
}

func mergeRulesetSectionsFromState(item registryRuleset, state config.RegistryArtifact, localContent string, localStatus string) (string, []string, error) {
	localNormalized, err := normalizeRulesetContentForRegistry(localContent, item.Metadata.Status)
	if err != nil {
		return "", nil, err
	}
	remoteNormalized, err := normalizeRulesetContentForRegistry(item.Content, item.Metadata.Status)
	if err != nil {
		return "", nil, err
	}
	merged, conflicts := mergeRulesetSectionsByHash(state, localNormalized, remoteNormalized, item.Slug)
	if len(conflicts) > 0 {
		return "", conflicts, nil
	}
	updated, err := setRulesetStatus(merged, localStatus)
	if err != nil {
		return "", nil, err
	}
	return updated, nil, nil
}

func mergeRulesetSections(baseContent, localContent, remoteContent string) (string, []string) {
	base := markdownSegments(baseContent)
	local := markdownSegments(localContent)
	remote := markdownSegments(remoteContent)

	baseByKey := markdownSegmentMap(base)
	localByKey := markdownSegmentMap(local)
	remoteByKey := markdownSegmentMap(remote)
	seen := make(map[string]bool)
	var conflicts []string
	var builder strings.Builder

	appendMerged := func(key string) {
		if seen[key] {
			return
		}
		seen[key] = true
		baseRaw := baseByKey[key]
		localRaw := localByKey[key]
		remoteRaw := remoteByKey[key]
		switch {
		case localRaw == baseRaw:
			builder.WriteString(remoteRaw)
		case remoteRaw == baseRaw:
			builder.WriteString(localRaw)
		case localRaw == remoteRaw:
			builder.WriteString(localRaw)
		default:
			conflicts = append(conflicts, key)
		}
	}

	for _, segment := range remote {
		appendMerged(segment.key)
	}
	for _, segment := range local {
		appendMerged(segment.key)
	}
	if len(conflicts) > 0 {
		sort.Strings(conflicts)
		return "", conflicts
	}
	return builder.String(), nil
}

func mergeRulesetSectionsByHash(state config.RegistryArtifact, localContent, remoteContent, slug string) (string, []string) {
	local := markdownSegments(localContent)
	remote := markdownSegments(remoteContent)
	localByKey := markdownSegmentMap(local)
	remoteByKey := markdownSegmentMap(remote)
	baseHashes := registrySectionHashMap(state.Sections)

	seen := make(map[string]bool)
	var conflicts []string
	var builder strings.Builder
	appendMerged := func(key string) {
		if seen[key] {
			return
		}
		seen[key] = true
		baseHash := baseHashes[key]
		localRaw, localExists := localByKey[key]
		remoteRaw, remoteExists := remoteByKey[key]
		localHash := ""
		if localExists {
			localHash = contentHash(localRaw)
		}
		remoteHash := ""
		if remoteExists {
			remoteHash = contentHash(remoteRaw)
		}
		localChanged := localExists && localHash != baseHash
		remoteChanged := remoteExists && remoteHash != baseHash
		switch {
		case !localExists && !remoteExists:
			return
		case !localChanged:
			builder.WriteString(remoteRaw)
		case !remoteChanged:
			builder.WriteString(localRaw)
		case localRaw == remoteRaw:
			builder.WriteString(localRaw)
		default:
			conflicts = append(conflicts, fmt.Sprintf("%s section %s changed locally and in registry", rulesetTarget(slug), key))
		}
	}

	for _, segment := range remote {
		appendMerged(segment.key)
	}
	for _, segment := range local {
		appendMerged(segment.key)
	}
	if len(conflicts) > 0 {
		sort.Strings(conflicts)
		return "", conflicts
	}
	return builder.String(), nil
}

func rulesetRegistrySectionArtifacts(content, registryStatus string) []config.RegistryArtifactSection {
	normalized, err := normalizeRulesetContentForRegistry(content, registryStatus)
	if err != nil {
		return nil
	}
	segments := markdownSegments(normalized)
	sections := make([]config.RegistryArtifactSection, 0, len(segments))
	for _, segment := range segments {
		sections = append(sections, config.RegistryArtifactSection{
			Key:           segment.key,
			InstalledHash: contentHash(segment.raw),
		})
	}
	return sections
}

func registrySectionHashMap(sections []config.RegistryArtifactSection) map[string]string {
	out := make(map[string]string, len(sections))
	for _, section := range sections {
		out[section.Key] = section.InstalledHash
	}
	return out
}
