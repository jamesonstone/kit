package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	stdreflect "reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestReviewLoopCommandRequiresPR(t *testing.T) {
	err := runReviewLoop(&cobra.Command{}, reviewLoopOptions{MaxSubagents: 1})
	if err == nil || !strings.Contains(err.Error(), "--pr is required") {
		t.Fatalf("expected --pr error, got %v", err)
	}
}

func TestReviewLoopResolvesRepairContextBeforeRemoteReviewWork(t *testing.T) {
	previousResolver := resolvePRRepairContext
	previousFetch := reviewLoopFetchPRContext
	previousWait := reviewLoopWaitForCodeRabbit
	previousLoad := reviewLoopLoadReviewTasks
	t.Cleanup(func() {
		resolvePRRepairContext = previousResolver
		reviewLoopFetchPRContext = previousFetch
		reviewLoopWaitForCodeRabbit = previousWait
		reviewLoopLoadReviewTasks = previousLoad
	})

	var calls []string
	repair := &repairContext{WorktreePath: "/tmp/kit/GH-89"}
	resolvePRRepairContext = func(
		_ context.Context,
		_ io.Reader,
		_ io.Writer,
		_ string,
		prRef string,
	) (*repairContext, error) {
		calls = append(calls, "resolve")
		if prRef != "90" {
			t.Fatalf("repair PR ref = %q, want 90", prRef)
		}
		return repair, nil
	}
	reviewLoopFetchPRContext = func(prRef string) (reviewLoopPRContext, error) {
		calls = append(calls, "fetch")
		if prRef != "90" {
			t.Fatalf("fetch PR ref = %q, want 90", prRef)
		}
		return reviewLoopPRContext{HeadRefOID: "abc123"}, nil
	}
	reviewLoopWaitForCodeRabbit = func(ctx reviewLoopPRContext) error {
		calls = append(calls, "wait")
		if ctx.LocalRoot != repair.WorktreePath || ctx.Repair != repair {
			t.Fatalf("wait context did not preserve repair lane: %#v", ctx)
		}
		return nil
	}
	reviewLoopLoadReviewTasks = func(
		prRef string,
		_ bool,
	) ([]dispatchReviewTask, string, bool, error) {
		calls = append(calls, "load")
		if prRef != "90" {
			t.Fatalf("load PR ref = %q, want 90", prRef)
		}
		return nil, "", false, nil
	}

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)
	cmd.SetIn(strings.NewReader(""))
	if err := runReviewLoop(cmd, reviewLoopOptions{
		PRRef:        "90",
		Watch:        true,
		MaxSubagents: 1,
	}); err != nil {
		t.Fatalf("runReviewLoop() error = %v", err)
	}
	if want := []string{"resolve", "fetch", "wait", "load"}; !stdreflect.DeepEqual(calls, want) {
		t.Fatalf("call order = %v, want %v", calls, want)
	}
	if !strings.Contains(out.String(), "No actionable current review feedback found.") {
		t.Fatalf("no-actionable behavior missing:\n%s", out.String())
	}
}

