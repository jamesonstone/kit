package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
