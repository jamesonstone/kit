package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
		"/plan",
		"You are in planning mode for feature: **sample-feature**",
		"Do NOT implement code, write production changes, or move into execution",
		"Ask clarifying questions until you reach ≥95% confidence that you understand the problem and desired solution",
		"Use numbered lists",
		"Ask questions in batches of up to 10",
		"For every question, include your current best recommended default, proposed solution, or assumption",
		"State uncertainties",
		"\"yes\" or \"y\" approves all recommended defaults in the batch",
		"\"yes 3, 4, 5\" or \"y 3, 4, 5\" approves only those numbered defaults in the batch",
		"If the user approves only specific question numbers, treat all other questions in that batch as unresolved",
		"After each batch of up to 10 questions, output your current percentage understanding so the user can see progress",
		"planning only — no implementation",
		"kit spec sample-feature",
		"/tmp/docs/specs/0001-sample/BRAINSTORM.md",
		"/tmp/project/docs/notes/0001-sample",
		"Inspect the feature notes directory",
		"ignore `.gitkeep`",
		"read only the notes relevant to the user thesis",
		"record specific note files that shaped the brainstorm",
		"leave the notes directory dependency as `optional`",
		"/tmp/project/docs/CONSTITUTION.md",
		"## DEPENDENCIES",
		"`Dependency`, `Type`, `Location`, `Used For`, and `Status`",
		"Use an RLM-style just-in-time prior-work pass over `/tmp/docs/specs` before broad repository reads",
		"/tmp/project/docs/PROJECT_PROGRESS_SUMMARY.md",
		"conditional reads only",
		"shared interface or contract",
		"inspect at most 5 prior feature directories",
		"do not paraphrase entire prior docs into chat",
		"for Figma or other MCP-driven design dependencies, store the exact design URL or file/node reference in `Location`",
		"`Status` = `stale`",
		"no section in `BRAINSTORM.md` may remain empty or contain only an HTML TODO comment",
		"`not applicable`, `not required`, or `no additional information required`",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected prompt to contain %q", check)
		}
	}

	if !strings.HasPrefix(prompt, "/plan\n\n") {
		t.Fatalf("expected prompt to start with /plan, got %q", prompt[:8])
	}
}

func TestRunBrainstorm_CreatesFeatureNotesDirAndSeedsDependency(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)

	restoreWD, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restoreWD()

	restoreEditor := stubBrainstormEditor(t, "Need better import validation for malformed CSV uploads.")
	defer restoreEditor()
	restoreFlags := setBrainstormFlagState(false, false, "", false, false, false, false)
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
		"Need better import validation for malformed CSV uploads.",
		"| Feature notes | notes | docs/notes/0001-sample-feature | optional pre-brainstorm research input | optional |",
	}
	for _, check := range checks {
		if !strings.Contains(string(content), check) {
			t.Fatalf("expected BRAINSTORM.md to contain %q, got %q", check, string(content))
		}
	}
}

func TestEnsureBrainstormNotesDependency_AppendsWithoutRemovingExistingRows(t *testing.T) {
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
		t.Fatal("expected notes dependency to be appended")
	}
	changed, err = ensureBrainstormNotesDependency(brainstormPath, notesRelPath)
	if err != nil {
		t.Fatalf("second ensureBrainstormNotesDependency() error = %v", err)
	}
	if changed {
		t.Fatal("expected second notes dependency ensure to be a no-op")
	}

	content, err := os.ReadFile(brainstormPath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	text := string(content)
	checks := []string{
		"| Existing note | notes | docs/notes/0001-sample/old.md | prior observation | stale |",
		"| Feature notes | notes | docs/notes/0001-sample | optional pre-brainstorm research input | optional |",
		"<!-- keep this comment -->",
	}
	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Fatalf("expected updated BRAINSTORM.md to contain %q, got %q", check, text)
		}
	}
	if count := strings.Count(text, "| Feature notes | notes | docs/notes/0001-sample | optional pre-brainstorm research input | optional |"); count != 1 {
		t.Fatalf("expected one feature notes dependency row, got %d in %q", count, text)
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
