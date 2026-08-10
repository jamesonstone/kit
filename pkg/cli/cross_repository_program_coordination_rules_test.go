package cli

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/document"
)

func TestCrossRepositoryProgramCoordinationRegistryRulesetIsValid(t *testing.T) {
	const slug = "cross-repository-program-coordination"
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
		"coding-agent",
		"cross-repository",
		"program",
		"deployment",
		"handoff",
		"resume",
		"dispatch",
	} {
		if !slices.Contains(ruleset.Metadata.AppliesTo, appliesTo) {
			t.Errorf("applies_to = %#v, want %q", ruleset.Metadata.AppliesTo, appliesTo)
		}
	}

	normalizedBody := strings.Join(strings.Fields(ruleset.Body), " ")
	for _, check := range []string{
		"### One Coordinator And One Ledger",
		"docs/programs/<program>/PROGRAM.md",
		"reconstruct the ledger from live repository, GitHub, runtime, and validation state",
		"Participant repositories may point to the ledger. They must not create competing ledgers",
		"### Stable Program Model",
		"the current ready frontier",
		"### Separate State Dimensions",
		"implementation: planned, in progress, blocked, or complete at an exact commit",
		"GitHub delivery: issue, branch, pull request, review/check, and merge state",
		"deployment/runtime: target environment",
		"validation: suites or gates run",
		"### Evidence Contract",
		"### Supervisor And Participant Authority",
		"may edit the coordinator ledger only when the supervisor explicitly assigns the exact coordinator checkout",
		"Dispatch only the current ready frontier",
		"### Checkpoints",
		"planned context compaction",
		"### Resume, Handoff, And Reconciliation",
		"change stale, unavailable, or unverified claims to `unknown` or `unobserved`",
		"### Completion",
	} {
		if !strings.Contains(normalizedBody, check) {
			t.Errorf("expected %s ruleset to contain %q", slug, check)
		}
	}

	index, err := os.ReadFile(filepath.Join("..", "..", "docs", "references", "README.md"))
	if err != nil {
		t.Fatalf("read references index: %v", err)
	}
	if !strings.Contains(string(index), "| `cross-repository-program-coordination` |") {
		t.Fatal("references index does not list cross-repository-program-coordination")
	}
}
