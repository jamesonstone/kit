// package rollup generates PROJECT_PROGRESS_SUMMARY.md.
package rollup

import (
	"fmt"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

func generateContent(summaries []FeatureSummary, cfg *config.Config) string {
	var b strings.Builder

	b.WriteString("# PROJECT PROGRESS SUMMARY\n\n")

	// feature progress table
	b.WriteString("## FEATURE PROGRESS TABLE\n\n")
	b.WriteString("| ID | FEATURE | PATH | PHASE | PAUSED | CREATED | SUMMARY |\n")
	b.WriteString("| -- | ------- | ---- | ----- | ------ | ------- | ------- |\n")

	for _, s := range summaries {
		created := formatFeatureDate(s.Created)
		paused := "no"
		if s.Paused {
			paused = "yes"
		}
		writef(&b, "| %s | %s | `%s` | %s | %s | %s | %s |\n",
			s.ID, s.Name, s.Path, featureSummaryStatus(s), paused, created, formatSummaryTableCell(s.Summary))
	}

	b.WriteString("\n")

	// project intent
	b.WriteString("## PROJECT INTENT\n\n")
	b.WriteString(projectIntentSummary() + "\n\n")

	// global constraints
	b.WriteString("## GLOBAL CONSTRAINTS\n\n")
	writef(&b, "See `%s` for project-wide constraints and principles.\n\n", cfg.ConstitutionPath)

	// feature summaries
	b.WriteString("## FEATURE SUMMARIES\n\n")

	for _, s := range summaries {
		writef(&b, "### %s\n\n", s.Name)
		writef(&b, "- **STATUS**: %s\n", featureSummaryStatus(s))
		if s.Paused {
			b.WriteString("- **PAUSED**: yes\n")
		} else {
			b.WriteString("- **PAUSED**: no\n")
		}
		writef(&b, "- **INTENT**: %s\n", s.Intent)
		writef(&b, "- **APPROACH**: %s\n", s.Approach)
		writef(&b, "- **OPEN ITEMS**: %s\n", s.OpenItems)
		var pointers []string
		if s.HasBrainstorm {
			pointers = append(pointers, fmt.Sprintf("`%s/BRAINSTORM.md`", s.Path))
		}
		pointers = append(pointers, fmt.Sprintf("`%s/SPEC.md`", s.Path))
		if s.WorkflowVersion != document.WorkflowVersionV3 {
			pointers = append(pointers,
				fmt.Sprintf("`%s/PLAN.md`", s.Path),
				fmt.Sprintf("`%s/TASKS.md`", s.Path),
			)
		}
		writef(&b, "- **POINTERS**: %s\n\n", strings.Join(pointers, ", "))
	}

	// last updated
	b.WriteString("## LAST UPDATED\n\n")
	writef(&b, "%s\n", time.Now().Format("2006-01-02 15:04:05 MST"))

	return b.String()
}

func projectIntentSummary() string {
	return "Kit is a coding-agent-first repository contract and evidence harness. Native agent planning owns research and design; Kit resolves local workflows and rules while canonical repository documents preserve consequential rationale, validation, and outcomes."
}

func writef(b *strings.Builder, format string, args ...any) {
	_, _ = fmt.Fprintf(b, format, args...)
}

// Update is an alias for Generate (updates the existing file).
func Update(projectRoot string, cfg *config.Config) error {
	return Generate(projectRoot, cfg)
}

func formatSummaryTableCell(summary string) string {
	normalized := strings.Join(strings.Fields(summary), " ")
	return strings.ReplaceAll(normalized, "|", `\|`)
}

func featureSummaryStatus(summary FeatureSummary) string {
	return string(summary.Phase)
}

func formatFeatureDate(date time.Time) string {
	if date.IsZero() {
		return "unknown"
	}
	return date.Format("2006-01-02")
}
