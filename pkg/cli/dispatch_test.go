package cli

import (
	"strings"
	"testing"
)

func TestBuildDispatchPrompt(t *testing.T) {
	tasks := []dispatchTask{
		{ID: "D001", Index: 1, Body: "Update middleware"},
		{ID: "D002", Index: 2, Body: "Refresh README"},
	}

	prompt := buildDispatchPrompt(tasks, defaultDispatchMaxSubagents, "/tmp/project", dispatchInputSourceEditor, dispatchPromptOptions{})
	checks := []string{
		"Prepare an Agent Team Plan",
		"Working directory: /tmp/project",
		"Input source: editor",
		"Effective max subagents: 3",
		"Default max concurrent lanes: 3",
		"Hard ceiling: 4",
		"### D001",
		"### D002",
		"one accountable supervisor",
		"agent-team-orchestration.md",
		"anticipate which files are likely to change",
		"proposed lanes",
		"subagents that will actually be spawned",
		"logical-only lanes that will not be spawned",
		"intentionally omitted implementation or verification lanes with reasons",
		"validation/review lanes",
		"risks and unknowns",
		"self-direct execution",
		"launching at most 3 concurrent subagents",
		"Keep all subagent work in the existing project directory",
		"do not create or use git worktrees",
		"stop and ask the user how to proceed instead of creating an alternate checkout",
		"single supervisor lane; no specialist or verification agents spawned",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected prompt to contain %q", check)
		}
	}

	if !strings.HasPrefix(prompt, "Prepare an Agent Team Plan") {
		t.Fatalf("expected prompt to start with dispatch header, got %q", prompt[:40])
	}
	if strings.Contains(prompt, "/plan") || strings.Contains(prompt, "planning mode") {
		t.Fatalf("expected prompt to avoid native plan-mode triggers, got %q", prompt)
	}
	if strings.Contains(prompt, "PR Reflection and Resolution Cycle") {
		t.Fatalf("expected non-PR dispatch prompt to omit PR reflection cycle, got %q", prompt)
	}
}

func TestDispatchCommandMaxSubagentsDefaultAndCeiling(t *testing.T) {
	flag := dispatchCmd.Flags().Lookup("max-subagents")
	if flag == nil {
		t.Fatal("expected dispatch to expose --max-subagents")
	}
	if flag.DefValue != "3" || !strings.Contains(flag.Usage, "hard ceiling 4") {
		t.Fatalf("unexpected --max-subagents flag metadata: def=%q usage=%q", flag.DefValue, flag.Usage)
	}
}

func TestBuildDispatchPromptIncludesCommonReviewInstruction(t *testing.T) {
	tasks := []dispatchTask{
		{ID: "D001", Index: 1, Body: "Source: internal/app.go:12\nReview task:\nFix the stale assertion"},
	}

	prompt := buildDispatchPrompt(
		tasks,
		4,
		"/tmp/project",
		dispatchInputSourcePR,
		dispatchPromptOptions{CommonReviewInstruction: coderabbitSharedReviewInstruction},
	)

	checks := []string{
		"Input source: pr-review",
		"Common Review Instruction",
		coderabbitSharedReviewInstruction,
		"Fix the stale assertion",
	}
	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected prompt to contain %q", check)
		}
	}
}

func TestBuildDispatchPromptIncludesPRReflectionCycle(t *testing.T) {
	tasks := []dispatchTask{
		{ID: "D001", Index: 1, Body: "Source: internal/app.go:12\nReview task:\nFix the stale assertion"},
	}

	prompt := buildDispatchPrompt(
		tasks,
		3,
		"/tmp/project",
		dispatchInputSourcePR,
		dispatchPromptOptions{PRTarget: "14"},
	)

	checks := []string{
		"PR Reflection and Resolution Cycle",
		"after validation and push-to-PR",
		"gh pr view \"14\" --json headRefOid -q .headRefOid",
		"git rev-parse HEAD",
		"Run a reflection cycle against the pushed diff",
		"no code has been pushed to the PR after your push",
		"kit dispatch --pr \"14\" --resolve --yes",
		"resolved conversation count or reason resolution was skipped",
	}
	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected prompt to contain %q, got:\n%s", check, prompt)
		}
	}
	if strings.Contains(prompt, "--coderabbit --resolve") {
		t.Fatalf("expected default PR resolution to include all active conversations, got:\n%s", prompt)
	}
}

func TestBuildDispatchPromptScopesPRReflectionCycleToCodeRabbit(t *testing.T) {
	tasks := []dispatchTask{
		{ID: "D001", Index: 1, Body: "Source: internal/app.go:12\nReview task:\nFix the stale assertion"},
	}

	prompt := buildDispatchPrompt(
		tasks,
		3,
		"/tmp/project",
		dispatchInputSourcePR,
		dispatchPromptOptions{CodeRabbitOnly: true, PRTarget: "14"},
	)

	checks := []string{
		"all active CodeRabbit-authored PR review conversations",
		"kit dispatch --pr \"14\" --coderabbit --resolve --yes",
	}
	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected prompt to contain %q, got:\n%s", check, prompt)
		}
	}
}

func TestNormalizeDispatchTasks(t *testing.T) {
	raw := strings.Join([]string{
		"Investigate auth failures",
		"on expired sessions",
		"",
		"- Update middleware",
		"  - preserve nested detail",
		"  - keep order",
		"",
		"1. Refresh CLI help",
		"2. Add README entry",
		"",
		"Confirm validation output",
	}, "\n")

	tasks, err := normalizeDispatchTasks(raw)
	if err != nil {
		t.Fatalf("expected task normalization to succeed: %v", err)
	}

	if len(tasks) != 5 {
		t.Fatalf("expected 5 normalized tasks, got %d", len(tasks))
	}

	wantBodies := []string{
		"Investigate auth failures\non expired sessions",
		"Update middleware\n  - preserve nested detail\n  - keep order",
		"Refresh CLI help",
		"Add README entry",
		"Confirm validation output",
	}

	for index, wantBody := range wantBodies {
		if tasks[index].ID != "D00"+string(rune('1'+index)) {
			t.Fatalf("expected stable task ID at index %d, got %q", index, tasks[index].ID)
		}
		if tasks[index].Body != wantBody {
			t.Fatalf("expected body %q, got %q", wantBody, tasks[index].Body)
		}
	}
}

func TestNormalizeDispatchTasksRejectsEmptyInput(t *testing.T) {
	if _, err := normalizeDispatchTasks(" \n\t "); err == nil {
		t.Fatalf("expected empty task input to fail")
	}
}
