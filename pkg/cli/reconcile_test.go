package cli

import (
	"bytes"
	"io"
	"os"
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
		"feature sample",
		"Only update Kit-managed docs and scaffold files; do not modify product code, tests, runtime config, generated artifacts, or implementation files.",
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
		"also refresh `PROJECT_PROGRESS_SUMMARY.md`",
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
	if strings.Contains(prompt, "/plan") {
		t.Fatalf("expected reconcile prompt to avoid native plan-mode triggers, got %q", prompt)
	}
}

func TestReadReconcileMenuUsesReadableDefaults(t *testing.T) {
	var out bytes.Buffer

	choice, err := readReconcileMenu(strings.NewReader("\n\n\n"), &out)
	if err != nil {
		t.Fatalf("readReconcileMenu() error = %v", err)
	}
	if !choice.IncludeFiles {
		t.Fatalf("IncludeFiles = false, want true")
	}
	if choice.Force {
		t.Fatalf("Force = true, want false")
	}
	if !choice.OutputPrompt {
		t.Fatalf("OutputPrompt = false, want true")
	}
	for _, check := range []string{"Reconcile Options", "include files?", "force these changes?", "output coding-agent prompt too?"} {
		if !strings.Contains(out.String(), check) {
			t.Fatalf("expected menu output to contain %q, got %q", check, out.String())
		}
	}
}

func TestReadReconcileMenuCanSkipFilesAndPrompt(t *testing.T) {
	choice, err := readReconcileMenu(strings.NewReader("n\nn\n"), io.Discard)
	if err != nil {
		t.Fatalf("readReconcileMenu() error = %v", err)
	}
	if choice.IncludeFiles {
		t.Fatalf("IncludeFiles = true, want false")
	}
	if choice.Force {
		t.Fatalf("Force = true, want false")
	}
	if choice.OutputPrompt {
		t.Fatalf("OutputPrompt = true, want false")
	}
}

func TestReadReconcileMenuRejectsInvalidAnswer(t *testing.T) {
	if got, err := readReconcileMenu(strings.NewReader("maybe\n"), io.Discard); err == nil {
		t.Fatalf("readReconcileMenu() = %#v, want error", got)
	}
}

func TestBuildReconcilePromptIncludesReferenceMigrationRules(t *testing.T) {
	report := &reconcileReport{
		ProjectRoot:        t.TempDir(),
		ReferenceMigration: true,
	}

	prompt := buildReconcilePrompt(report)
	checks := []string{
		"Migration target: replace deprecated front matter `dependencies` with canonical graph-like `references`",
		"reference migration: enabled",
		"map `location` to `target`",
		"add a graph `relation`",
		"add `read_policy`",
		"prefer `read_policy: must` for constraints",
	}
	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected migration prompt to contain %q, got %q", check, prompt)
		}
	}
}

func TestBuildReconcilePromptIncludesVerificationMigrationRules(t *testing.T) {
	report := &reconcileReport{
		ProjectRoot:           t.TempDir(),
		VerificationMigration: true,
		Findings: []reconcileFinding{
			{
				Severity:          reconcileSeverityWarning,
				FilePath:          "/tmp/TASKS.md",
				Issue:             "active feature tasks do not declare executable verification fields: T001 missing VERIFY",
				UpdateInstruction: "add executable verification fields where commands are known",
			},
		},
	}

	prompt := buildReconcilePrompt(report)
	checks := []string{
		"Migration target: add executable verification fields to active task details",
		"verification migration: enabled",
		"verification migration is advisory",
		"do not mark legacy docs invalid",
		"do not guess verification commands from prose",
		"leave uncertain commands as `not yet declared`",
		"run `kit legacy verify <feature> --dry-run`, refresh `.kit/state.json`, then rerun `kit check <feature>` and `kit check --project`",
	}
	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected verification migration prompt to contain %q, got %q", check, prompt)
		}
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

func TestBuildReconcilePrompt_UsesVersionedInstructionShortcut(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	report := &reconcileReport{
		ProjectRoot: projectRoot,
		Findings: []reconcileFinding{
			{
				Severity: reconcileSeverityError,
				FilePath: filepath.Join(projectRoot, "AGENTS.md"),
				Issue:    "missing Kit-managed repository instruction file",
			},
		},
	}

	prompt := buildReconcilePrompt(report)
	if !strings.Contains(prompt, "`kit scaffold agents --version 2 --append-only`") {
		t.Fatalf("expected prompt to use versioned scaffold shortcut, got %q", prompt)
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

func TestBuildReconcileReportFeatureScopeFindsFrontMatterIdentityDrift(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-sample")
	content := strings.Replace(
		withFeatureFrontMatter(validSpecWithRelationships("none\n"), "spec", "0001-sample"),
		`  id: "0001"
  slug: sample`,
		`  id: "0002"
  slug: other`,
		1,
	)
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), content)
	writeFile(t, filepath.Join(featurePath, "PLAN.md"), withFeatureFrontMatter(validPlan(), "plan", "0001-sample"))
	writeFile(t, filepath.Join(featurePath, "TASKS.md"), withFeatureFrontMatter(validTasks(), "tasks", "0001-sample"))
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
	for _, check := range []string{
		"front matter feature.id `0002` does not match containing feature directory id `0001`",
		"front matter feature.slug `other` does not match containing feature directory slug `sample`",
	} {
		if !strings.Contains(issues, check) {
			t.Fatalf("expected identity drift finding %q, got %q", check, issues)
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
