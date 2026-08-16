package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReviewLoopRenderIncludesOnlyFixTasks(t *testing.T) {
	withFakeEditor(t)
	var copied string
	previousCopy := clipboardCopyFunc
	defer func() { clipboardCopyFunc = previousCopy }()
	clipboardCopyFunc = func(text string) error {
		copied = text
		return nil
	}

	classified := []reviewLoopClassifiedFinding{
		{
			Kind: reviewLoopFix,
			Finding: reviewLoopFinding{Task: dispatchReviewTask{
				Path: "internal/app.go", Line: 12, Author: "coderabbitai", URL: "https://example.com/1", Body: "Fix app routing.",
			}},
			Reason: "current and actionable",
		},
		{
			Kind: reviewLoopStale,
			Finding: reviewLoopFinding{Task: dispatchReviewTask{
				Path: "internal/old.go", Line: 99, Author: "coderabbitai", URL: "https://example.com/2", Body: "Remove stale code.",
			}},
			Reason: "line no longer exists",
		},
	}
	ctx := reviewLoopPRContext{
		Target:     dispatchPRTarget{Owner: "Patient-Driven-Care", Repo: "cortex", Number: 67},
		URL:        "https://github.com/Patient-Driven-Care/cortex/pull/67",
		HeadRefOID: "abc123",
	}

	out := &bytes.Buffer{}
	err := runReviewLoopPrompt(out, reviewLoopOptions{}, ctx, classified, coderabbitSharedReviewInstruction)
	if err != nil {
		t.Fatalf("runReviewLoopPrompt() error = %v", err)
	}

	summary := out.String()
	if !strings.Contains(summary, "[STALE] internal/old.go:99") {
		t.Fatalf("expected stale finding in summary, got %q", summary)
	}
	if !strings.Contains(copied, "Fix app routing.") {
		t.Fatalf("expected copied dispatch prompt to include fix task, got %q", copied)
	}
	if strings.Contains(copied, "Remove stale code.") {
		t.Fatalf("expected non-fix task to be excluded from dispatch prompt, got %q", copied)
	}
}

func TestReviewLoopRenderNoActionableSkipsEditor(t *testing.T) {
	previousRunner := editorInputRunner
	defer func() { editorInputRunner = previousRunner }()
	editorInputRunner = func(_ freeTextInputConfig, _ string, _ string) (string, bool, error) {
		return "", false, fmt.Errorf("editor should not run")
	}

	classified := []reviewLoopClassifiedFinding{
		{
			Kind: reviewLoopNeedsHuman,
			Finding: reviewLoopFinding{Task: dispatchReviewTask{
				Path: "internal/app.go", Line: 12, Body: "Needs human decision.",
			}},
			Reason: "ambiguous",
		},
	}
	ctx := reviewLoopPRContext{Target: dispatchPRTarget{Number: 67}, HeadRefOID: "abc123"}

	out := &bytes.Buffer{}
	if err := runReviewLoopPrompt(out, reviewLoopOptions{OutputOnly: true}, ctx, classified, ""); err != nil {
		t.Fatalf("runReviewLoopPrompt() error = %v", err)
	}
	if !strings.Contains(out.String(), "No actionable current review feedback found.") {
		t.Fatalf("expected no-actionable message, got %q", out.String())
	}
}

func TestReviewLoopLineExists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(path, []byte("one\ntwo"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !reviewLoopLineExists(path, 2) {
		t.Fatal("expected line 2 to exist")
	}
	if reviewLoopLineExists(path, 3) {
		t.Fatal("expected line 3 to be missing")
	}
}
