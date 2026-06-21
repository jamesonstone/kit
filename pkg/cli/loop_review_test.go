package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestLoopReviewCommandIsRegisteredUnderLoop(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"loop", "review"})
	if err != nil {
		t.Fatalf("rootCmd.Find(loop review) error = %v", err)
	}
	if cmd == nil || cmd.Name() != "review" {
		t.Fatalf("expected loop review command, got %#v", cmd)
	}
	if flag := cmd.Flags().Lookup("subagents"); flag == nil {
		t.Fatal("expected loop review to expose --subagents")
	}

	workflow, _, err := rootCmd.Find([]string{"loop", "workflow"})
	if err != nil {
		t.Fatalf("rootCmd.Find(loop workflow) error = %v", err)
	}
	if workflow == nil || workflow.Name() != "workflow" {
		t.Fatalf("expected loop workflow command, got %#v", workflow)
	}
}

func TestParseLoopReviewAgentResult(t *testing.T) {
	result := parseLoopReviewAgentResult(`Correctness: 96%
Status: done

- Issue: nil path; Fix: added guard.
done
`)
	if !result.Done {
		t.Fatal("Done = false, want true")
	}
	if result.Correctness != 96 {
		t.Fatalf("Correctness = %d, want 96", result.Correctness)
	}
	if len(result.Bullets) != 1 || !strings.Contains(result.Bullets[0], "nil path") {
		t.Fatalf("Bullets = %#v", result.Bullets)
	}

	notDone := parseLoopReviewAgentResult("Correctness: 99%\n- Issue: x; Fix: y.\n")
	if notDone.Done {
		t.Fatal("Done = true without final done line")
	}
}

func TestResolveLoopReviewBaseFallsBackToMain(t *testing.T) {
	restore := installReviewLoopFakes(t, nil, fakeReviewLoopRunner{
		output: func(_ string, name string, args ...string) ([]byte, error) {
			if name != "git" || strings.Join(args, " ") != "rev-parse --verify main" {
				return nil, fmt.Errorf("missing ref")
			}
			return []byte("main\n"), nil
		},
	})
	defer restore()

	base, err := resolveLoopReviewBase(t.TempDir(), "")
	if err != nil {
		t.Fatalf("resolveLoopReviewBase() error = %v", err)
	}
	if base != "main" {
		t.Fatalf("base = %q, want main", base)
	}
}

func TestExecuteLoopReviewStopsWhenAgentReportsDone(t *testing.T) {
	projectRoot := setupLoopReviewProject(t, `#!/bin/sh
set -eu
cat >/dev/null
printf 'Correctness: 96%%\nStatus: done\n\n- Issue: nil path; Fix: added guard.\ndone\n'
`)
	restore := installLoopReviewGitFake(t)
	defer restore()

	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	report, err := executeLoopReview(context.Background(), loopReviewOptions{
		ProjectRoot: projectRoot,
		Config:      cfg,
		Agent:       cfg.Loop.Agent,
	})
	if err != nil {
		t.Fatalf("executeLoopReview() error = %v", err)
	}
	if report.Status != "complete" || report.Correctness != 96 {
		t.Fatalf("report = %#v", report)
	}
	if len(report.Iterations) != 1 {
		t.Fatalf("Iterations = %d, want 1", len(report.Iterations))
	}
	if _, err := os.Stat(filepath.Join(projectRoot, filepath.FromSlash(report.ArtifactDir), "run.json")); err != nil {
		t.Fatalf("expected run artifact: %v", err)
	}
}

func TestExecuteLoopReviewContinuesUntilDoneAndConfidence(t *testing.T) {
	projectRoot := setupLoopReviewProject(t, `#!/bin/sh
set -eu
cat >/dev/null
if [ "$KIT_LOOP_ITERATION" = "1" ]; then
  printf 'Correctness: 90%%\nStatus: fixing\n\n- Issue: missing test; Fix: added next pass.\n'
else
  printf 'Correctness: 97%%\nStatus: done\n\n- Issue: missing test; Fix: added focused test.\ndone\n'
fi
`)
	restore := installLoopReviewGitFake(t)
	defer restore()

	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	report, err := executeLoopReview(context.Background(), loopReviewOptions{
		ProjectRoot: projectRoot,
		Config:      cfg,
		Agent:       cfg.Loop.Agent,
	})
	if err != nil {
		t.Fatalf("executeLoopReview() error = %v", err)
	}
	if len(report.Iterations) != 2 {
		t.Fatalf("Iterations = %d, want 2", len(report.Iterations))
	}
	if report.Correctness != 97 {
		t.Fatalf("Correctness = %d, want 97", report.Correctness)
	}
}

