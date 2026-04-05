package cli

import (
	"fmt"
	"strings"
)

func printReconcileSummary(report *reconcileReport) {
	fmt.Print(renderReconcileSummary(report, styleForStdout()))
}

func renderReconcileSummary(report *reconcileReport, style humanOutputStyle) string {
	fileSummaries := summarizeReconcileFiles(report.Findings)
	errors, warnings := reconcileSeverityCounts(report.Findings)

	var sb strings.Builder
	if divider := style.sectionDivider(); divider != "" {
		sb.WriteString(divider)
		sb.WriteString("\n")
	}
	sb.WriteString(style.title("🧩", "Reconcile Audit"))
	sb.WriteString("\n")
	if divider := style.sectionDivider(); divider != "" {
		sb.WriteString(divider)
		sb.WriteString("\n")
	}

	scope := "whole project"
	if report.Feature != nil {
		scope = fmt.Sprintf("feature %s", report.Feature.Slug)
	}

	sb.WriteString(fmt.Sprintf("%s %s\n", style.label("Scope:"), scope))
	sb.WriteString(fmt.Sprintf("%s %d (%d errors, %d warnings) across %d files\n\n",
		style.label("Findings:"), len(report.Findings), errors, warnings, len(fileSummaries)))

	sb.WriteString(renderReconcileSummaryTable(fileSummaries))
	sb.WriteString("\n")
	sb.WriteString(style.muted("raw prompt stays compact; the clipboard payload omits this terminal summary"))
	sb.WriteString("\n\n")
	return sb.String()
}

func renderReconcileSummaryTable(fileSummaries []reconcileFileSummary) string {
	headers := []string{"Severity", "Issues", "File", "Focus"}
	rows := make([][]string, 0, len(fileSummaries))
	widths := []int{len(headers[0]), len(headers[1]), len(headers[2]), len(headers[3])}

	for _, summary := range fileSummaries {
		row := []string{
			reconcileSeverityBadge(summary.ErrorCount, summary.WarnCount),
			fmt.Sprintf("%d", len(summary.Issues)),
			summary.Path,
			strings.Join(limitStrings(summary.Actions, 2), "; "),
		}
		rows = append(rows, row)
		for i := range row {
			if len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
	}

	var sb strings.Builder
	sb.WriteString(formatReconcileTableRow(headers, widths))
	sb.WriteString("\n")
	sb.WriteString(formatReconcileDivider(widths))
	sb.WriteString("\n")
	for _, row := range rows {
		sb.WriteString(formatReconcileTableRow(row, widths))
		sb.WriteString("\n")
	}
	return sb.String()
}

func formatReconcileTableRow(values []string, widths []int) string {
	return fmt.Sprintf(
		"%-*s  %-*s  %-*s  %-*s",
		widths[0], values[0],
		widths[1], values[1],
		widths[2], values[2],
		widths[3], values[3],
	)
}

func formatReconcileDivider(widths []int) string {
	parts := make([]string, 0, len(widths))
	for _, width := range widths {
		parts = append(parts, strings.Repeat("─", width))
	}
	return strings.Join(parts, "  ")
}
