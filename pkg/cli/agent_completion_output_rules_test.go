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
		"## Rules",
		"### Shape",
		"Write each response in the shape its content calls for",
		"Match length to consequence rather than to effort spent",
		"### What A Terminal Response Conveys",
		"What the user now has",
		"What remains unfinished, and why",
		"Anything blocking completion, and what would clear it",
		"the exact command or prompt when there is one",
		"Whether the work is finished, partly finished, blocked, or failed, said plainly",
		"the identifiers a reader needs to find or undo it",
		"Blockers and unfinished scope belong where the reader will see them, as prominent as the successes",
		"### Reporting State Truthfully",
		"Report each check as observed",
		"`PENDING`, `UNKNOWN`, `SKIPPED`, and `NOT_APPLICABLE`",
		"Report a check as passing only when it ran and passed",
		"When something could not be validated, say so and say why",
		"Distinguish a verified fact from an inference and from a hypothesis",
		"### Evidence",
		"not an index of everything checked",
		"Include an identifier when the reader needs it to act",
		"smallest evidence set that proves each terminal node",
		"Satisfy them on content; they say nothing about layout, and a heading alone satisfies none of them",
		"These are failures of content, not of layout",
		"Carrying one response template across unrelated tasks",
		"These differ in shape because their content differs",
	} {
		if !strings.Contains(normalized, check) {
			t.Errorf("expected %s ruleset to contain %q", slug, check)
		}
	}
	for _, forbidden := range []string{
		"## What happened",
		"## Deviations",
		"## Next steps",
		"**Status:",
		"**None.**",
		"retired envelope",
		"## Proportionality Gate",
		"## Three-Section Completion Contract",
		"## Density Budget",
		"Target twelve rendered lines or fewer",
		"| Type | Action required | Why | Continue with |",
		"| Item | Result | Evidence |",
	} {
		if strings.Contains(ruleset.Body, forbidden) {
			t.Errorf("%s ruleset still contains centered detail table %q", slug, forbidden)
		}
	}
}

func TestAgentCompletionOutputExamplesModelUnstructuredResponses(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "references", "rules", "agent-completion-output.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read agent completion output ruleset: %v", err)
	}
	body := string(content)

	examples := strings.Index(body, "## Examples")
	if examples < 0 {
		t.Fatal("ruleset has no examples section")
	}
	verification := strings.Index(body, "## Verification")
	if verification <= examples {
		t.Fatal("verification section is missing or out of order")
	}
	shown := body[examples:verification]

	for _, label := range []string{
		"A small answer stays a small answer:",
		"A clean delivery, said once:",
		"Partial work, where the gap leads:",
		"A blocker, with the unblock:",
	} {
		if !strings.Contains(shown, label) {
			t.Errorf("examples do not cover %q", label)
		}
	}

	for _, forbidden := range []string{
		"## What happened",
		"## Deviations",
		"## Next steps",
		"**Status:",
		"**None.**",
		"**Required — User:**",
	} {
		if strings.Contains(shown, forbidden) {
			t.Errorf("examples still model the retired envelope %q", forbidden)
		}
	}

	if !strings.Contains(shown, "resume diagnosis using the authorized production logs") {
		t.Error("blocked example does not give a copy-ready continuation")
	}
	if !strings.Contains(shown, "Tell me which you want and I'll finish it.") {
		t.Error("partial example does not name what the reader must decide")
	}
}

func TestAgentCompletionOutputIsIntegratedWithRelatedRules(t *testing.T) {
	checks := map[string][]string{
		"docs/references/README.md": {
			"Use `rules/agent-completion-output.md`",
			"| `agent-completion-output` |",
			"it requires no response format and names only the facts a response must not leave out",
		},
		"docs/references/rules/github-pr-delivery.md": {
			"Follow `agent-completion-output`, which prescribes no format",
			"must be recoverable from the response, including identity, assignment, and hosted-state evidence; where they appear is free",
		},
		"docs/references/rules/testing-and-environment-validation.md": {
			"Follow `agent-completion-output` for terminal reporting, which prescribes no format",
			"Keep observed validation results, gaps or non-passing evidence, and any required rerun or remediation visible and distinct",
		},
		"docs/references/rules/agent-team-orchestration.md": {
			"State `task_outcome` plainly",
			"report degraded or unsatisfied conformance as its own fact",
		},
		"docs/references/rules/cross-repository-program-coordination.md": {
			"Render the terminal program result through `agent-completion-output`, which prescribes no format",
			"state unresolved dependencies and exact handoffs plainly",
		},
		"docs/references/rules/constitution-curation.md": {
			"State the Constitution curation result once",
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
