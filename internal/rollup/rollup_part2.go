package rollup

import (
	"fmt"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/config"
)

func generateContent(summaries []FeatureSummary, cfg *config.Config) string {
	var b strings.Builder

	b.WriteString("# PROJECT PROGRESS SUMMARY\n\n")

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

	b.WriteString("## PROJECT INTENT\n\n")
	b.WriteString(projectIntentSummary() + "\n\n")

	b.WriteString("## GLOBAL CONSTRAINTS\n\n")
	writef(&b, "See `%s` for project-wide constraints and principles.\n\n", cfg.ConstitutionPath)

	b.WriteString("## FEATURE SUMMARIES\n\n")

	for _, s := range summaries {
		writef(&b, "### %s\n\n", s.Name)
		writef(&b, "- **STATUS**: %s\n", featureSummaryStatus(s))
		if s.Paused {
			b.WriteString("- **PAUSED**: yes\n")
		} else {
			b.WriteString("- **PAUSED**: no\n")
		}
		if s.Removed && !s.RemovedAt.IsZero() {
			writef(&b, "- **REMOVED AT**: %s\n", s.RemovedAt.Format("2006-01-02"))
		}
		writef(&b, "- **INTENT**: %s\n", s.Intent)
		writef(&b, "- **APPROACH**: %s\n", s.Approach)
		writef(&b, "- **OPEN ITEMS**: %s\n", s.OpenItems)
		if s.Removed {
			pointers := fmt.Sprintf("removed; original docs path was `%s`", s.Path)
			if s.HasNotes {
				pointers += fmt.Sprintf("; retained notes at `%s`", s.NotesPath)
			}
			writef(&b, "- **POINTERS**: %s\n\n", pointers)
			continue
		}
		var pointers []string
		if s.HasBrainstorm {
			pointers = append(pointers, fmt.Sprintf("`%s/BRAINSTORM.md`", s.Path))
		}
		pointers = append(pointers,
			fmt.Sprintf("`%s/SPEC.md`", s.Path),
			fmt.Sprintf("`%s/PLAN.md`", s.Path),
			fmt.Sprintf("`%s/TASKS.md`", s.Path),
		)
		writef(&b, "- **POINTERS**: %s\n\n", strings.Join(pointers, ", "))
	}

	b.WriteString("## LAST UPDATED\n\n")
	writef(&b, "%s\n", time.Now().Format("2006-01-02 15:04:05 MST"))

	return b.String()
}

func projectIntentSummary() string {
	return "Kit is a document-first workflow harness for disciplined thought work. It keeps durable project context in canonical markdown artifacts so humans and coding agents can move from research to specification, planning, tasks, implementation, reflection, and completion with explicit traceability."
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
	if summary.Removed {
		return "removed"
	}
	return string(summary.Phase)
}

func formatFeatureDate(date time.Time) string {
	if date.IsZero() {
		return "unknown"
	}
	return date.Format("2006-01-02")
}
