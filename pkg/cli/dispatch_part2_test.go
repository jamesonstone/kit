package cli

import (
	"strings"
	"testing"
)

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
