package prfix

import (
	"os"
	"strings"
	"testing"
)

func TestPRFixPromptGolden(t *testing.T) {
	target := Target{Repository{"acme", "app"}, 9}
	lane := Lane{
		Repository: "acme/app", PRNumber: 9, PRURL: target.URL(), HeadBranch: "GH-9",
		ExpectedHead: "abc123", LocalHead: "abc123", WorktreePath: "/tmp/acme/app/GH-9",
		PushTarget: "origin/GH-9", DirtyOwnership: "exclude", DirtyPaths: []string{"local.txt"},
	}
	feedback := []Feedback{{
		Kind: "review-thread", ThreadID: "PRRT_1", Path: "internal/app.go", Line: 12,
		Author: "reviewer", URL: "https://example.com/thread", Task: "Fix the valid issue.",
		Fingerprint: "sha256:one",
	}}
	tasks := RenderFeedback(feedback)
	got, err := BuildPrompt(target, lane, feedback, tasks, 3)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/pr-fix.golden.md")
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("prompt changed\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
	for _, required := range []string{
		"kit contract resolve --workflow pr-feedback-repair --json",
		"Agent Team Plan", "hard ceiling is 4", "read-only verification lane",
		"Verify every finding against current `HEAD`", "full pushed diff",
		"kit pr fix --pr", "--resolve", "--thread \"PRRT_1\"",
	} {
		if !strings.Contains(got, required) {
			t.Errorf("prompt missing %q", required)
		}
	}
}

func TestPromptBoundsSubagentsAndResolutionIDs(t *testing.T) {
	if err := ValidateMaxSubagents(0); err == nil {
		t.Fatal("zero subagents accepted")
	}
	if err := ValidateMaxSubagents(5); err == nil {
		t.Fatal("hard ceiling exceeded")
	}
	items := []Feedback{{ThreadID: "B"}, {ThreadID: "A"}, {ThreadID: "B"}, {Kind: "review"}}
	if got := strings.Join(ResolutionThreadIDs(items), ","); got != "A,B" {
		t.Fatalf("thread IDs = %s", got)
	}
}
