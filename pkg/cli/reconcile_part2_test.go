package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/templates"
)

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

func TestBuildReconcileReportWarnsForActiveVerificationFields(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-sample")
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), withFeatureFrontMatter(validSpecWithRelationships("none\n"), "spec", "0001-sample"))
	writeFile(t, filepath.Join(featurePath, "PLAN.md"), withFeatureFrontMatter(validPlan(), "plan", "0001-sample"))
	writeFile(t, filepath.Join(featurePath, "TASKS.md"), withFeatureFrontMatter(legacyTasksWithoutExecutableFields(false), "tasks", "0001-sample"))
	writeFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), validProgressSummary("0001", "sample"))

	feat := &feature.Feature{
		Number:  1,
		Slug:    "sample",
		DirName: "0001-sample",
		Path:    featurePath,
		Phase:   feature.PhaseImplement,
	}

	report, err := buildReconcileReport(projectRoot, cfg, feat)
	if err != nil {
		t.Fatalf("buildReconcileReport() error = %v", err)
	}

	var found bool
	for _, finding := range report.Findings {
		if strings.Contains(finding.Issue, "active feature tasks do not declare executable verification fields") {
			found = true
			if finding.Severity != reconcileSeverityWarning {
				t.Fatalf("Severity = %q, want warning", finding.Severity)
			}
			if !strings.Contains(finding.UpdateInstruction, "propose runnable checks separately from confirmed checks") {
				t.Fatalf("expected prose-only acceptance guidance, got %q", finding.UpdateInstruction)
			}
		}
	}
	if !found {
		t.Fatalf("expected executable verification advisory, got %q", findingsIssues(report.Findings))
	}
}

func TestBuildReconcileReportSkipsCompletedLegacyVerificationFields(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-sample")
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), withFeatureFrontMatter(validSpecWithRelationships("none\n"), "spec", "0001-sample"))
	writeFile(t, filepath.Join(featurePath, "PLAN.md"), withFeatureFrontMatter(validPlan(), "plan", "0001-sample"))
	writeFile(t, filepath.Join(featurePath, "TASKS.md"), withFeatureFrontMatter(legacyTasksWithoutExecutableFields(true), "tasks", "0001-sample"))
	writeFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), validProgressSummary("0001", "sample"))

	feat := &feature.Feature{
		Number:  1,
		Slug:    "sample",
		DirName: "0001-sample",
		Path:    featurePath,
		Phase:   feature.PhaseComplete,
	}

	report, err := buildReconcileReport(projectRoot, cfg, feat)
	if err != nil {
		t.Fatalf("buildReconcileReport() error = %v", err)
	}

	if strings.Contains(findingsIssues(report.Findings), "active feature tasks do not declare executable verification fields") {
		t.Fatalf("expected completed legacy feature to skip advisory, got %q", findingsIssues(report.Findings))
	}
}

func TestBuildReconcileReportSkipsHistoricalFeatureScopedVerificationFields(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	historicalPath := filepath.Join(projectRoot, "docs", "specs", "0001-historical")
	writeFile(t, filepath.Join(historicalPath, "SPEC.md"), withFeatureFrontMatter(validSpecWithRelationships("none\n"), "spec", "0001-historical"))
	writeFile(t, filepath.Join(historicalPath, "PLAN.md"), withFeatureFrontMatter(validPlan(), "plan", "0001-historical"))
	writeFile(t, filepath.Join(historicalPath, "TASKS.md"), withFeatureFrontMatter(legacyTasksWithoutExecutableFields(false), "tasks", "0001-historical"))
	activePath := filepath.Join(projectRoot, "docs", "specs", "0002-active")
	writeFile(t, filepath.Join(activePath, "SPEC.md"), withFeatureFrontMatter(validSpecWithRelationships("none\n"), "spec", "0002-active"))
	writeFile(t, filepath.Join(activePath, "PLAN.md"), withFeatureFrontMatter(validPlan(), "plan", "0002-active"))
	writeFile(t, filepath.Join(activePath, "TASKS.md"), withFeatureFrontMatter(legacyTasksWithoutExecutableFields(false), "tasks", "0002-active"))
	writeFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), validProgressSummaryForFeatures(
		[]string{"0001-historical", "0002-active"},
	))

	feat := &feature.Feature{
		Number:  1,
		Slug:    "historical",
		DirName: "0001-historical",
		Path:    historicalPath,
		Phase:   feature.PhaseReflect,
	}

	report, err := buildReconcileReport(projectRoot, cfg, feat)
	if err != nil {
		t.Fatalf("buildReconcileReport() error = %v", err)
	}

	if strings.Contains(findingsIssues(report.Findings), "active feature tasks do not declare executable verification fields") {
		t.Fatalf("expected historical feature-scoped reconcile to skip advisory, got %q", findingsIssues(report.Findings))
	}
}

func TestBuildReconcileReportProjectScopeWarnsOnlyActiveFeature(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	writeFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), validConstitution())
	writeFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), validProgressSummaryForFeatures(
		[]string{"0001-historical", "0002-active"},
	))
	writeFile(t, filepath.Join(projectRoot, "AGENTS.md"), templates.AgentsMD)
	writeFile(t, filepath.Join(projectRoot, "CLAUDE.md"), templates.ClaudeMD)
	writeFile(t, filepath.Join(projectRoot, ".github", "copilot-instructions.md"), templates.CopilotInstructionsMD)
	writeInitScaffoldArtifacts(t, projectRoot)
	for _, support := range templates.InstructionSupportFiles(config.InstructionScaffoldVersionTOC) {
		writeFile(t, filepath.Join(projectRoot, support.RelativePath), support.Content)
	}

	historicalPath := filepath.Join(projectRoot, "docs", "specs", "0001-historical")
	writeFile(t, filepath.Join(historicalPath, "SPEC.md"), withFeatureFrontMatter(validSpecWithRelationships("none\n"), "spec", "0001-historical"))
	writeFile(t, filepath.Join(historicalPath, "PLAN.md"), withFeatureFrontMatter(validPlan(), "plan", "0001-historical"))
	writeFile(t, filepath.Join(historicalPath, "TASKS.md"), withFeatureFrontMatter(legacyTasksWithoutExecutableFields(true), "tasks", "0001-historical"))
	activePath := filepath.Join(projectRoot, "docs", "specs", "0002-active")
	writeFile(t, filepath.Join(activePath, "SPEC.md"), withFeatureFrontMatter(validSpecWithRelationships("none\n"), "spec", "0002-active"))
	writeFile(t, filepath.Join(activePath, "PLAN.md"), withFeatureFrontMatter(validPlan(), "plan", "0002-active"))
	writeFile(t, filepath.Join(activePath, "TASKS.md"), withFeatureFrontMatter(legacyTasksWithoutExecutableFields(false), "tasks", "0002-active"))

	report, err := buildReconcileReport(projectRoot, cfg, nil)
	if err != nil {
		t.Fatalf("buildReconcileReport() error = %v", err)
	}

	issues := findingsIssues(report.Findings)
	if !strings.Contains(issues, "active feature tasks do not declare executable verification fields") {
		t.Fatalf("expected active verification advisory, got %q", issues)
	}
	if strings.Contains(issues, "0001-historical") {
		t.Fatalf("expected historical feature to be left alone, got %q", issues)
	}
}
