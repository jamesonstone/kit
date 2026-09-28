package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/document"
	"github.com/jamesonstone/kit/v3/internal/feature"
)

type reconcileSeverity string

const (
	reconcileSeverityError   reconcileSeverity = "error"
	reconcileSeverityWarning reconcileSeverity = "warning"
)

type reconcileFinding struct {
	Severity          reconcileSeverity
	NonBlocking       bool
	AllowsCodeChanges bool
	FilePath          string
	Issue             string
	UpdateInstruction string
}

type reconcileReport struct {
	ProjectRoot      string
	Feature          *feature.Feature
	Findings         []reconcileFinding
	DeliverySnapshot []managedFileDeliverySnapshot
	SourceFileAudit  *sourceFileAuditSummary
}

func (r *reconcileReport) cleanResult() string {
	if r.Feature != nil {
		return fmt.Sprintf("No reconciliation needed for feature %s.", r.Feature.Slug)
	}
	result := "No reconciliation needed. Kit-managed project state already matches the current contract for this scope."
	if evidence := sourceFileAuditEvidence(r.SourceFileAudit); evidence != "" {
		result += " " + evidence + "."
	}
	return result
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
		sourceAudit := inspectSourceFileSizes(projectRoot, cfg.SourceFileLineLimit)
		report.SourceFileAudit = &sourceAudit.Summary
		report.Findings = append(report.Findings, sourceAudit.Findings...)
		report.Findings = append(report.Findings, auditDuplicateFeatureNumbers(cfg.SpecsPath(projectRoot), projectRoot, features)...)
		report.Findings = append(report.Findings, auditInitScaffoldArtifacts(projectRoot)...)
		report.Findings = append(report.Findings, auditConstitution(projectRoot)...)
		report.Findings = append(report.Findings, auditRulesets(projectRoot)...)
		for i := range features {
			report.Findings = append(report.Findings, auditFeatureDocuments(projectRoot, &features[i], targets)...)
		}
		if activeVerificationFeature != nil {
			report.Findings = append(report.Findings, auditExecutableVerificationAdvisory(projectRoot, activeVerificationFeature)...)
		}
		report.Findings = append(report.Findings, auditInstructionFiles(projectRoot, cfg)...)
	} else {
		report.Findings = append(report.Findings, auditFeatureDocuments(projectRoot, feat, targets)...)
		if activeVerificationFeature != nil && activeVerificationFeature.DirName == feat.DirName {
			report.Findings = append(report.Findings, auditExecutableVerificationAdvisory(projectRoot, activeVerificationFeature)...)
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
			"create `docs/CONSTITUTION.md` and populate the current Kit sections before reconciling feature docs",
		)}
	}
	if content, err := os.ReadFile(path); err == nil && isBootstrapConstitution(string(content)) {
		return nil
	}

	return auditStructuredDocument(path, document.TypeConstitution, projectRoot, nil)
}

func auditInitScaffoldArtifacts(projectRoot string) []reconcileFinding {
	findings := auditGitignoreScaffold(projectRoot)
	// The local-only .env and .envrc are per-checkout state that clean clones
	// never have; their absence needs no action, so they are not audited.
	for _, artifact := range []struct {
		relativePath string
		description  string
	}{
		{relativePath: codeRabbitConfigPath, description: "CodeRabbit review configuration"},
		{relativePath: pullRequestTemplatePath, description: "GitHub pull request template"},
		{relativePath: autoAssignWorkflowPath, description: "GitHub issue and pull request auto-assignment workflow"},
	} {
		absolutePath := filepath.Join(projectRoot, filepath.FromSlash(artifact.relativePath))
		if document.Exists(absolutePath) {
			continue
		}
		findings = append(findings, newFinding(
			reconcileSeverityWarning,
			absolutePath,
			fmt.Sprintf("missing Kit init scaffold artifact `%s`", artifact.relativePath),
			fmt.Sprintf("run `kit reconcile` to create the missing %s, then review it before committing", artifact.description),
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
			"run `kit reconcile` to create `.gitignore` with the current Kit-local environment, cache, and scratch artifact entries",
		)}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return []reconcileFinding{newFinding(
			reconcileSeverityWarning,
			path,
			"failed to read `.gitignore` for Kit-managed init scaffold entries",
			"fix `.gitignore` readability, then run `kit reconcile` to append any missing Kit-managed entries",
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
		"run `kit reconcile` to append the missing ignore entries while preserving existing project-specific ignores",
	)}
}

func quotedGitignorePatterns(patterns []string) []string {
	quoted := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		quoted = append(quoted, fmt.Sprintf("`%s`", pattern))
	}
	return quoted
}
