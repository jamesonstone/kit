package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

func TestRunSpecInteractive_UsesEditorByDefault(t *testing.T) {
	t.Setenv("EDITOR", "")
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0010-sample")
	specPath := filepath.Join(featurePath, "SPEC.md")
	writeFile(t, specPath, "# SPEC\n")

	restore := chdirForTest(t, projectRoot)
	defer restore()

	previousWait := awaitEditorLaunchConfirmation
	previousRunner := editorInputRunner
	previousDeliveryPrompt := promptSpecDeliveryIntent
	defer func() {
		awaitEditorLaunchConfirmation = previousWait
		editorInputRunner = previousRunner
		promptSpecDeliveryIntent = previousDeliveryPrompt
	}()

	var sequence []string
	awaitEditorLaunchConfirmation = func(_ *os.File, _ io.Writer) error {
		sequence = append(sequence, "wait")
		return nil
	}
	editorInputRunner = func(_ freeTextInputConfig, fieldName, _ string) (string, bool, error) {
		sequence = append(sequence, fieldName)
		return fieldName + " answer", true, nil
	}
	promptSpecDeliveryIntent = func() (string, error) {
		sequence = append(sequence, "delivery-intent")
		return specDeliveryIntentIssueBranchPRLater, nil
	}

	cfg := config.Default()
	feat := &feature.Feature{Slug: "sample", DirName: "0010-sample", Path: featurePath}

	var answers *specAnswers
	output := captureStdout(t, func() {
		var err error
		answers, err = runSpecInteractive(
			specPath,
			"",
			feat,
			projectRoot,
			cfg,
			newFreeTextInputConfig(false, "", false, true),
			true,
			true,
		)
		if err != nil {
			t.Fatalf("runSpecInteractive() error = %v", err)
		}
	})

	wantSequence := []string{
		"wait",
		"feature thesis",
		"delivery-intent",
	}
	if strings.Join(sequence, "|") != strings.Join(wantSequence, "|") {
		t.Fatalf("unexpected editor prompt sequence: got %v want %v", sequence, wantSequence)
	}
	if answers == nil || answers.Problem != "feature thesis answer" {
		t.Fatalf("expected thesis answer to be returned, got %#v", answers)
	}

	checks := []string{
		"Spec Thesis",
		"A default editor will open for this response.",
		"What to write",
		"What Kit handles next",
		"coding agent will infer, research, clarify, and fill every other SPEC.md section",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
	text := readFile(t, specPath)
	if !strings.Contains(text, "feature thesis answer") {
		t.Fatalf("expected SPEC.md to contain thesis, got:\n%s", text)
	}
	doc := document.Parse(text, specPath, document.TypeSpec)
	if got := doc.DeliveryIntent(); got != specDeliveryIntentIssueBranchPRLater {
		t.Fatalf("delivery intent = %q, want %q", got, specDeliveryIntentIssueBranchPRLater)
	}
	if clarification, ok := doc.ClarificationState(); !ok || clarification.Status != document.ClarificationStatusOpen {
		t.Fatalf("expected thesis capture to reset clarification state, got %#v ok=%v", clarification, ok)
	}
	if !strings.Contains(text, "User intends to create a new issue, branch, and PR later") {
		t.Fatalf("expected Delivery Decision to describe issue/branch/PR intent, got:\n%s", text)
	}
}

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
