package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/document"
	"github.com/jamesonstone/kit/v3/internal/feature"
)

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
