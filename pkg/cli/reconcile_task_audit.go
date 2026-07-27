package cli

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

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

func reconcileTaskFields(content string) map[string]bool {
	fields := make(map[string]bool)
	for _, line := range strings.Split(content, "\n") {
		match := reconcileTaskFieldPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		fields[strings.ToUpper(strings.TrimSpace(match[1]))] = true
	}
	return fields
}

func nextReconcileSection(content string, start int) int {
	matches := reconcileSectionPattern.FindAllStringIndex(content[start:], -1)
	if len(matches) == 0 {
		return -1
	}
	return start + matches[0][0]
}

func missingExecutableFields(fields map[string]bool) []string {
	required := []string{"VERIFY", "EXPECTED FILES", "RISK", "ROLLBACK"}
	var missing []string
	for _, field := range required {
		if !fields[field] {
			missing = append(missing, field)
		}
	}
	return missing
}
