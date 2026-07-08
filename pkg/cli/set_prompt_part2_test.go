package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestRunSetPromptWithOptions_EditorCancelFailsWithoutSaving(t *testing.T) {
	setupPromptTestEnvironment(t)
	stubSetPromptEditorResult(t, "", false)

	err := runSetPromptWithOptions([]string{"custom", "review"}, false, true)
	if err == nil {
		t.Fatalf("expected editor cancel error")
	}
	if !strings.Contains(err.Error(), "prompt content entry cancelled") {
		t.Fatalf("unexpected error = %v", err)
	}
	if _, found, loadErr := config.LoadGlobal(); loadErr != nil || found {
		t.Fatalf("expected no global config, found = %v err = %v", found, loadErr)
	}
}

func TestRunSetPromptWithOptions_EditorEmptyContentFailsWithoutSaving(t *testing.T) {
	setupPromptTestEnvironment(t)
	stubSetPromptEditorResult(t, "", true)

	err := runSetPromptWithOptions([]string{"custom", "review"}, false, true)
	if err == nil {
		t.Fatalf("expected empty content error")
	}
	if !strings.Contains(err.Error(), "prompt content cannot be empty") {
		t.Fatalf("unexpected error = %v", err)
	}
	if _, found, loadErr := config.LoadGlobal(); loadErr != nil || found {
		t.Fatalf("expected no global config, found = %v err = %v", found, loadErr)
	}
}

func TestRunSetPromptWithOptions_MissingEditorFails(t *testing.T) {
	setupPromptTestEnvironment(t)
	t.Setenv("EDITOR", "")
	restoreLookPath := stubLookPath(map[string]string{})
	defer restoreLookPath()

	previousWait := awaitEditorLaunchConfirmation
	t.Cleanup(func() {
		awaitEditorLaunchConfirmation = previousWait
	})
	awaitEditorLaunchConfirmation = func(_ *os.File, _ io.Writer) error {
		return nil
	}

	err := runSetPromptWithOptions([]string{"custom", "review"}, false, true)
	if err == nil {
		t.Fatalf("expected missing editor error")
	}
	if !strings.Contains(err.Error(), "no vim-compatible editor found") {
		t.Fatalf("unexpected error = %v", err)
	}
}

func TestRunSetPromptWithOptions_WizardNormalizesAndStoresDescription(t *testing.T) {
	projectRoot, _ := setupPromptTestProject(t)
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	stubSetPromptEditor(t, "wizard body")

	output := withStdin(t, "Custom Noun\nReview Flow\nreview description\n", func() string {
		return captureStdout(t, func() {
			if err := runSetPromptWithOptions(nil, false, false); err != nil {
				t.Fatalf("runSetPromptWithOptions() error = %v", err)
			}
		})
	})

	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	prompt := cfg.Prompts["custom-noun"]["review-flow"]
	if prompt.Content != "wizard body" {
		t.Fatalf("prompt content = %q, want wizard body", prompt.Content)
	}
	if prompt.Description != "review description" {
		t.Fatalf("prompt description = %q, want review description", prompt.Description)
	}
	if !strings.Contains(output, "Prompt noun: ") || !strings.Contains(output, "Description (optional): ") {
		t.Fatalf("expected wizard prompts, got %q", output)
	}
}

func stubSetPromptEditor(t *testing.T, content string) *int {
	return stubSetPromptEditorResult(t, content, true)
}

func stubSetPromptEditorResult(t *testing.T, content string, changed bool) *int {
	t.Helper()

	previousWait := awaitEditorLaunchConfirmation
	previousRunner := editorInputRunner
	calls := 0

	t.Cleanup(func() {
		awaitEditorLaunchConfirmation = previousWait
		editorInputRunner = previousRunner
	})

	awaitEditorLaunchConfirmation = func(_ *os.File, _ io.Writer) error {
		return nil
	}
	editorInputRunner = func(_ freeTextInputConfig, _ string, _ string) (string, bool, error) {
		calls++
		return content, changed, nil
	}

	return &calls
}
