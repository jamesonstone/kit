package cli

import (
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

func setupLoopReviewProject(t *testing.T, agentScript string) string {
	t.Helper()
	projectRoot := t.TempDir()
	agentPath := filepath.Join(projectRoot, "agent.sh")
	if err := os.WriteFile(agentPath, []byte(agentScript), 0o755); err != nil {
		t.Fatalf("WriteFile(agent) error = %v", err)
	}
	cfg := config.Default()
	cfg.Loop.Agent = config.LoopAgentConfig{Command: agentPath}
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	return projectRoot
}

func installLoopReviewGitFake(t *testing.T) func() {
	t.Helper()
	return installReviewLoopFakes(t, nil, fakeReviewLoopRunner{
		output: func(_ string, name string, args ...string) ([]byte, error) {
			if name != "git" {
				return nil, fmt.Errorf("unexpected command: %s %v", name, args)
			}
			joined := strings.Join(args, " ")
			switch joined {
			case "rev-parse --verify origin/main":
				return []byte("origin/main\n"), nil
			case "diff --name-only origin/main...HEAD":
				return []byte("internal/app.go\n"), nil
			case "diff --name-only", "diff --cached --name-only":
				return nil, nil
			case "diff --stat origin/main...HEAD":
				return []byte(" internal/app.go | 2 +-\n"), nil
			case "diff --stat", "diff --cached --stat":
				return nil, nil
			default:
				return nil, fmt.Errorf("unexpected git args: %s", joined)
			}
		},
	})
}

func installLoopReviewPRFake(t *testing.T, status reviewLoopCheckStatus) func() {
	t.Helper()
	return installReviewLoopFakes(t, nil, fakeReviewLoopRunner{
		output: func(_ string, name string, args ...string) ([]byte, error) {
			if name == "gh" {
				return reviewLoopPRPayload("abc123"), nil
			}
			if name != "git" {
				return nil, fmt.Errorf("unexpected command: %s %v", name, args)
			}
			return installLoopReviewGitOutput(args...)
		},
		outputAllowError: func(_ string, name string, args ...string) ([]byte, error) {
			if name != "gh" {
				return nil, fmt.Errorf("unexpected command: %s %v", name, args)
			}
			switch status {
			case reviewLoopCheckPending:
				return []byte(`[{"name":"CodeRabbit","state":"PENDING","bucket":"pending"}]`), nil
			case reviewLoopCheckComplete:
				return []byte(`[{"name":"CodeRabbit","state":"SUCCESS","bucket":"pass"}]`), nil
			default:
				return []byte(`[]`), nil
			}
		},
	})
}

func installLoopReviewGitOutput(args ...string) ([]byte, error) {
	joined := strings.Join(args, " ")
	switch joined {
	case "rev-parse --verify origin/main":
		return []byte("origin/main\n"), nil
	case "diff --name-only origin/main...HEAD":
		return []byte("internal/app.go\n"), nil
	case "diff --name-only", "diff --cached --name-only":
		return nil, nil
	case "diff --stat origin/main...HEAD":
		return []byte(" internal/app.go | 2 +-\n"), nil
	case "diff --stat", "diff --cached --stat":
		return nil, nil
	default:
		return nil, fmt.Errorf("unexpected git args: %s", joined)
	}
}
