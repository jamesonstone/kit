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
		"direct questions, definitions, confirmations, rewrites, brief explanations, small read-only lookups, concise recommendations",
		"Do not emit status tokens, canonical section headings, synthetic None items",
		"Do not use word count, token count, elapsed time, or tool-call count as an applicability threshold",
		"When uncertain, prefer natural prose",
		"## Three-Section Completion Contract",
		"emit exactly these headings in order",
		"## What happened",
		"## Deviations",
		"## Next steps",
		"**Status: <PASS|PARTIAL|BLOCKED|FAIL> — <one-sentence outcome>.**",
		"Use one nested evidence layer only",
		"Use one `**None.**` bullet when there are no deviations",
		"Every required follow-up includes a copy-ready prompt or command",
		"Use one `**None.**` bullet when no action remains",
		"## Task-Specific Content",
		"Task types define which facts must be retained, not additional output sections",
		"Use exactly the three canonical headings",
		"Do not repeat a fact across sections",
		"Repository-memory decision, rationale, and artifacts become one concise What happened bullet",
		"`PENDING`, `UNKNOWN`, `SKIPPED`, `NOT_APPLICABLE`",
	} {
		if !strings.Contains(normalized, check) {
			t.Errorf("expected %s ruleset to contain %q", slug, check)
		}
	}
	for _, forbidden := range []string{
		"Make every terminal task response immediately scannable and actionable",
		"before every terminal task completion",
		"## Structured Completion Envelope",
		"### Operator Action List",
		"## Required Profiles",
		"## Left-Aligned Detail Contract",
		"Order items as `Blocker`, `Incomplete`, `Next`, `Optional`, then `None`",
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
	structuredStart := strings.Index(body, "Substantial implementation:")
	if conversationalStart < 0 || structuredStart <= conversationalStart {
		t.Fatal("completion examples are missing or out of order")
	}
	conversational := body[conversationalStart:structuredStart]
	for _, forbidden := range []string{
		"# PASS —",
		"## What happened",
		"## Deviations",
		"## Next steps",
		"**None.**",
		"## Repository Memory",
	} {
		if strings.Contains(conversational, forbidden) {
			t.Errorf("conversational example contains structured output %q", forbidden)
		}
	}

	complexStart := strings.Index(body, "Complex production coordination:")
	if complexStart <= structuredStart {
		t.Fatal("complex production example is missing or out of order")
	}
	structured := body[structuredStart:complexStart]
	for _, required := range []string{
		"## What happened",
		"**Status: PASS — completion output is proportional and ready for review.**",
		"## Deviations",
		"CodeRabbit remains `PENDING`",
		"## Next steps",
		"**Optional — User:** Review PR #123 after CodeRabbit completes.",
	} {
		if !strings.Contains(structured, required) {
			t.Errorf("structured example does not contain %q", required)
		}
	}

	complexEnd := strings.Index(body[complexStart:], "Blocked diagnosis:")
	if complexEnd < 0 {
		t.Fatal("blocked diagnosis example is missing")
	}
	complex := body[complexStart : complexStart+complexEnd]
	for _, required := range []string{
		"**Status: PASS — PRs #290, #181, and #230 merged",
		"Production validation passed and manual upsert remained default-off",
		"Non-fatal workflow warnings",
		"## Next steps\n\n- **None.**",
	} {
		if !strings.Contains(complex, required) {
			t.Errorf("complex example does not contain %q", required)
		}
	}
	for _, forbidden := range []string{
		"## Operational Result",
		"## Validation",
		"## Feature State",
		"## Residual Notes",
		"## Coordination",
		"## Repository Memory",
	} {
		if strings.Contains(complex, forbidden) {
			t.Errorf("complex example contains superseded section %q", forbidden)
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
			"Follow the `agent-completion-output` three-section contract",
			"fields below into concise What happened bullets",
		},
		"docs/references/rules/testing-and-environment-validation.md": {
			"Follow `agent-completion-output` for terminal reporting",
			"validation results under What happened",
		},
		"docs/references/rules/agent-team-orchestration.md": {
			"Map `task_outcome` to the first What happened status bullet",
		},
		"docs/references/rules/cross-repository-program-coordination.md": {
			"`agent-completion-output` three-section contract",
		},
		"docs/references/rules/constitution-curation.md": {
			"Constitution curation result in one concise What happened bullet",
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
