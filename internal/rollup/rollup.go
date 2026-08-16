// package rollup generates PROJECT_PROGRESS_SUMMARY.md.
package rollup

import (
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/document"
	"github.com/jamesonstone/kit/v3/internal/feature"
)

// FeatureSummary contains extracted information about a feature for the rollup.
type FeatureSummary struct {
	ID              string
	Name            string
	Path            string
	Phase           feature.Phase
	Paused          bool
	Created         time.Time
	HasBrainstorm   bool
	WorkflowVersion int
	Summary         string
	Intent          string
	Approach        string
	OpenItems       string
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
		if doc.Metadata != nil {
			summary.WorkflowVersion = doc.Metadata.WorkflowVersion
		}
		summary.Summary = doc.SummaryText()

		// extract problem section as intent
		summary.Intent = doc.IntentText("PROBLEM")
		if summary.Summary == "" {
			summary.Summary = summary.Intent
		}

		// extract open questions
		if section := doc.GetSection("OPEN-QUESTIONS"); section != nil {
			summary.OpenItems = document.ExtractFirstParagraph(section)
		}
		if summary.WorkflowVersion == document.WorkflowVersionV3 {
			summary.Intent = summary.Summary
			summary.Approach = document.ExtractFirstParagraph(doc.GetSection("ACCEPTED PLAN"))
			if f.Phase == feature.PhaseComplete {
				summary.OpenItems = "none"
			} else {
				summary.OpenItems = "see SPEC.md"
			}
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
			if summary.Intent == "" {
				summary.Intent = doc.IntentText("USER THESIS")
			}
			if summary.OpenItems == "" {
				if section := doc.GetSection("QUESTIONS"); section != nil {
					summary.OpenItems = document.ExtractFirstParagraph(section)
				}
			}
		}
	}

	// V3 keeps its accepted plan inside SPEC.md. Legacy workflows may use PLAN.md.
	if summary.WorkflowVersion != document.WorkflowVersionV3 {
		planPath := filepath.Join(f.Path, "PLAN.md")
		if doc, err := document.ParseFile(planPath, document.TypePlan); err == nil {
			summary.Approach = doc.IntentText("APPROACH")
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
		if summary.WorkflowVersion == document.WorkflowVersionV3 {
			summary.Approach = "(accepted plan not captured yet)"
		} else {
			summary.Approach = "(see PLAN.md)"
		}
	}
	if summary.OpenItems == "" {
		summary.OpenItems = "none"
	}

	return summary
}

func sortFeatureSummaries(summaries []FeatureSummary) {
	sort.SliceStable(summaries, func(i, j int) bool {
		if summaries[i].ID != summaries[j].ID {
			return summaries[i].ID < summaries[j].ID
		}
		return summaries[i].Name < summaries[j].Name
	})
}
