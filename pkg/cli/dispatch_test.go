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

func TestResolveDispatchPRTargetParsesURLMarkdownAndOwnerNumber(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want dispatchPRTarget
	}{
		{
			name: "url",
			raw:  "https://github.com/Patient-Driven-Care/cortex/pull/67",
			want: dispatchPRTarget{Owner: "Patient-Driven-Care", Repo: "cortex", Number: 67},
		},
		{
			name: "markdown link",
			raw:  "[Patient-Driven-Care/cortex#67](https://github.com/Patient-Driven-Care/cortex/pull/67)",
			want: dispatchPRTarget{Owner: "Patient-Driven-Care", Repo: "cortex", Number: 67},
		},
		{
			name: "owner repo number",
			raw:  "Patient-Driven-Care/cortex#67",
			want: dispatchPRTarget{Owner: "Patient-Driven-Care", Repo: "cortex", Number: 67},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveDispatchPRTarget(tc.raw)
			if err != nil {
				t.Fatalf("expected parse to succeed: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %#v, got %#v", tc.want, got)
			}
		})
	}
}

func TestResolveDispatchPRTargetUsesCurrentRepoForNumber(t *testing.T) {
	previous := dispatchCurrentRepoResolver
	dispatchCurrentRepoResolver = func() (string, string, error) {
		return "Patient-Driven-Care", "cortex", nil
	}
	defer func() {
		dispatchCurrentRepoResolver = previous
	}()

	got, err := resolveDispatchPRTarget("67")
	if err != nil {
		t.Fatalf("expected current-repo PR number parse to succeed: %v", err)
	}

	want := dispatchPRTarget{Owner: "Patient-Driven-Care", Repo: "cortex", Number: 67}
	if got != want {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
}

func TestParseGitHubRemoteURL(t *testing.T) {
	cases := []struct {
		raw   string
		owner string
		repo  string
	}{
		{raw: "git@github.com:jamesonstone/kit.git", owner: "jamesonstone", repo: "kit"},
		{raw: "https://github.com/Patient-Driven-Care/cortex.git", owner: "Patient-Driven-Care", repo: "cortex"},
		{raw: "ssh://git@github.com/Patient-Driven-Care/cortex.git", owner: "Patient-Driven-Care", repo: "cortex"},
	}

	for _, tc := range cases {
		owner, repo, err := parseGitHubRemoteURL(tc.raw)
		if err != nil {
			t.Fatalf("expected %q to parse: %v", tc.raw, err)
		}
		if owner != tc.owner || repo != tc.repo {
			t.Fatalf("expected %s/%s, got %s/%s", tc.owner, tc.repo, owner, repo)
		}
	}
}

func TestBuildDispatchPRInputFiltersExtractsAndDedupes(t *testing.T) {
	threads := []dispatchGitHubReviewThread{
		reviewThreadFixture("internal/app.go", 12, false, false, "coderabbitai", coderabbitCommentBody("Fix app routing."), "https://example.com/1"),
		reviewThreadFixture("internal/app.go", 12, false, false, "coderabbitai", coderabbitCommentBody("Fix app routing."), "https://example.com/duplicate"),
		reviewThreadFixture("internal/stale.go", 2, true, false, "coderabbitai", coderabbitCommentBody("Skip resolved."), "https://example.com/2"),
		reviewThreadFixture("internal/old.go", 3, false, true, "coderabbitai", coderabbitCommentBody("Skip outdated."), "https://example.com/3"),
		reviewThreadFixture("internal/human.go", 4, false, false, "octocat", "Please update the docs.\n\n<!-- fingerprinting:test -->", "https://example.com/4"),
	}

	input := buildDispatchPRInput(threads, false)
	if input.CommonReviewInstruction != coderabbitSharedReviewInstruction {
		t.Fatalf("expected shared CodeRabbit instruction once, got %q", input.CommonReviewInstruction)
	}
	if strings.Count(input.RawTasks, "Fix app routing.") != 1 {
		t.Fatalf("expected duplicate CodeRabbit prompt to be collapsed, got %q", input.RawTasks)
	}
	if strings.Contains(input.RawTasks, "Verify each finding") {
		t.Fatalf("expected repeated boilerplate to be stripped from task bodies: %q", input.RawTasks)
	}
	if strings.Contains(input.RawTasks, "Skip resolved") || strings.Contains(input.RawTasks, "Skip outdated") {
		t.Fatalf("expected resolved/outdated threads to be filtered: %q", input.RawTasks)
	}
	if !strings.Contains(input.RawTasks, "Please update the docs.") {
		t.Fatalf("expected non-CodeRabbit review task to be included: %q", input.RawTasks)
	}

	coderabbitInput := buildDispatchPRInput(threads, true)
	if strings.Contains(coderabbitInput.RawTasks, "Please update the docs.") {
		t.Fatalf("expected --coderabbit input to exclude non-CodeRabbit authors: %q", coderabbitInput.RawTasks)
	}
}

func TestSplitDispatchPRInputFromEditorKeepsInstructionOutOfTasks(t *testing.T) {
	raw := strings.Join([]string{
		coderabbitSharedReviewInstruction,
		"",
		"- Source: internal/app.go:12",
		"  Review task:",
		"  Fix app routing.",
	}, "\n")

	tasks, instruction := splitDispatchPRInputFromEditor(raw, coderabbitSharedReviewInstruction)
	if instruction != coderabbitSharedReviewInstruction {
		t.Fatalf("expected common instruction to be preserved, got %q", instruction)
	}
	if strings.Contains(tasks, "Verify each finding") {
		t.Fatalf("expected tasks to omit common instruction, got %q", tasks)
	}
	if !strings.Contains(tasks, "Fix app routing.") {
		t.Fatalf("expected review task body, got %q", tasks)
	}
}

func TestResolveDispatchInputSourcePrecedence(t *testing.T) {
	if got := resolveDispatchInputSource("tasks.md", false); got != dispatchInputSourceFile {
		t.Fatalf("expected --file to win, got %s", got)
	}

	if got := resolveDispatchInputSource("", false); got != dispatchInputSourceStdin {
		t.Fatalf("expected stdin source, got %s", got)
	}

	if got := resolveDispatchInputSource("", true); got != dispatchInputSourceEditor {
		t.Fatalf("expected editor source, got %s", got)
	}
}

func TestDispatchCommandExposesPRFlags(t *testing.T) {
	dispatch, _, err := rootCmd.Find([]string{"dispatch"})
	if err != nil {
		t.Fatalf("rootCmd.Find(dispatch) error = %v", err)
	}
	if dispatch.Flags().Lookup("pr") == nil {
		t.Fatal("expected dispatch to expose --pr")
	}
	if dispatch.Flags().Lookup("coderabbit") == nil {
		t.Fatal("expected dispatch to expose --coderabbit")
	}
	if dispatch.Flags().Lookup("resolve") == nil {
		t.Fatal("expected dispatch to expose --resolve")
	}
	if dispatch.Flags().Lookup("yes") == nil {
		t.Fatal("expected dispatch to expose --yes")
	}
}

func reviewThreadFixture(
	path string,
	line int,
	resolved bool,
	outdated bool,
	author string,
	body string,
	url string,
) dispatchGitHubReviewThread {
	thread := dispatchGitHubReviewThread{
		IsResolved: resolved,
		IsOutdated: outdated,
		Path:       path,
		Line:       line,
	}
	thread.Comments.Nodes = []dispatchGitHubReviewComment{
		{
			Body: body,
			URL:  url,
		},
	}
	thread.Comments.Nodes[0].Author.Login = author
	return thread
}

func coderabbitCommentBody(task string) string {
	return strings.Join([]string{
		"_⚠️ Potential issue_ | _🟠 Major_ | _⚡ Quick win_",
		"",
		"**Finding title**",
		"",
		"<details>",
		"<summary>🤖 Prompt for AI Agents</summary>",
		"",
		"```",
		coderabbitSharedReviewInstruction,
		"",
		task,
		"```",
		"",
		"</details>",
		"",
		"<!-- fingerprinting:test -->",
	}, "\n")
}

func TestValidateDispatchMaxSubagents(t *testing.T) {
	if err := validateDispatchMaxSubagents(1); err != nil {
		t.Fatalf("expected positive max-subagents to be valid: %v", err)
	}
	if err := validateDispatchMaxSubagents(hardDispatchMaxSubagents); err != nil {
		t.Fatalf("expected hard ceiling max-subagents to be valid: %v", err)
	}

	if err := validateDispatchMaxSubagents(0); err == nil {
		t.Fatalf("expected max-subagents validation to fail for zero")
	}
	if err := validateDispatchMaxSubagents(hardDispatchMaxSubagents + 1); err == nil {
		t.Fatalf("expected max-subagents validation to fail above hard ceiling")
	}
}
