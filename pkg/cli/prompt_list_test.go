package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
