package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
)

const registryArtifactSchemaVersion = 1

type rulesetRegistrySyncResult struct {
	content   string
	state     string
	hash      string
	conflicts []string
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
		return syncRulesetRegistryContentFromSections(item, state, localContent, localStatus)
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

	return syncRulesetRegistryContentFromFetchedBase(ctx, item, state, localContent, localStatus)
}

func syncRulesetRegistryContentFromSections(
	item registryRuleset,
	state config.RegistryArtifact,
	localContent string,
	localStatus string,
) (rulesetRegistrySyncResult, error) {
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

func syncRulesetRegistryContentFromFetchedBase(
	ctx context.Context,
	item registryRuleset,
	state config.RegistryArtifact,
	localContent string,
	localStatus string,
) (rulesetRegistrySyncResult, error) {
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
	return syncRulesetRegistryContentFromBase(item, state, baseContent, localContent, localStatus)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
