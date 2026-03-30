package cli

import "testing"

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

	if output != "Copied agent instructions to the clipboard.\n" {
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
