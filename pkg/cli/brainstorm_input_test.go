package cli

import (
	"io"
	"os"
	"testing"
)

func TestPromptBrainstormThesis_UsesEditorByDefault(t *testing.T) {
	previousWait := awaitEditorLaunchConfirmation
	previousRunner := editorInputRunner
	defer func() {
		awaitEditorLaunchConfirmation = previousWait
		editorInputRunner = previousRunner
	}()

	waitCalls := 0
	runCalls := 0
	awaitEditorLaunchConfirmation = func(_ *os.File, _ io.Writer) error {
		waitCalls++
		return nil
	}
	editorInputRunner = func(_ freeTextInputConfig, fieldName, _ string) (string, bool, error) {
		runCalls++
		if fieldName != "brainstorm thesis" {
			t.Fatalf("unexpected field name %q", fieldName)
		}
		return "captured thesis", true, nil
	}

	got, err := promptBrainstormThesis(newFreeTextInputConfig(false, "", false, true))
	if err != nil {
		t.Fatalf("promptBrainstormThesis() error = %v", err)
	}

	if got != "captured thesis" {
		t.Fatalf("expected captured thesis, got %q", got)
	}
	if waitCalls != 1 || runCalls != 1 {
		t.Fatalf("expected one editor launch, got wait=%d run=%d", waitCalls, runCalls)
	}
}
