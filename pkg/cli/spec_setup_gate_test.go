package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/templates"
)

func TestRunSpecSetupGateAllowsBypassForFreshProject(t *testing.T) {
	t.Setenv("EDITOR", "")
	projectRoot := t.TempDir()
	restore := chdirForTest(t, projectRoot)
	defer restore()
	restoreSpecFlags := restoreSpecFlagState()
	defer restoreSpecFlags()

	var reasons []string
	restoreSpecSetupTestHooks(t)
	promptSpecSetupGate = func(got []string) (specSetupGateDecision, error) {
		reasons = append([]string{}, got...)
		return specSetupGateContinue, nil
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
		if err := runSpec(cmd, []string{"first-feature"}); err != nil {
			t.Fatalf("runSpec() error = %v", err)
		}
	})

	for _, check := range []string{config.ConfigFileName + " is missing", "docs/CONSTITUTION.md is missing"} {
		if !containsString(reasons, check) {
			t.Fatalf("expected setup gate reasons to include %q, got %#v", check, reasons)
		}
	}
	for _, path := range []string{
		filepath.Join(projectRoot, config.ConfigFileName),
		filepath.Join(projectRoot, "docs", "CONSTITUTION.md"),
		filepath.Join(projectRoot, "docs", "specs", "0001-first-feature", "SPEC.md"),
		filepath.Join(projectRoot, "docs", "notes", "0001-first-feature", "README.md"),
		filepath.Join(projectRoot, "docs", "notes", "0001-first-feature", "inbox", ".gitkeep"),
		filepath.Join(projectRoot, "docs", "notes", "0001-first-feature", "references", ".gitkeep"),
		filepath.Join(projectRoot, "docs", "notes", "0001-first-feature", "responses", ".gitkeep"),
		filepath.Join(projectRoot, "docs", "notes", "0001-first-feature", "private", ".gitignore"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist after bypass: %v", path, err)
		}
	}
	if !strings.Contains(output, "**THESIS**: feature thesis answer") {
		t.Fatalf("expected spec prompt output to contain captured thesis, got:\n%s", output)
	}
}

func TestRunSpecSetupGateCopiesInitPromptAndStops(t *testing.T) {
	projectRoot := t.TempDir()
	restore := chdirForTest(t, projectRoot)
	defer restore()
	restoreSpecFlags := restoreSpecFlagState()
	defer restoreSpecFlags()

	var copied string
	restoreSpecSetupTestHooks(t)
	promptSpecSetupGate = func(_ []string) (specSetupGateDecision, error) {
		return specSetupGateReinit, nil
	}
	clipboardCopyFunc = func(text string) error {
		copied = text
		return nil
	}
	editorInputRunner = func(_ freeTextInputConfig, fieldName, _ string) (string, bool, error) {
		t.Fatalf("editorInputRunner called after re-init decision for %q", fieldName)
		return "", false, nil
	}

	cmd := newSpecProfileTestCommand()
	specLegacySupervisor = true
	output := captureStdout(t, func() {
		if err := runSpec(cmd, []string{"first-feature"}); err != nil {
			t.Fatalf("runSpec() error = %v", err)
		}
	})

	if !strings.Contains(copied, "Treat the exact generated starter at ") ||
		!strings.Contains(copied, "docs/CONSTITUTION.md as a valid bootstrap Constitution") {
		t.Fatalf("expected copied init prompt to target Constitution, got:\n%s", copied)
	}
	if !strings.Contains(output, "Paste the copied prompt into your agent") {
		t.Fatalf("expected re-init next steps, got:\n%s", output)
	}
	for _, path := range []string{
		filepath.Join(projectRoot, config.ConfigFileName),
		filepath.Join(projectRoot, "docs", "specs", "0001-first-feature", "SPEC.md"),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected %s not to be written after re-init decision, got %v", path, err)
		}
	}
}

