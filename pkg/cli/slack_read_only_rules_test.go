package cli

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/document"
)

func TestSlackReadOnlyRegistryRulesetIsValid(t *testing.T) {
	const slug = "slack-read-only"
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
		"slack", "messaging", "communication", "coding-agent", "automation", "collaboration",
	} {
		if !slices.Contains(ruleset.Metadata.AppliesTo, appliesTo) {
			t.Errorf("applies_to = %#v, want %q", ruleset.Metadata.AppliesTo, appliesTo)
		}
	}

	normalized := strings.Join(strings.Fields(ruleset.Body), " ")
	for _, check := range []string{
		"Treat all Slack access as read-only by default",
		"You may read and search Slack without additional approval",
		"read the entire thread for context",
		"inspect or search the channel containing that thread",
		"Never send a Slack message without the human's explicit approval",
		"Drafting a Slack message is not authorization to send it",
		"draft a response",
		"mean draft only",
		"Ask whether the human authorizes sending that specific message",
		`"send it," "send this," or "yes, send that message."`,
		"Approval is single-use and message-specific",
		"handle this",
		"When uncertain whether the human authorized a Slack write action, do not perform it",
		"adding or removing reactions",
		"changing channel information",
		"Reading, searching, retrieving context, and analyzing Slack content do not require approval",
	} {
		if !strings.Contains(normalized, check) {
			t.Errorf("expected %s ruleset to contain %q", slug, check)
		}
	}
}

func TestSlackReadOnlyIsIntegratedWithRelatedRules(t *testing.T) {
	checks := map[string][]string{
		"docs/references/README.md": {
			"Use `rules/slack-read-only.md`",
			"| `slack-read-only` |",
		},
		"docs/agents/RLM.md": {
			"Load `docs/references/rules/slack-read-only.md` before any Slack write, and when a Slack link, channel, thread, or search is part of the task",
		},
		"docs/CONSTITUTION.md": {
			"Load `docs/references/rules/slack-read-only.md`",
			"Treat Slack as read-only by default",
		},
		"docs/references/workflows/implementation-delivery.md": {
			"slug: slack-read-only",
			"required: false",
		},
		"AGENTS.md": {
			"## Slack: Read-Only by Default, Explicit Approval Required to Send",
			"Drafting a Slack message is not authorization to send it",
		},
		"CLAUDE.md": {
			"## Slack: Read-Only by Default, Explicit Approval Required to Send",
		},
		".github/copilot-instructions.md": {
			"## Slack: Read-Only by Default, Explicit Approval Required to Send",
		},
		"docs/agents/GUARDRAILS.md": {
			"## Slack: Read-Only by Default, Explicit Approval Required to Send",
			"Follow `docs/references/rules/slack-read-only.md` before any Slack write",
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
