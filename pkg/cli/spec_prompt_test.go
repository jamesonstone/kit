package cli

import (
	"bytes"
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
func assertV2SpecPromptContract(t *testing.T, output string) {
	t.Helper()

	checks := []string{
		"## Goal",
		"## User Context",
		"## Repository Context",
		"## Source And State Contract",
		"`SPEC.md` is the single durable feature artifact",
		"Context `### Source Map`",
		"`SRC-###`, `REQ-###`, `AC-###`",
		"clarification state",
		"## Clarification And Autonomy",
		"research repository-discoverable facts first",
		"Ask only about material choices that remain non-discoverable",
		"Outside `clarify`, do not re-ask settled questions or request routine permission",
		"## Constraints And Approval Boundaries",
		"Safe repository reads and reversible in-scope edits need no extra approval",
		"git status --short",
		"Before Git/GitHub delivery mutation",
		"Delivery Contract",
		"## Phase Outcomes",
		"Do not skip a phase gate",
		"validates phase state in code",
		"## Agent Routing",
		"docs/references/rules/agent-team-orchestration.md",
		"read-only verifier",
		"## Success Criteria",
		"implementation evidence",
		"exact validation evidence",
		"Never claim a check ran when it did not",
		"## Output Contract",
		"Open Questions",
		"## Final Response Contract",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected v2 spec prompt to contain %q", check)
		}
	}
	for _, forbidden := range []string{
		"Programmatic Tool Calling",
		"persisted reasoning",
		"Pro mode",
		"text.verbosity",
		"Ask clarification questions until",
	} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("v2 spec prompt unexpectedly contains %q", forbidden)
		}
	}
	assertFinalResponseContractHeadings(t, output,
		"Outcome",
		"Evidence",
		"Artifacts And State",
		"Agent Team",
		"Open Items",
	)
}

func assertV2SpecPromptExcludesV1StageAssumptions(t *testing.T, output string) {
	t.Helper()

	unwanted := []string{
		"Only update SPEC.md and supporting documentation",
		"Run 'kit plan",
		"Run `kit plan",
		"usually `kit plan",
		"Run 'kit legacy plan",
		"Run `kit legacy plan",
		"usually `kit legacy plan",
		"Avoid implementation details (focus on WHAT, not HOW)",
		"write the selected skills into canonical front matter `skills`; use the legacy `## SKILLS` table",
		"keep the legacy `none | n/a | n/a | no additional skills required | no` row",
	}

	for _, check := range unwanted {
		if strings.Contains(output, check) {
			t.Fatalf("v2 spec prompt reintroduced v1 stage assumption %q", check)
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = writer
	defer func() {
		os.Stdout = original
	}()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		t.Fatalf("io.Copy() error = %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("reader.Close() error = %v", err)
	}

	return buf.String()
}

func chdirForTest(t *testing.T, dir string) func() {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("os.Chdir() error = %v", err)
	}

	return func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("os.Chdir() restore error = %v", err)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := document.Write(path, content); err != nil {
		t.Fatalf("document.Write(%q) error = %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	return string(content)
}

func defaultKitConfig() string {
	return "goal_percentage: 95\nspecs_dir: docs/specs\nskills_dir: .agents/skills\nconstitution_path: docs/CONSTITUTION.md\nallow_out_of_order: false\nagents:\n  - AGENTS.md\n  - CLAUDE.md\n  - .github/copilot-instructions.md\nfeature_naming:\n  numeric_width: 4\n  separator: '-'\n"
}

func documentTemplateWithSummary() string {
	return "# SPEC\n\n## SUMMARY\n\nsummary\n"
}
