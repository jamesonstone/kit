package cli

import (
	"fmt"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/jamesonstone/kit/internal/document"
)

var (
	reconcileCommentPattern    = regexp.MustCompile(`(?s)<!--.*?-->`)
	reconcileTaskListPattern   = regexp.MustCompile(`(?m)^\s*-\s*\[[ xX]\]\s*(T\d{3}):`)
	reconcileTaskDetailPattern = regexp.MustCompile(`(?m)^###\s*(T\d{3})\s*$`)
)

func auditStructuredDocument(path string, docType document.DocumentType, projectRoot string, relationshipTargets map[string]bool) []reconcileFinding {
	doc, err := document.ParseFile(path, docType)
	if err != nil {
		return []reconcileFinding{newFinding(
			reconcileSeverityError,
			path,
			fmt.Sprintf("failed to parse `%s`", filepath.Base(path)),
			templateSource(projectRoot),
			"fix the markdown structure before reconciliation continues",
			[]string{fmt.Sprintf("sed -n '1,260p' %s", path)},
		)}
	}

	var findings []reconcileFinding
	for _, section := range expectedSectionsFor(docType) {
		entry := doc.GetSection(section)
		if entry == nil {
			findings = append(findings, newFinding(
				reconcileSeverityError,
				path,
				fmt.Sprintf("missing required section `## %s`", section),
				contractSourceForSection(projectRoot, docType, section),
				fmt.Sprintf("add `## %s` and populate it with current repository-backed content", section),
				searchHintsForSection(projectRoot, path, section),
			))
			continue
		}
		if !meaningfulSectionContent(entry.Content) {
			findings = append(findings, newFinding(
				reconcileSeverityError,
				path,
				fmt.Sprintf("required section `## %s` is empty or placeholder-only", section),
				contractSourceForSection(projectRoot, docType, section),
				fmt.Sprintf("replace placeholder-only content in `## %s` with current repo-backed content", section),
				searchHintsForSection(projectRoot, path, section),
			))
		}
	}

	for _, expectation := range tableExpectationsFor(docType) {
		entry := doc.GetSection(expectation.Section)
		if entry == nil {
			continue
		}
		if issue := validateTableSection(entry.Content, expectation.Headers, expectation.RequireRows); issue != "" {
			findings = append(findings, newFinding(
				reconcileSeverityError,
				path,
				fmt.Sprintf("malformed `%s` table in `## %s`", filepath.Base(path), expectation.Section),
				contractSourceForSection(projectRoot, docType, expectation.Section),
				fmt.Sprintf("reshape `## %s` to match the current Kit table contract", expectation.Section),
				searchHintsForTable(projectRoot, path, expectation.Section),
			))
		}
	}

	if docType == document.TypeBrainstorm || docType == document.TypeSpec {
		findings = append(findings, auditRelationships(path, doc, projectRoot, relationshipTargets)...)
	}

	return findings
}

func auditRelationships(path string, doc *document.Document, projectRoot string, relationshipTargets map[string]bool) []reconcileFinding {
	section := doc.GetSection("RELATIONSHIPS")
	if section == nil || !meaningfulSectionContent(section.Content) {
		return nil
	}

	relationships, err := document.ParseRelationshipsSection(section)
	if err != nil {
		return []reconcileFinding{newFinding(
			reconcileSeverityError,
			path,
			fmt.Sprintf("invalid `RELATIONSHIPS` content: %v", err),
			constitutionSource(projectRoot),
			"rewrite `## RELATIONSHIPS` to use `none` or explicit `- builds on:`, `- depends on:`, or `- related to:` bullets",
			searchHintsForSection(projectRoot, path, "RELATIONSHIPS"),
		)}
	}

	var findings []reconcileFinding
	for _, relation := range relationships {
		if relationshipTargets != nil && !relationshipTargets[relation.Target] {
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				path,
				fmt.Sprintf("relationship target `%s` does not exist in `docs/specs/`", relation.Target),
				constitutionSource(projectRoot),
				"remove or correct the stale relationship target after checking the current feature directory names",
				[]string{
					fmt.Sprintf("rg -n \"^## RELATIONSHIPS|%s\" %s", relation.Target, filepath.Join(projectRoot, "docs", "specs")),
					fmt.Sprintf("ls %s", filepath.Join(projectRoot, "docs", "specs")),
				},
			))
		}
	}

	return findings
}

func auditTaskAlignment(path, projectRoot string) []reconcileFinding {
	doc, err := document.ParseFile(path, document.TypeTasks)
	if err != nil {
		return nil
	}

	tableIDs, ok := progressTableTaskIDs(doc.GetSection("PROGRESS TABLE"))
	if !ok {
		return nil
	}

	listIDs := reconcileTaskListPattern.FindAllStringSubmatch(doc.Content, -1)
	detailIDs := reconcileTaskDetailPattern.FindAllStringSubmatch(doc.Content, -1)
	listSet := matchesToSet(listIDs)
	detailSet := matchesToSet(detailIDs)

	var findings []reconcileFinding
	for _, id := range tableIDs {
		if !listSet[id] {
			findings = append(findings, newFinding(
				reconcileSeverityError,
				path,
				fmt.Sprintf("task `%s` exists in `PROGRESS TABLE` but not in `TASK LIST`", id),
				initProjectSource(projectRoot),
				"align the task list so every progress-table task has a matching checkbox entry",
				searchHintsForTaskAlignment(path),
			))
		}
		if !detailSet[id] {
			findings = append(findings, newFinding(
				reconcileSeverityError,
				path,
				fmt.Sprintf("task `%s` exists in `PROGRESS TABLE` but not in `TASK DETAILS`", id),
				initProjectSource(projectRoot),
				"add or restore the missing task-details block so every progress-table task has a matching `###` section",
				searchHintsForTaskAlignment(path),
			))
		}
	}

	for id := range listSet {
		if !slices.Contains(tableIDs, id) {
			findings = append(findings, newFinding(
				reconcileSeverityError,
				path,
				fmt.Sprintf("task `%s` exists in `TASK LIST` but not in `PROGRESS TABLE`", id),
				initProjectSource(projectRoot),
				"align the progress table so every checkbox task has a matching row",
				searchHintsForTaskAlignment(path),
			))
		}
	}

	return findings
}
