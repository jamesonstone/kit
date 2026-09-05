package cli

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/document"
)

func TestInfrastructureChangeApprovalRegistryRulesetIsValid(t *testing.T) {
	const slug = "infrastructure-change-approval"
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
		"cloud",
		"infrastructure-as-code",
		"aws",
		"gcp",
		"azure",
		"kubernetes",
		"terraform",
		"pulumi",
		"cloudformation",
	} {
		if !slices.Contains(ruleset.Metadata.AppliesTo, appliesTo) {
			t.Errorf("applies_to = %#v, want %q", ruleset.Metadata.AppliesTo, appliesTo)
		}
	}

	normalizedBody := strings.Join(strings.Fields(ruleset.Body), " ")
	for _, check := range []string{
		"### Read-Only Discovery",
		"Discovery must not alter cloud resources, Kubernetes objects, remote state",
		"### Consolidated Change Outline",
		"target identity: provider, account, project or subscription",
		"material impact and risk: availability, data, security or IAM, cost",
		"rollback or recovery",
		"include the complete infrastructure outline in that plan instead of creating a separate approval ceremony",
		"### Name-Aware Material AWS Targets",
		"account name (account ID)",
		"Region long name (Region code)",
		"display name unavailable",
		"Do not create a separate identity prompt or approval ceremony",
		"### One Confirmation And One-Pass Execution",
		"Obtain one explicit user confirmation of the complete outline",
		"Standing merge/deploy authority never authorizes a covered mutation",
		"### Standard Deployments Under Standing Authority",
		"Generic task acceptance does not authorize deployment",
		"It excludes novel provider commands, new targets, workflow mutation, IAM",
		"Only an explicit human resume or replacement grant restores authority",
		"remaining task work to completion in one pass without asking for command-by-command approval",
		"### Deletion And Removal Exception",
		"always requires explicit user confirmation after the consolidated outline",
		"One confirmation covers every deletion or removal named in the batch",
		"### Follow-Up Batches And Material Deviations",
		"collect all then-known changes into one follow-up outline",
		"Do not re-confirm actions already included in an approved batch",
		"### Routine Application Operations",
		"already-provisioned",
		"are not infrastructure-approval batches",
		"never authorize deletion or removal",
	} {
		if !strings.Contains(normalizedBody, check) {
			t.Errorf("expected %s ruleset to contain %q", slug, check)
		}
	}

	index, err := os.ReadFile(filepath.Join("..", "..", "docs", "references", "README.md"))
	if err != nil {
		t.Fatalf("read references index: %v", err)
	}
	if !strings.Contains(string(index), "| `infrastructure-change-approval` |") {
		t.Fatal("references index does not list infrastructure-change-approval")
	}
}
