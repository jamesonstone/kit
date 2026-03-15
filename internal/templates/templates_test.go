package templates

import (
	"strings"
	"testing"
)

func TestInstructionFile_ReturnsKnownTemplates(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "agents", path: "AGENTS.md", want: AgentsMD},
		{name: "claude", path: "CLAUDE.md", want: ClaudeMD},
		{name: "copilot", path: ".github/copilot-instructions.md", want: CopilotInstructionsMD},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InstructionFile(tt.path); got != tt.want {
				t.Fatalf("InstructionFile(%q) did not return the expected template", tt.path)
			}
		})
	}
}

func TestInstructionFile_FallsBackToAgentTemplate(t *testing.T) {
	got := InstructionFile("nested/GEMINI.md")
	if !strings.HasPrefix(got, "# GEMINI\n\n") {
		t.Fatalf("expected fallback template to use file stem, got %q", got[:min(len(got), 16)])
	}
	if !strings.Contains(got, "## Change Classification (Required First Step)") {
		t.Fatalf("expected fallback template to use the comprehensive shared instructions")
	}
}

func TestCopilotInstructionsMD_FrontLoadsCriticalRules(t *testing.T) {
	firstWindow := CopilotInstructionsMD
	if len(firstWindow) > 4000 {
		firstWindow = firstWindow[:4000]
	}

	checks := []string{
		"## Fast rules for chat and code review",
		"classify every request first",
		"`BRAINSTORM.md` when present, then `SPEC.md` → `PLAN.md` → `TASKS.md`",
		"do NOT run `coderabbit --prompt-only`, `git add`, or `git commit` without explicit approval",
	}

	for _, check := range checks {
		if !strings.Contains(firstWindow, check) {
			t.Fatalf("expected first 4000 characters to contain %q", check)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
