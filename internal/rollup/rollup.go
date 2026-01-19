// package rollup generates PROJECT_PROGRESS_SUMMARY.md.
package rollup

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

// FeatureSummary contains extracted information about a feature for the rollup.
type FeatureSummary struct {
	ID        string
	Name      string
	Path      string
	Phase     feature.Phase
	Created   time.Time
	Summary   string
	Intent    string
	Approach  string
	OpenItems string
}

// Generate creates or updates the PROJECT_PROGRESS_SUMMARY.md file.
func Generate(projectRoot string, cfg *config.Config) error {
	specsDir := cfg.SpecsPath(projectRoot)
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return fmt.Errorf("failed to list features: %w", err)
	}

	summaries := make([]FeatureSummary, 0, len(features))
	for _, f := range features {
		summary := extractFeatureSummary(f, cfg.SpecsDir)
		summaries = append(summaries, summary)
	}

	content := generateContent(summaries, cfg)
	summaryPath := cfg.ProgressSummaryPath(projectRoot)

	if err := document.Write(summaryPath, content); err != nil {
		return fmt.Errorf("failed to write PROJECT_PROGRESS_SUMMARY.md: %w", err)
	}

	return nil
}

func extractFeatureSummary(f feature.Feature, specsDir string) FeatureSummary {
	summary := FeatureSummary{
		ID:      fmt.Sprintf("%04d", f.Number),
		Name:    f.Slug,
		Path:    filepath.Join(specsDir, f.DirName),
		Phase:   f.Phase,
		Created: f.CreatedAt,
	}

	// try to extract info from SPEC.md
	specPath := filepath.Join(f.Path, "SPEC.md")
	if doc, err := document.ParseFile(specPath, document.TypeSpec); err == nil {
		// extract problem section as summary
		if section := doc.GetSection("PROBLEM"); section != nil {
			summary.Summary = document.ExtractFirstParagraph(section)
			summary.Intent = summary.Summary
		}

		// extract open questions
		if section := doc.GetSection("OPEN-QUESTIONS"); section != nil {
			summary.OpenItems = document.ExtractFirstParagraph(section)
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
		summary.Intent = "(see SPEC.md)"
	}
	if summary.Approach == "" {
		summary.Approach = "(see PLAN.md)"
	}
	if summary.OpenItems == "" {
		summary.OpenItems = "none"
	}

	return summary
}

func generateContent(summaries []FeatureSummary, cfg *config.Config) string {
	var b strings.Builder

	b.WriteString("# PROJECT PROGRESS SUMMARY\n\n")

	// feature progress table
	b.WriteString("## FEATURE PROGRESS TABLE\n\n")
	b.WriteString("| ID | FEATURE | PATH | PHASE | CREATED | SUMMARY |\n")
	b.WriteString("| -- | ------- | ---- | ----- | ------- | ------- |\n")

	for _, s := range summaries {
		created := s.Created.Format("2006-01-02")
		// truncate summary for table
		tableSummary := s.Summary
		if len(tableSummary) > 60 {
			tableSummary = tableSummary[:57] + "..."
		}
		b.WriteString(fmt.Sprintf("| %s | %s | `%s` | %s | %s | %s |\n",
			s.ID, s.Name, s.Path, s.Phase, created, tableSummary))
	}

	b.WriteString("\n")

	// project intent
	b.WriteString("## PROJECT INTENT\n\n")
	b.WriteString("<!-- TODO: describe the overall project purpose -->\n\n")

	// global constraints
	b.WriteString("## GLOBAL CONSTRAINTS\n\n")
	b.WriteString(fmt.Sprintf("See `%s` for project-wide constraints and principles.\n\n", cfg.ConstitutionPath))

	// feature summaries
	b.WriteString("## FEATURE SUMMARIES\n\n")

	for _, s := range summaries {
		b.WriteString(fmt.Sprintf("### %s\n\n", s.Name))
		b.WriteString(fmt.Sprintf("- **STATUS**: %s\n", s.Phase))
		b.WriteString(fmt.Sprintf("- **INTENT**: %s\n", s.Intent))
		b.WriteString(fmt.Sprintf("- **APPROACH**: %s\n", s.Approach))
		b.WriteString(fmt.Sprintf("- **OPEN ITEMS**: %s\n", s.OpenItems))
		b.WriteString(fmt.Sprintf("- **POINTERS**: `%s/SPEC.md`, `%s/PLAN.md`, `%s/TASKS.md`\n\n",
			s.Path, s.Path, s.Path))
	}

	// last updated
	b.WriteString("## LAST UPDATED\n\n")
	b.WriteString(fmt.Sprintf("%s\n", time.Now().Format("2006-01-02 15:04:05 MST")))

	return b.String()
}

// Update is an alias for Generate (updates the existing file).
func Update(projectRoot string, cfg *config.Config) error {
	return Generate(projectRoot, cfg)
}