func TestExecuteLoopReviewStopsOnAgentCommandFailure(t *testing.T) {
	projectRoot := setupLoopReviewProject(t, `#!/bin/sh
set -eu
cat >/dev/null
printf 'OpenAI Codex v0.140.0\n' >&2
printf 'ERROR: model is not supported\n' >&2
exit 2
`)
	restore := installLoopReviewGitFake(t)
	defer restore()

	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	report, err := executeLoopReview(context.Background(), loopReviewOptions{
		ProjectRoot: projectRoot,
		Config:      cfg,
		Agent:       cfg.Loop.Agent,
	})
	if err == nil {
		t.Fatal("executeLoopReview() error = nil, want agent failure")
	}
	if len(report.Iterations) != 1 {
		t.Fatalf("Iterations = %d, want 1", len(report.Iterations))
	}
	if !strings.Contains(report.StopReason, "agent command failed at iteration 1") {
		t.Fatalf("StopReason = %q", report.StopReason)
	}
	if !strings.Contains(report.StopReason, "ERROR: model is not supported") {
		t.Fatalf("StopReason missing stderr context: %q", report.StopReason)
	}
	if strings.Contains(report.StopReason, "OpenAI Codex v0.140.0") {
		t.Fatalf("StopReason used banner instead of actionable error: %q", report.StopReason)
	}
	if strings.Contains(report.StopReason, "max iterations reached") {
		t.Fatalf("StopReason incorrectly reports max iterations: %q", report.StopReason)
	}
}

func TestExecuteLoopReviewEmitsProgressAndStreamsAgentOutput(t *testing.T) {
	projectRoot := setupLoopReviewProject(t, `#!/bin/sh
set -eu
printf 'agent-visible-status\n' >&2
cat >/dev/null
printf 'Correctness: 96%%\nStatus: done\n\n- Issue: nil path; Fix: added guard.\ndone\n'
`)
	restore := installLoopReviewGitFake(t)
	defer restore()

	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	var progress bytes.Buffer
	report, err := executeLoopReview(context.Background(), loopReviewOptions{
		ProjectRoot: projectRoot,
		Config:      cfg,
		Agent:       cfg.Loop.Agent,
		Progress:    &loopReviewSynchronizedWriter{writer: &progress},
	})
	if err != nil {
		t.Fatalf("executeLoopReview() error = %v", err)
	}
	output := progress.String()
	for _, want := range []string{
		"run " + report.RunID + " started",
		"🤖 single-agent mode enabled",
		"target resolved: base=origin/main changed_files=1",
		"artifacts: " + report.ArtifactDir,
		"iteration 1/10: prompt written",
		"iteration 1/10: running agent:",
		"agent process started",
		"agent stderr: agent-visible-status",
		"agent stdout: Correctness: 96%",
		"iteration 1/10: parsed result done=true correctness=96%",
		"run " + report.RunID + " complete: correctness=96%",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("progress output missing %q:\n%s", want, output)
		}
	}
}

func TestExecuteLoopReviewIngestsPRFeedbackIntoNextPass(t *testing.T) {
	projectRoot := setupLoopReviewProject(t, `#!/bin/sh
set -eu
if [ "$KIT_LOOP_ITERATION" = "2" ]; then
  grep -q 'Fix app routing' || exit 7
fi
cat >/dev/null
printf 'Correctness: 96%%\nStatus: done\n\n- Issue: app routing; Fix: adjusted branch.\ndone\n'
`)
	restore := installLoopReviewPRFake(t, reviewLoopCheckComplete)
	defer restore()
	reviewLoopLoadReviewTasks = func(_ string, _ bool) ([]dispatchReviewTask, string, bool, error) {
		return []dispatchReviewTask{{Path: "internal/app.go", Line: 12, Body: "Fix app routing.", URL: "https://example.com/1"}}, "", true, nil
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	report, err := executeLoopReview(context.Background(), loopReviewOptions{
		ProjectRoot: projectRoot,
		Config:      cfg,
		Agent:       cfg.Loop.Agent,
		PRRef:       "Patient-Driven-Care/cortex#67",
	})
	if err != nil {
		t.Fatalf("executeLoopReview() error = %v", err)
	}
	if len(report.Iterations) != 2 {
		t.Fatalf("Iterations = %d, want 2", len(report.Iterations))
	}
	if report.PRStatus != "CodeRabbit complete" {
		t.Fatalf("PRStatus = %q", report.PRStatus)
	}
}

func TestExecuteLoopReviewExitsProvisionallyWhenCodeRabbitPending(t *testing.T) {
	projectRoot := setupLoopReviewProject(t, `#!/bin/sh
set -eu
cat >/dev/null
printf 'Correctness: 96%%\nStatus: done\n\n- No high, medium, or correctness-impacting issues found.\ndone\n'
`)
	restore := installLoopReviewPRFake(t, reviewLoopCheckPending)
	defer restore()
	reviewLoopLoadReviewTasks = func(_ string, _ bool) ([]dispatchReviewTask, string, bool, error) {
		return nil, "", false, nil
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	report, err := executeLoopReview(context.Background(), loopReviewOptions{
		ProjectRoot: projectRoot,
		Config:      cfg,
		Agent:       cfg.Loop.Agent,
		PRRef:       "Patient-Driven-Care/cortex#67",
	})
	if err != nil {
		t.Fatalf("executeLoopReview() error = %v", err)
	}
	if report.Status != "complete" || report.PRStatus != "local done, CodeRabbit pending" {
		t.Fatalf("report = %#v", report)
	}
	if !strings.Contains(report.StopReason, "Rerun") && !strings.Contains(report.StopReason, "rerun") {
		t.Fatalf("expected rerun guidance, got %q", report.StopReason)
	}
}
