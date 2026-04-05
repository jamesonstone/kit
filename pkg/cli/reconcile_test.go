package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/templates"
	"github.com/spf13/cobra"
)

func TestBuildReconcilePromptIncludesScopeRulesAndVerification(t *testing.T) {
	previousSingleAgent := singleAgent
	singleAgent = false
	t.Cleanup(func() { singleAgent = previousSingleAgent })

	projectRoot := t.TempDir()
	report := &reconcileReport{
		ProjectRoot: projectRoot,
		Feature: &feature.Feature{
			Slug:    "sample",
			DirName: "0001-sample",
			Path:    filepath.Join(projectRoot, "docs", "specs", "0001-sample"),
		},
		NeedsRollup: true,
		Findings: []reconcileFinding{
			{
				Severity:          reconcileSeverityError,
				FilePath:          filepath.Join(projectRoot, "docs", "specs", "0001-sample", "TASKS.md"),
				Issue:             "task `T001` exists in `PROGRESS TABLE` but not in `TASK DETAILS`",
				ContractSource:    initProjectSource(projectRoot),
				UpdateInstruction: "add the missing task-details block",
				SearchHints:       []string{"rg -n \"^### T[0-9]{3}$\" /tmp/TASKS.md"},
			},
		},
	}

	prompt := buildReconcilePrompt(report)
	checks := []string{
		"/plan",
		"feature sample",
		"docs only; no product code",
		"use subagents and queue work according to overlapping file changes",
		"contract order:",
		"Audit snapshot:",
		"Files to fix:",
		"| Severity | File | Issues | Focus |",
		"align task IDs",
		"Notable issues:",
		"`TASKS.md`: task `T001` exists in `PROGRESS TABLE` but not in `TASK DETAILS`",
		"Search shortcuts:",
		"`kit check sample`",
		"`kit rollup`",
		"`Findings`",
		"`Updates`",
		"`Verification`",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected prompt to contain %q", check)
		}
	}

	if strings.Contains(prompt, "`Open Questions`") || strings.Contains(prompt, "`Search Plan`") {
		t.Fatalf("expected compact response contract, got %q", prompt)
	}
}

func TestBuildReconcilePrompt_OmitsSubagentInstructionForSingleAgent(t *testing.T) {
	previousSingleAgent := singleAgent
	singleAgent = true
	t.Cleanup(func() { singleAgent = previousSingleAgent })

	report := &reconcileReport{
		ProjectRoot: t.TempDir(),
		Findings: []reconcileFinding{
			{
				Severity: reconcileSeverityError,
				FilePath: "/tmp/TASKS.md",
				Issue:    "task `T001` exists in `PROGRESS TABLE` but not in `TASK DETAILS`",
			},
		},
	}

	prompt := buildReconcilePrompt(report)
	if strings.Contains(prompt, "use subagents and queue work according to overlapping file changes") {
		t.Fatalf("expected single-agent prompt to omit explicit subagent instruction, got %q", prompt)
	}
}

func TestBuildReconcilePromptGroupsFindingsByFile(t *testing.T) {
	projectRoot := t.TempDir()
	target := filepath.Join(projectRoot, "docs", "specs", "0001-sample", "TASKS.md")
	report := &reconcileReport{
		ProjectRoot: projectRoot,
		Findings: []reconcileFinding{
			{
				Severity:          reconcileSeverityError,
				FilePath:          target,
				Issue:             "task `T001` exists in `PROGRESS TABLE` but not in `TASK DETAILS`",
				ContractSource:    initProjectSource(projectRoot),
				UpdateInstruction: "align task details",
				SearchHints:       []string{"rg -n \"task\" " + target},
			},
			{
				Severity:          reconcileSeverityError,
				FilePath:          target,
				Issue:             "task `T001` exists in `TASK LIST` but not in `PROGRESS TABLE`",
				ContractSource:    initProjectSource(projectRoot),
				UpdateInstruction: "align task list",
				SearchHints:       []string{"rg -n \"task\" " + target},
			},
		},
	}

	prompt := buildReconcilePrompt(report)
	if got := strings.Count(prompt, "`"+target+"`"); got != 1 {
		t.Fatalf("expected grouped file path once, got %d in %q", got, prompt)
	}
}