func TestReviewLoopStopsBeforeRemoteReviewWorkWhenRepairContextFails(t *testing.T) {
	previousResolver := resolvePRRepairContext
	previousFetch := reviewLoopFetchPRContext
	t.Cleanup(func() {
		resolvePRRepairContext = previousResolver
		reviewLoopFetchPRContext = previousFetch
	})

	wantErr := errors.New("repository mismatch")
	resolvePRRepairContext = func(
		context.Context,
		io.Reader,
		io.Writer,
		string,
		string,
	) (*repairContext, error) {
		return nil, wantErr
	}
	reviewLoopFetchPRContext = func(string) (reviewLoopPRContext, error) {
		t.Fatal("remote PR metadata must not be fetched before repair context validation")
		return reviewLoopPRContext{}, nil
	}

	err := runReviewLoop(&cobra.Command{}, reviewLoopOptions{
		PRRef:        "90",
		MaxSubagents: 1,
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("runReviewLoop() error = %v, want %v", err, wantErr)
	}
}

func TestDispatchLoopRoutesToReviewLoop(t *testing.T) {
	previousExecutor := reviewLoopExecutor
	previousPR := dispatchPR
	previousCodeRabbit := dispatchCodeRabbit
	previousWatch := dispatchWatch
	previousFile := dispatchFile
	previousCopy := dispatchCopy
	previousMax := dispatchMaxSubagents
	defer func() {
		reviewLoopExecutor = previousExecutor
		dispatchPR = previousPR
		dispatchCodeRabbit = previousCodeRabbit
		dispatchWatch = previousWatch
		dispatchFile = previousFile
		dispatchCopy = previousCopy
		dispatchMaxSubagents = previousMax
	}()

	dispatchPR = "67"
	dispatchCodeRabbit = true
	dispatchWatch = true
	dispatchFile = ""
	dispatchCopy = true
	dispatchMaxSubagents = 4

	var got reviewLoopOptions
	reviewLoopExecutor = func(_ *cobra.Command, opts reviewLoopOptions) error {
		got = opts
		return nil
	}

	if err := runDispatchReviewLoopAlias(&cobra.Command{}, true); err != nil {
		t.Fatalf("dispatch --loop alias error = %v", err)
	}
	if got.PRRef != "67" || !got.CodeRabbitOnly || !got.Watch || !got.Copy || !got.OutputOnly || got.MaxSubagents != 4 {
		t.Fatalf("unexpected alias options: %#v", got)
	}
}

func TestDispatchLoopRejectsIncompatibleInputs(t *testing.T) {
	previousPR := dispatchPR
	previousFile := dispatchFile
	defer func() {
		dispatchPR = previousPR
		dispatchFile = previousFile
	}()

	dispatchPR = "67"
	dispatchFile = "tasks.md"
	err := runDispatchReviewLoopAlias(&cobra.Command{}, false)
	if err == nil || !strings.Contains(err.Error(), "--file cannot be used with --loop") {
		t.Fatalf("expected --file conflict, got %v", err)
	}

	dispatchPR = ""
	dispatchFile = ""
	err = runDispatchReviewLoopAlias(&cobra.Command{}, false)
	if err == nil || !strings.Contains(err.Error(), "--loop requires --pr") {
		t.Fatalf("expected missing --pr error, got %v", err)
	}
}

func TestReviewLoopClassifications(t *testing.T) {
	tmp := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(previousDir)
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("current.go", []byte("line 1\nline 2\nline 3\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx := reviewLoopPRContext{IssueHints: []string{"#67"}}
	tasks := []dispatchReviewTask{
		{Path: "current.go", Line: 1, Body: "Fix the current issue."},
		{Path: "current.go", Line: 2, Body: "This is valid but out of scope for this PR."},
		{Path: "current.go", Line: 3, Body: "This is a false positive after checking the code."},
		{Path: "current.go", Line: 0, Body: "Needs human decision."},
		{Path: "missing.go", Line: 1, Body: "Old comment."},
	}

	classified := classifyReviewLoopFindings(ctx, tasks)
	counts := map[reviewLoopClassification]int{}
	for _, finding := range classified {
		counts[finding.Kind]++
		if strings.TrimSpace(finding.Reason) == "" {
			t.Fatalf("expected reason for %#v", finding)
		}
	}

	for _, kind := range []reviewLoopClassification{
		reviewLoopFix,
		reviewLoopValidOutOfScope,
		reviewLoopFalsePositive,
		reviewLoopNeedsHuman,
		reviewLoopStale,
	} {
		if counts[kind] == 0 {
			t.Fatalf("expected classification %s in %#v", kind, counts)
		}
	}
}

func TestReviewLoopClassifiesRepoRelativePathsFromSubdirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "cmd"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "app.go"), []byte("line 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(previousDir)
	if err := os.Chdir(filepath.Join(root, "cmd")); err != nil {
		t.Fatal(err)
	}

	classified := classifyReviewLoopFindings(
		reviewLoopPRContext{LocalRoot: root},
		[]dispatchReviewTask{{Path: "internal/app.go", Line: 1, Body: "Fix app routing."}},
	)
	if len(classified) != 1 {
		t.Fatalf("classified length = %d, want 1", len(classified))
	}
	if classified[0].Kind != reviewLoopFix {
		t.Fatalf("classification = %s, want %s; reason: %s", classified[0].Kind, reviewLoopFix, classified[0].Reason)
	}
}

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
	err := runReviewLoopPrompt(out, reviewLoopOptions{MaxSubagents: 2}, ctx, classified, coderabbitSharedReviewInstruction)
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
	if err := runReviewLoopPrompt(out, reviewLoopOptions{MaxSubagents: 1, OutputOnly: true}, ctx, classified, ""); err != nil {
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
