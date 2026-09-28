package cli

import (
	"strings"
	"testing"
)

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