func TestBuildReconcileReportFeatureScopeFindsRelationshipAndTaskDrift(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-sample")
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), validSpecWithRelationships("- related to: 0002-missing\n"))
	writeFile(t, filepath.Join(featurePath, "PLAN.md"), validPlan())
	writeFile(t, filepath.Join(featurePath, "TASKS.md"), invalidTasksMissingDetail())
	writeFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), validProgressSummary("0001", "sample"))

	feat := &feature.Feature{
		Number:  1,
		Slug:    "sample",
		DirName: "0001-sample",
		Path:    featurePath,
	}

	report, err := buildReconcileReport(projectRoot, cfg, feat)
	if err != nil {
		t.Fatalf("buildReconcileReport() error = %v", err)
	}

	issues := findingsIssues(report.Findings)
	checks := []string{
		"relationship target `0002-missing` does not exist in `docs/specs/`",
		"task `T001` exists in `PROGRESS TABLE` but not in `TASK DETAILS`",
	}

	for _, check := range checks {
		if !strings.Contains(issues, check) {
			t.Fatalf("expected findings to include %q, got %q", check, issues)
		}
	}
}

func TestBuildReconcileReportProjectScopeFindsInstructionFileDrift(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	writeFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), validConstitution())
	writeFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), validProgressSummary("", ""))

	report, err := buildReconcileReport(projectRoot, cfg, nil)
	if err != nil {
		t.Fatalf("buildReconcileReport() error = %v", err)
	}

	issues := findingsIssues(report.Findings)
	if !strings.Contains(issues, "missing Kit-managed repository instruction file") {
		t.Fatalf("expected instruction-file drift finding, got %q", issues)
	}
}

func TestRunReconcileRejectsAllWithFeatureArg(t *testing.T) {
	reconcileAll = true
	t.Cleanup(func() { reconcileAll = false })

	cmd := &cobra.Command{}
	err := runReconcile(cmd, []string{"sample"})
	if err == nil || !strings.Contains(err.Error(), "--all cannot be used with a feature argument") {
		t.Fatalf("expected --all validation error, got %v", err)
	}
}

func TestRunReconcileCleanFeaturePrintsSuccess(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-sample")
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), validSpecWithRelationships("none\n"))
	writeFile(t, filepath.Join(featurePath, "PLAN.md"), validPlan())
	writeFile(t, filepath.Join(featurePath, "TASKS.md"), validTasks())
	writeFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), validProgressSummary("0001", "sample"))

	setWorkingDirectory(t, projectRoot)

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.Flags().Bool("output-only", false, "")

	reconcileAll = false
	if err := runReconcile(cmd, []string{"sample"}); err != nil {
		t.Fatalf("runReconcile() error = %v", err)
	}

	if got := out.String(); !strings.Contains(got, "No reconciliation needed for feature sample.") {
		t.Fatalf("expected clean success output, got %q", got)
	}
}

func TestRenderReconcileSummaryShowsCompactTable(t *testing.T) {
	projectRoot := t.TempDir()
	report := &reconcileReport{
		ProjectRoot: projectRoot,
		Findings: []reconcileFinding{
			{
				Severity:          reconcileSeverityError,
				FilePath:          filepath.Join(projectRoot, "docs", "specs", "0001-sample", "TASKS.md"),
				Issue:             "task `T001` exists in `PROGRESS TABLE` but not in `TASK DETAILS`",
				UpdateInstruction: "align task IDs",
			},
			{
				Severity:          reconcileSeverityWarning,
				FilePath:          filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"),
				Issue:             "progress summary is missing the feature summary heading for `0001-sample`",
				UpdateInstruction: "refresh rollup",
			},
		},
	}

	summary := renderReconcileSummary(report, humanOutputStyle{})
	checks := []string{
		"Reconcile Audit",
		"Scope: whole project",
		"Findings: 2 (1 errors, 1 warnings) across 2 files",
		"Severity  Issues",
		"E1",
		"W1",
		"raw prompt stays compact",
	}

	for _, check := range checks {
		if !strings.Contains(summary, check) {
			t.Fatalf("expected summary to contain %q, got %q", check, summary)
		}
	}
}

func TestReconcileProjectScopeWithCurrentInstructionFilesIsClean(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	writeFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), validConstitution())
	writeFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), validProgressSummary("", ""))
	writeFile(t, filepath.Join(projectRoot, "AGENTS.md"), templates.AgentsMD)
	writeFile(t, filepath.Join(projectRoot, "CLAUDE.md"), templates.ClaudeMD)
	writeFile(t, filepath.Join(projectRoot, ".github", "copilot-instructions.md"), templates.CopilotInstructionsMD)

	report, err := buildReconcileReport(projectRoot, cfg, nil)
	if err != nil {
		t.Fatalf("buildReconcileReport() error = %v", err)
	}

	if len(report.Findings) != 0 {
		t.Fatalf("expected clean project report, got %#v", report.Findings)
	}
}
