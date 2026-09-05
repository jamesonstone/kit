package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
	"github.com/spf13/cobra"
)

func TestRunCheckRejectsProjectWithFeatureArg(t *testing.T) {
	checkProject = true
	checkAll = false
	t.Cleanup(func() {
		checkProject = false
		checkAll = false
	})

	cmd := &cobra.Command{}
	err := runCheck(cmd, []string{"sample"})
	if err == nil || !strings.Contains(err.Error(), "--project cannot be used with a feature argument") {
		t.Fatalf("expected --project validation error, got %v", err)
	}
}

func TestRunCheckProjectFailsOnRepoDrift(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	writeFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), validConstitution())
	writeFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), validProgressSummary("", ""))
	setWorkingDirectory(t, projectRoot)

	checkProject = true
	checkAll = false
	t.Cleanup(func() {
		checkProject = false
		checkAll = false
	})

	cmd := &cobra.Command{}
	err := runCheck(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "project validation failed") {
		t.Fatalf("expected project validation failure, got %v", err)
	}
}

func TestRunCheckProjectPassesWhenRepoIsCoherent(t *testing.T) {
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
	writeStandingAuthorityPolicies(t, projectRoot)
	setWorkingDirectory(t, projectRoot)

	checkProject = true
	checkAll = false
	t.Cleanup(func() {
		checkProject = false
		checkAll = false
	})

	cmd := &cobra.Command{}
	if err := runCheck(cmd, nil); err != nil {
		t.Fatalf("runCheck() error = %v", err)
	}
}

func TestRunCheckProjectPassesWithV3InstructionContract(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	writeFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), validConstitution())
	writeFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), validProgressSummary("", ""))
	writeInitScaffoldArtifacts(t, projectRoot)
	for _, relativePath := range instructionArtifactPaths(cfg, instructionFileSelection{}, config.InstructionScaffoldVersionMemory, true) {
		content, _, err := instructionArtifactContent(relativePath, config.InstructionScaffoldVersionMemory)
		if err != nil {
			t.Fatalf("instructionArtifactContent(%q) error = %v", relativePath, err)
		}
		writeFile(t, filepath.Join(projectRoot, relativePath), content)
	}
	writeStandingAuthorityPolicies(t, projectRoot)
	setWorkingDirectory(t, projectRoot)

	checkProject = true
	checkAll = false
	t.Cleanup(func() {
		checkProject = false
		checkAll = false
	})

	cmd := &cobra.Command{}
	if err := runCheck(cmd, nil); err != nil {
		t.Fatalf("runCheck() error = %v", err)
	}
}

func TestRunCheckProjectDoesNotFailForCustomizedV2MigrationAdvisory(t *testing.T) {
	projectRoot := setupCoherentProjectForCheck(t)
	agentsPath := filepath.Join(projectRoot, "AGENTS.md")
	writeFile(t, agentsPath, readFile(t, agentsPath)+"\n## Local Policy\n\nPreserve this customization.\n")
	setWorkingDirectory(t, projectRoot)

	checkProject = true
	checkAll = false
	t.Cleanup(func() {
		checkProject = false
		checkAll = false
	})

	cmd := &cobra.Command{}
	if err := runCheck(cmd, nil); err != nil {
		t.Fatalf("runCheck() error = %v", err)
	}
}

func TestBuildReconcileReportAcceptsBootstrapConstitution(t *testing.T) {
	projectRoot := setupCoherentProjectForCheck(t)
	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	writeFile(t, filepath.Join(projectRoot, cfg.ConstitutionPath), templates.Constitution)

	report, err := buildReconcileReport(projectRoot, cfg, nil)
	if err != nil {
		t.Fatalf("buildReconcileReport() error = %v", err)
	}
	for _, finding := range report.Findings {
		if filepath.Base(finding.FilePath) == "CONSTITUTION.md" {
			t.Fatalf("expected generated starter Constitution to be valid bootstrap state, got %#v", finding)
		}
	}
}

func TestBuildReconcileReportRejectsPartiallyCustomizedEmptyConstitution(t *testing.T) {
	projectRoot := setupCoherentProjectForCheck(t)
	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	writeFile(t, filepath.Join(projectRoot, cfg.ConstitutionPath), `# CONSTITUTION

## PRINCIPLES

## CONSTRAINTS

Project constraints are defined.

## NON-GOALS

No test non-goals.

## DEFINITIONS

Test definition.
`)

	report, err := buildReconcileReport(projectRoot, cfg, nil)
	if err != nil {
		t.Fatalf("buildReconcileReport() error = %v", err)
	}
	if issues := findingsIssues(report.Findings); !strings.Contains(issues, "required section `## PRINCIPLES` is empty or placeholder-only") {
		t.Fatalf("expected partially customized Constitution to remain actionable, got:\n%s", issues)
	}
}

func TestRunCheckProjectFailsWhenV2RootIsVerboseManual(t *testing.T) {
	projectRoot := setupCoherentProjectForCheck(t)
	writeFile(t, filepath.Join(projectRoot, "AGENTS.md"), templates.LegacyAgentsMD)
	setWorkingDirectory(t, projectRoot)

	checkProject = true
	checkAll = false
	t.Cleanup(func() {
		checkProject = false
		checkAll = false
	})

	cmd := &cobra.Command{}
	err := runCheck(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "project validation failed") {
		t.Fatalf("expected project validation failure, got %v", err)
	}
}

func TestRunCheckProjectFailsWhenRLMGuidanceDrifts(t *testing.T) {
	projectRoot := setupCoherentProjectForCheck(t)
	writeFile(t, filepath.Join(projectRoot, "docs", "agents", "RLM.md"), "# RLM\n\n## Purpose\n\n- stale guidance\n")
	setWorkingDirectory(t, projectRoot)

	checkProject = true
	checkAll = false
	t.Cleanup(func() {
		checkProject = false
		checkAll = false
	})

	cmd := &cobra.Command{}
	err := runCheck(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "project validation failed") {
		t.Fatalf("expected project validation failure, got %v", err)
	}
}

func TestRunCheckProjectFailsWhenRootRequiresVendorTool(t *testing.T) {
	projectRoot := setupCoherentProjectForCheck(t)
	writeFile(t, filepath.Join(projectRoot, "AGENTS.md"), templates.AgentsMD+"\n- must use Codex for every change\n")
	setWorkingDirectory(t, projectRoot)

	checkProject = true
	checkAll = false
	t.Cleanup(func() {
		checkProject = false
		checkAll = false
	})

	cmd := &cobra.Command{}
	err := runCheck(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "project validation failed") {
		t.Fatalf("expected project validation failure, got %v", err)
	}
}
