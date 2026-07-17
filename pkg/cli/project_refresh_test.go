package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
)

func TestProjectRefreshStatusDueByFeatureInterval(t *testing.T) {
	projectRoot, cfg := setupProjectRefreshTestProject(t)
	createCompletedV2Feature(t, projectRoot, "0001-alpha")
	createCompletedV2Feature(t, projectRoot, "0002-beta")
	createCompletedV2Feature(t, projectRoot, "0003-gamma")
	cfg.ProjectRefresh.Constitution.FeatureInterval = 2
	cfg.ProjectRefresh.Constitution.LastCompletedFeatureCount = 1

	status, err := calculateProjectRefreshStatus(projectRoot, cfg, time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("calculateProjectRefreshStatus() error = %v", err)
	}
	if !status.Due {
		t.Fatalf("expected refresh to be due, got %#v", status)
	}
	if status.FeaturesSinceLastReview != 2 {
		t.Fatalf("FeaturesSinceLastReview = %d, want 2", status.FeaturesSinceLastReview)
	}
}

func TestProjectRefreshStatusDueByMaxAge(t *testing.T) {
	projectRoot, cfg := setupProjectRefreshTestProject(t)
	cfg.ProjectRefresh.Constitution.MaxAgeDays = 30
	cfg.ProjectRefresh.Constitution.LastReviewedAt = "2026-05-01T00:00:00Z"

	status, err := calculateProjectRefreshStatus(projectRoot, cfg, time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("calculateProjectRefreshStatus() error = %v", err)
	}
	if !status.Due {
		t.Fatalf("expected refresh to be due by age, got %#v", status)
	}
	if !strings.Contains(strings.Join(status.Reasons, " "), "day(s) ago") {
		t.Fatalf("expected age due reason, got %#v", status.Reasons)
	}
}

func TestProjectRefreshStatusNotDueBelowThresholds(t *testing.T) {
	projectRoot, cfg := setupProjectRefreshTestProject(t)
	createCompletedV2Feature(t, projectRoot, "0001-alpha")
	cfg.ProjectRefresh.Constitution.FeatureInterval = 5
	cfg.ProjectRefresh.Constitution.MaxAgeDays = 30
	cfg.ProjectRefresh.Constitution.LastReviewedAt = "2026-06-20T00:00:00Z"

	status, err := calculateProjectRefreshStatus(projectRoot, cfg, time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("calculateProjectRefreshStatus() error = %v", err)
	}
	if status.Due {
		t.Fatalf("expected refresh to not be due, got %#v", status)
	}
}

func TestRunProjectRefreshOutputOnlyEmitsPromptWithoutWritingConstitution(t *testing.T) {
	projectRoot, _ := setupProjectRefreshTestProject(t)
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	before := readFile(t, constitutionPath)
	setWorkingDirectory(t, projectRoot)

	output := captureStdout(t, func() {
		if err := runProjectRefreshCommand(&cobra.Command{}, projectRefreshOptions{OutputOnly: true}, time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)); err != nil {
			t.Fatalf("runProjectRefreshCommand() error = %v", err)
		}
	})

	for _, check := range []string{
		"## Project Refresh",
		"Normal post-validation work performs continuous Constitution curation",
		"follow `docs/references/rules/constitution-curation.md`",
		"treat cadence as a review trigger",
		"docs/CONSTITUTION.md",
		"docs/PROJECT_PROGRESS_SUMMARY.md",
		"Feature docs:",
		"Discovery commands:",
		"Analyze for durable changes:",
		"Update guidance:",
		"Verification:",
		"`kit reconcile --all --include-files`",
		"`Findings`",
	} {
		if !strings.Contains(output, check) {
			t.Fatalf("expected prompt to contain %q, got:\n%s", check, output)
		}
	}
	after := readFile(t, constitutionPath)
	if after != before {
		t.Fatalf("output-only refresh modified Constitution:\nbefore=%q\nafter=%q", before, after)
	}
}

