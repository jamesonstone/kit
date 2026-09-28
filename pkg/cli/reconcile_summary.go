package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// writeReconcileFindings prints what remains after reconcile as concise facts:
// severity, file, issue, and the fix. Agents and people act on these directly.
func writeReconcileFindings(out io.Writer, report *reconcileReport) error {
	errors, warnings := reconcileSeverityCounts(report.Findings)
	scope := "whole project"
	if report.Feature != nil {
		scope = "feature " + report.Feature.Slug
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Reconcile findings (%s): %d error(s), %d warning(s)\n", scope, errors, warnings)
	if evidence := sourceFileAuditEvidence(report.SourceFileAudit); evidence != "" {
		sb.WriteString(evidence + "\n")
	}
	for _, finding := range report.Findings {
		severity := string(finding.Severity)
		if finding.NonBlocking {
			severity += ", non-blocking"
		}
		fmt.Fprintf(&sb, "- [%s] %s: %s\n", severity, relativeFindingPath(report.ProjectRoot, finding.FilePath), finding.Issue)
		if finding.UpdateInstruction != "" {
			fmt.Fprintf(&sb, "  fix: %s\n", finding.UpdateInstruction)
		}
	}
	_, err := io.WriteString(out, sb.String())
	return err
}

func reconcileSeverityCounts(findings []reconcileFinding) (int, int) {
	errors, warnings := 0, 0
	for _, finding := range findings {
		if finding.Severity == reconcileSeverityError {
			errors++
		} else {
			warnings++
		}
	}
	return errors, warnings
}

func relativeFindingPath(projectRoot, path string) string {
	if rel, err := filepath.Rel(projectRoot, path); err == nil && !strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(rel)
	}
	return path
}
