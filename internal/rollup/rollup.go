// package rollup generates PROJECT_PROGRESS_SUMMARY.md.
package rollup

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

// FeatureSummary contains extracted information about a feature for the rollup.
type FeatureSummary struct {
	ID            string
	Name          string
	Path          string
	Phase         feature.Phase
	Paused        bool
	Removed       bool
	Created       time.Time
	RemovedAt     time.Time
	HasNotes      bool
	NotesPath     string
	HasBrainstorm bool
	Summary       string
	Intent        string
	Approach      string
	OpenItems     string
}

func removedFeatureNotesPath(dirName string) string {
	return filepath.ToSlash(filepath.Join("docs", "notes", dirName))
}

// Generate creates or updates the PROJECT_PROGRESS_SUMMARY.md file.
func Generate(projectRoot string, cfg *config.Config) error {
	specsDir := cfg.SpecsPath(projectRoot)
	features, err := feature.ListFeaturesWithState(specsDir, cfg)
	if err != nil {
		return fmt.Errorf("failed to list features: %w", err)
	}

	summaries := make([]FeatureSummary, 0, len(features))
	liveFeatureDirs := make(map[string]struct{}, len(features))
	for _, f := range features {
		liveFeatureDirs[f.DirName] = struct{}{}
		summary := extractFeatureSummary(f, cfg.SpecsDir)
		summaries = append(summaries, summary)
	}
	for _, removed := range cfg.RemovedFeatures {
		if removed.DirName == "" {
			continue
		}
		if _, exists := liveFeatureDirs[removed.DirName]; exists {
			continue
		}
		summaries = append(summaries, removedFeatureSummary(projectRoot, removed, cfg.SpecsDir))
	}
	sortFeatureSummaries(summaries)

	content := generateContent(summaries, cfg)
	summaryPath := cfg.ProgressSummaryPath(projectRoot)

	if err := document.Write(summaryPath, content); err != nil {
		return fmt.Errorf("failed to write PROJECT_PROGRESS_SUMMARY.md: %w", err)
	}

	return nil
}

func extractFeatureSummary(f feature.Feature, specsDir string) FeatureSummary {
	summary := FeatureSummary{
		ID:            fmt.Sprintf("%04d", f.Number),
		Name:          f.Slug,
		Path:          filepath.Join(specsDir, f.DirName),
		Phase:         f.Phase,
		Paused:        f.Paused,
		Created:       f.CreatedAt,
		HasBrainstorm: document.Exists(filepath.Join(f.Path, "BRAINSTORM.md")),
	}

	brainstormPath := filepath.Join(f.Path, "BRAINSTORM.md")

	// try to extract info from SPEC.md
	specPath := filepath.Join(f.Path, "SPEC.md")
	if doc, err := document.ParseFile(specPath, document.TypeSpec); err == nil {
		if section := doc.GetSection("SUMMARY"); section != nil {
			summary.Summary = document.ExtractFirstParagraph(section)
		}

		// extract problem section as intent
		if section := doc.GetSection("PROBLEM"); section != nil {
			summary.Intent = document.ExtractFirstParagraph(section)
			if summary.Summary == "" {
				summary.Summary = summary.Intent
			}
		}

		// extract open questions
		if section := doc.GetSection("OPEN-QUESTIONS"); section != nil {
			summary.OpenItems = document.ExtractFirstParagraph(section)
		}
	}

	if summary.Summary == "" && summary.HasBrainstorm {
		if brainstormSummary, err := feature.ExtractBrainstormSummary(brainstormPath); err == nil {
			summary.Summary = brainstormSummary
			if summary.Intent == "" {
				summary.Intent = brainstormSummary
			}
		}

		if doc, err := document.ParseFile(brainstormPath, document.TypeBrainstorm); err == nil {
			if summary.OpenItems == "" {
				if section := doc.GetSection("QUESTIONS"); section != nil {
					summary.OpenItems = document.ExtractFirstParagraph(section)
				}
			}
		}
	}

	// try to extract approach from PLAN.md
	planPath := filepath.Join(f.Path, "PLAN.md")
	if doc, err := document.ParseFile(planPath, document.TypePlan); err == nil {
		if section := doc.GetSection("APPROACH"); section != nil {
			summary.Approach = document.ExtractFirstParagraph(section)
		}
	}

	// set defaults for missing fields
	if summary.Summary == "" {
		summary.Summary = "(no description)"
	}
	if summary.Intent == "" {
		summary.Intent = summary.Summary
	}
	if summary.Approach == "" {
		summary.Approach = "(see PLAN.md)"
	}
	if summary.OpenItems == "" {
		summary.OpenItems = "none"
	}

	return summary
}

func removedFeatureSummary(projectRoot string, removed config.RemovedFeature, specsDir string) FeatureSummary {
	number := removed.Number
	slug := removed.Slug
	if number == 0 || slug == "" {
		parsedNumber, parsedSlug, ok := feature.ParseDirName(removed.DirName)
		if ok {
			if number == 0 {
				number = parsedNumber
			}
			if slug == "" {
				slug = parsedSlug
			}
		}
	}

	createdAt := parseConfigTimestamp(removed.CreatedAt)
	removedAt := parseConfigTimestamp(removed.RemovedAt)
	summary := "Removed by kit rm."
	if !removedAt.IsZero() {
		summary = fmt.Sprintf("Removed by kit rm on %s.", removedAt.Format("2006-01-02"))
	}
	notesPath := removedFeatureNotesPath(removed.DirName)

	return FeatureSummary{
		ID:        fmt.Sprintf("%04d", number),
		Name:      slug,
		Path:      filepath.Join(specsDir, removed.DirName),
		Removed:   true,
		Created:   createdAt,
		RemovedAt: removedAt,
		HasNotes:  document.Exists(filepath.Join(projectRoot, notesPath)),
		NotesPath: notesPath,
		Summary:   summary,
		Intent:    summary,
		Approach:  "Feature directory and docs were deleted wholesale by `kit rm`; tombstone retained for project history.",
		OpenItems: "none",
	}
}

func parseConfigTimestamp(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed
	}
	if parsed, err := time.Parse("2006-01-02", value); err == nil {
		return parsed
	}
	return time.Time{}
}

func sortFeatureSummaries(summaries []FeatureSummary) {
	sort.SliceStable(summaries, func(i, j int) bool {
		if summaries[i].ID != summaries[j].ID {
			return summaries[i].ID < summaries[j].ID
		}
		return summaries[i].Name < summaries[j].Name
	})
}

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

	// last updated
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
