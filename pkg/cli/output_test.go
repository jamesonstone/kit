package cli

import (
	"strings"
	"testing"
)

func TestFormatAgentInstructionBlock(t *testing.T) {
	block := formatAgentInstructionBlock("line one\nline two\n")
	want := "---\nline one\nline two\n---\n"

	if block != want {
		t.Fatalf("formatAgentInstructionBlock() = %q, want %q", block, want)
	}
}

func TestFormatAgentInstructionBlockEmptyInput(t *testing.T) {
	block := formatAgentInstructionBlock("")
	want := "---\n\n---\n"

	if block != want {
		t.Fatalf("formatAgentInstructionBlock() = %q, want %q", block, want)
	}
}

func TestFormatAgentInstructionBlockWhitespaceOnly(t *testing.T) {
	block := formatAgentInstructionBlock("   \t\n")
	want := "---\n   \t\n---\n"

	if block != want {
		t.Fatalf("formatAgentInstructionBlock() = %q, want %q", block, want)
	}
}

func TestFormatAgentInstructionBlockMultipleNewlines(t *testing.T) {
	block := formatAgentInstructionBlock("a\n\nb\n")
	want := "---\na\n\nb\n---\n"

	if block != want {
		t.Fatalf("formatAgentInstructionBlock() = %q, want %q", block, want)
	}
}

func TestFormatAgentInstructionBlockAddsTrailingNewline(t *testing.T) {
	block := formatAgentInstructionBlock("line one")
	want := "---\nline one\n---\n"

	if block != want {
		t.Fatalf("formatAgentInstructionBlock() = %q, want %q", block, want)
	}
}

func TestWritePromptWithClipboardDefault_CopiesAndAcknowledges(t *testing.T) {
	previous := clipboardCopyFunc
	defer func() {
		clipboardCopyFunc = previous
	}()

	var copied string
	clipboardCopyFunc = func(text string) error {
		copied = text
		return nil
	}

	output := captureStdout(t, func() {
		if err := writePromptWithClipboardDefault("prompt text", false, false); err != nil {
			t.Fatalf("writePromptWithClipboardDefault() error = %v", err)
		}
	})

	if copied != "prompt text" {
		t.Fatalf("expected clipboard copy %q, got %q", "prompt text", copied)
	}

	if output != "Copied the prepared text to the clipboard.\n" {
		t.Fatalf("unexpected stdout: %q", output)
	}
}

func TestWritePromptWithClipboardDefault_OutputOnlySkipsDefaultCopy(t *testing.T) {
	previous := clipboardCopyFunc
	defer func() {
		clipboardCopyFunc = previous
	}()

	copied := false
	clipboardCopyFunc = func(text string) error {
		copied = true
		return nil
	}

	output := captureStdout(t, func() {
		if err := writePromptWithClipboardDefault("prompt text", true, false); err != nil {
			t.Fatalf("writePromptWithClipboardDefault() error = %v", err)
		}
	})

	if copied {
		t.Fatalf("expected output-only mode to skip clipboard copy")
	}

	if output != "prompt text" {
		t.Fatalf("unexpected stdout: %q", output)
	}
}

func TestWritePromptWithClipboardDefault_OutputOnlyAndCopyDoesBoth(t *testing.T) {
	previous := clipboardCopyFunc
	defer func() {
		clipboardCopyFunc = previous
	}()

	var copied string
	clipboardCopyFunc = func(text string) error {
		copied = text
		return nil
	}

	output := captureStdout(t, func() {
		if err := writePromptWithClipboardDefault("prompt text", true, true); err != nil {
			t.Fatalf("writePromptWithClipboardDefault() error = %v", err)
		}
	})

	if copied != "prompt text" {
		t.Fatalf("expected clipboard copy %q, got %q", "prompt text", copied)
	}

	if output != "prompt text" {
		t.Fatalf("unexpected stdout: %q", output)
	}
}

func TestOutputPromptWithoutSubagentsWithClipboardDefault_SkipsSubagentSuffix(t *testing.T) {
	previousCopy := clipboardCopyFunc
	previousSingleAgent := singleAgent
	defer func() {
		clipboardCopyFunc = previousCopy
		singleAgent = previousSingleAgent
	}()

	var copied string
	clipboardCopyFunc = func(text string) error {
		copied = text
		return nil
	}
	singleAgent = false

	output := captureStdout(t, func() {
		if err := outputPromptWithoutSubagentsWithClipboardDefault("prompt text", false, false); err != nil {
			t.Fatalf("outputPromptWithoutSubagentsWithClipboardDefault() error = %v", err)
		}
	})

	if strings.Contains(copied, "## Subagent Orchestration") {
		t.Fatalf("expected dispatch-style helper to skip subagent suffix, got %q", copied)
	}

	if output != "Copied the prepared text to the clipboard.\n" {
		t.Fatalf("unexpected stdout: %q", output)
	}
}

func TestOutputPromptWithoutSubagentsWithFrontendProfileKeepsFrontendSuffix(t *testing.T) {
	previousCopy := clipboardCopyFunc
	previousSingleAgent := singleAgent
	defer func() {
		clipboardCopyFunc = previousCopy
		singleAgent = previousSingleAgent
	}()
	restorePromptProfileState(t, promptProfileFrontend, true)

	var copied string
	clipboardCopyFunc = func(text string) error {
		copied = text
		return nil
	}
	singleAgent = false

	_ = captureStdout(t, func() {
		if err := outputPromptWithoutSubagentsWithClipboardDefault("prompt text", false, false); err != nil {
			t.Fatalf("outputPromptWithoutSubagentsWithClipboardDefault() error = %v", err)
		}
	})

	if !strings.Contains(copied, "## Frontend Profile") {
		t.Fatalf("expected frontend profile suffix, got %q", copied)
	}
	if strings.Contains(copied, "## Subagent Orchestration") {
		t.Fatalf("expected dispatch-style helper to skip subagent suffix, got %q", copied)
	}
}

func TestHelpTemplateIncludesHumanReadableHeadings(t *testing.T) {
	got := helpTemplate(true)

	checks := []string{
		"🚀 Usage",
		"🧰 Available Commands",
		"⚙️ Flags",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("expected help template to contain %q", check)
		}
	}
}
