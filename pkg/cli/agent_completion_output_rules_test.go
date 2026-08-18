package cli

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/document"
)

func TestAgentCompletionOutputRegistryRulesetIsValid(t *testing.T) {
	const slug = "agent-completion-output"
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
		"completion", "implementation", "research", "diagnosis", "planning",
		"validation", "testing", "review", "operations", "deployment",
		"monitoring", "coordination", "handoff",
	} {
		if !slices.Contains(ruleset.Metadata.AppliesTo, appliesTo) {
			t.Errorf("applies_to = %#v, want %q", ruleset.Metadata.AppliesTo, appliesTo)
		}
	}

	normalized := strings.Join(strings.Fields(ruleset.Body), " ")
	for _, check := range []string{
		"# <PASS|PARTIAL|BLOCKED|FAIL> — <one-sentence outcome>",
		"Type | Action required | Why | Continue with",
		"Order rows as `Blocker`, `Incomplete`, `Next`, `Optional`, then `None`",
		"every PASS response includes one `None` row",
		"copy-ready prompt or command",
		"### Implementation And Delivery",
		"### Research And Discovery",
		"### Diagnosis And Troubleshooting",
		"### Planning And Design",
		"### Validation And Testing",
		"### Review And Audit",
		"### Operations, Deployment, And Monitoring",
		"### Coordination And Handoff",
		"### Fallback",
		"higher-priority host wrapper",
		"`PENDING`, `UNKNOWN`, `SKIPPED`",
	} {
		if !strings.Contains(normalized, check) {
			t.Errorf("expected %s ruleset to contain %q", slug, check)
		}
	}
}

func TestAgentCompletionOutputIsIntegratedWithRelatedRules(t *testing.T) {
	checks := map[string][]string{
		"docs/references/README.md": {
			"Use `rules/agent-completion-output.md`",
			"| `agent-completion-output` |",
		},
		"docs/references/rules/github-pr-delivery.md": {
			"Follow `agent-completion-output` with the implementation/delivery profile",
		},
		"docs/references/rules/testing-and-environment-validation.md": {
			"Follow `agent-completion-output` for terminal reporting",
		},
		"docs/references/rules/agent-team-orchestration.md": {
			"Map `task_outcome` to the overall status heading",
		},
		"docs/references/rules/cross-repository-program-coordination.md": {
			"`agent-completion-output` coordination/handoff profile",
		},
		"docs/references/rules/constitution-curation.md": {
			"`agent-completion-output` Repository Memory table",
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
