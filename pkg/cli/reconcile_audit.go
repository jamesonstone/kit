package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

type reconcileSeverity string

const (
	reconcileSeverityError   reconcileSeverity = "error"
	reconcileSeverityWarning reconcileSeverity = "warning"
)

type reconcileFinding struct {
	Severity          reconcileSeverity
	FilePath          string
	Issue             string
	ContractSource    string
	UpdateInstruction string
	SearchHints       []string
}

type reconcileReport struct {
	ProjectRoot           string
	Feature               *feature.Feature
	Findings              []reconcileFinding
	NeedsRollup           bool
	ReferenceMigration    bool
	VerificationMigration bool
}

func (r *reconcileReport) cleanResult() string {
	if r.Feature != nil {
		return fmt.Sprintf("No reconciliation needed for feature %s.", r.Feature.Slug)
	}
	return "No reconciliation needed. Kit-managed documents and scaffold artifacts already match the current contract for this scope."
}

func buildReconcileReport(projectRoot string, cfg *config.Config, feat *feature.Feature) (*reconcileReport, error) {
	report := &reconcileReport{
		ProjectRoot: projectRoot,
		Feature:     feat,
	}

	features, err := feature.ListFeaturesWithState(cfg.SpecsPath(projectRoot), cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to list features: %w", err)
	}
	activeVerificationFeature := activeFeatureForVerificationAdvisory(features)

	targets := make(map[string]bool, len(features))
	for _, item := range features {
		targets[item.DirName] = true
	}

	if feat == nil {
		report.Findings = append(report.Findings, auditDuplicateFeatureNumbers(cfg.SpecsPath(projectRoot), projectRoot, features)...)
		report.Findings = append(report.Findings, auditInitScaffoldArtifacts(projectRoot)...)
		report.Findings = append(report.Findings, auditConstitution(projectRoot)...)
		report.Findings = append(report.Findings, auditRulesets(projectRoot)...)
		report.Findings = append(report.Findings, auditProjectProgressSummary(projectRoot, features)...)
		for i := range features {
			report.Findings = append(report.Findings, auditFeatureDocuments(projectRoot, &features[i], targets)...)
		}
		if activeVerificationFeature != nil {
			report.Findings = append(report.Findings, auditExecutableVerificationAdvisory(projectRoot, activeVerificationFeature)...)
		}
		report.Findings = append(report.Findings, auditInstructionFiles(projectRoot, cfg)...)
	} else {
		report.Findings = append(report.Findings, auditFeatureDocuments(projectRoot, feat, targets)...)
		report.Findings = append(report.Findings, auditFeatureRollupCoverage(projectRoot, feat)...)
		if activeVerificationFeature != nil && activeVerificationFeature.DirName == feat.DirName {
			report.Findings = append(report.Findings, auditExecutableVerificationAdvisory(projectRoot, activeVerificationFeature)...)
		}
	}

	for _, finding := range report.Findings {
		if filepath.Base(finding.FilePath) == "PROJECT_PROGRESS_SUMMARY.md" {
			report.NeedsRollup = true
			break
		}
	}

	sort.SliceStable(report.Findings, func(i, j int) bool {
		if report.Findings[i].Severity != report.Findings[j].Severity {
			return report.Findings[i].Severity < report.Findings[j].Severity
		}
		if report.Findings[i].FilePath != report.Findings[j].FilePath {
			return report.Findings[i].FilePath < report.Findings[j].FilePath
		}
		return report.Findings[i].Issue < report.Findings[j].Issue
	})

	return report, nil
}

func auditConstitution(projectRoot string) []reconcileFinding {
	path := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	if !document.Exists(path) {
		return []reconcileFinding{newFinding(
			reconcileSeverityError,
			path,
			"missing Kit-managed root document `CONSTITUTION.md`",
			templateSource(projectRoot),
			"create `docs/CONSTITUTION.md` and populate the current Kit sections before reconciling feature docs",
			[]string{
				fmt.Sprintf("sed -n '1,240p' %s", templateSource(projectRoot)),
				fmt.Sprintf("sed -n '1,240p' %s", initProjectSource(projectRoot)),
			},
		)}
	}

	return auditStructuredDocument(path, document.TypeConstitution, projectRoot, nil)
}

func auditInitScaffoldArtifacts(projectRoot string) []reconcileFinding {
	var findings []reconcileFinding
	findings = append(findings, auditGitignoreScaffold(projectRoot)...)

	for _, artifact := range []struct {
		relativePath string
		description  string
		localOnly    bool
	}{
		{relativePath: envPath, description: "blank local environment file", localOnly: true},
		{relativePath: envrcPath, description: "local direnv bootstrap file", localOnly: true},
		{relativePath: codeRabbitConfigPath, description: "CodeRabbit review configuration"},
		{relativePath: pullRequestTemplatePath, description: "GitHub pull request template"},
		{relativePath: autoAssignWorkflowPath, description: "GitHub issue and pull request auto-assignment workflow"},
	} {
		absolutePath := filepath.Join(projectRoot, filepath.FromSlash(artifact.relativePath))
		if document.Exists(absolutePath) {
			continue
		}
		update := fmt.Sprintf("run `kit init` to create the missing %s, then review the generated file before committing it", artifact.description)
		if artifact.localOnly {
			update = fmt.Sprintf("run `kit init` to create the missing %s and keep it covered by `.gitignore`", artifact.description)
		}
		findings = append(findings, newFinding(
			reconcileSeverityWarning,
			absolutePath,
			fmt.Sprintf("missing Kit init scaffold artifact `%s`", artifact.relativePath),
			initProjectSource(projectRoot),
			update,
			[]string{
				"kit init",
				fmt.Sprintf("test -f %s", absolutePath),
			},
		))
	}

	return findings
}

func auditGitignoreScaffold(projectRoot string) []reconcileFinding {
	path := filepath.Join(projectRoot, gitignorePath)
	if !document.Exists(path) {
		return []reconcileFinding{newFinding(
			reconcileSeverityWarning,
			path,
			"missing `.gitignore` for Kit-managed init scaffold entries",
			initProjectSource(projectRoot),
			"run `kit init` to create `.gitignore` with the current Kit-local environment, cache, and scratch artifact entries",
			[]string{
				"kit init",
				fmt.Sprintf("sed -n '1,160p' %s", path),
			},
		)}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return []reconcileFinding{newFinding(
			reconcileSeverityWarning,
			path,
			"failed to read `.gitignore` for Kit-managed init scaffold entries",
			initProjectSource(projectRoot),
			"fix `.gitignore` readability, then run `kit init` to append any missing Kit-managed entries",
			[]string{fmt.Sprintf("sed -n '1,160p' %s", path)},
		)}
	}

	missing := missingGitignorePatterns(string(data))
	if len(missing) == 0 {
		return nil
	}

	return []reconcileFinding{newFinding(
		reconcileSeverityWarning,
		path,
		fmt.Sprintf("missing Kit-managed `.gitignore` entries: %s", strings.Join(quotedGitignorePatterns(missing), ", ")),
		initProjectSource(projectRoot),
		"run `kit init` to append the missing ignore entries while preserving existing project-specific ignores",
		[]string{
			"kit init",
			fmt.Sprintf("sed -n '1,160p' %s", path),
		},
	)}
}

func quotedGitignorePatterns(patterns []string) []string {
	quoted := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		quoted = append(quoted, fmt.Sprintf("`%s`", pattern))
	}
	return quoted
}
