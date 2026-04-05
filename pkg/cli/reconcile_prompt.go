package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type reconcileFileSummary struct {
	Path        string
	ErrorCount  int
	WarnCount   int
	Issues      []string
	Actions     []string
	SearchHints []string
}

type reconcileCategorySummary struct {
	Name        string
	SearchHints []string
}

func buildReconcilePrompt(report *reconcileReport) string {
	var sb strings.Builder

	scope := "whole project"
	verifyCmd := "`kit check --all`"
	if report.Feature != nil {
		scope = fmt.Sprintf("feature %s", report.Feature.Slug)
		verifyCmd = fmt.Sprintf("`kit check %s`", report.Feature.Slug)
	}

	fileSummaries := summarizeReconcileFiles(report.Findings)
	categorySummaries := summarizeReconcileCategories(report.Findings)
	errorCount, warningCount := reconcileSeverityCounts(report.Findings)

	sb.WriteString("/plan\n\n")
	sb.WriteString(fmt.Sprintf("Reconcile Kit-managed docs for the %s.\n\n", scope))
	sb.WriteString("Rules:\n")
	sb.WriteString("- docs only; no product code, test, or runtime changes\n")
	sb.WriteString("- preserve project wording when it already satisfies the current contract\n")
	if !singleAgent {
		sb.WriteString("- use subagents and queue work according to overlapping file changes; keep overlapping files in the same lane\n")
	}
	sb.WriteString(fmt.Sprintf("- contract order: %s -> %s -> %s\n\n", templateSource(report.ProjectRoot), constitutionSource(report.ProjectRoot), initProjectSource(report.ProjectRoot)))

	sb.WriteString("Audit snapshot:\n")
	sb.WriteString(fmt.Sprintf("- findings: %d (%d errors, %d warnings)\n", len(report.Findings), errorCount, warningCount))
	sb.WriteString(fmt.Sprintf("- files to touch: %d\n", len(fileSummaries)))
	sb.WriteString(fmt.Sprintf("- verify after edits: %s\n", verifyCmd))
	if report.NeedsRollup {
		sb.WriteString("- also run: `kit rollup`\n")
	} else {
		sb.WriteString("- also run `kit rollup` only if `PROJECT_PROGRESS_SUMMARY.md` changes\n")
	}
	sb.WriteString("\n")

	sb.WriteString("Files to fix:\n")
	sb.WriteString("| Severity | File | Issues | Focus |\n")
	sb.WriteString("| -------- | ---- | ------ | ----- |\n")
	for _, summary := range fileSummaries {
		sb.WriteString(fmt.Sprintf("| %s | `%s` | %d | %s |\n",
			reconcileSeverityBadge(summary.ErrorCount, summary.WarnCount),
			summary.Path,
			len(summary.Issues),
			strings.Join(limitStrings(summary.Actions, 2), "; "),
		))
	}
	sb.WriteString("\n")

	sb.WriteString("Notable issues:\n")
	for _, summary := range fileSummaries {
		sb.WriteString(fmt.Sprintf("- `%s`: %s\n",
			filepath.Base(summary.Path),
			strings.Join(limitStrings(summary.Issues, issueLimitForScope(report)), "; "),
		))
	}
	sb.WriteString("\n")

	sb.WriteString("Search shortcuts:\n")
	for _, category := range categorySummaries {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", category.Name, strings.Join(wrapCode(limitStrings(category.SearchHints, 2)), "; ")))
	}
	if hasInstructionFileFinding(report.Findings) {
		sb.WriteString("- instruction files: `kit scaffold-agents --append-only`\n")
	}
	sb.WriteString("\n")

	sb.WriteString("Reply with exactly these sections:\n")
	sb.WriteString("- `Findings`: bullets for what was stale\n")
	sb.WriteString("- `Updates`: bullets for what changed; include unresolved questions only if any remain\n")
	sb.WriteString("- `Verification`: bullets for commands run and whether they passed\n")

	return sb.String()
}

