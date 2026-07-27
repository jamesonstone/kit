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

func TestOutputCompiledPrompt_IncludesSkillsDiscoveryInputs(t *testing.T) {
	projectRoot := t.TempDir()
	homeDir := filepath.Join(projectRoot, "home")
	codexDir := filepath.Join(homeDir, ".codex")

	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", codexDir)

	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	writeFile(t, filepath.Join(projectRoot, "AGENTS.md"), "# AGENTS\n")
	writeFile(t, filepath.Join(projectRoot, "CLAUDE.md"), "# CLAUDE\n")
	writeFile(t, filepath.Join(projectRoot, ".github", "copilot-instructions.md"), "# COPILOT\n")
	writeFile(t, filepath.Join(homeDir, ".claude", "CLAUDE.md"), "# global claude\n")
	writeFile(t, filepath.Join(codexDir, "AGENTS.md"), "# global codex agents\n")
	writeFile(t, filepath.Join(codexDir, "instructions.md"), "# global codex instructions\n")

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0009-spec-skills-discovery")
	specPath := filepath.Join(featurePath, "SPEC.md")
	brainstormPath := filepath.Join(featurePath, "BRAINSTORM.md")
	writeFile(t, brainstormPath, "# BRAINSTORM\n")
	writeFile(t, specPath, documentTemplateWithSummary())

	restore := chdirForTest(t, projectRoot)
	defer restore()

	cfg := config.Default()

	output := captureStdout(t, func() {
		err := outputCompiledPrompt(
			specPath,
			brainstormPath,
			"spec-skills-discovery",
			projectRoot,
			cfg,
			&specAnswers{Problem: "skills are undocumented"},
			true,
		)
		if err != nil {
			t.Fatalf("outputCompiledPrompt() error = %v", err)
		}
	})

	checks := []string{
		filepath.Join(projectRoot, "AGENTS.md"),
		filepath.Join(projectRoot, "CLAUDE.md"),
		filepath.Join(projectRoot, ".github", "copilot-instructions.md"),
		filepath.Join(projectRoot, ".agents", "skills", "*", "SKILL.md"),
		"`SPEC.md` is the single durable feature artifact",
		"Load only sources relevant to the current decision",
		"feature identity, relationships, references, skills, and delivery intent",
		filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"),
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
	assertV2SpecPromptContract(t, output)
	assertV2SpecPromptExcludesV1StageAssumptions(t, output)
}

func TestRunSpecTemplate_IncludesSkillsSectionGuidance(t *testing.T) {
	projectRoot := t.TempDir()
	homeDir := filepath.Join(projectRoot, "home")
	codexDir := filepath.Join(homeDir, ".codex")

	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", codexDir)

	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	writeFile(t, filepath.Join(projectRoot, "AGENTS.md"), "# AGENTS\n")
	writeFile(t, filepath.Join(projectRoot, "CLAUDE.md"), "# CLAUDE\n")
	writeFile(t, filepath.Join(projectRoot, ".github", "copilot-instructions.md"), "# COPILOT\n")

	restore := chdirForTest(t, projectRoot)
	defer restore()

	cfg := config.Default()
	specPath := filepath.Join(projectRoot, "docs", "specs", "0010-sample", "SPEC.md")

	output := captureStdout(t, func() {
		err := runSpecTemplate(specPath, "", "sample", projectRoot, cfg, true, false)
		if err != nil {
			t.Fatalf("runSpecTemplate() error = %v", err)
		}
	})

	checks := []string{
		"`SPEC.md` is the single durable feature artifact",
		"Load only sources relevant to the current decision",
		"feature identity, relationships, references, skills, and delivery intent",
		filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"),
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
	assertV2SpecPromptContract(t, output)
	assertV2SpecPromptExcludesV1StageAssumptions(t, output)
}

func TestRunSpecTemplate_IncludesRLMGuidanceWhenBrainstormHintsLargeRepo(t *testing.T) {
	projectRoot := t.TempDir()
	homeDir := filepath.Join(projectRoot, "home")
	codexDir := filepath.Join(homeDir, ".codex")

	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", codexDir)

	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	writeFile(t, filepath.Join(projectRoot, "AGENTS.md"), "# AGENTS\n")
	writeFile(t, filepath.Join(projectRoot, "CLAUDE.md"), "# CLAUDE\n")
	writeFile(t, filepath.Join(projectRoot, ".github", "copilot-instructions.md"), "# COPILOT\n")

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0011-repository-audit")
	specPath := filepath.Join(featurePath, "SPEC.md")
	brainstormPath := filepath.Join(featurePath, "BRAINSTORM.md")
	writeFile(t, brainstormPath, "scan repository for auth and FHIR integration points\n")

	restore := chdirForTest(t, projectRoot)
	defer restore()

	cfg := config.Default()
	output := captureStdout(t, func() {
		err := runSpecTemplate(specPath, brainstormPath, "repository-audit", projectRoot, cfg, true, false)
		if err != nil {
			t.Fatalf("runSpecTemplate() error = %v", err)
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
