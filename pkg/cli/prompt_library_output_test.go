package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/promptlib"
)

func TestOutputPromptLibraryPrompt_DefaultCopiesAndPrintsMetadataAndBody(t *testing.T) {
	previous := clipboardCopyFunc
	t.Cleanup(func() {
		clipboardCopyFunc = previous
	})

	var copied string
	clipboardCopyFunc = func(text string) error {
		copied = text
		return nil
	}

	prompt := promptlib.EffectivePrompt{
		Prompt: promptlib.Prompt{
			Identity:    promptlib.Identity{Noun: "coding-agent", Verb: "short"},
			Content:     "prompt text",
			Description: "short prompt",
		},
		Kind:     promptlib.SourceLocal,
		Location: "/repo/.kit.yaml",
		Shadowed: []promptlib.SourcePrompt{
			{Kind: promptlib.SourceGlobal},
			{Kind: promptlib.SourceBuiltin},
		},
	}

	output := captureStdout(t, func() {
		if err := outputPromptLibraryPrompt(prompt, false, false); err != nil {
			t.Fatalf("outputPromptLibraryPrompt() error = %v", err)
		}
	})

	if copied != "prompt text" {
		t.Fatalf("clipboard copy = %q, want prompt text", copied)
	}
	checks := []string{
		"Copied the prepared text to the clipboard.",
		"Prompt Library",
		"Command: coding-agent short",
		"Origin: local (/repo/.kit.yaml)",
		"Overrides: local overrides global, builtin",
		"---\nprompt text\n---",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q, got %q", check, output)
		}
	}
}

func TestOutputPromptLibraryPrompt_OutputOnlySkipsDefaultCopy(t *testing.T) {
	previous := clipboardCopyFunc
	t.Cleanup(func() {
		clipboardCopyFunc = previous
	})

	copied := false
	clipboardCopyFunc = func(text string) error {
		copied = true
		return nil
	}

	prompt := promptlib.EffectivePrompt{
		Prompt: promptlib.Prompt{
			Identity: promptlib.Identity{Noun: "coding-agent", Verb: "short"},
			Content:  "prompt text",
		},
		Kind: promptlib.SourceBuiltin,
	}

	output := captureStdout(t, func() {
		if err := outputPromptLibraryPrompt(prompt, true, false); err != nil {
			t.Fatalf("outputPromptLibraryPrompt() error = %v", err)
		}
	})

	if copied {
		t.Fatalf("expected output-only to skip default clipboard copy")
	}
	if output != "prompt text" {
		t.Fatalf("output = %q, want raw prompt text", output)
	}
}

func TestOutputPromptLibraryPrompt_OutputOnlyCopyDoesBoth(t *testing.T) {
	previous := clipboardCopyFunc
	t.Cleanup(func() {
		clipboardCopyFunc = previous
	})

	var copied string
	clipboardCopyFunc = func(text string) error {
		copied = text
		return nil
	}

	prompt := promptlib.EffectivePrompt{
		Prompt: promptlib.Prompt{
			Identity: promptlib.Identity{Noun: "coding-agent", Verb: "short"},
			Render: func() (string, error) {
				return "rendered prompt", nil
			},
		},
		Kind: promptlib.SourceBuiltin,
	}

	output := captureStdout(t, func() {
		if err := outputPromptLibraryPrompt(prompt, true, true); err != nil {
			t.Fatalf("outputPromptLibraryPrompt() error = %v", err)
		}
	})

	if copied != "rendered prompt" {
		t.Fatalf("clipboard copy = %q, want rendered prompt", copied)
	}
	if output != "rendered prompt" {
		t.Fatalf("output = %q, want rendered prompt", output)
	}
}

func TestOutputPromptLibraryPrompt_CopyFailureSkipsDefaultOutput(t *testing.T) {
	previous := clipboardCopyFunc
	t.Cleanup(func() {
		clipboardCopyFunc = previous
	})

	clipboardCopyFunc = func(text string) error {
		return errors.New("clipboard unavailable")
	}

	prompt := promptlib.EffectivePrompt{
		Prompt: promptlib.Prompt{
			Identity: promptlib.Identity{Noun: "coding-agent", Verb: "short"},
			Content:  "prompt text",
		},
		Kind: promptlib.SourceBuiltin,
	}

	output := captureStdout(t, func() {
		err := outputPromptLibraryPrompt(prompt, false, false)
		if err == nil {
			t.Fatalf("expected copy failure")
		}
		if !strings.Contains(err.Error(), "failed to copy prompt to clipboard") {
			t.Fatalf("unexpected error = %v", err)
		}
	})

	if output != "" {
		t.Fatalf("expected no default output when copy fails, got %q", output)
	}
}
