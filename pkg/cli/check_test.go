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
	cfg.InstructionScaffoldVersion = config.CurrentInstructionScaffoldVersion
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	writeFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), validConstitution())
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
	cfg.InstructionScaffoldVersion = config.CurrentInstructionScaffoldVersion
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	writeFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), validConstitution())
	writeInitScaffoldArtifacts(t, projectRoot)
	writeCurrentInstructionArtifacts(t, projectRoot)
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
	writeInitScaffoldArtifacts(t, projectRoot)
	writeCurrentInstructionArtifacts(t, projectRoot)
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
