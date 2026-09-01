package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/document"
)

func TestAgentTeamOrchestrationRegistryRulesetIsCapabilityAware(t *testing.T) {
	const slug = "agent-team-orchestration"
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
	for _, appliesTo := range []string{
		"coding-agent", "workflow", "dispatch", "subagent", "verification",
	} {
		if !slices.Contains(ruleset.Metadata.AppliesTo, appliesTo) {
			t.Errorf("applies_to = %#v, want %q", ruleset.Metadata.AppliesTo, appliesTo)
		}
	}

	normalized := strings.Join(strings.Fields(ruleset.Body), " ")
	for _, check := range []string{
		"SCOPED -> CAPABILITY_NEGOTIATING -> MAPPING -> SYNTHESIZING -> DRILL_DOWN -> SYNTHESIZING -> PLAN_READY -> EXECUTION_READY -> IMPLEMENTING -> INTEGRATING -> VERIFYING -> REPAIRING -> VERIFYING -> SOURCE_VERIFIED -> PR_READY -> MERGE_AUTHORIZED -> MERGED -> RELEASE_VERIFIED -> PROVENANCE_PR_READY -> COMPLETE",
		"`SYNTHESIZING -> DRILL_DOWN -> SYNTHESIZING` loop repeats",
		"`VERIFYING -> REPAIRING -> VERIFYING` loop repeats",
		"`SOURCE_VERIFIED` requires the integrated source, tests, documentation, and evidence to agree",
		"`PR_READY` never implies `MERGE_READY`",
		"accepted task or active `/goal` to cover the exact current pull-request set",
		"`MERGED` requires observed evidence that the exact authorized pull request merged",
		"`RELEASE_VERIFIED` requires evidence tying the expected merged source to the exact release tag",
		"`PROVENANCE_PR_READY` requires a separately authorized and owned issue/branch/worktree/pull-request lane",
		"It never invents `MERGE_READY` for that separate pull request",
		"must not fabricate them or skip through them to claim lifecycle completion",
		"Build a Capability Manifest",
		"separate_execution: confirmed | unavailable | unknown",
		"parallel_execution: confirmed | unavailable | unknown",
		"stable_agent_references: confirmed | unavailable | unknown",
		"same_agent_follow_up: confirmed | unavailable | unknown",
		"model_selection: confirmed | unavailable | unknown",
		"effort_selection: confirmed | unavailable | unknown",
		"fresh_verification: confirmed | unavailable | unknown",
		"wait_status_controls: confirmed | unavailable | unknown",
		"effective_capacity: <host-confirmed value> | host-managed | unavailable | unknown",
		"selected_topology: single-supervisor | root-with-children | host-managed",
		"delegation_depth: zero | one | unknown",
		"Profiles describe required behavior, not vendor products or fixed model IDs",
		"`architect`", "`orchestrator`", "`mapper`", "`specialist`", "`precision`", "`verifier`",
		"Only the root supervisor may launch children",
		"Delegation depth is exactly one",
		"The active host owns admission, scheduling, available slots, and effective capacity",
		"Kit defines no default or ceiling for concurrent agents",
		"Preserve each `lane_id` and its stable host agent reference",
		"provide a full rebrief",
		"fresh independent `verifier`",
		"distinct read-only self-review",
		"verification_independent: false",
		"Read the top-level project `goal_percentage` as the `PLAN_READY` threshold",
		"Default to `95` only when it is absent",
		"zero unresolved material questions",
		"native task context, not specification front matter",
		"requested_profile", "effective_profile",
		"execution_kind: actual_agent | logical_lane | omitted",
		"task_outcome: workflow-native outcome",
		"orchestration_conformance: full | degraded | unsatisfied",
		"Kit commands do not inspect a host's live agent roster",
		"First-Pass Topology Decision",
		"Before `PLAN_READY`, the root supervisor must have evaluated and recorded a",
		"Recording the decision is mandatory even when",
		"This first-pass evaluation is mandatory and does not wait for",
		"Serialize work sharing a repository, a migration or schema registry",
		"contract under active revision, deployment state, runtime authority",
		"Handoff Reconciliation",
		"Never adopt a child or logical lane's handoff solely from its narrative",
		"reconcile against git and remote heads",
		"treat a lane's own summary as a starting hypothesis to reconcile",
	} {
		if !strings.Contains(normalized, check) {
			t.Errorf("expected %s ruleset to contain %q", slug, check)
		}
	}
}

func TestAgentTeamOrchestrationFallbackOrderAndNeutrality(t *testing.T) {
	body := readRepositoryFile(t, "docs/references/rules/agent-team-orchestration.md")
	normalized := strings.Join(strings.Fields(body), " ")

	previous := -1
	for _, step := range []string{
		"an equal-or-stronger eligible configuration",
		"a narrowed, low-risk lane paired with stronger verification",
		"a runtime-selected and explicitly unverified configuration",
		"`BLOCKED` when no acceptable configuration remains",
	} {
		index := strings.Index(normalized, step)
		if index < 0 {
			t.Fatalf("fallback order is missing %q", step)
		}
		if index <= previous {
			t.Fatalf("fallback step %q is out of order", step)
		}
		previous = index
	}

	lowerBody := strings.ToLower(body)
	for _, forbidden := range []string{
		"--max-subagents", "default maximum concurrent lanes", "hard ceiling",
		"never exceed 4", "max concurrency is 3", "codex", "claude", "copilot",
		"warp", "openai", "anthropic", "gpt-", "opus", "sonnet", "haiku",
		"luna", "terra",
	} {
		if strings.Contains(lowerBody, forbidden) {
			t.Errorf("provider-neutral ruleset unexpectedly contains %q", forbidden)
		}
	}

	numericValue := regexp.MustCompile(`\b[0-9]+\b`)
	for _, line := range strings.Split(lowerBody, "\n") {
		hasAgentSubject := strings.Contains(line, "agent") ||
			strings.Contains(line, "child") || strings.Contains(line, "lane")
		hasLimitPolicy := strings.Contains(line, "default") ||
			strings.Contains(line, "maximum") || strings.Contains(line, "hard ceiling") ||
			strings.Contains(line, "never exceed") || strings.Contains(line, "at most")
		if hasAgentSubject && hasLimitPolicy && numericValue.MatchString(line) {
			t.Errorf("ruleset contains fixed numeric concurrency policy %q", line)
		}
	}

	index, err := os.ReadFile(filepath.Join("..", "..", "docs", "references", "README.md"))
	if err != nil {
		t.Fatalf("read references index: %v", err)
	}
	if !strings.Contains(string(index), "| `agent-team-orchestration` |") {
		t.Fatal("references index does not list agent-team-orchestration")
	}
}
