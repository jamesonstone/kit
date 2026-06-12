package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
)

const registryArtifactSchemaVersion = 1

var markdownHeadingLinePattern = regexp.MustCompile(`^(#{1,6})\s+(.+?)\s*$`)

type rulesetRegistrySyncResult struct {
	content   string
	state     string
	hash      string
	conflicts []string
}

type markdownSegment struct {
	key string
	raw string
}

type markdownHeadingMatch struct {
	start int
	level int
	name  string
}

type markdownHeadingPathEntry struct {
	level int
	key   string
}

func rulesetRegistryRepoFullName() string {
	return rulesetRegistryOwner + "/" + rulesetRegistryRepo
}

func normalizedRulesetContentHash(content, registryStatus string) (string, error) {
	normalized, err := normalizeRulesetContentForRegistry(content, registryStatus)
	if err != nil {
		return "", err
	}
	return contentHash(normalized), nil
}

func normalizeRulesetContentForRegistry(content, registryStatus string) (string, error) {
	normalized, err := setRulesetStatus(content, registryStatus)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(normalized) + "\n", nil
}

func rulesetLocalStatus(content, fallback string) string {
	parsed := parseRuleset(content, "")
	if parsed.ParseErr == nil && validRulesetStatus(parsed.Metadata.Status) {
		return parsed.Metadata.Status
	}
	if validRulesetStatus(fallback) {
		return fallback
	}
	return rulesetReferenceStatus
}

func registryArtifactForRuleset(item registryRuleset, state string, installedHash string, content string) config.RegistryArtifact {
	artifact := config.RegistryArtifact{
		Kind:          rulesetKind,
		Slug:          item.Slug,
		Path:          rulesetTarget(item.Slug),
		SourceRepo:    firstNonEmpty(item.SourceRepo, rulesetRegistryRepoFullName()),
		SourceBranch:  firstNonEmpty(item.SourceBranch, rulesetRegistryBranch),
		SourceCommit:  item.SourceCommit,
		SourcePath:    firstNonEmpty(item.SourcePath, rulesetTarget(item.Slug)),
		InstalledHash: installedHash,
		State:         state,
	}
	if state == registryArtifactStateManaged && installedHash != "" && installedHash != item.NormalizedHash {
		artifact.Sections = rulesetRegistrySectionArtifacts(content, item.Metadata.Status)
	}
	return artifact
}

func recordRulesetRegistryState(cfg *config.Config, item registryRuleset, state string, installedHash string, content string) {
	if cfg == nil {
		return
	}
	cfg.Registry.SchemaVersion = registryArtifactSchemaVersion
	cfg.Registry.Source = config.RegistrySource{
		Repo:   rulesetRegistryRepoFullName(),
		Branch: rulesetRegistryBranch,
	}
	cfg.UpsertRegistryArtifact(registryArtifactForRuleset(item, state, installedHash, content))
}

func rulesetRegistryState(cfg *config.Config, slug string) (config.RegistryArtifact, bool) {
	if cfg == nil {
		return config.RegistryArtifact{}, false
	}
	return cfg.RegistryArtifact(rulesetKind, slug)
}

