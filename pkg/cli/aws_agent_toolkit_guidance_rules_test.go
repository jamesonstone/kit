package cli

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/document"
)

func TestAWSAgentToolkitGuidanceRulesetIsValid(t *testing.T) {
	const slug = "aws-agent-toolkit-guidance"
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
		"aws",
		"aws-cli",
		"aws-mcp",
		"agent-toolkit",
		"infrastructure",
		"documentation",
		"secrets",
	} {
		if !slices.Contains(ruleset.Metadata.AppliesTo, appliesTo) {
			t.Errorf("applies_to = %#v, want %q", ruleset.Metadata.AppliesTo, appliesTo)
		}
	}

	normalizedBody := strings.Join(strings.Fields(ruleset.Body), " ")
	for _, check := range []string{
		"setup-instructions/setup.md",
		"rules/aws-agent-rules.md",
		"load the most relevant AWS Agent Toolkit skill",
		"Do not assume the host exposes a tool literally named `retrieve_skill`",
		"current official AWS documentation",
		"Prefer the AWS MCP Server",
		"Use the AWS CLI directly",
		"the user explicitly requires the CLI",
		"run `kit aws verify` before the first AWS-dependent command",
		"After Kit verification fails, never fall back",
		"infrastructure-change-approval.md",
		"Prefer AWS CDK or CloudFormation",
		"AWS Well-Architected Framework",
		"load the current Agent Toolkit `aws-secrets-manager` skill",
		"Never call `secretsmanager get-secret-value`",
		"`asm-exec`",
	} {
		if !strings.Contains(normalizedBody, check) {
			t.Errorf("expected %s ruleset to contain %q", slug, check)
		}
	}

	index, err := os.ReadFile(filepath.Join("..", "..", "docs", "references", "README.md"))
	if err != nil {
		t.Fatalf("read references index: %v", err)
	}
	if !strings.Contains(string(index), "| `aws-agent-toolkit-guidance` |") {
		t.Fatal("references index does not list aws-agent-toolkit-guidance")
	}
}
