package cli

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/document"
)

func TestDeletionSafetyRegistryRulesetIsValid(t *testing.T) {
	const slug = "deletion-safety"
	path := filepath.Join("..", "..", "docs", "references", "rules", slug+".md")
	ruleset, err := parseRulesetFile(path)
	if err != nil {
		t.Fatalf("parseRulesetFile() error = %v", err)
	}
	if issues := validateRulesetDocument(ruleset, slug); len(issues) > 0 {
		t.Fatalf("%s ruleset issues = %#v", slug, issues)
	}
	if ruleset.Metadata.RegistryScope != rulesetRegistryScopeDownstream {
		t.Fatalf("registry_scope = %q, want downstream", ruleset.Metadata.RegistryScope)
	}
	if ruleset.Metadata.ReadPolicyDefault != document.ReferenceReadPolicyMust {
		t.Fatalf("read_policy_default = %q, want must", ruleset.Metadata.ReadPolicyDefault)
	}
	for _, appliesTo := range []string{
		"implementation", "data", "persistence", "filesystem", "identity",
		"api", "ui", "automation", "cleanup", "retention", "migration",
		"operations", "cloud", "infrastructure",
	} {
		if !slices.Contains(ruleset.Metadata.AppliesTo, appliesTo) {
			t.Errorf("applies_to = %#v, want %q", ruleset.Metadata.AppliesTo, appliesTo)
		}
	}

	normalizedBody := strings.Join(strings.Fields(ruleset.Body), " ")
	for _, check := range []string{
		"### Soft Delete",
		"supported and authorized restore path",
		"Task-owned ephemeral scratch that never became authoritative state",
		"If ownership or significance is uncertain, treat the target as covered",
		"### Hard Delete",
		"retention expiry",
		"### Soft Delete Is The Default",
		"Interpret an unqualified request to delete or remove covered state as soft delete",
		"retention deadline makes covered state eligible for review or purge; it is not independent authority",
		"### Separate Hard-Delete Surface",
		"server-side authorization and confirmation enforcement",
		"materialized target IDs or an immutable snapshot or version token",
		"compare the current target set or immutable version with the confirmed snapshot",
		"### Specific Manual Confirmation Before Hard Delete",
		"exact target identities or a bounded selector, current resolved count",
		"why soft delete or continued quarantine is insufficient",
		"ask the human to confirm irreversible deletion of those exact targets",
		"the initial request, even when it asked to delete or destroy",
		"Target-set or version drift, including changed identities with the same count",
		"### Composition With Other Gates",
		"one manual response may satisfy both only when the combined outline",
		"### Implementation And Test Expectations",
		"Hard-delete tests prove the separate privileged path rejects missing",
	} {
		if !strings.Contains(normalizedBody, check) {
			t.Errorf("expected %s ruleset to contain %q", slug, check)
		}
	}
}

func TestDeletionSafetyIsIntegratedWithRelatedRules(t *testing.T) {
	checks := map[string][]string{
		"docs/references/README.md": {
			"Use `rules/deletion-safety.md`",
			"| `deletion-safety` |",
		},
		"docs/references/rules/safety-guardrails.md": {
			"Follow `deletion-safety`",
			"post-outline specific manual confirmation",
		},
		"docs/references/rules/infrastructure-change-approval.md": {
			"Follow `deletion-safety` first",
			"one exact post-outline manual confirmation",
		},
		"docs/references/rules/testing-and-environment-validation.md": {
			"Follow `deletion-safety` for cleanup",
			"pre-run test approval or a generic cleanup policy does not count",
		},
	}
	for path, required := range checks {
		content, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		normalized := strings.Join(strings.Fields(string(content)), " ")
		for _, check := range required {
			if !strings.Contains(normalized, check) {
				t.Errorf("expected %s to contain %q", path, check)
			}
		}
	}
}
