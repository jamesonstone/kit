package cli

import (
	"strings"
	"testing"
)

func TestExtractRelevantCILogExcerptRedactsAndLimits(t *testing.T) {
	raw := strings.Join([]string{
		"line 1",
		"line 2",
		"TOKEN=super-secret-value",
		"line 4",
		"line 5",
		"Error: build failed",
		"ghp_abcd1234",
		"line 8",
		"line 9",
	}, "\n")
	excerpt, truncated := extractRelevantCILogExcerpt(raw, 8)
	if !truncated {
		t.Fatal("expected log to be truncated")
	}
	joined := strings.Join(excerpt, "\n")
	if strings.Contains(joined, "super-secret-value") || strings.Contains(joined, "ghp_abcd1234") {
		t.Fatalf("expected secrets to be redacted, got:\n%s", joined)
	}
	if !strings.Contains(joined, "TOKEN=[REDACTED]") || !strings.Contains(joined, "Error: build failed") {
		t.Fatalf("expected relevant redacted lines, got:\n%s", joined)
	}
}

func TestOpenCIDispatchPromptUsesEditorInitialContent(t *testing.T) {
	previousEditor := editorInputRunner
	previousClipboard := clipboardCopyFunc
	defer func() {
		editorInputRunner = previousEditor
		clipboardCopyFunc = previousClipboard
	}()
	var initial string
	editorInputRunner = func(_ freeTextInputConfig, _ string, initialContent string) (string, bool, error) {
		initial = initialContent
		return initialContent, false, nil
	}
	var copied string
	clipboardCopyFunc = func(text string) error {
		copied = text
		return nil
	}

	diagnosis := ciDiagnosis{
		Target: ciTarget{Repository: "jamesonstone/kit", Kind: "run", RunID: 101},
		Runs: []ciRunFailure{{
			RunID:    101,
			Workflow: "Tests",
			FailedJobs: []ciJobFailure{{
				Name:       "test",
				LogExcerpt: []string{"Error: expected nil"},
			}},
		}},
		FailureFound:   true,
		RootCause:      "The first relevant failing log line is: Error: expected nil",
		Recommendation: "Fix the test failure.",
	}
	diagnosis.AgentPrompt = buildCIAgentPrompt(diagnosis)

	err := openCIDispatchPrompt(
		ciOptions{InputConfig: newFreeTextInputConfig(false, "", false, true)},
		diagnosis,
	)
	if err != nil {
		t.Fatalf("openCIDispatchPrompt() error = %v", err)
	}
	if !strings.Contains(initial, "Error: expected nil") {
		t.Fatalf("expected editor initial content to include evidence, got:\n%s", initial)
	}
	if !strings.Contains(copied, "Prepare an Agent Team Plan") || !strings.Contains(copied, "Max concurrent subagents: 3") {
		t.Fatalf("expected dispatch prompt copied, got:\n%s", copied)
	}
}