func TestRunSpecSetupGateAcceptsStarterConstitution(t *testing.T) {
	t.Setenv("EDITOR", "")
	projectRoot := t.TempDir()
	cfg := defaultInitConfig()
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeFile(t, filepath.Join(projectRoot, cfg.ConstitutionPath), templates.Constitution)
	writeInstructionArtifactsForSpecSetupTest(t, projectRoot, cfg)

	restore := chdirForTest(t, projectRoot)
	defer restore()
	restoreSpecFlags := restoreSpecFlagState()
	defer restoreSpecFlags()

	restoreSpecSetupTestHooks(t)
	promptSpecSetupGate = func(_ []string) (specSetupGateDecision, error) {
		t.Fatal("promptSpecSetupGate called for valid bootstrap Constitution")
		return "", nil
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

	if err := runSpec(cmd, []string{"starter-constitution"}); err != nil {
		t.Fatalf("runSpec() error = %v", err)
	}
	if got := readFile(t, filepath.Join(projectRoot, cfg.ConstitutionPath)); got != templates.Constitution {
		t.Fatalf("expected spec creation not to overwrite starter Constitution, got:\n%s", got)
	}
}

func TestAssessSpecSetupDetectsEmptyConstitutionSections(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := defaultInitConfig()
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeFile(t, filepath.Join(projectRoot, cfg.ConstitutionPath), `# CONSTITUTION

## PRINCIPLES

## CONSTRAINTS

Project constraints are defined.

## NON-GOALS

No test non-goals.

## DEFINITIONS

Test definition.
`)
	writeInstructionArtifactsForSpecSetupTest(t, projectRoot, cfg)

	status := assessSpecSetup(projectRoot, cfg, false)
	if !containsString(status.Reasons, `docs/CONSTITUTION.md section "PRINCIPLES" has no project-specific content`) {
		t.Fatalf("expected empty section setup reason, got %#v", status.Reasons)
	}
}

func TestRunSpecSetupGateSkippedForReadyProject(t *testing.T) {
	t.Setenv("EDITOR", "")
	projectRoot, _ := setupLifecycleTestProject(t)
	restore := chdirForTest(t, projectRoot)
	defer restore()
	restoreSpecFlags := restoreSpecFlagState()
	defer restoreSpecFlags()
	restoreSpecSetupTestHooks(t)

	promptSpecSetupGate = func(_ []string) (specSetupGateDecision, error) {
		t.Fatal("promptSpecSetupGate called for ready project")
		return "", nil
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
	if err := runSpec(cmd, []string{"ready-project"}); err != nil {
		t.Fatalf("runSpec() error = %v", err)
	}
}

func TestNormalizeSpecSetupGateDecision(t *testing.T) {
	for _, raw := range []string{"", "c", "continue", "bypass", "skip", "y"} {
		got, err := normalizeSpecSetupGateDecision(raw)
		if err != nil {
			t.Fatalf("normalizeSpecSetupGateDecision(%q) error = %v", raw, err)
		}
		if got != specSetupGateContinue {
			t.Fatalf("normalizeSpecSetupGateDecision(%q) = %q, want continue", raw, got)
		}
	}
	for _, raw := range []string{"r", "re-init", "reinit", "init"} {
		got, err := normalizeSpecSetupGateDecision(raw)
		if err != nil {
			t.Fatalf("normalizeSpecSetupGateDecision(%q) error = %v", raw, err)
		}
		if got != specSetupGateReinit {
			t.Fatalf("normalizeSpecSetupGateDecision(%q) = %q, want re-init", raw, got)
		}
	}
	if _, err := normalizeSpecSetupGateDecision("later"); err == nil {
		t.Fatal("expected invalid setup gate decision to fail")
	}
}

func restoreSpecSetupTestHooks(t *testing.T) {
	t.Helper()

	previousSetupGate := promptSpecSetupGate
	previousWait := awaitEditorLaunchConfirmation
	previousRunner := editorInputRunner
	previousDeliveryPrompt := promptSpecDeliveryIntent
	previousClipboard := clipboardCopyFunc
	t.Cleanup(func() {
		promptSpecSetupGate = previousSetupGate
		awaitEditorLaunchConfirmation = previousWait
		editorInputRunner = previousRunner
		promptSpecDeliveryIntent = previousDeliveryPrompt
		clipboardCopyFunc = previousClipboard
	})
}

func writeInstructionArtifactsForSpecSetupTest(t *testing.T, projectRoot string, cfg *config.Config) {
	t.Helper()

	for _, relativePath := range instructionArtifactPaths(
		cfg,
		instructionFileSelection{},
		cfg.EffectiveInstructionScaffoldVersion(),
		true,
	) {
		content, _, err := instructionArtifactContent(relativePath, cfg.EffectiveInstructionScaffoldVersion())
		if err != nil {
			t.Fatalf("instructionArtifactContent(%q) error = %v", relativePath, err)
		}
		writeFile(t, filepath.Join(projectRoot, relativePath), content)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
