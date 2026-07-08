package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/document"
)

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
