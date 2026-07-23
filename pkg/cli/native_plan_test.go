package cli

import (
	"errors"
	"strings"
	"testing"
)

func TestBuildPlanChallengePromptMapsReviewToCodexPlanControls(t *testing.T) {
	plan := "1. Inspect the service.\n2. Add the endpoint.\n3. Run tests."
	prompt := buildPlanChallengePrompt(plan)

	for _, want := range []string{
		"# Adversarial Plan Challenge",
		"candidate implementation plan generated",
		"by Codex for Mac",
		"Review the plan; do not implement it.",
		"Challenge only material issues:",
		"misunderstood goal or incomplete observable acceptance",
		"missing edge cases, failure modes, dependencies, migrations, or rollback",
		"validation that would not prove the requested outcome",
		"IMPLEMENT THIS PLAN",
		"TELL CODEX WHAT TO DO DIFFERENT:",
		"suitable for pasting into Codex's",
		`do different" field`,
		"<candidate-codex-plan>",
		plan,
		"</candidate-codex-plan>",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("challenge prompt missing %q:\n%s", want, prompt)
		}
	}
}

func TestRunPlanChallengeCopiesSupplementedPromptByDefault(t *testing.T) {
	var copied string
	withPlanChallengeClipboard(
		t,
		func() (string, error) { return "Inspect, implement, and validate.", nil },
		func(text string) error {
			copied = text
			return nil
		},
	)

	output := captureStdout(t, func() {
		if err := runPlanChallenge(planChallengeOptions{}); err != nil {
			t.Fatalf("runPlanChallenge() error = %v", err)
		}
	})

	if copied != buildPlanChallengePrompt("Inspect, implement, and validate.") {
		t.Fatalf("copied prompt did not match composed challenge:\n%s", copied)
	}
	if !strings.Contains(output, "Copied the prepared text to the clipboard.") {
		t.Fatalf("expected clipboard acknowledgement, got %q", output)
	}
}

func TestRunPlanChallengeOutputOnlyDoesNotReplaceClipboard(t *testing.T) {
	const plan = "Preserve current behavior."
	copyCalled := false
	withPlanChallengeClipboard(
		t,
		func() (string, error) { return plan, nil },
		func(string) error {
			copyCalled = true
			return nil
		},
	)

	output := captureStdout(t, func() {
		if err := runPlanChallenge(planChallengeOptions{OutputOnly: true}); err != nil {
			t.Fatalf("runPlanChallenge() error = %v", err)
		}
	})

	if copyCalled {
		t.Fatal("--output-only replaced the clipboard")
	}
	if output != buildPlanChallengePrompt(plan) {
		t.Fatalf("raw output did not match challenge prompt:\n%s", output)
	}
}

func TestRunPlanChallengeOutputOnlyCopyWritesAndPrints(t *testing.T) {
	const plan = "Add focused coverage."
	var copied string
	withPlanChallengeClipboard(
		t,
		func() (string, error) { return plan, nil },
		func(text string) error {
			copied = text
			return nil
		},
	)

	output := captureStdout(t, func() {
		if err := runPlanChallenge(planChallengeOptions{OutputOnly: true, Copy: true}); err != nil {
			t.Fatalf("runPlanChallenge() error = %v", err)
		}
	})

	want := buildPlanChallengePrompt(plan)
	if copied != want || output != want {
		t.Fatalf("dual output mismatch:\ncopied=%q\noutput=%q", copied, output)
	}
}

func TestRunPlanChallengeRejectsEmptyClipboard(t *testing.T) {
	copyCalled := false
	withPlanChallengeClipboard(
		t,
		func() (string, error) { return " \n\t", nil },
		func(string) error {
			copyCalled = true
			return nil
		},
	)

	err := runPlanChallenge(planChallengeOptions{})
	if err == nil || !strings.Contains(err.Error(), "copy the complete Codex /plan output") {
		t.Fatalf("expected actionable empty-clipboard error, got %v", err)
	}
	if copyCalled {
		t.Fatal("empty clipboard attempted to write a challenge prompt")
	}
}

func TestRunPlanChallengeReportsClipboardReadFailure(t *testing.T) {
	readErr := errors.New("pbpaste unavailable")
	withPlanChallengeClipboard(
		t,
		func() (string, error) { return "", readErr },
		func(string) error { return nil },
	)

	err := runPlanChallenge(planChallengeOptions{})
	if !errors.Is(err, readErr) || !strings.Contains(err.Error(), "read copied plan from clipboard") {
		t.Fatalf("expected wrapped clipboard read error, got %v", err)
	}
}

func TestRunPlanChallengeReportsClipboardWriteFailure(t *testing.T) {
	writeErr := errors.New("pbcopy unavailable")
	withPlanChallengeClipboard(
		t,
		func() (string, error) { return "Candidate plan", nil },
		func(string) error { return writeErr },
	)

	err := runPlanChallenge(planChallengeOptions{})
	if !errors.Is(err, writeErr) || !strings.Contains(err.Error(), "failed to copy to clipboard") {
		t.Fatalf("expected wrapped clipboard write error, got %v", err)
	}
}

func TestNativePlanCommandRegistersChallenge(t *testing.T) {
	cmd := newNativePlanCommand()
	challenge, _, err := cmd.Find([]string{"challenge"})
	if err != nil {
		t.Fatalf("find challenge command: %v", err)
	}
	if challenge.CommandPath() != "plan challenge" {
		t.Fatalf("challenge command path = %q, want %q", challenge.CommandPath(), "plan challenge")
	}
	if challenge.Flag("output-only") == nil || challenge.Flag("copy") == nil {
		t.Fatal("challenge command did not register clipboard output flags")
	}
}

func withPlanChallengeClipboard(
	t *testing.T,
	read func() (string, error),
	copy func(string) error,
) {
	t.Helper()

	previousRead := clipboardReadFunc
	previousCopy := clipboardCopyFunc
	clipboardReadFunc = read
	clipboardCopyFunc = copy
	t.Cleanup(func() {
		clipboardReadFunc = previousRead
		clipboardCopyFunc = previousCopy
	})
}
