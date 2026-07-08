package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/document"
)

func assertV2SpecPromptContract(t *testing.T, output string) {
	t.Helper()

	checks := []string{
		"## Durable Repository Facts",
		"## Instruction Entrypoints",
		"## Supporting Inputs",
		"## Source Of Truth Precedence",
		"Safety, permission, and system constraints",
		"## SPEC.md Contract",
		"`SPEC.md` is the single durable feature artifact",
		"CONSTITUTION",
		"PROJECT PROGRESS",
		"durable repository facts",
		"RLM",
		"Kit's just-in-time context-routing pattern",
		"KIT MANAGED RULESETS",
		"pointer-loaded durable repo-local rulesets managed by Kit",
		"Use this fixed section order: Thesis, Context, Clarifications, Requirements, Assumptions, Acceptance Criteria, Implementation Plan, Task Checklist, Validation Map, Reflection Notes, Documentation Updates, Delivery Decision, Evidence.",
		"Valid phases are `clarify`, `ready`, `implement`, `validate`, `reflect`, `deliver`, `complete`, and `blocked`.",
		"## Supervisor Responsibilities",
		"## Prompt-Only And V1 Compatibility",
		"## Clarification-First Operating Model",
		"Start in Clarification Mode unless `SPEC.md` front matter already has `clarification.status: ready`",
		"Execution Mode begins only after clarification state is ready",
		"Keep the current conversation as live context after clarification completes.",
		"do not guess user intent in the clarify stage",
		"## First-Action Checklist",
		"git status --short",
		"## Dirty Worktree And Ownership Gate",
		"Classify existing changes as user-owned, in-scope, unrelated, or unknown",
		"## Pre-Instruction Report",
		"confidence percentage, unresolved question count, and whether any readiness gate blocks implementation",
		"`clarification.status`, `clarification.confidence`, and `clarification.unresolved_questions`",
		"## Clarification Loop",
		"Maintain front matter `clarification.status` as `open` while questions remain",
		"repo evidence before implementation begins",
		"record the exact accepted defaults in `SPEC.md`",
		"When the gate becomes ready, set `clarification.status: ready`",
		"## Source Map Mechanics",
		"Required columns: ID, Source, Selector, Claim / Fact, Used For, Maps To, Status.",
		"Source Map gate",
		"## Objective Readiness Gates",
		"every acceptance criterion has a stable `AC-###` ID",
		"## Acceptance Criteria Discipline",
		"stable acceptance criterion IDs such as `AC-001`",
		"## Phase Transition Rules",
		"Do not skip phases.",
		"## Agent Team Model",
		"The supervisor agent owns `SPEC.md`, clarification, scope, acceptance criteria, lane assignment, integration, validation synthesis, delivery gating, and final response.",
		"docs/references/rules/agent-team-orchestration.md",
		"Default to a subagent team for implementation and verification.",
		"Use a single supervisor lane only when the work is trivial, tightly coupled, the active runtime cannot spawn subagents, or `--single-agent` is explicitly active.",
		"do not keep work single-lane merely because subagents were not explicitly re-requested.",
		"Treat planned lanes as logical work routing until separate agents are actually spawned.",
		"single supervisor lane; no specialist or verification agents spawned",
		"Do not describe logical lanes as agents, spawned lanes, or verification agents unless separate agents actually ran.",
		"Default max concurrent lanes: 3.",
		"Hard ceiling: 4, only when predicted file overlap is clearly low.",
		"Do not use \"as many agents as possible.\"",
		"Verification lanes are read-only by default.",
		"## Agent Team Plan",
		"implementation lanes that will actually be spawned as subagents",
		"read-only verification lanes that will actually be spawned as subagents",
		"logical-only lanes that will not be spawned",
		"reason for each omitted implementation or verification subagent",
		"predicted touched files per lane",
		"## Implementation Rules",
		"## Acceptance Coverage Audit",
		"Each acceptance criterion row must include criterion id, implementation evidence, validation command or review evidence, documentation impact, verifier result, and gap status.",
		"## Validation And Verification Phase",
		"Map validation 1:1 to Acceptance Criteria in `SPEC.md`",
		"Use at least one read-only verification subagent by default after implementation",
		"Read-only verification lanes must not edit files",
		"For each verifier gap, record `gap id -> acceptance criterion id -> Source Map id -> fix diff area -> rerun evidence -> verifier closure`",
		"If validation is impossible, record reason, risk, substitute evidence, user-visible impact, owner or next action, and whether delivery is blocked.",
		"## Reflection Phase",
		"`kit loop workflow` writes runtime-owned `REFLECT.json` verdict evidence next to `SPEC.md`",
		"must not fabricate or self-report verdict values",
		"## Zero-Error Completion Gate",
		"No known errors remain",
		"## Documentation Update Rules",
		"## Delivery Intent And Hard Gate",
		"## SPEC.md Update Requirements",
		"Update Context `### Source Map` whenever a material claim is added",
		"## Response Scope",
		"Clarification-loop replies should use numbered questions",
		"## Final Response Contract",
		"state the exception that justified single-lane execution",
		"do not present logical planning lanes as spawned agents",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected v2 spec prompt to contain %q", check)
		}
	}
	assertFinalResponseContractHeadings(t, output,
		"Summary",
		"SPEC.md State",
		"Acceptance Coverage",
		"Validation Evidence",
		"Zero-Error Gate",
		"Agent Team",
		"Delivery",
		"Open Items",
	)
}

func assertV2SpecPromptExcludesV1StageAssumptions(t *testing.T, output string) {
	t.Helper()

	unwanted := []string{
		"Only update SPEC.md and supporting documentation",
		"Run 'kit plan",
		"Run `kit plan",
		"usually `kit plan",
		"Run 'kit legacy plan",
		"Run `kit legacy plan",
		"usually `kit legacy plan",
		"Avoid implementation details (focus on WHAT, not HOW)",
		"write the selected skills into canonical front matter `skills`; use the legacy `## SKILLS` table",
		"keep the legacy `none | n/a | n/a | no additional skills required | no` row",
	}

	for _, check := range unwanted {
		if strings.Contains(output, check) {
			t.Fatalf("v2 spec prompt reintroduced v1 stage assumption %q", check)
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = writer
	defer func() {
		os.Stdout = original
	}()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		t.Fatalf("io.Copy() error = %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("reader.Close() error = %v", err)
	}

	return buf.String()
}

func chdirForTest(t *testing.T, dir string) func() {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("os.Chdir() error = %v", err)
	}

	return func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("os.Chdir() restore error = %v", err)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := document.Write(path, content); err != nil {
		t.Fatalf("document.Write(%q) error = %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	return string(content)
}

func defaultKitConfig() string {
	return "goal_percentage: 95\nspecs_dir: docs/specs\nskills_dir: .agents/skills\nconstitution_path: docs/CONSTITUTION.md\nallow_out_of_order: false\nagents:\n  - AGENTS.md\n  - CLAUDE.md\n  - .github/copilot-instructions.md\nfeature_naming:\n  numeric_width: 4\n  separator: '-'\n"
}

func documentTemplateWithSummary() string {
	return "# SPEC\n\n## SUMMARY\n\nsummary\n"
}
