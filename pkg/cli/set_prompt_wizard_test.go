package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

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
