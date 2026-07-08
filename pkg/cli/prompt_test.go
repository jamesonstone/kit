package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPromptWithOptions_DirectBuiltInCopiesAndPrintsMetadata(t *testing.T) {
	setupPromptTestEnvironment(t)

	var copied string
	withClipboardCopy(t, func(text string) error {
		copied = text
		return nil
	})

	output := captureStdout(t, func() {
		if err := runPromptWithOptions([]string{"coding-agent", "short"}, false, false); err != nil {
			t.Fatalf("runPromptWithOptions() error = %v", err)
		}
	})

	if copied == "" {
		t.Fatalf("expected prompt to be copied")
	}
	if !strings.HasPrefix(copied, "---\n") {
		t.Fatalf("expected copied coding-agent prompt to start with instruction delimiter, got %q", copied)
	}
	if strings.Contains(copied, "pbcopy") || strings.Contains(copied, "osascript") {
		t.Fatalf("expected copied prompt body to exclude wrapper commands, got %q", copied)
	}
	checks := []string{
		"Copied the prepared text to the clipboard.",
		"Prompt Library",
		"Command: coding-agent short",
		"Origin: builtin (built-in)",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q, got %q", check, output)
		}
	}
}

func TestRunPromptWithOptions_DirectCodingAgentBuiltIns(t *testing.T) {
	tests := []struct {
		verb      string
		wantText  string
		rejectOne string
	}{
		{verb: "short", wantText: "Clarify before implementing.", rejectOne: "pbcopy"},
		{verb: "long", wantText: "Stay in clarification and information-gathering workflow.", rejectOne: "osascript"},
		{verb: "instructions", wantText: "Output a concise, comprehensive set of coding agent instructions.", rejectOne: "old_clipboard"},
	}

	for _, tt := range tests {
		t.Run(tt.verb, func(t *testing.T) {
			setupPromptTestEnvironment(t)
			output := captureStdout(t, func() {
				if err := runPromptWithOptions([]string{"coding-agent", tt.verb}, true, false); err != nil {
					t.Fatalf("runPromptWithOptions() error = %v", err)
				}
			})
			if !strings.HasPrefix(output, "---\n") {
				t.Fatalf("expected output to start with instruction delimiter, got %q", output)
			}
			if !strings.Contains(output, tt.wantText) {
				t.Fatalf("expected output to contain %q, got %q", tt.wantText, output)
			}
			if strings.Contains(output, tt.rejectOne) {
				t.Fatalf("expected output to exclude %q, got %q", tt.rejectOne, output)
			}
		})
	}
}

func TestRunPromptWithOptions_DirectKitSpecBuiltInRendersV2Prompt(t *testing.T) {
	projectRoot := newPromptContextProject(t)
	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-alpha")
	writePromptContextFeatureDocs(t, featurePath, true)
	setWorkingDirectory(t, projectRoot)

	output := captureStdout(t, func() {
		if err := runPromptWithOptions([]string{"kit", "spec"}, true, false); err != nil {
			t.Fatalf("runPromptWithOptions() error = %v", err)
		}
	})

	assertV2SpecPromptContract(t, output)
	assertV2SpecPromptExcludesV1StageAssumptions(t, output)
	if !strings.Contains(output, "Kit v2 `kit spec` workflow for feature `alpha`") {
		t.Fatalf("expected kit spec prompt output, got %q", output)
	}
}

func TestRunPromptWithOptions_OutputOnlyCopyCopiesRawBuiltIn(t *testing.T) {
	setupPromptTestEnvironment(t)

	var copied string
	withClipboardCopy(t, func(text string) error {
		copied = text
		return nil
	})

	output := captureStdout(t, func() {
		if err := runPromptWithOptions([]string{"coding-agent", "short"}, true, true); err != nil {
			t.Fatalf("runPromptWithOptions() error = %v", err)
		}
	})

	if copied == "" {
		t.Fatalf("expected prompt to be copied")
	}
	if copied != output {
		t.Fatalf("clipboard payload = %q, want stdout %q", copied, output)
	}
	if !strings.HasPrefix(output, "---\n") {
		t.Fatalf("expected output to start with instruction delimiter, got %q", output)
	}
}

func TestRunPromptWithOptions_OutputOnlyPrintsRawBuiltIn(t *testing.T) {
	setupPromptTestEnvironment(t)

	copied := false
	withClipboardCopy(t, func(text string) error {
		copied = true
		return nil
	})

	output := captureStdout(t, func() {
		if err := runPromptWithOptions([]string{"coding-agent", "short"}, true, false); err != nil {
			t.Fatalf("runPromptWithOptions() error = %v", err)
		}
	})

	if copied {
		t.Fatalf("expected --output-only to skip clipboard copy")
	}
	if output == "" {
		t.Fatalf("expected raw prompt output")
	}
	if !strings.HasPrefix(output, "---\n") {
		t.Fatalf("expected raw coding-agent output to start with instruction delimiter, got %q", output)
	}
	if strings.Contains(output, "Prompt Library") || strings.Contains(output, "Command:") {
		t.Fatalf("expected raw output without metadata, got %q", output)
	}
}

func TestRunPromptWithOptions_DirectNoMatchSuggestsNearestVerb(t *testing.T) {
	setupPromptTestEnvironment(t)

	err := runPromptWithOptions([]string{"coding-agent", "shrt"}, true, false)
	if err == nil {
		t.Fatalf("expected no-match error")
	}
	if !strings.Contains(err.Error(), "nearest verbs for \"coding-agent\": short") {
		t.Fatalf("expected nearest verb suggestion, got %v", err)
	}
}

func TestRunPromptWithOptions_DirectNoMatchSuggestsNearestNoun(t *testing.T) {
	setupPromptTestEnvironment(t)

	err := runPromptWithOptions([]string{"coding-agnt", "short"}, true, false)
	if err == nil {
		t.Fatalf("expected no-match error")
	}
	if !strings.Contains(err.Error(), "nearest nouns: coding-agent") {
		t.Fatalf("expected nearest noun suggestion, got %v", err)
	}
}

func TestRunPromptWithOptions_SelectsNounAndVerbDeterministically(t *testing.T) {
	setupPromptTestEnvironment(t)

	var copied string
	withClipboardCopy(t, func(text string) error {
		copied = text
		return nil
	})

	output := withStdin(t, "1\n3\n", func() string {
		return captureStdout(t, func() {
			if err := runPromptWithOptions(nil, false, false); err != nil {
				t.Fatalf("runPromptWithOptions() error = %v", err)
			}
		})
	})

	if copied == "" {
		t.Fatalf("expected selected prompt to be copied")
	}
	checks := []string{
		"Select a prompt noun:",
		"[1] coding-agent",
		"Select a prompt verb for coding-agent:",
		"[3] short",
		"Short clarification-before-implementation prompt.",
		"Command: coding-agent short",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q, got %q", check, output)
		}
	}
}
