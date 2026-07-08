package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
)

func reconcileFindingCategory(finding reconcileFinding) string {
	lowerIssue := strings.ToLower(finding.Issue)
	base := filepath.Base(finding.FilePath)

	switch {
	case strings.Contains(lowerIssue, "init scaffold") || strings.Contains(lowerIssue, ".gitignore"):
		return "init scaffold"
	case strings.Contains(lowerIssue, "executable verification"):
		return "verification"
	case strings.Contains(lowerIssue, "reference") || strings.Contains(lowerIssue, "dependencies are deprecated"):
		return "references"
	case strings.Contains(lowerIssue, "relationship"):
		return "relationships"
	case strings.Contains(lowerIssue, "task `") || strings.Contains(lowerIssue, "task details") || base == "TASKS.md":
		return "tasks"
	case strings.Contains(lowerIssue, "instruction file"):
		return "instruction files"
	case strings.Contains(lowerIssue, "progress summary") || base == "PROJECT_PROGRESS_SUMMARY.md":
		return "progress summary"
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
	case strings.Contains(issue, "init scaffold") || strings.Contains(issue, ".gitignore"):
		return "refresh init scaffold"
	case strings.Contains(issue, "executable verification"):
		return "add verification fields"
	case strings.Contains(issue, "reference") || strings.Contains(issue, "dependencies are deprecated"):
		return "migrate references"
	case strings.Contains(issue, "task `") || strings.Contains(issue, "task details"):
		return "align task IDs"
	case strings.Contains(issue, "relationship"):
		return "fix relationships"
	case strings.Contains(issue, "instruction file"):
		return "refresh instruction file"
	case strings.Contains(issue, "progress summary"):
		return "refresh progress summary"
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

func reconcileInstructionShortcut(projectRoot string) string {
	cfg := config.LoadOrDefault(projectRoot)
	version := detectInstructionScaffoldVersion(projectRoot, cfg)
	if config.IsInstructionScaffoldVersionSupported(version) {
		return fmt.Sprintf("kit scaffold agents --version %d --append-only", version)
	}

	return "kit scaffold agents --append-only"
}
