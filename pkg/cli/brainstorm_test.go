package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/document"
)

func TestBuildBrainstormPrompt(t *testing.T) {
	prompt := buildBrainstormPrompt(
		"/tmp/docs/specs/0001-sample/BRAINSTORM.md",
		"sample-feature",
		"/tmp/project",
		"Need better import validation for malformed CSV uploads.",
		95,
	)

	checks := []string{
		"Research feature `sample-feature`",
		"durable research record. Do not implement product code",
		"## Research Contract",
		"Resolve discoverable ambiguity from repository evidence",
		"material non-discoverable choices",
		"recommended default and why the answer changes the result",
		"kit spec sample-feature",
		"/tmp/docs/specs/0001-sample/BRAINSTORM.md",
		"/tmp/project/docs/notes/0001-sample",
		"Ignore placeholders such as .gitkeep",
		"/tmp/project/docs/CONSTITUTION.md",
		"Keep front matter references and feature relationships current",
		"## Success Criteria",
		"`not applicable`",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected prompt to contain %q", check)
		}
	}
	assertFinalResponseContractHeadings(t, prompt,
		"Summary",
		"Artifacts Updated",
		"Key Decisions",
		"Open Questions",
		"Next Step",
	)

	if !strings.HasPrefix(prompt, "Research feature `sample-feature`") {
		t.Fatalf("expected prompt to start with research header, got %q", prompt[:64])
	}
	if strings.Contains(prompt, "/plan") || strings.Contains(prompt, "planning mode") {
		t.Fatalf("expected prompt to avoid native plan-mode triggers, got %q", prompt)
	}
}

func TestRunBrainstorm_CreatesFeatureNotesDirAndSeedsReference(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)

	restoreWD, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restoreWD()

	restoreEditor := stubBrainstormEditor(t, "Need better import validation for malformed CSV uploads.")
	defer restoreEditor()
	restoreFlags := setBrainstormFlagState(false, "", false, false, false, false)
	defer restoreFlags()

	cmd := newBrainstormTestCommand()
	if err := cmd.Flags().Set("output-only", "true"); err != nil {
		t.Fatalf("Flags().Set() error = %v", err)
	}
	_ = captureStdout(t, func() {
		if err := runBrainstorm(cmd, []string{"sample-feature"}); err != nil {
			t.Fatalf("runBrainstorm() error = %v", err)
		}
	})

	notesPath := filepath.Join(projectRoot, "docs", "notes", "0001-sample-feature")
	if _, err := os.Stat(filepath.Join(notesPath, ".gitkeep")); err != nil {
		t.Fatalf("expected feature notes .gitkeep, got %v", err)
	}

	brainstormPath := filepath.Join(projectRoot, "docs", "specs", "0001-sample-feature", "BRAINSTORM.md")
	content, err := os.ReadFile(brainstormPath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	checks := []string{
		"kit_metadata_version: 1",
		"artifact: brainstorm",
		"dir: 0001-sample-feature",
		"Need better import validation for malformed CSV uploads.",
		"name: Feature notes",
		"target: docs/notes/0001-sample-feature",
		"relation: informs",
		"read_policy: conditional",
		"used_for: optional pre-brainstorm research input",
		"status: optional",
	}
	for _, check := range checks {
		if !strings.Contains(string(content), check) {
			t.Fatalf("expected BRAINSTORM.md to contain %q, got %q", check, string(content))
		}
	}
}

