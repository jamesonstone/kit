package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
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
		"missing Kit init scaffold artifact `.coderabbit.yaml`",
		"missing Kit init scaffold artifact `.github/pull_request_template.md`",
		"missing Kit init scaffold artifact `.github/workflows/auto-assign.yml`",
	} {
		if !strings.Contains(issues, check) {
			t.Fatalf("expected init scaffold finding %q, got %q", check, issues)
		}
	}
}

func TestBuildReconcileReportIgnoresMissingLocalEnvironmentFiles(t *testing.T) {
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

func TestReconcileProjectScopeWithCurrentInstructionFilesIsClean(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.CurrentInstructionScaffoldVersion
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	writeFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), validConstitution())
	writeInitScaffoldArtifacts(t, projectRoot)
	writeCurrentInstructionArtifacts(t, projectRoot)
	writeStandingAuthorityPolicies(t, projectRoot)

	report, err := buildReconcileReport(projectRoot, cfg, nil)
	if err != nil {
		t.Fatalf("buildReconcileReport() error = %v", err)
	}

	if len(report.Findings) != 0 {
		t.Fatalf("expected clean project report, got %#v", report.Findings)
	}
}

// The local-only .env and .envrc need no action when absent, so they are not
// findings at all; tracked scaffold files still are.
func TestMissingLocalEnvironmentFilesAreNotFindings(t *testing.T) {
	projectRoot := t.TempDir()
	sawCodeRabbit := false
	for _, finding := range auditInitScaffoldArtifacts(projectRoot) {
		base := filepath.Base(finding.FilePath)
		if base == ".env" || base == ".envrc" {
			t.Fatalf("missing local-only %s reported: %#v", base, finding)
		}
		if base == ".coderabbit.yaml" {
			sawCodeRabbit = !finding.NonBlocking && strings.Contains(finding.UpdateInstruction, "kit reconcile")
		}
	}
	if !sawCodeRabbit {
		t.Fatal("missing tracked scaffold file must still block and point to kit reconcile")
	}
}
