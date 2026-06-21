package cli

import (
	"os"
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

func TestRunPromptWithOptions_SelectsVerbForProvidedNoun(t *testing.T) {
	setupPromptTestEnvironment(t)

	var copied string
	withClipboardCopy(t, func(text string) error {
		copied = text
		return nil
	})

	output := withStdin(t, "3\n", func() string {
		return captureStdout(t, func() {
			if err := runPromptWithOptions([]string{"coding-agent"}, false, false); err != nil {
				t.Fatalf("runPromptWithOptions() error = %v", err)
			}
		})
	})

	if copied == "" {
		t.Fatalf("expected selected prompt to be copied")
	}
	if strings.Contains(output, "Select a prompt noun:") {
		t.Fatalf("expected noun argument to skip noun selector, got %q", output)
	}
	if !strings.Contains(output, "Command: coding-agent short") {
		t.Fatalf("expected selected prompt output, got %q", output)
	}
}

func TestRunPromptList_RendersEffectiveBuiltIns(t *testing.T) {
	setupPromptTestEnvironment(t)

	output := captureStdout(t, func() {
		if err := runPromptList(promptListCmd, nil); err != nil {
			t.Fatalf("runPromptList() error = %v", err)
		}
	})

	checks := []string{
		"Prompt Library",
		"COMMAND",
		"DESCRIPTION",
		"ORIGIN",
		"OVERRIDES",
		"coding-agent short",
		"kit spec",
		"workflow spec",
		"support dispatch",
		"none",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q, got %q", check, output)
		}
	}
	if strings.Index(output, "coding-agent instructions") > strings.Index(output, "project init") {
		t.Fatalf("expected prompt list to be sorted by noun then verb, got %q", output)
	}
	for _, removed := range []string{"workflow brainstorm", "workflow plan", "workflow tasks", "workflow implement", "workflow reflect"} {
		if strings.Contains(output, removed) {
			t.Fatalf("expected prompt list to omit removed v1 workflow prompt %q, got %q", removed, output)
		}
	}
}

func TestRunPromptList_RendersShadowMetadata(t *testing.T) {
	projectRoot, homeDir := setupPromptTestProject(t)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	writeFile(t, filepath.Join(homeDir, ".config", "kit", ".kit.yaml"), `prompts:
  coding-agent:
    short:
      content: global prompt
      description: global short
`)
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig()+`prompts:
  coding-agent:
    short:
      content: local prompt
      description: local short
`)

	output := captureStdout(t, func() {
		if err := runPromptList(promptListCmd, nil); err != nil {
			t.Fatalf("runPromptList() error = %v", err)
		}
	})

	checks := []string{
		"coding-agent short",
		"local short",
		"local (" + filepath.Join(cwd, ".kit.yaml") + ")",
		"local overrides global, builtin",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q, got %q", check, output)
		}
	}
}

func setupPromptTestEnvironment(t *testing.T) {
	t.Helper()

	tempDir := t.TempDir()
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	setWorkingDirectory(t, tempDir)
}

func setupPromptTestProject(t *testing.T) (string, string) {
	t.Helper()

	projectRoot := t.TempDir()
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	setWorkingDirectory(t, projectRoot)
	return projectRoot, homeDir
}

func withClipboardCopy(t *testing.T, copyFunc func(string) error) {
	t.Helper()

	previous := clipboardCopyFunc
	t.Cleanup(func() {
		clipboardCopyFunc = previous
	})
	clipboardCopyFunc = copyFunc
}

func TestReadPromptSelection_RejectsInvalidInput(t *testing.T) {
	previous := os.Stdin
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	if _, err := writer.WriteString("9\n"); err != nil {
		t.Fatalf("writer.WriteString() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	os.Stdin = reader
	t.Cleanup(func() {
		os.Stdin = previous
		_ = reader.Close()
	})

	if _, err := readPromptSelection(3); err == nil {
		t.Fatalf("expected invalid selection error")
	}
}
