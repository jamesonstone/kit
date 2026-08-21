package cli

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/document"
)

func TestHumanAuthorshipRegistryRulesetIsValid(t *testing.T) {
	const slug = "human-authorship"
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
	if ruleset.Metadata.ReadPolicyDefault != document.ReferenceReadPolicyConditional {
		t.Fatalf("read_policy_default = %q, want conditional", ruleset.Metadata.ReadPolicyDefault)
	}
	for _, appliesTo := range []string{"git", "github", "commit", "pull-request", "issue", "attribution"} {
		if !slices.Contains(ruleset.Metadata.AppliesTo, appliesTo) {
			t.Errorf("applies_to = %#v, want %q", ruleset.Metadata.AppliesTo, appliesTo)
		}
	}

	normalized := strings.Join(strings.Fields(ruleset.Body), " ")
	for _, check := range []string{
		"Keep displayed authorship human-only",
		"Do not load it for ordinary implementation",
		"The human user is the only git author",
		"Never insert agent or tool attribution",
		"Co-authored-by:",
		"Generated with …",
		"a human DCO trailer for the user is allowed",
		"Inlining this ruleset into",
	} {
		if !strings.Contains(normalized, check) {
			t.Errorf("expected %s ruleset to contain %q", slug, check)
		}
	}
}

func TestHumanAuthorshipIsIntegratedWithRelatedRules(t *testing.T) {
	checks := map[string][]string{
		"docs/references/README.md": {
			"Use `rules/human-authorship.md`",
			"| `human-authorship` |",
		},
		"docs/references/rules/github-pr-delivery.md": {
			"Load `docs/references/rules/human-authorship.md` before any commit, pull request, issue, comment, or other attribution text",
		},
		"docs/references/rules/safety-guardrails.md": {
			"Load `docs/references/rules/human-authorship.md` before any commit, pull request, issue, comment, or other attribution text",
		},
		"docs/agents/RLM.md": {
			"Load `docs/references/rules/human-authorship.md` before any commit, pull request, issue, review comment, or other attribution text",
		},
		"docs/CONSTITUTION.md": {
			"load `docs/references/rules/human-authorship.md`",
		},
		"docs/references/workflows/implementation-delivery.md": {
			"slug: human-authorship",
			"required: false",
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
