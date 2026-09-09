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
		"## Density Budget",
		"A structured report is a briefing for someone deciding what to do next",
		"Target twelve rendered lines or fewer for ordinary work",
		"Keep What happened to five bullets or fewer",
		"Keep each bullet to one sentence, plus at most one short clause carrying its evidence",
		"Nest at most one line under a bullet",
		"## Three-Section Completion Contract",
		"use these headings and no others",
		"## What happened",
		"## Deviations",
		"## Next steps",
		"**Status: <PASS|PARTIAL|BLOCKED|FAIL> — <one-sentence outcome>.**",
		"Prose is the default shape for this section",
		"### Evidence",
		"Include an identifier when the reader needs it to act",
		"Leave out: a commit SHA cited as proof that something was checked",
		"At most one link per claim",
		"### Formatting",
		"Reserve bold for the status line, a blocker, and a required action",
		"Omit this section entirely when nothing diverged",
		"Omit this section entirely when no action remains",
		"always, when the status is PARTIAL, BLOCKED, or FAIL",
		"Every required follow-up includes a copy-ready prompt or command",
		"Speculative offers of work the user has not asked for are not next steps",
		"## Task-Specific Content",
		"Task type decides which facts earn a place, not how many sections the report has",
		"A composing contract names which facts must survive, never how many lines they may occupy",
		"For merge or release orchestration, keep What happened to state changes",
		"smallest evidence set that proves each terminal node",
		"Do not include a chronological command log, repeated checks, unchanged polling, or routine tool details",
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
		"Always include this section. Use one `**None.**` bullet",
		"Use one nested evidence layer only",
		"Start with a short bold lead; put identifiers and evidence after it",
		"Prefer exact links, identifiers, commands, timestamps, and counts over vague",
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

	const (
		conversationalLabel = "Small conversational answer:"
		implementationLabel = "Substantial implementation, nothing diverged and nothing remains:"
		coordinationLabel   = "Complex production coordination, where separate bullets earn their place:"
		blockedLabel        = "Blocked diagnosis:"
	)
	offsets := make([]int, 0, 4)
	for _, label := range []string{conversationalLabel, implementationLabel, coordinationLabel, blockedLabel} {
		at := strings.Index(body, label)
		if at < 0 {
			t.Fatalf("completion example %q is missing", label)
		}
		if len(offsets) > 0 && at <= offsets[len(offsets)-1] {
			t.Fatalf("completion example %q is out of order", label)
		}
		offsets = append(offsets, at)
	}

	conversational := body[offsets[0]:offsets[1]]
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

	implementation := body[offsets[1]:offsets[2]]
	if !strings.Contains(implementation, "**Status: PASS — completion output is proportional and ready for review.**") {
		t.Error("implementation example does not open with the status line")
	}
	for _, forbidden := range []string{"## Deviations", "## Next steps"} {
		if strings.Contains(implementation, forbidden) {
			t.Errorf("implementation example emits empty section %q instead of omitting it", forbidden)
		}
	}

	coordination := body[offsets[2]:offsets[3]]
	for _, required := range []string{
		"**Status: PASS — PRs #290, #181, and #230 merged",
		"upsert stayed default-off as intended",
		"## Deviations",
	} {
		if !strings.Contains(coordination, required) {
			t.Errorf("coordination example does not contain %q", required)
		}
	}
	if strings.Contains(coordination, "## Next steps") {
		t.Error("coordination example emits an empty Next steps section")
	}
	for _, forbidden := range []string{
		"## Operational Result",
		"## Validation",
		"## Feature State",
		"## Residual Notes",
		"## Coordination",
		"## Repository Memory",
	} {
		if strings.Contains(coordination, forbidden) {
			t.Errorf("coordination example contains superseded section %q", forbidden)
		}
	}

	blocked := body[offsets[3]:]
	for _, required := range []string{
		"**Status: BLOCKED —",
		"## Deviations",
		"remains `UNKNOWN`",
		"## Next steps",
		"**Required — User:** Grant read-only production log access.",
		"Continue with: `Resume diagnosis using the authorized production logs.`",
	} {
		if !strings.Contains(blocked, required) {
			t.Errorf("blocked example does not contain %q", required)
		}
	}

	if strings.Contains(body, "- **None.**") {
		t.Error("ruleset still models a None bullet")
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
