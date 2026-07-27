package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/templates"
	"github.com/spf13/cobra"
)

func TestBuildReconcileReportProjectScopeFindsMissingGitignoreEntries(t *testing.T) {
	projectRoot := setupCoherentProjectForCheck(t)
	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	writeFile(t, filepath.Join(projectRoot, gitignorePath), "# custom ignores\ncustom.log\n.kit/runs/\n")

	report, err := buildReconcileReport(projectRoot, cfg, nil)
	if err != nil {
		t.Fatalf("buildReconcileReport() error = %v", err)
	}

	issues := findingsIssues(report.Findings)
	checks := []string{
		"missing Kit-managed `.gitignore` entries",
		"`.env`",
		"`.envrc`",
		"`.kit/cache/`",
	}
	for _, check := range checks {
		if !strings.Contains(issues, check) {
			t.Fatalf("expected gitignore scaffold finding %q, got %q", check, issues)
		}
	}
}

func TestBuildReconcileReportProjectScopeFindsMissingInitScaffoldArtifacts(t *testing.T) {
	projectRoot := setupCoherentProjectForCheck(t)
	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	for _, relativePath := range []string{envPath, envrcPath, codeRabbitConfigPath, pullRequestTemplatePath, autoAssignWorkflowPath} {
		if err := os.Remove(filepath.Join(projectRoot, relativePath)); err != nil {
			t.Fatalf("os.Remove(%s) error = %v", relativePath, err)
		}
	}

	report, err := buildReconcileReport(projectRoot, cfg, nil)
	if err != nil {
		t.Fatalf("buildReconcileReport() error = %v", err)
	}

	issues := findingsIssues(report.Findings)
	for _, check := range []string{
		"missing Kit init scaffold artifact `.env`",
		"missing Kit init scaffold artifact `.envrc`",
		"missing Kit init scaffold artifact `.coderabbit.yaml`",
		"missing Kit init scaffold artifact `.github/pull_request_template.md`",
		"missing Kit init scaffold artifact `.github/workflows/auto-assign.yml`",
	} {
		if !strings.Contains(issues, check) {
			t.Fatalf("expected init scaffold finding %q, got %q", check, issues)
		}
	}
}

func TestBuildReconcileReportAllowsMissingLocalEnvironmentFilesInLinkedCheckout(t *testing.T) {
	projectRoot := setupCoherentProjectForCheck(t)
	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	for _, relativePath := range []string{envPath, envrcPath, codeRabbitConfigPath} {
		if err := os.Remove(filepath.Join(projectRoot, relativePath)); err != nil {
			t.Fatalf("os.Remove(%s) error = %v", relativePath, err)
		}
	}
	commonDir := filepath.Join(t.TempDir(), "common")
	worktreeGitDir := filepath.Join(commonDir, "worktrees", "linked")
	writeFile(t, filepath.Join(worktreeGitDir, "commondir"), "../..\n")
	writeFile(t, filepath.Join(projectRoot, ".git"), "gitdir: "+worktreeGitDir+"\n")

	report, err := buildReconcileReport(projectRoot, cfg, nil)
	if err != nil {
		t.Fatalf("buildReconcileReport() error = %v", err)
	}

	issues := findingsIssues(report.Findings)
	for _, unexpected := range []string{
		"missing Kit init scaffold artifact `.env`",
		"missing Kit init scaffold artifact `.envrc`",
	} {
		if strings.Contains(issues, unexpected) {
			t.Fatalf("linked checkout should not require local-only scaffold %q, got %q", unexpected, issues)
		}
	}
	if !strings.Contains(issues, "missing Kit init scaffold artifact `.coderabbit.yaml`") {
		t.Fatalf("linked checkout must still require non-local scaffold artifacts, got %q", issues)
	}
}

func TestUsesLinkedWorktreeCheckoutRejectsSubmoduleGitFile(t *testing.T) {
	projectRoot := t.TempDir()
	moduleGitDir := filepath.Join(t.TempDir(), "modules", "example")
	if err := os.MkdirAll(moduleGitDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}
	writeFile(t, filepath.Join(projectRoot, ".git"), "gitdir: "+moduleGitDir+"\n")

	if usesLinkedWorktreeCheckout(projectRoot) {
		t.Fatal("submodule-style .git file must not be treated as linked-worktree metadata")
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
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), withFeatureFrontMatter(validSpecWithRelationships("none\n"), "spec", "0001-sample"))
	writeFile(t, filepath.Join(featurePath, "PLAN.md"), withFeatureFrontMatter(validPlan(), "plan", "0001-sample"))
	writeFile(t, filepath.Join(featurePath, "TASKS.md"), withFeatureFrontMatter(validTasks(), "tasks", "0001-sample"))
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
				UpdateInstruction: "refresh progress summary",
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
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	writeFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), validConstitution())
	writeFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), validProgressSummary("", ""))
	writeFile(t, filepath.Join(projectRoot, "AGENTS.md"), templates.AgentsMD)
	writeFile(t, filepath.Join(projectRoot, "CLAUDE.md"), templates.ClaudeMD)
	writeFile(t, filepath.Join(projectRoot, ".github", "copilot-instructions.md"), templates.CopilotInstructionsMD)
	writeInitScaffoldArtifacts(t, projectRoot)
	for _, support := range templates.InstructionSupportFiles(config.InstructionScaffoldVersionTOC) {
		writeFile(t, filepath.Join(projectRoot, support.RelativePath), support.Content)
	}

	report, err := buildReconcileReport(projectRoot, cfg, nil)
	if err != nil {
		t.Fatalf("buildReconcileReport() error = %v", err)
	}

	if len(report.Findings) != 0 {
		t.Fatalf("expected clean project report, got %#v", report.Findings)
	}
}
