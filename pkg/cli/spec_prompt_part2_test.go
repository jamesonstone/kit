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

func TestRunSpecWithoutSelectionCandidatesStartsInteractiveCreation(t *testing.T) {
	t.Setenv("EDITOR", "")
	projectRoot, _ := setupLifecycleTestProject(t)
	restore := chdirForTest(t, projectRoot)
	defer restore()
	restoreSpecFlags := restoreSpecFlagState()
	defer restoreSpecFlags()

	previousPrompt := promptSpecFeatureRef
	previousWait := awaitEditorLaunchConfirmation
	previousRunner := editorInputRunner
	previousDeliveryPrompt := promptSpecDeliveryIntent
	defer func() {
		promptSpecFeatureRef = previousPrompt
		awaitEditorLaunchConfirmation = previousWait
		editorInputRunner = previousRunner
		promptSpecDeliveryIntent = previousDeliveryPrompt
	}()

	promptSpecFeatureRef = func() (string, error) {
		return "sample", nil
	}
	awaitEditorLaunchConfirmation = func(_ *os.File, _ io.Writer) error {
		return nil
	}
	editorInputRunner = func(_ freeTextInputConfig, fieldName, _ string) (string, bool, error) {
		return fieldName + " answer", true, nil
	}
	promptSpecDeliveryIntent = func() (string, error) {
		return specDeliveryIntentIdeaOnly, nil
	}

	cmd := newSpecProfileTestCommand()
	if err := cmd.Flags().Set("output-only", "true"); err != nil {
		t.Fatalf("Flags().Set(output-only) error = %v", err)
	}

	output := captureStdout(t, func() {
		if err := runSpec(cmd, nil); err != nil {
			t.Fatalf("runSpec() error = %v", err)
		}
	})

	specPath := filepath.Join(projectRoot, "docs", "specs", "0001-sample", "SPEC.md")
	if _, err := os.Stat(specPath); err != nil {
		t.Fatalf("expected SPEC.md to be created at %s: %v", specPath, err)
	}
	for _, check := range []string{
		"Spec Thesis",
		"**THESIS**: feature thesis answer",
		"**DELIVERY INTENT**: no - idea-only SPEC.md capture",
	} {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q, got:\n%s", check, output)
		}
	}
	text := readFile(t, specPath)
	if !strings.Contains(text, "feature thesis answer") {
		t.Fatalf("expected SPEC.md to contain the captured thesis, got:\n%s", text)
	}
	if !strings.Contains(text, "Idea capture only") {
		t.Fatalf("expected SPEC.md to record idea-only delivery decision, got:\n%s", text)
	}
	doc := document.Parse(text, specPath, document.TypeSpec)
	if got := doc.DeliveryIntent(); got != specDeliveryIntentIdeaOnly {
		t.Fatalf("delivery intent = %q, want %q", got, specDeliveryIntentIdeaOnly)
	}
	if clarification, ok := doc.ClarificationState(); !ok || clarification.Status != document.ClarificationStatusOpen {
		t.Fatalf("expected new SPEC.md to include open clarification state, got %#v ok=%v", clarification, ok)
	}
}

func TestRunSpecExistingSpecDoesNotPromptForThesisByDefault(t *testing.T) {
	t.Setenv("EDITOR", "")
	projectRoot, _ := setupLifecycleTestProject(t)
	restore := chdirForTest(t, projectRoot)
	defer restore()
	restoreSpecFlags := restoreSpecFlagState()
	defer restoreSpecFlags()

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-sample")
	specPath := filepath.Join(featurePath, "SPEC.md")
	writeFile(t, specPath, "# SPEC\n\n## THESIS\n\nOriginal thesis\n\n## DELIVERY DECISION\n\nOriginal delivery decision\n")

	previousRunner := editorInputRunner
	previousDeliveryPrompt := promptSpecDeliveryIntent
	defer func() {
		editorInputRunner = previousRunner
		promptSpecDeliveryIntent = previousDeliveryPrompt
	}()
	editorInputRunner = func(_ freeTextInputConfig, fieldName, _ string) (string, bool, error) {
		t.Fatalf("editorInputRunner called for existing SPEC.md field %q", fieldName)
		return "", false, nil
	}
	promptSpecDeliveryIntent = func() (string, error) {
		t.Fatal("promptSpecDeliveryIntent called for existing SPEC.md")
		return "", nil
	}

	cmd := newSpecProfileTestCommand()
	if err := cmd.Flags().Set("output-only", "true"); err != nil {
		t.Fatalf("Flags().Set(output-only) error = %v", err)
	}

	output := captureStdout(t, func() {
		if err := runSpec(cmd, []string{"sample"}); err != nil {
			t.Fatalf("runSpec() error = %v", err)
		}
	})

	if strings.Contains(output, "Spec Thesis") {
		t.Fatalf("existing SPEC.md unexpectedly reopened thesis prompt, got:\n%s", output)
	}
	text := readFile(t, specPath)
	if !strings.Contains(text, "Original thesis") || !strings.Contains(text, "Original delivery decision") {
		t.Fatalf("existing SPEC.md content was not preserved, got:\n%s", text)
	}
}

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
		"## Context Routing",
		"use the repository RLM pattern",
		"load the smallest source that resolves the current decision",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
}
