package cli

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestBuildBrainstormPrompt(t *testing.T) {
	prompt := buildBrainstormPrompt(
		"/tmp/docs/specs/0001-sample/BRAINSTORM.md",
		"sample-feature",
		"/tmp/project",
		"Need better import validation for malformed CSV uploads.",
		95,
	)

	checks := []string{
		"/plan",
		"You are in planning mode for feature: **sample-feature**",
		"Do NOT implement code, write production changes, or move into execution",
		"Ask clarifying questions until you reach ≥95% confidence that you understand the problem and desired solution",
		"Use numbered lists",
		"Ask questions in batches of up to 10",
		"For every question, include your current best recommended default, proposed solution, or assumption",
		"State uncertainties",
		"\"yes\" or \"y\" approves all recommended defaults in the batch",
		"\"yes 3, 4, 5\" or \"y 3, 4, 5\" approves only those numbered defaults in the batch",
		"If the user approves only specific question numbers, treat all other questions in that batch as unresolved",
		"After each batch of up to 10 questions, output your current percentage understanding so the user can see progress",
		"planning only — no implementation",
		"kit spec sample-feature",
		"/tmp/docs/specs/0001-sample/BRAINSTORM.md",
		"/tmp/project/docs/CONSTITUTION.md",
		"## DEPENDENCIES",
		"`Dependency`, `Type`, `Location`, `Used For`, and `Status`",
		"for Figma or other MCP-driven design dependencies, store the exact design URL or file/node reference in `Location`",
		"`Status` = `stale`",
		"no section in `BRAINSTORM.md` may remain empty or contain only an HTML TODO comment",
		"`not applicable`, `not required`, or `no additional information required`",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected prompt to contain %q", check)
		}
	}

	if !strings.HasPrefix(prompt, "/plan\n\n") {
		t.Fatalf("expected prompt to start with /plan, got %q", prompt[:8])
	}
}

func TestPromptBrainstormThesis_UsesEditorByDefault(t *testing.T) {
	previousWait := awaitEditorLaunchConfirmation
	previousRunner := editorInputRunner
	defer func() {
		awaitEditorLaunchConfirmation = previousWait
		editorInputRunner = previousRunner
	}()

	waitCalls := 0
	runCalls := 0
	awaitEditorLaunchConfirmation = func(_ *os.File, _ io.Writer) error {
		waitCalls++
		return nil
	}
	editorInputRunner = func(_ freeTextInputConfig, fieldName, _ string) (string, bool, error) {
		runCalls++
		if fieldName != "brainstorm thesis" {
			t.Fatalf("unexpected field name %q", fieldName)
		}
		return "captured thesis", true, nil
	}

	got, err := promptBrainstormThesis(newFreeTextInputConfig(false, "", false, true))
	if err != nil {
		t.Fatalf("promptBrainstormThesis() error = %v", err)
	}

	if got != "captured thesis" {
		t.Fatalf("expected captured thesis, got %q", got)
	}
	if waitCalls != 1 || runCalls != 1 {
		t.Fatalf("expected one editor launch, got wait=%d run=%d", waitCalls, runCalls)
	}
}
