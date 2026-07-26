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
		"## Goal",
		"## User Context",
		"## Repository Context",
		"## Source And State Contract",
		"`SPEC.md` is the single durable feature artifact",
		"Context `### Source Map`",
		"`SRC-###`, `REQ-###`, `AC-###`",
		"clarification state",
		"## Clarification And Autonomy",
		"research repository-discoverable facts first",
		"Ask only about material choices that remain non-discoverable",
		"Keep `clarification.status: open` only while one or more such questions remain",
		"Record residual uncertainty that does not require a user decision as an assumption or named risk",
		"Confidence is a reporting signal and does not determine `clarification.status`",
		"Outside `clarify`, do not re-ask settled questions or request routine permission",
		"## Constraints And Approval Boundaries",
		"Safe repository reads and reversible in-scope edits need no extra approval",
		"git status --short",
		"Before Git/GitHub delivery mutation",
		"Delivery Contract",
		"## Phase Outcomes",
		"Do not skip a phase gate",
		"validates phase state in code",
		"## Agent Routing",
		"docs/references/rules/agent-team-orchestration.md",
		"read-only verifier",
		"single supervisor lane; no specialist or verification agents spawned",
		"## Success Criteria",
		"implementation evidence",
		"exact validation evidence",
		"Never claim a check ran when it did not",
		"## Output Contract",
		"Open Questions",
		"## Final Response Contract",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected v2 spec prompt to contain %q", check)
		}
	}
	for _, forbidden := range []string{
		"Programmatic Tool Calling",
		"persisted reasoning",
		"Pro mode",
		"text.verbosity",
		"Ask clarification questions until",
		"confidence is at least",
	} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("v2 spec prompt unexpectedly contains %q", forbidden)
		}
	}
	assertFinalResponseContractHeadings(t, output,
		"Outcome",
		"Evidence",
		"Artifacts And State",
		"Agent Team",
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
