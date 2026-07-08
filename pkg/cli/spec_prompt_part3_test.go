package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

func TestRunSpecReviseThesisAppendsDatedNoteAndDeliveryIntent(t *testing.T) {
	t.Setenv("EDITOR", "")
	projectRoot, _ := setupLifecycleTestProject(t)
	restore := chdirForTest(t, projectRoot)
	defer restore()
	restoreSpecFlags := restoreSpecFlagState()
	defer restoreSpecFlags()

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-sample")
	specPath := filepath.Join(featurePath, "SPEC.md")
	writeFile(t, specPath, "# SPEC\n\n## THESIS\n\nOriginal thesis\n\n## DELIVERY DECISION\n\nOriginal delivery decision\n")

	previousWait := awaitEditorLaunchConfirmation
	previousRunner := editorInputRunner
	previousDeliveryPrompt := promptSpecDeliveryIntent
	defer func() {
		awaitEditorLaunchConfirmation = previousWait
		editorInputRunner = previousRunner
		promptSpecDeliveryIntent = previousDeliveryPrompt
	}()
	awaitEditorLaunchConfirmation = func(_ *os.File, _ io.Writer) error {
		return nil
	}
	editorInputRunner = func(_ freeTextInputConfig, fieldName, _ string) (string, bool, error) {
		if fieldName != "feature thesis" {
			t.Fatalf("fieldName = %q, want feature thesis", fieldName)
		}
		return "Revised thesis", true, nil
	}
	promptSpecDeliveryIntent = func() (string, error) {
		return specDeliveryIntentContinueCurrent, nil
	}

	cmd := newSpecProfileTestCommand()
	if err := cmd.Flags().Set("output-only", "true"); err != nil {
		t.Fatalf("Flags().Set(output-only) error = %v", err)
	}
	if err := cmd.Flags().Set("revise-thesis", "true"); err != nil {
		t.Fatalf("Flags().Set(revise-thesis) error = %v", err)
	}

	output := captureStdout(t, func() {
		if err := runSpec(cmd, []string{"sample"}); err != nil {
			t.Fatalf("runSpec() error = %v", err)
		}
	})

	for _, check := range []string{
		"Spec Thesis",
		"**THESIS**: Revised thesis",
		"**DELIVERY INTENT**: continue - coding agent should continue on the current branch/current issue/current PR",
	} {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q, got:\n%s", check, output)
		}
	}
	text := readFile(t, specPath)
	for _, check := range []string{
		"Original thesis",
		"### Thesis Revision - ",
		"Revised thesis",
		"User intends for the coding agent to continue",
	} {
		if !strings.Contains(text, check) {
			t.Fatalf("expected SPEC.md to contain %q, got:\n%s", check, text)
		}
	}
	doc := document.Parse(text, specPath, document.TypeSpec)
	if got := doc.DeliveryIntent(); got != specDeliveryIntentContinueCurrent {
		t.Fatalf("delivery intent = %q, want %q", got, specDeliveryIntentContinueCurrent)
	}
	if clarification, ok := doc.ClarificationState(); !ok || clarification.Status != document.ClarificationStatusOpen {
		t.Fatalf("expected thesis revision to reopen clarification state, got %#v ok=%v", clarification, ok)
	}
}

func TestOutputCompiledPrompt_IncludesRLMGuidanceWhenContextRequiresIt(t *testing.T) {
	projectRoot := t.TempDir()
	homeDir := filepath.Join(projectRoot, "home")
	codexDir := filepath.Join(homeDir, ".codex")

	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", codexDir)

	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	writeFile(t, filepath.Join(projectRoot, "AGENTS.md"), "# AGENTS\n")
	writeFile(t, filepath.Join(projectRoot, "CLAUDE.md"), "# CLAUDE\n")
	writeFile(t, filepath.Join(projectRoot, ".github", "copilot-instructions.md"), "# COPILOT\n")

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0012-codebase-audit")
	specPath := filepath.Join(featurePath, "SPEC.md")
	brainstormPath := filepath.Join(featurePath, "BRAINSTORM.md")
	writeFile(t, brainstormPath, "# BRAINSTORM\n")
	writeFile(t, specPath, documentTemplateWithSummary())

	restore := chdirForTest(t, projectRoot)
	defer restore()

	cfg := config.Default()
	answers := &specAnswers{Problem: "Need codebase-wide analysis of all FHIR and auth flows."}

	output := captureStdout(t, func() {
		err := outputCompiledPrompt(specPath, brainstormPath, "codebase-audit", projectRoot, cfg, answers, true)
		if err != nil {
			t.Fatalf("outputCompiledPrompt() error = %v", err)
		}
	})

	checks := []string{
		"# Use RLM Pattern",
		"parallelization_mode: \"rlm\"",
		"immediate decision → smallest artifact → required facts → act or recurse",
		"add `rlm` to canonical front matter `skills`",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
}
