package cli

import (
	"fmt"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
)

const registryArtifactSchemaVersion = 1

type rulesetRegistrySyncResult struct {
	content   string
	state     string
	hash      string
	conflicts []string
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

func registryArtifactForRuleset(item registryRuleset, state string, installedHash string) config.RegistryArtifact {
	return config.RegistryArtifact{
		Kind:          rulesetKind,
		Slug:          item.Slug,
		Path:          rulesetTarget(item.Slug),
		InstalledHash: installedHash,
		State:         state,
	}
}

func recordRulesetRegistryState(cfg *config.Config, item registryRuleset, state string, installedHash string) {
	if cfg == nil {
		return
	}
	cfg.Registry.SchemaVersion = registryArtifactSchemaVersion
	cfg.UpsertRegistryArtifact(registryArtifactForRuleset(item, state, installedHash))
}

func rulesetRegistryState(cfg *config.Config, slug string) (config.RegistryArtifact, bool) {
	if cfg == nil {
		return config.RegistryArtifact{}, false
	}
	return cfg.RegistryArtifact(rulesetKind, slug)
}

// syncRulesetRegistryContent reconciles a local rule with the rule embedded in
// this binary: identical or Kit-written-and-unmodified files take the embedded
// version; locally edited files are preserved unless forced.
func syncRulesetRegistryContent(
	item registryRuleset,
	state config.RegistryArtifact,
	localContent string,
	force bool,
) (rulesetRegistrySyncResult, error) {
	localStatus := rulesetLocalStatus(localContent, item.Metadata.Status)
	embeddedHash := item.NormalizedHash
	if embeddedHash == "" {
		var err error
		if embeddedHash, err = normalizedRulesetContentHash(item.Content, item.Metadata.Status); err != nil {
			return rulesetRegistrySyncResult{}, err
		}
	}
	embedded := func() (rulesetRegistrySyncResult, error) {
		updated, err := setRulesetStatus(item.Content, localStatus)
		if err != nil {
			return rulesetRegistrySyncResult{}, err
		}
		return rulesetRegistrySyncResult{content: updated, state: registryArtifactStateManaged, hash: embeddedHash}, nil
	}
	if force {
		return embedded()
	}
	localHash, err := normalizedRulesetContentHash(localContent, item.Metadata.Status)
	if err != nil {
		return rulesetRegistrySyncResult{
			content:   localContent,
			state:     registryArtifactStateLocalCustom,
			hash:      state.InstalledHash,
			conflicts: []string{fmt.Sprintf("%s has invalid local ruleset content: %v", rulesetTarget(item.Slug), err)},
		}, nil
	}
	switch {
	case localHash == embeddedHash:
		return rulesetRegistrySyncResult{content: localContent, state: registryArtifactStateManaged, hash: embeddedHash}, nil
	case state.State != registryArtifactStateLocalCustom && strings.TrimSpace(state.InstalledHash) != "" && localHash == state.InstalledHash:
		return embedded()
	default:
		return rulesetRegistrySyncResult{
			content:   localContent,
			state:     registryArtifactStateLocalCustom,
			hash:      localHash,
			conflicts: []string{fmt.Sprintf("%s has local custom content; use --force to accept Kit's version", rulesetTarget(item.Slug))},
		}, nil
	}
}
