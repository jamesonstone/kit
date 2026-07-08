package cli

import (
	"context"
	"os"
	"testing"
)

func stubRulesetRegistryContent(t *testing.T, contentByCommit map[string]string) {
	t.Helper()
	previous := rulesetRegistryContentFetcher
	t.Cleanup(func() {
		rulesetRegistryContentFetcher = previous
	})
	rulesetRegistryContentFetcher = func(_ context.Context, _ string, commit string, _ string) (string, error) {
		content, ok := contentByCommit[commit]
		if !ok {
			return "", os.ErrNotExist
		}
		return content, nil
	}
}

func resetReconcileFlags(t *testing.T) {
	t.Helper()
	previousOutputOnly := reconcileOutputOnly
	previousAll := reconcileAll
	previousCopy := reconcileCopy
	previousMigrateReferences := reconcileMigrateReferences
	previousMigrateVerification := reconcileMigrateVerification
	t.Cleanup(func() {
		reconcileOutputOnly = previousOutputOnly
		reconcileAll = previousAll
		reconcileCopy = previousCopy
		reconcileMigrateReferences = previousMigrateReferences
		reconcileMigrateVerification = previousMigrateVerification
	})
	reconcileOutputOnly = false
	reconcileAll = false
	reconcileCopy = false
	reconcileMigrateReferences = false
	reconcileMigrateVerification = false
}