func summarizeReconcileFiles(findings []reconcileFinding) []reconcileFileSummary {
	summaries := make(map[string]*reconcileFileSummary)
	order := make([]string, 0, len(findings))

	for _, finding := range findings {
		summary, ok := summaries[finding.FilePath]
		if !ok {
			summary = &reconcileFileSummary{Path: finding.FilePath}
			summaries[finding.FilePath] = summary
			order = append(order, finding.FilePath)
		}

		if finding.Severity == reconcileSeverityError {
			summary.ErrorCount++
		} else {
			summary.WarnCount++
		}
		summary.Issues = appendUniqueString(summary.Issues, finding.Issue)
		summary.Actions = appendUniqueString(summary.Actions, shortActionForFinding(finding))
		for _, hint := range finding.SearchHints {
			summary.SearchHints = appendUniqueString(summary.SearchHints, hint)
		}
	}

	result := make([]reconcileFileSummary, 0, len(summaries))
	for _, path := range order {
		result = append(result, *summaries[path])
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].ErrorCount != result[j].ErrorCount {
			return result[i].ErrorCount > result[j].ErrorCount
		}
		if result[i].WarnCount != result[j].WarnCount {
			return result[i].WarnCount > result[j].WarnCount
		}
		return result[i].Path < result[j].Path
	})

	return result
}

func summarizeReconcileCategories(findings []reconcileFinding) []reconcileCategorySummary {
	grouped := map[string][]string{}
	order := []string{}

	for _, finding := range findings {
		category := reconcileFindingCategory(finding)
		if _, ok := grouped[category]; !ok {
			order = append(order, category)
		}
		for _, hint := range finding.SearchHints {
			grouped[category] = appendUniqueString(grouped[category], hint)
		}
	}

	result := make([]reconcileCategorySummary, 0, len(order))
	for _, category := range order {
		result = append(result, reconcileCategorySummary{
			Name:        category,
			SearchHints: grouped[category],
		})
	}

	return result
}

func reconcileFindingCategory(finding reconcileFinding) string {
	lowerIssue := strings.ToLower(finding.Issue)
	base := filepath.Base(finding.FilePath)

	switch {
	case strings.Contains(lowerIssue, "relationship"):
		return "relationships"
	case strings.Contains(lowerIssue, "task `") || strings.Contains(lowerIssue, "task details") || base == "TASKS.md":
		return "tasks"
	case strings.Contains(lowerIssue, "instruction file"):
		return "instruction files"
	case strings.Contains(lowerIssue, "progress summary") || base == "PROJECT_PROGRESS_SUMMARY.md":
		return "rollup"
	case strings.Contains(lowerIssue, "table"):
		return "tables"
	case strings.Contains(lowerIssue, "section"):
		return "sections"
	default:
		return strings.TrimSuffix(base, filepath.Ext(base))
	}
}

func reconcileSeverityBadge(errors, warnings int) string {
	switch {
	case errors > 0 && warnings > 0:
		return fmt.Sprintf("E%d/W%d", errors, warnings)
	case errors > 0:
		return fmt.Sprintf("E%d", errors)
	default:
		return fmt.Sprintf("W%d", warnings)
	}
}

func reconcileSeverityCounts(findings []reconcileFinding) (int, int) {
	var errors int
	var warnings int
	for _, finding := range findings {
		if finding.Severity == reconcileSeverityError {
			errors++
		} else {
			warnings++
		}
	}
	return errors, warnings
}

func shortActionForFinding(finding reconcileFinding) string {
	issue := strings.ToLower(finding.Issue)
	switch {
	case strings.Contains(issue, "task `") || strings.Contains(issue, "task details"):
		return "align task IDs"
	case strings.Contains(issue, "relationship"):
		return "fix relationships"
	case strings.Contains(issue, "instruction file"):
		return "refresh instruction file"
	case strings.Contains(issue, "progress summary"):
		return "refresh rollup"
	case strings.Contains(issue, "table"):
		return "repair required table"
	case strings.Contains(issue, "missing required section"):
		return "add missing section"
	case strings.Contains(issue, "placeholder-only") || strings.Contains(issue, "empty"):
		return "fill required section"
	case strings.Contains(issue, "missing `spec.md`"):
		return "create SPEC.md"
	case strings.Contains(issue, "missing `plan.md`"):
		return "create PLAN.md"
	default:
		return "reconcile document"
	}
}

func issueLimitForScope(report *reconcileReport) int {
	if report.Feature != nil {
		return 2
	}
	return 1
}

func limitStrings(values []string, max int) []string {
	if len(values) <= max {
		return values
	}
	return values[:max]
}

func appendUniqueString(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func wrapCode(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, fmt.Sprintf("`%s`", value))
	}
	return out
}

func hasInstructionFileFinding(findings []reconcileFinding) bool {
	for _, finding := range findings {
		if reconcileFindingCategory(finding) == "instruction files" {
			return true
		}
	}
	return false
}
