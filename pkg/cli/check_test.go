package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/templates"
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

func TestBuildReconcileReportReportsLegacyMissingFrontMatter(t *testing.T) {
	projectRoot := setupCoherentProjectForCheck(t)
	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	writeFile(t, filepath.Join(projectRoot, "docs", "specs", "0001-alpha", "SPEC.md"), validSpecWithRelationships("none\n"))

	report, err := buildReconcileReport(projectRoot, cfg, nil)
	if err != nil {
		t.Fatalf("buildReconcileReport() error = %v", err)
	}
	issues := findingsIssues(report.Findings)
	if !strings.Contains(issues, "feature artifact is missing canonical YAML front matter") {
		t.Fatalf("expected legacy front matter migration warning, got:\n%s", issues)
	}
}

func TestCheckFeatureFailsOnMalformedPresentFrontMatter(t *testing.T) {
	projectRoot := t.TempDir()
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	featurePath := filepath.Join(specsDir, "0001-alpha")
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), `---
kit_metadata_version: 1
artifact: spec
feature:
  id: "0001"
  slug: alpha
  dir: 0001-alpha
# missing closing delimiter
# SPEC

## SUMMARY

summary
`)

	err := checkFeature(projectRoot, specsDir, "alpha")
	if err == nil || !strings.Contains(err.Error(), "validation failed") {
		t.Fatalf("expected malformed front matter validation failure, got %v", err)
	}
}

func TestCheckFeatureAllowsLegacyDocsWithoutFrontMatter(t *testing.T) {
	projectRoot := t.TempDir()
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	featurePath := filepath.Join(specsDir, "0001-alpha")
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), validSpecWithRelationships("none\n"))

	if err := checkFeature(projectRoot, specsDir, "alpha"); err != nil {
		t.Fatalf("expected legacy missing front matter to be tolerated, got %v", err)
	}
}

func TestCheckFeatureFailsWhenFrontMatterIdentityDriftsFromDirectory(t *testing.T) {
	projectRoot := t.TempDir()
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	featurePath := filepath.Join(specsDir, "0001-alpha")
	content := strings.Replace(
		withFeatureFrontMatter(validSpecWithRelationships("none\n"), "spec", "0001-alpha"),
		`  id: "0001"
  slug: alpha`,
		`  id: "0002"
  slug: beta`,
		1,
	)
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), content)

	err := checkFeature(projectRoot, specsDir, "alpha")
	if err == nil || !strings.Contains(err.Error(), "validation failed") {
		t.Fatalf("expected feature identity validation failure, got %v", err)
	}
}

func TestRunCheckProjectFailsOnDuplicateFeatureNumbers(t *testing.T) {
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
	writeFile(t, filepath.Join(projectRoot, "docs", "specs", "0012-alpha", "SPEC.md"), "# SPEC\n\n## RELATIONSHIPS\n\nnone\n")
	writeFile(t, filepath.Join(projectRoot, "docs", "specs", "0012-beta", "SPEC.md"), "# SPEC\n\n## RELATIONSHIPS\n\nnone\n")
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
		t.Fatalf("expected duplicate-number project validation failure, got %v", err)
	}
}

func setupCoherentProjectForCheck(t *testing.T) string {
	t.Helper()

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

	return projectRoot
}

func writeInitScaffoldArtifacts(t *testing.T, projectRoot string) {
	t.Helper()

	writeFile(t, filepath.Join(projectRoot, gitignorePath), templates.Gitignore)
	writeFile(t, filepath.Join(projectRoot, envPath), "")
	writeFile(t, filepath.Join(projectRoot, envrcPath), templates.Envrc)
	writeFile(t, filepath.Join(projectRoot, codeRabbitConfigPath), templates.CodeRabbitConfig)
	writeFile(t, filepath.Join(projectRoot, pullRequestTemplatePath), templates.PullRequestTemplate)
}
