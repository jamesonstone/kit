package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/feature"
)

func TestBuildReconcilePromptIncludesScopeRulesAndVerification(t *testing.T) {
	projectRoot := t.TempDir()
	report := &reconcileReport{
		ProjectRoot: projectRoot,
		Feature: &feature.Feature{
			Slug:    "sample",
			DirName: "0001-sample",
			Path:    filepath.Join(projectRoot, "docs", "specs", "0001-sample"),
		},
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
		"contract order:",
		"Audit snapshot:",
		"Files to fix:",
		"| Severity | File | Issues | Focus |",
		"align task IDs",
		"Notable issues:",
		"`TASKS.md`: task `T001` exists in `PROGRESS TABLE` but not in `TASK DETAILS`",
		"Search shortcuts:",
		"Delivery of command-created files:",
		"apply only the listed manual findings until a fresh snapshot exists",
		"require the rerun to emit a new exact command-owned snapshot",
		"if it cannot, do not adopt managed-file changes and report the blocker",
		"explicitly stage only those paths",
		"match the snapshot exactly",
		"create or update the ready pull request",
		"disturb unrelated root-checkout or worktree changes",
		"`kit check sample`",
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
		"after migration, rerun `kit check <feature>` and `kit check --project`; retain executable verification details only as historical evidence",
	}
	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected verification migration prompt to contain %q, got %q", check, prompt)
		}
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
	cfg.InstructionScaffoldVersion = config.CurrentInstructionScaffoldVersion
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
	if !strings.Contains(prompt, "`kit reconcile --include-files --dry-run --diff`") {
		t.Fatalf("expected prompt to use the supported reconcile shortcut, got %q", prompt)
	}
}
