package cli

import (
	"context"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/templates"
)

func stubRulesetRegistry(t *testing.T, rulesets ...registryRuleset) {
	t.Helper()
	previous := rulesetRegistryFetcher
	t.Cleanup(func() {
		rulesetRegistryFetcher = previous
	})
	rulesetRegistryFetcher = func(_ context.Context) ([]registryRuleset, error) {
		return rulesets, nil
	}
}

func registryRulesetForTest(slug string, appliesTo []string) registryRuleset {
	content := templates.BuildRulesetWithOptions(templates.RulesetOptions{
		Slug:              slug,
		Description:       "Description for " + slug,
		AppliesTo:         appliesTo,
		ReadPolicyDefault: "conditional",
	})
	return registryRulesetWithContentForTest(slug, content, "test-"+slug+"-commit")
}

func registryRulesetWithContentForTest(slug, content, commit string) registryRuleset {
	parsed := parseRuleset(content, slug+".md")
	hash, err := normalizedRulesetContentHash(content, parsed.Metadata.Status)
	if err != nil {
		panic(err)
	}
	return registryRuleset{
		Slug:           slug,
		Content:        content,
		Metadata:       parsed.Metadata,
		NormalizedHash: hash,
	}
}

func resetReconcileFlags(t *testing.T) {
	t.Helper()
	previousOutputOnly := reconcileOutputOnly
	previousAll := reconcileAll
	previousCopy := reconcileCopy
	previousMigrateReferences := reconcileMigrateReferences
	previousMigrateVerification := reconcileMigrateVerification
	previousIncludeFiles := reconcileIncludeFiles
	previousForce := reconcileForce
	previousDryRun := reconcileDryRun
	previousDiff := reconcileDiff
	previousRefreshFiles := reconcileRefreshFiles
	t.Cleanup(func() {
		reconcileOutputOnly = previousOutputOnly
		reconcileAll = previousAll
		reconcileCopy = previousCopy
		reconcileMigrateReferences = previousMigrateReferences
		reconcileMigrateVerification = previousMigrateVerification
		reconcileIncludeFiles = previousIncludeFiles
		reconcileForce = previousForce
		reconcileDryRun = previousDryRun
		reconcileDiff = previousDiff
		reconcileRefreshFiles = previousRefreshFiles
	})
	reconcileOutputOnly = false
	reconcileAll = false
	reconcileCopy = false
	reconcileMigrateReferences = false
	reconcileMigrateVerification = false
	reconcileIncludeFiles = false
	reconcileForce = false
	reconcileDryRun = false
	reconcileDiff = false
	reconcileRefreshFiles = nil
}
