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

	prompt := buildDispatchPrompt(tasks, "/tmp/project", dispatchInputSourceEditor, dispatchPromptOptions{})
	checks := []string{
		"Prepare an Agent Team Plan",
		"Working directory: `/tmp/project`",
		"Input source: editor",
		"### D001",
		"### D002",
		"one accountable supervisor",
		"predict touched files/interfaces",
		"Cluster by file overlap",
		"Runtime Capability Negotiation",
		"host-confirmed separate execution",
		"Preserve `unknown` literally",
		"actual agent only when the runtime explicitly creates a separate execution",
		"delegation depth at one",
		"`single-supervisor`, `root-with-children`, or `host-managed`",
		"requested and effective profiles separately",
		"runtime-selected/unverified",
		"parallel execution",
		"stable agent reference",
		"continuity loss",
		"fresh independent read-only verifier",
		"supervisor self-review",
		"assigned checkout or prepared worktree",
		"may not independently create, switch, move, or remove worktrees",
		"Agent Team Plan Output",
		"Capability Manifest",
		"`actual_agent`, `logical_lane`, or `omitted`",
		"`parallel-confirmed`, `sequential`, or `unconfirmed`",
		"verification_independent: true | false | unknown",
		"task_outcome",
		"orchestration_conformance",
		"Publish the plan before spawning",
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
	for _, retired := range []string{"Max concurrent subagents", "at most 3", "hard ceiling", "never exceed 4"} {
		if strings.Contains(prompt, retired) {
			t.Fatalf("expected prompt to omit retired fixed-cap policy %q, got %q", retired, prompt)
		}
	}
}

func TestDispatchCommandRejectsRemovedMaxSubagentsFlag(t *testing.T) {
	if flag := dispatchCmd.Flags().Lookup("max-subagents"); flag != nil {
		t.Fatalf("expected dispatch not to expose --max-subagents, got %#v", flag)
	}
	err := dispatchCmd.ParseFlags([]string{"--max-subagents", "2"})
	if err == nil || !strings.Contains(err.Error(), "unknown flag: --max-subagents") {
		t.Fatalf("expected removed flag to be rejected as unknown, got %v", err)
	}
}

func TestBuildDispatchPromptIncludesCommonReviewInstruction(t *testing.T) {
	tasks := []dispatchTask{
		{ID: "D001", Index: 1, Body: "Source: internal/app.go:12\nReview task:\nFix the stale assertion"},
	}

	prompt := buildDispatchPrompt(
		tasks,
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

func TestBuildDispatchPromptCarriesResolvedRepairLane(t *testing.T) {
	prompt := buildDispatchPrompt(
		[]dispatchTask{{ID: "D001", Index: 1, Body: "Fix the review finding"}},
		"/tmp/kit/GH-67",
		dispatchInputSourcePR,
		dispatchPromptOptions{
			PRTarget: "https://github.com/jamesonstone/kit/pull/67",
			RepairContext: &repairContext{
				Repository:      "jamesonstone/kit",
				PRNumber:        67,
				PRURL:           "https://github.com/jamesonstone/kit/pull/67",
				HeadBranch:      "GH-67",
				ExpectedHeadOID: "remote-head",
				LocalHeadOID:    "local-head",
				WorktreePath:    "/tmp/kit/GH-67",
				WorktreeCreated: true,
				ExistingChanges: repairChangesExclude,
				DirtyStatus:     " M pkg/cli/pr.go",
				PushTarget:      "origin/GH-67",
			},
		},
	)

	for _, check := range []string{
		"## Repair Lane",
		"https://github.com/jamesonstone/kit/pull/67",
		"GH-67",
		"remote-head",
		"local-head",
		"/tmp/kit/GH-67",
		"created",
		"exclude",
		" M pkg/cli/pr.go",
		"Preserve them, do not stage or modify their paths",
		"origin/GH-67",
		"never create a second repair branch or pull request",
	} {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected repair prompt to contain %q, got:\n%s", check, prompt)
		}
	}
}

func TestBuildDispatchPromptShellQuotesRepairWorktree(t *testing.T) {
	worktreePath := "/tmp/kit/$(touch injected); lane's"
	prompt := buildDispatchPrompt(
		[]dispatchTask{{ID: "D001", Index: 1, Body: "Fix the review finding"}},
		worktreePath,
		dispatchInputSourcePR,
		dispatchPromptOptions{
			RepairContext: &repairContext{
				HeadBranch:   "GH-67",
				WorktreePath: worktreePath,
				PushTarget:   "origin/GH-67",
			},
		},
	)

	quotedPath := `'/tmp/kit/$(touch injected); lane'"'"'s'`
	if got := strings.Count(prompt, "git -C "+quotedPath); got != 2 {
		t.Fatalf("repair prompt contains shell-quoted worktree path %d times, want 2:\n%s", got, prompt)
	}
	if strings.Contains(prompt, `git -C "/tmp/kit/$(touch injected); lane's"`) {
		t.Fatalf("repair prompt used command-substitution-capable double quotes:\n%s", prompt)
	}
}

func TestBuildDispatchPromptScopesPRReflectionCycleToCodeRabbit(t *testing.T) {
	tasks := []dispatchTask{
		{ID: "D001", Index: 1, Body: "Source: internal/app.go:12\nReview task:\nFix the stale assertion"},
	}

	prompt := buildDispatchPrompt(
		tasks,
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