func syncRulesetRegistryContent(
	ctx context.Context,
	item registryRuleset,
	state config.RegistryArtifact,
	localContent string,
	force bool,
) (rulesetRegistrySyncResult, error) {
	localStatus := rulesetLocalStatus(localContent, item.Metadata.Status)
	remoteHash := item.NormalizedHash
	if remoteHash == "" {
		var err error
		remoteHash, err = normalizedRulesetContentHash(item.Content, item.Metadata.Status)
		if err != nil {
			return rulesetRegistrySyncResult{}, err
		}
	}
	localHash, localHashErr := normalizedRulesetContentHash(localContent, item.Metadata.Status)
	if force {
		updated, err := setRulesetStatus(item.Content, localStatus)
		if err != nil {
			return rulesetRegistrySyncResult{}, err
		}
		return rulesetRegistrySyncResult{
			content: updated,
			state:   registryArtifactStateManaged,
			hash:    remoteHash,
		}, nil
	}
	if localHashErr != nil {
		return rulesetRegistrySyncResult{
			content:   localContent,
			state:     registryArtifactStateLocalCustom,
			hash:      state.InstalledHash,
			conflicts: []string{fmt.Sprintf("%s has invalid local ruleset content: %v", rulesetTarget(item.Slug), localHashErr)},
		}, nil
	}
	if localHash == remoteHash {
		return rulesetRegistrySyncResult{
			content: localContent,
			state:   registryArtifactStateManaged,
			hash:    remoteHash,
		}, nil
	}
	if state.State == registryArtifactStateLocalCustom || strings.TrimSpace(state.InstalledHash) == "" {
		return rulesetRegistrySyncResult{
			content:   localContent,
			state:     registryArtifactStateLocalCustom,
			hash:      localHash,
			conflicts: []string{fmt.Sprintf("%s has local custom content; use --force to accept registry content", rulesetTarget(item.Slug))},
		}, nil
	}
	if state.SourceCommit != "" && item.SourceCommit != "" && state.SourceCommit == item.SourceCommit {
		if localHash == state.InstalledHash {
			return rulesetRegistrySyncResult{
				content: localContent,
				state:   registryArtifactStateManaged,
				hash:    state.InstalledHash,
			}, nil
		}
		return rulesetRegistrySyncResult{
			content:   localContent,
			state:     registryArtifactStateLocalCustom,
			hash:      state.InstalledHash,
			conflicts: []string{fmt.Sprintf("%s has local custom content; use --force to accept registry content", rulesetTarget(item.Slug))},
		}, nil
	}
	if len(state.Sections) > 0 {
		merged, conflicts, err := mergeRulesetSectionsFromState(item, state, localContent, localStatus)
		if err != nil {
			return rulesetRegistrySyncResult{}, err
		}
		if len(conflicts) > 0 {
			return rulesetRegistrySyncResult{
				content:   localContent,
				state:     registryArtifactStateConflict,
				hash:      state.InstalledHash,
				conflicts: conflicts,
			}, nil
		}
		mergedHash, err := normalizedRulesetContentHash(merged, item.Metadata.Status)
		if err != nil {
			return rulesetRegistrySyncResult{}, err
		}
		return rulesetRegistrySyncResult{
			content: merged,
			state:   registryArtifactStateManaged,
			hash:    mergedHash,
		}, nil
	}
	if localHash == state.InstalledHash {
		updated, err := setRulesetStatus(item.Content, localStatus)
		if err != nil {
			return rulesetRegistrySyncResult{}, err
		}
		return rulesetRegistrySyncResult{
			content: updated,
			state:   registryArtifactStateManaged,
			hash:    remoteHash,
		}, nil
	}
	if remoteHash == state.InstalledHash {
		return rulesetRegistrySyncResult{
			content: localContent,
			state:   registryArtifactStateLocalCustom,
			hash:    state.InstalledHash,
		}, nil
	}

	baseContent, err := rulesetRegistryContentFetcher(ctx, firstNonEmpty(state.SourceRepo, item.SourceRepo), state.SourceCommit, firstNonEmpty(state.SourcePath, item.SourcePath))
	if err != nil {
		return rulesetRegistrySyncResult{
			content:   localContent,
			state:     registryArtifactStateConflict,
			hash:      state.InstalledHash,
			conflicts: []string{fmt.Sprintf("%s cannot fetch registry base %s: %v", rulesetTarget(item.Slug), state.SourceCommit, err)},
		}, nil
	}
	baseHash, err := normalizedRulesetContentHash(baseContent, item.Metadata.Status)
	if err != nil {
		return rulesetRegistrySyncResult{}, err
	}
	if baseHash != state.InstalledHash {
		return rulesetRegistrySyncResult{
			content:   localContent,
			state:     registryArtifactStateConflict,
			hash:      state.InstalledHash,
			conflicts: []string{fmt.Sprintf("%s registry base hash mismatch for %s", rulesetTarget(item.Slug), state.SourceCommit)},
		}, nil
	}
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

func contentHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func markdownSegments(content string) []markdownSegment {
	matches := markdownHeadingMatches(content)
	if len(matches) == 0 {
		return []markdownSegment{{key: "preamble", raw: content}}
	}
	segments := make([]markdownSegment, 0, len(matches)+1)
	if matches[0].start > 0 {
		segments = append(segments, markdownSegment{key: "preamble", raw: content[:matches[0].start]})
	}
	var path []markdownHeadingPathEntry
	for i, match := range matches {
		start := match.start
		end := len(content)
		if i+1 < len(matches) {
			end = matches[i+1].start
		}
		key := markdownHeadingPathKey(match, &path)
		segments = append(segments, markdownSegment{key: key, raw: content[start:end]})
	}
	return segments
}

func markdownHeadingPathKey(match markdownHeadingMatch, path *[]markdownHeadingPathEntry) string {
	current := markdownHeadingSegmentKey(match.level, match.name)
	for len(*path) > 0 && (*path)[len(*path)-1].level >= match.level {
		*path = (*path)[:len(*path)-1]
	}
	parts := make([]string, 0, len(*path)+1)
	for _, entry := range *path {
		parts = append(parts, entry.key)
	}
	parts = append(parts, current)
	*path = append(*path, markdownHeadingPathEntry{level: match.level, key: current})
	return strings.Join(parts, " > ")
}

func markdownHeadingSegmentKey(level int, name string) string {
	return strings.Repeat("#", level) + " " + strings.ToLower(strings.TrimSpace(name))
}

func markdownHeadingMatches(content string) []markdownHeadingMatch {
	lines := strings.SplitAfter(content, "\n")
	var matches []markdownHeadingMatch
	offset := 0
	inFence := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			offset += len(line)
			continue
		}
		if !inFence {
			withoutNewline := strings.TrimRight(line, "\r\n")
			parts := markdownHeadingLinePattern.FindStringSubmatch(withoutNewline)
			if len(parts) == 3 {
				matches = append(matches, markdownHeadingMatch{
					start: offset,
					level: len(parts[1]),
					name:  strings.TrimSpace(parts[2]),
				})
			}
		}
		offset += len(line)
	}
	return matches
}

func markdownSegmentMap(segments []markdownSegment) map[string]string {
	out := make(map[string]string, len(segments))
	for _, segment := range segments {
		out[segment.key] = segment.raw
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
