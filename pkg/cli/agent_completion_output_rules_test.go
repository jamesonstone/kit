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
		"## Proportionality Gate",
		"### Conversational Responses",
		"### Structured Handoff Triggers",
		"direct question, definition, confirmation, rewrite, brief explanation, small read-only lookup, concise recommendation",
		"Do not emit a `PASS`, `PARTIAL`, `BLOCKED`, or `FAIL` heading",
		"Do not add `## Next actions`, a synthetic `None` item, a task profile, or a Repository Memory block",
		"Do not use word count, token count, elapsed time, or number of tool calls as an applicability threshold",
		"When uncertain, prefer a natural response",
		"The user explicitly requests the canonical structured completion report",
		"# <PASS|PARTIAL|BLOCKED|FAIL> — <one-sentence outcome>",
		"## Next actions",
		"Order items as `Blocker`, `Incomplete`, `Next`, `Optional`, then `None`",
		"every structured PASS response includes one `None` item",
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
		"## Left-Aligned Detail Contract",
		"Do not use a Markdown pipe table",
		"left-aligned section headings and CommonMark list or key/value blocks",
		"**Decision:** `created|updated|refactored|deleted|not required`",
		"higher-priority host wrapper",
		"`PENDING`, `UNKNOWN`, `SKIPPED`, `NOT_APPLICABLE`",
	} {
		if !strings.Contains(normalized, check) {
			t.Errorf("expected %s ruleset to contain %q", slug, check)
		}
	}
	for _, forbidden := range []string{
		"Make every terminal task response immediately scannable and actionable",
		"before every terminal task completion",
		"| Type | Action required | Why | Continue with |",
		"| Item | Result | Evidence |",
		"| Question | Finding | Evidence and confidence | Implication |",
		"| Check | Scope | Status | Evidence or gap |",
		"| Workstream | Owner | State | Dependency or next handoff |",
	} {
		if strings.Contains(ruleset.Body, forbidden) {
			t.Errorf("%s ruleset still contains centered detail table %q", slug, forbidden)
		}
	}
}

func TestAgentCompletionOutputExamplesPreserveProportionalBoundary(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "references", "rules", "agent-completion-output.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read agent completion output ruleset: %v", err)
	}
	body := string(content)
	conversationalStart := strings.Index(body, "Small conversational answer:")
	structuredStart := strings.Index(body, "Completed implementation:")
	if conversationalStart < 0 || structuredStart <= conversationalStart {
		t.Fatal("completion examples are missing or out of order")
	}
	conversational := body[conversationalStart:structuredStart]
	for _, forbidden := range []string{
		"# PASS —",
		"## Next actions",
		"**None —",
		"## Repository Memory",
	} {
		if strings.Contains(conversational, forbidden) {
			t.Errorf("conversational example contains structured output %q", forbidden)
		}
	}

	structured := body[structuredStart:]
	for _, required := range []string{
		"# PASS — canonical completion output is implemented and ready for review",
		"## Next actions",
		"**None — No action required.**",
		"## Delivery",
	} {
		if !strings.Contains(structured, required) {
			t.Errorf("structured example does not contain %q", required)
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
			"`agent-completion-output` Repository Memory key/value block",
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
