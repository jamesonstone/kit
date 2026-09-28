package cli

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/jamesonstone/kit/v3/internal/document"
)

var (
	reconcileCommentPattern     = regexp.MustCompile(`(?s)<!--.*?-->`)
	reconcileTaskIDPattern      = regexp.MustCompile(`^T\d{3}$`)
	reconcileTaskListPattern    = regexp.MustCompile(`(?m)^\s*-\s*\[[ xX]\]\s*(T\d{3}):`)
	reconcileTaskDetailPattern  = regexp.MustCompile(`(?m)^###\s*(T\d{3})\s*$`)
	reconcileTaskHeadingPattern = regexp.MustCompile(`(?m)^###\s*(T\d{3})\s*$`)
	reconcileTaskFieldPattern   = regexp.MustCompile(`^\s*-\s+\*\*([A-Z][A-Z -]*)\*\*:\s*(.*)$`)
	reconcileSectionPattern     = regexp.MustCompile(`(?m)^##\s+`)
)

func auditStructuredDocument(path string, docType document.DocumentType, projectRoot string, relationshipTargets map[string]bool) []reconcileFinding {
	doc, err := document.ParseFile(path, docType)
	if err != nil {
		return []reconcileFinding{newFinding(
			reconcileSeverityError,
			path,
			fmt.Sprintf("failed to parse `%s`", filepath.Base(path)),
			"fix the markdown structure before reconciliation continues",
		)}
	}

	var findings []reconcileFinding
	findings = append(findings, auditMetadataDiagnostics(path, doc, projectRoot)...)
	findings = append(findings, auditMetadataMigrationState(path, doc, projectRoot)...)
	findings = append(findings, auditRulesetReferences(projectRoot, path, doc)...)

	for _, section := range doc.RequiredSections() {
		entry := doc.GetSection(section)
		if entry == nil {
			findings = append(findings, newFinding(
				reconcileSeverityError,
				path,
				fmt.Sprintf("missing required section `## %s`", section),
				fmt.Sprintf("add `## %s` and populate it with current repository-backed content", section),
			))
			continue
		}
		if (docType == document.TypeConstitution || doc.RequiresPopulatedSection(section)) && !meaningfulSectionContent(entry.Content) {
			findings = append(findings, newFinding(
				reconcileSeverityError,
				path,
				fmt.Sprintf("required section `## %s` is empty or placeholder-only", section),
				fmt.Sprintf("replace placeholder-only content in `## %s` with current repo-backed content", section),
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
				fmt.Sprintf("reshape `## %s` to match the current Kit table contract", expectation.Section),
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
		finding := newFinding(
			severity,
			path,
			fmt.Sprintf("front matter metadata issue: %s", diagnostic.Message),
			diagnostic.Fix,
		)
		if doc.Type == document.TypeSpec && doc.Metadata != nil && doc.Metadata.WorkflowVersion == document.WorkflowVersionV2 && severity == reconcileSeverityWarning {
			finding.NonBlocking = true
		}
		findings = append(findings, finding)
	}
	for _, conflict := range doc.MetadataConflictWarnings {
		findings = append(findings, newFinding(
			reconcileSeverityWarning,
			path,
			fmt.Sprintf("front matter/body metadata conflict: %s", conflict.Message),
			"treat front matter as canonical and update or remove the stale body metadata",
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
			"update front matter feature identity to match the canonical feature directory",
		))
	}
	if doc.Metadata.Feature.Slug != "" && doc.Metadata.Feature.Slug != expected.Slug {
		findings = append(findings, newFinding(
			reconcileSeverityError,
			path,
			fmt.Sprintf("front matter feature.slug `%s` does not match containing feature directory slug `%s`", doc.Metadata.Feature.Slug, expected.Slug),
			"update front matter feature identity to match the canonical feature directory",
		))
	}
	if doc.Metadata.Feature.Dir != "" && doc.Metadata.Feature.Dir != expected.Dir {
		findings = append(findings, newFinding(
			reconcileSeverityError,
			path,
			fmt.Sprintf("front matter feature.dir `%s` does not match containing feature directory `%s`", doc.Metadata.Feature.Dir, expected.Dir),
			"update front matter feature identity to match the canonical feature directory",
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
		"add typed front matter for artifact identity, feature identity, relationships, references, and skills as applicable",
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
				"rewrite `## RELATIONSHIPS` to use `none` or explicit `- builds on:`, `- depends on:`, or `- related to:` bullets",
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
				"remove or correct the stale relationship target after checking the current feature directory names",
			))
		}
	}

	return findings
}
