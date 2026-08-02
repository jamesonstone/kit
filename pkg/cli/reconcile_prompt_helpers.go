package cli

import (
	"fmt"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
)

func shortActionForFinding(finding reconcileFinding) string {
	issue := strings.ToLower(finding.Issue)
	switch {
	case finding.AllowsCodeChanges:
		return "split by responsibility"
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

func reconcileAllowsCodeChanges(findings []reconcileFinding) bool {
	for _, finding := range findings {
		if finding.AllowsCodeChanges {
			return true
		}
	}
	return false
}

func reconcileWorkflowScopeRule(findings []reconcileFinding) string {
	if !reconcileAllowsCodeChanges(findings) {
		return docsOnlyWorkflowRule("Kit-managed docs and scaffold files")
	}
	return "Only update files listed by this audit and directly required tests or canonical docs; source/test edits are authorized only for behavior-preserving semantic splits under the source-file-size rule. Do not change product behavior, dependencies, public interfaces, runtime configuration, or generated artifacts."
}

func reconcilePromptOpening(findings []reconcileFinding, scope string) string {
	if reconcileAllowsCodeChanges(findings) {
		return fmt.Sprintf("Reconcile Kit-managed project state for the %s.", scope)
	}
	return fmt.Sprintf("Reconcile Kit-managed docs for the %s.", scope)
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
