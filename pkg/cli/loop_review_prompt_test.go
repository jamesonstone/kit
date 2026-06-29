package cli

import (
	"strings"
	"testing"
)

func TestBuildLoopReviewPromptDefaultsToSingleAgent(t *testing.T) {
	oldSingleAgent := singleAgent
	singleAgent = false
	t.Cleanup(func() {
		singleAgent = oldSingleAgent
	})

	prompt := buildLoopReviewPrompt(loopReviewOptions{MinConfidence: 95}, loopReviewTarget{
		BaseRef:      "origin/main",
		ChangedFiles: []string{"internal/app.go"},
		DiffStat:     " internal/app.go | 2 +-",
	}, nil, "")
	if strings.Contains(prompt, "## Subagent Orchestration") {
		t.Fatalf("did not expect subagent guidance by default:\n%s", prompt)
	}
	if strings.Contains(prompt, "## Review Subagent Pre-Analysis") {
		t.Fatalf("did not expect subagent pre-analysis by default:\n%s", prompt)
	}
	if !strings.Contains(prompt, "## Required Final Output") {
		t.Fatalf("expected final output contract in prompt:\n%s", prompt)
	}
	for _, want := range []string{
		"Use Kit RLM",
		"docs/CONSTITUTION.md",
		"docs/references/rules/*",
		"verify every finding against current code",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected prompt to contain %q:\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, "kit dispatch --pr") {
		t.Fatalf("did not expect generic loop review to include review-thread resolution instructions:\n%s", prompt)
	}
	if !strings.HasSuffix(strings.TrimSpace(prompt), "```") {
		t.Fatalf("expected final output contract to remain last section:\n%s", prompt)
	}
}

func TestBuildLoopReviewPromptIncludesResolutionGuidanceForPRFix(t *testing.T) {
	prompt := buildLoopReviewPrompt(
		loopReviewOptions{MinConfidence: 95, ResolvePRFeedback: true},
		loopReviewTarget{
			BaseRef:      "origin/main",
			ChangedFiles: []string{"internal/app.go"},
			DiffStat:     " internal/app.go | 2 +-",
		},
		&reviewLoopPRContext{
			Target: dispatchPRTarget{Owner: "jamesonstone", Repo: "kit", Number: 27},
			URL:    "https://github.com/jamesonstone/kit/pull/27",
		},
		"",
	)
	for _, want := range []string{
		"resolve all matching current unresolved review threads",
		"including human and CodeRabbit feedback",
		"kit dispatch --pr 27 --resolve --yes",
		"without `--coderabbit`",
		"Resolve only feedback you verified as fixed or intentionally no-op",
		"Review threads: <resolved count and skipped/remaining reason>.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected PR fix prompt to contain %q:\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, "Do not stage, commit, push, post PR comments, resolve review threads, or mutate GitHub.") {
		t.Fatalf("expected PR fix prompt to replace the blanket resolve prohibition:\n%s", prompt)
	}
}

func TestBuildLoopReviewPromptIncludesSubagentGuidanceWhenRequested(t *testing.T) {
	oldSingleAgent := singleAgent
	singleAgent = false
	t.Cleanup(func() {
		singleAgent = oldSingleAgent
	})

	prompt := buildLoopReviewPrompt(loopReviewOptions{MinConfidence: 95, UseSubagents: true}, loopReviewTarget{
		BaseRef:      "origin/main",
		ChangedFiles: []string{"internal/app.go"},
		DiffStat:     " internal/app.go | 2 +-",
	}, nil, "")
	for _, want := range []string{
		"## Subagent Orchestration",
		"## Review Subagent Pre-Analysis",
		"actual subagents spawned",
		"never exceed 4",
		"## Required Final Output",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected prompt to contain %q:\n%s", want, prompt)
		}
	}
	if !strings.HasSuffix(strings.TrimSpace(prompt), "```") {
		t.Fatalf("expected final output contract to remain last section:\n%s", prompt)
	}

	singleAgent = true
	prompt = buildLoopReviewPrompt(loopReviewOptions{MinConfidence: 95, UseSubagents: true}, loopReviewTarget{BaseRef: "origin/main"}, nil, "")
	if strings.Contains(prompt, "## Subagent Orchestration") || strings.Contains(prompt, "## Review Subagent Pre-Analysis") {
		t.Fatalf("did not expect subagent guidance with --single-agent:\n%s", prompt)
	}
}