func TestRunBrainstormFrontendProfileCreatesDesignMaterialsAndSeedsReferences(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)

	restoreWD, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restoreWD()

	restoreEditor := stubBrainstormEditor(t, "Need a responsive dashboard redesign.")
	defer restoreEditor()
	restoreFlags := setBrainstormFlagState(false, "", false, false, false, false)
	defer restoreFlags()
	restorePromptProfileState(t, promptProfileFrontend, true)

	cmd := newBrainstormTestCommand()
	if err := cmd.Flags().Set("output-only", "true"); err != nil {
		t.Fatalf("Flags().Set() error = %v", err)
	}
	output := captureStdout(t, func() {
		if err := runBrainstorm(cmd, []string{"dashboard-redesign"}); err != nil {
			t.Fatalf("runBrainstorm() error = %v", err)
		}
	})

	featureDir := "0001-dashboard-redesign"
	designPath := filepath.Join(projectRoot, "docs", "notes", featureDir, "design")
	for _, path := range []string{
		filepath.Join(designPath, ".gitkeep"),
		filepath.Join(designPath, "screenshots", ".gitkeep"),
		filepath.Join(designPath, "references", ".gitkeep"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected frontend design placeholder %s, got %v", path, err)
		}
	}

	brainstormPath := filepath.Join(projectRoot, "docs", "specs", featureDir, "BRAINSTORM.md")
	content, err := os.ReadFile(brainstormPath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	text := string(content)
	checks := []string{
		"name: Feature notes",
		"target: docs/notes/0001-dashboard-redesign",
		"name: Frontend profile",
		"target: --profile=frontend",
		"name: Design materials",
		"target: docs/notes/0001-dashboard-redesign/design",
	}
	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Fatalf("expected BRAINSTORM.md to contain %q, got:\n%s", check, text)
		}
	}

	promptChecks := []string{
		"DESIGN MATERIALS",
		designPath,
		"Ignore placeholders such as .gitkeep",
		"## Frontend Profile",
	}
	for _, check := range promptChecks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected frontend brainstorm prompt to contain %q, got:\n%s", check, output)
		}
	}
}

func TestEnsureBrainstormNotesDependency_AppendsReferenceWithoutRemovingExistingBodyRows(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	brainstormPath := filepath.Join(projectRoot, "docs", "specs", "0001-sample", "BRAINSTORM.md")
	original := `# BRAINSTORM

## SUMMARY

summary

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| Existing note | notes | docs/notes/0001-sample/old.md | prior observation | stale |

<!-- keep this comment -->

## QUESTIONS

questions
`
	writeFile(t, brainstormPath, original)

	_, notesRelPath, err := ensureFeatureNotesDir(projectRoot, "0001-sample")
	if err != nil {
		t.Fatalf("ensureFeatureNotesDir() error = %v", err)
	}
	changed, err := ensureBrainstormNotesDependency(brainstormPath, notesRelPath)
	if err != nil {
		t.Fatalf("ensureBrainstormNotesDependency() error = %v", err)
	}
	if !changed {
		t.Fatal("expected notes reference to be appended")
	}
	changed, err = ensureBrainstormNotesDependency(brainstormPath, notesRelPath)
	if err != nil {
		t.Fatalf("second ensureBrainstormNotesDependency() error = %v", err)
	}
	if changed {
		t.Fatal("expected second notes reference ensure to be a no-op")
	}

	content, err := os.ReadFile(brainstormPath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	text := string(content)
	checks := []string{
		"| Existing note | notes | docs/notes/0001-sample/old.md | prior observation | stale |",
		"name: Feature notes",
		"target: docs/notes/0001-sample",
		"<!-- keep this comment -->",
	}
	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Fatalf("expected updated BRAINSTORM.md to contain %q, got %q", check, text)
		}
	}
	doc := document.Parse(text, brainstormPath, document.TypeBrainstorm)
	count := 0
	for _, reference := range doc.References() {
		if reference.Name == featureNotesReferenceName && reference.Target == "docs/notes/0001-sample" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected one feature notes reference in front matter, got %d in %q", count, text)
	}
}

func TestEnsureBrainstormNotesDependency_ErrorsOnMalformedFrontMatter(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	brainstormPath := filepath.Join(projectRoot, "docs", "specs", "0001-sample", "BRAINSTORM.md")
	writeFile(t, brainstormPath, `---
kit_metadata_version: 1
artifact: brainstorm
feature:
  id: "0001"
  slug: sample
  dir: 0001-sample
# BRAINSTORM
`)

	changed, err := ensureBrainstormNotesDependency(brainstormPath, "docs/notes/0001-sample")
	if err == nil {
		t.Fatal("ensureBrainstormNotesDependency() error = nil, want malformed front matter error")
	}
	if changed {
		t.Fatal("ensureBrainstormNotesDependency() changed = true, want false")
	}
}

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