func TestRunProjectRefreshConstitutionOnlyPromptScopesUpdates(t *testing.T) {
	projectRoot, _ := setupProjectRefreshTestProject(t)
	setWorkingDirectory(t, projectRoot)

	output := captureStdout(t, func() {
		if err := runProjectRefreshCommand(&cobra.Command{}, projectRefreshOptions{ConstitutionOnly: true, OutputOnly: true}, time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)); err != nil {
			t.Fatalf("runProjectRefreshCommand() error = %v", err)
		}
	})

	if !strings.Contains(output, "`docs/CONSTITUTION.md`: refresh durable project-wide principles") {
		t.Fatalf("expected Constitution update guidance, got:\n%s", output)
	}
	for _, unwanted := range []string{
		"`docs/PROJECT_PROGRESS_SUMMARY.md`: update",
		"feature docs under `docs/specs/`: leave alone unless",
		"`docs/agents/*`, `AGENTS.md`, `CLAUDE.md`",
	} {
		if strings.Contains(output, unwanted) {
			t.Fatalf("Constitution-only prompt should not include %q, got:\n%s", unwanted, output)
		}
	}
}

func TestRunProjectRefreshCopyUsesSharedPromptCopyBehavior(t *testing.T) {
	projectRoot, _ := setupProjectRefreshTestProject(t)
	setWorkingDirectory(t, projectRoot)
	previous := clipboardCopyFunc
	t.Cleanup(func() {
		clipboardCopyFunc = previous
	})
	var copied string
	clipboardCopyFunc = func(text string) error {
		copied = text
		return nil
	}

	output := captureStdout(t, func() {
		if err := runProjectRefreshCommand(&cobra.Command{}, projectRefreshOptions{OutputOnly: true, Copy: true}, time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)); err != nil {
			t.Fatalf("runProjectRefreshCommand() error = %v", err)
		}
	})

	if !strings.Contains(copied, "## Project Refresh") {
		t.Fatalf("expected copied prompt, got %q", copied)
	}
	if output != copied {
		t.Fatalf("output-only --copy should print the same prompt it copies")
	}
}

func TestRunProjectRefreshNowRecordsCadenceWithoutWritingConstitution(t *testing.T) {
	projectRoot, _ := setupProjectRefreshTestProject(t)
	createCompletedV2Feature(t, projectRoot, "0001-alpha")
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	before := readFile(t, constitutionPath)
	setWorkingDirectory(t, projectRoot)

	cmd := &cobra.Command{}
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	now := time.Date(2026, 6, 28, 12, 30, 0, 0, time.UTC)
	if err := runProjectRefreshCommand(cmd, projectRefreshOptions{Now: true}, now); err != nil {
		t.Fatalf("runProjectRefreshCommand(--now) error = %v", err)
	}

	updated, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	refresh := updated.ProjectRefresh.Constitution
	if refresh.LastReviewedAt != now.Format(time.RFC3339) {
		t.Fatalf("LastReviewedAt = %q, want %q", refresh.LastReviewedAt, now.Format(time.RFC3339))
	}
	if refresh.LastCompletedFeatureCount != 1 {
		t.Fatalf("LastCompletedFeatureCount = %d, want 1", refresh.LastCompletedFeatureCount)
	}
	after := readFile(t, constitutionPath)
	if after != before {
		t.Fatalf("--now modified Constitution:\nbefore=%q\nafter=%q", before, after)
	}
}

func setupProjectRefreshTestProject(t *testing.T) (string, *config.Config) {
	t.Helper()

	projectRoot := t.TempDir()
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), validConstitution())
	writeFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), validProgressSummary("", ""))
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	return projectRoot, cfg
}

func createCompletedV2Feature(t *testing.T, projectRoot string, dirName string) {
	t.Helper()

	writeFile(
		t,
		filepath.Join(projectRoot, "docs", "specs", dirName, "SPEC.md"),
		validV2SpecWithPhase(dirName, "complete"),
	)
}
