package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	stdreflect "reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestReviewLoopCommandRequiresPR(t *testing.T) {
	err := runReviewLoop(&cobra.Command{}, reviewLoopOptions{})
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
		PRRef: "90",
		Watch: true,
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
		PRRef: "90",
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
	defer func() {
		reviewLoopExecutor = previousExecutor
		dispatchPR = previousPR
		dispatchCodeRabbit = previousCodeRabbit
		dispatchWatch = previousWatch
		dispatchFile = previousFile
		dispatchCopy = previousCopy
	}()

	dispatchPR = "67"
	dispatchCodeRabbit = true
	dispatchWatch = true
	dispatchFile = ""
	dispatchCopy = true

	var got reviewLoopOptions
	reviewLoopExecutor = func(_ *cobra.Command, opts reviewLoopOptions) error {
		got = opts
		return nil
	}

	if err := runDispatchReviewLoopAlias(&cobra.Command{}, true); err != nil {
		t.Fatalf("dispatch --loop alias error = %v", err)
	}
	if got.PRRef != "67" || !got.CodeRabbitOnly || !got.Watch || !got.Copy || !got.OutputOnly {
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
	defer func() { _ = os.Chdir(previousDir) }()
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
	defer func() { _ = os.Chdir(previousDir) }()
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
