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
	findings = append(findings, auditMetadataDiagnostics(path, doc, projectRoot)...)
	findings = append(findings, auditMetadataMigrationState(path, doc, projectRoot)...)

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
		if doc.FrontMatterPresent && metadataSectionTableMigratedToFrontMatter(expectation.Section) {
			continue
		}
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

func auditMetadataDiagnostics(path string, doc *document.Document, projectRoot string) []reconcileFinding {
	var findings []reconcileFinding
	for _, diagnostic := range doc.MetadataDiagnostics {
		severity := reconcileSeverityWarning
		if diagnostic.Severity == document.MetadataDiagnosticError {
			severity = reconcileSeverityError
		}
		findings = append(findings, newFinding(
			severity,
			path,
			fmt.Sprintf("front matter metadata issue: %s", diagnostic.Message),
			contractSourceForSection(projectRoot, doc.Type, "FRONT MATTER"),
			diagnostic.Fix,
			[]string{fmt.Sprintf("sed -n '1,80p' %s", path)},
		))
	}
	for _, conflict := range doc.MetadataConflictWarnings {
		findings = append(findings, newFinding(
			reconcileSeverityWarning,
			path,
			fmt.Sprintf("front matter/body metadata conflict: %s", conflict.Message),
			contractSourceForSection(projectRoot, doc.Type, "FRONT MATTER"),
			"treat front matter as canonical and update or remove the stale body metadata",
			[]string{fmt.Sprintf("sed -n '1,140p' %s", path)},
		))
	}
	if doc.Metadata != nil {
		actualDir := filepath.Base(filepath.Dir(path))
		if actualDir != "." {
			expected := document.FeatureMetadataFromDir(actualDir)
			findings = append(findings, metadataIdentityFindings(path, doc, projectRoot, expected)...)
		}
	}
	return findings
}

func metadataIdentityFindings(path string, doc *document.Document, projectRoot string, expected document.FeatureMetadata) []reconcileFinding {
	if doc.Metadata == nil {
		return nil
	}

	var findings []reconcileFinding
	if doc.Metadata.Feature.ID != "" && doc.Metadata.Feature.ID != expected.ID {
		findings = append(findings, newFinding(
			reconcileSeverityError,
			path,
			fmt.Sprintf("front matter feature.id `%s` does not match containing feature directory id `%s`", doc.Metadata.Feature.ID, expected.ID),
			contractSourceForSection(projectRoot, doc.Type, "FRONT MATTER"),
			"update front matter feature identity to match the canonical feature directory",
			[]string{fmt.Sprintf("sed -n '1,80p' %s", path)},
		))
	}
	if doc.Metadata.Feature.Slug != "" && doc.Metadata.Feature.Slug != expected.Slug {
		findings = append(findings, newFinding(
			reconcileSeverityError,
			path,
			fmt.Sprintf("front matter feature.slug `%s` does not match containing feature directory slug `%s`", doc.Metadata.Feature.Slug, expected.Slug),
			contractSourceForSection(projectRoot, doc.Type, "FRONT MATTER"),
			"update front matter feature identity to match the canonical feature directory",
			[]string{fmt.Sprintf("sed -n '1,80p' %s", path)},
		))
	}
	if doc.Metadata.Feature.Dir != "" && doc.Metadata.Feature.Dir != expected.Dir {
		findings = append(findings, newFinding(
			reconcileSeverityError,
			path,
			fmt.Sprintf("front matter feature.dir `%s` does not match containing feature directory `%s`", doc.Metadata.Feature.Dir, expected.Dir),
			contractSourceForSection(projectRoot, doc.Type, "FRONT MATTER"),
			"update front matter feature identity to match the canonical feature directory",
			[]string{fmt.Sprintf("sed -n '1,80p' %s", path)},
		))
	}
	return findings
}

func auditMetadataMigrationState(path string, doc *document.Document, projectRoot string) []reconcileFinding {
	if !featureArtifactType(doc.Type) || doc.FrontMatterPresent {
		return nil
	}
	return []reconcileFinding{newFinding(
		reconcileSeverityWarning,
		path,
		"feature artifact is missing canonical YAML front matter and is using legacy markdown metadata fallback",
		contractSourceForSection(projectRoot, doc.Type, "FRONT MATTER"),
		"add typed front matter for artifact identity, feature identity, relationships, dependencies, and skills as applicable",
		[]string{fmt.Sprintf("sed -n '1,80p' %s", path)},
	)}
}

func featureArtifactType(docType document.DocumentType) bool {
	switch docType {
	case document.TypeBrainstorm, document.TypeSpec, document.TypePlan, document.TypeTasks:
		return true
	default:
		return false
	}
}

func metadataSectionTableMigratedToFrontMatter(section string) bool {
	switch section {
	case "RELATIONSHIPS", "DEPENDENCIES", "SKILLS":
		return true
	default:
		return false
	}
}

func auditRelationships(path string, doc *document.Document, projectRoot string, relationshipTargets map[string]bool) []reconcileFinding {
	section := doc.GetSection("RELATIONSHIPS")
	if section == nil || !meaningfulSectionContent(section.Content) {
		return nil
	}

	var relationships []document.Relationship
	if doc.FrontMatterPresent {
		var parseWarnings []document.RelationshipParseWarning
		relationships, parseWarnings = doc.Relationships()
		if len(parseWarnings) > 0 {
			return nil
		}
	} else {
		parsedRelationships, err := document.ParseRelationshipsSection(section)
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
		relationships = parsedRelationships
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
