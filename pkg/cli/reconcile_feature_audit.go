package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

func auditProjectProgressSummary(projectRoot string, features []feature.Feature) []reconcileFinding {
	path := filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md")
	if !document.Exists(path) {
		return []reconcileFinding{newFinding(
			reconcileSeverityError,
			path,
			"missing `PROJECT_PROGRESS_SUMMARY.md`",
			templateSource(projectRoot),
			"create the progress summary from the template and current feature docs",
			[]string{
				fmt.Sprintf("sed -n '1,220p' %s", templateSource(projectRoot)),
				fmt.Sprintf("ls %s", filepath.Join(projectRoot, "docs", "specs")),
			},
		)}
	}

	findings := auditStructuredDocument(path, document.TypeProgressSummary, projectRoot, nil)
	content, err := os.ReadFile(path)
	if err != nil {
		return append(findings, newFinding(
			reconcileSeverityError,
			path,
			"failed to read `PROJECT_PROGRESS_SUMMARY.md`",
			templateSource(projectRoot),
			"fix file readability before reconciliation can continue",
			[]string{fmt.Sprintf("sed -n '1,240p' %s", path)},
		))
	}

	body := string(content)
	for i := range features {
		findings = append(findings, auditFeatureRollupCoverageFromContent(projectRoot, body, &features[i])...)
	}
	return findings
}

func auditFeatureRollupCoverage(projectRoot string, feat *feature.Feature) []reconcileFinding {
	summaryPath := filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md")
	content, err := os.ReadFile(summaryPath)
	if err != nil {
		return nil
	}
	return auditFeatureRollupCoverageFromContent(projectRoot, string(content), feat)
}

func auditDuplicateFeatureNumbers(specsPath, projectRoot string, features []feature.Feature) []reconcileFinding {
	duplicates := feature.DuplicateNumberGroups(features)
	if len(duplicates) == 0 {
		return nil
	}

	var findings []reconcileFinding
	for number, group := range duplicates {
		names := make([]string, 0, len(group))
		for _, feat := range group {
			names = append(names, feat.DirName)
		}
		findings = append(findings, newFinding(
			reconcileSeverityError,
			specsPath,
			fmt.Sprintf("feature number `%04d` is duplicated by %s", number, strings.Join(names, ", ")),
			initProjectSource(projectRoot),
			"renumber or merge the conflicting feature directories so each numeric prefix is unique across `docs/specs/`",
			[]string{
				fmt.Sprintf("ls %s", specsPath),
				fmt.Sprintf("rg -n \"^# (BRAINSTORM|SPEC|PLAN|TASKS)\" %s", specsPath),
			},
		))
	}

	return findings
}

func auditFeatureRollupCoverageFromContent(projectRoot, content string, feat *feature.Feature) []reconcileFinding {
	summaryPath := filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md")
	rowSnippet := fmt.Sprintf("| %04d | %s |", feat.Number, feat.Slug)
	headingSnippet := fmt.Sprintf("### %s\n", feat.Slug)
	var findings []reconcileFinding

	if !strings.Contains(content, rowSnippet) {
		findings = append(findings, newFinding(
			reconcileSeverityWarning,
			summaryPath,
			fmt.Sprintf("progress summary is missing the feature-table row for `%s`", feat.DirName),
			templateSource(projectRoot),
			"refresh `PROJECT_PROGRESS_SUMMARY.md` after reconciling feature docs",
			[]string{
				fmt.Sprintf("rg -n \"^\\| %04d \\| %s \\|\" %s", feat.Number, feat.Slug, summaryPath),
				fmt.Sprintf("ls %s", filepath.Join(projectRoot, "docs", "specs")),
			},
		))
	}

	if !strings.Contains(content, headingSnippet) {
		findings = append(findings, newFinding(
			reconcileSeverityWarning,
			summaryPath,
			fmt.Sprintf("progress summary is missing the feature summary heading for `%s`", feat.DirName),
			templateSource(projectRoot),
			"refresh `PROJECT_PROGRESS_SUMMARY.md` after reconciliation so every current feature has a summary section",
			[]string{
				fmt.Sprintf("rg -n \"^### %s$\" %s", feat.Slug, summaryPath),
				fmt.Sprintf("ls %s", filepath.Join(projectRoot, "docs", "specs")),
			},
		))
	}

	return findings
}

func auditFeatureDocuments(projectRoot string, feat *feature.Feature, relationshipTargets map[string]bool) []reconcileFinding {
	paths := map[string]string{
		"brainstorm": filepath.Join(feat.Path, "BRAINSTORM.md"),
		"spec":       filepath.Join(feat.Path, "SPEC.md"),
		"plan":       filepath.Join(feat.Path, "PLAN.md"),
		"tasks":      filepath.Join(feat.Path, "TASKS.md"),
	}

	var findings []reconcileFinding
	specExists := document.Exists(paths["spec"])
	planExists := document.Exists(paths["plan"])
	tasksExists := document.Exists(paths["tasks"])

	if !specExists && (planExists || tasksExists) {
		findings = append(findings, newFinding(
			reconcileSeverityError,
			paths["spec"],
			"missing `SPEC.md` even though later-phase feature artifacts exist",
			templateSource(projectRoot),
			"create `SPEC.md` and backfill the current feature contract before keeping later artifacts",
			genericFeatureSearchHints(projectRoot, feat, paths["spec"], "SPEC"),
		))
	}
	if !planExists && tasksExists {
		findings = append(findings, newFinding(
			reconcileSeverityError,
			paths["plan"],
			"missing `PLAN.md` even though `TASKS.md` exists",
			templateSource(projectRoot),
			"create `PLAN.md` and restore the implementation approach before keeping the task list",
			genericFeatureSearchHints(projectRoot, feat, paths["plan"], "PLAN"),
		))
	}

	if document.Exists(paths["brainstorm"]) {
		findings = append(findings, auditStructuredDocument(paths["brainstorm"], document.TypeBrainstorm, projectRoot, relationshipTargets)...)
	}
	if specExists {
		findings = append(findings, auditStructuredDocument(paths["spec"], document.TypeSpec, projectRoot, relationshipTargets)...)
	}
	if planExists {
		findings = append(findings, auditStructuredDocument(paths["plan"], document.TypePlan, projectRoot, relationshipTargets)...)
	}
	if tasksExists {
		findings = append(findings, auditStructuredDocument(paths["tasks"], document.TypeTasks, projectRoot, relationshipTargets)...)
		findings = append(findings, auditTaskAlignment(paths["tasks"], projectRoot)...)
	}

	return findings
}
