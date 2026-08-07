package cli

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/document"
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
		"Obtain explicit user confirmation",
		"initial request counts as confirmation only when it contains the complete required outline",
		"execute the approved implementation, application, validation, and routine failure recovery to completion",
		"### Material Deviations",
		"Do not split a known batch into repeated approval prompts",
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
