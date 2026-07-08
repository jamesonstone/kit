package cli

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

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

func activeFeatureForVerificationAdvisory(features []feature.Feature) *feature.Feature {
	for i := len(features) - 1; i >= 0; i-- {
		if featureNeedsVerificationAdvisory(&features[i]) {
			return &features[i]
		}
	}
	return nil
}

func auditExecutableVerificationAdvisory(projectRoot string, feat *feature.Feature) []reconcileFinding {
	if !featureNeedsVerificationAdvisory(feat) {
		return nil
	}

	tasksPath := filepath.Join(feat.Path, "TASKS.md")
	doc, err := document.ParseFile(tasksPath, document.TypeTasks)
	if err != nil {
		return nil
	}

	taskStatuses, ok := progressTableTaskStatuses(doc.GetSection("PROGRESS TABLE"))
	if !ok {
		return nil
	}
	details := reconcileTaskDetails(doc.Content)
	var missing []string
	for _, taskID := range verificationAdvisoryTaskIDs(taskStatuses, feat.Phase) {
		missingFields := missingExecutableFields(details[taskID])
		if len(missingFields) == 0 {
			continue
		}
		missing = append(missing, fmt.Sprintf("%s missing %s", taskID, strings.Join(missingFields, ", ")))
	}
	if len(missing) == 0 {
		return nil
	}

	return []reconcileFinding{newFinding(
		reconcileSeverityWarning,
		tasksPath,
		fmt.Sprintf("active feature tasks do not declare executable verification fields: %s", strings.Join(missing, "; ")),
		templateSource(projectRoot),
		"add `VERIFY`, `EXPECTED FILES`, `RISK`, and `ROLLBACK` to active task details where commands are known; if acceptance criteria are prose-only, propose runnable checks separately from confirmed checks and leave uncertain commands as `not yet declared`",
		[]string{
			fmt.Sprintf("sed -n '1,260p' %s", tasksPath),
			fmt.Sprintf("kit legacy verify %s --dry-run", feat.Slug),
			fmt.Sprintf("kit check %s", feat.Slug),
		},
	)}
}

func featureNeedsVerificationAdvisory(feat *feature.Feature) bool {
	if feat == nil || feat.Paused {
		return false
	}
	return feat.Phase == feature.PhaseImplement || feat.Phase == feature.PhaseReflect
}

func progressTableTaskStatuses(section *document.Section) (map[string]string, bool) {
	if section == nil {
		return nil, false
	}
	rows := tableRows(section.Content)
	if len(rows) == 0 {
		return nil, false
	}
	result := make(map[string]string)
	for _, row := range rows {
		if len(row) < 3 || !reconcileTaskIDPattern.MatchString(row[0]) {
			continue
		}
		result[row[0]] = strings.ToLower(strings.TrimSpace(row[2]))
	}
	return result, len(result) > 0
}

func verificationAdvisoryTaskIDs(taskStatuses map[string]string, phase feature.Phase) []string {
	ids := make([]string, 0, len(taskStatuses))
	for id, status := range taskStatuses {
		if phase == feature.PhaseImplement && status == "done" {
			continue
		}
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}

func reconcileTaskDetails(content string) map[string]map[string]bool {
	details := make(map[string]map[string]bool)
	matches := reconcileTaskHeadingPattern.FindAllStringSubmatchIndex(content, -1)
	for i, match := range matches {
		id := content[match[2]:match[3]]
		start := match[1]
		end := len(content)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		if sectionEnd := nextReconcileSection(content, start); sectionEnd >= 0 && sectionEnd < end {
			end = sectionEnd
		}
		details[id] = reconcileTaskFields(content[start:end])
	}
	return details
}
