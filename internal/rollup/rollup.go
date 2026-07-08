package rollup

import (
	"fmt"
	"path/filepath"
	"sort"
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

	specPath := filepath.Join(f.Path, "SPEC.md")
	if doc, err := document.ParseFile(specPath, document.TypeSpec); err == nil {
		summary.Summary = doc.SummaryText()

		summary.Intent = doc.IntentText("PROBLEM")
		if summary.Summary == "" {
			summary.Summary = summary.Intent
		}

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

	planPath := filepath.Join(f.Path, "PLAN.md")
	if doc, err := document.ParseFile(planPath, document.TypePlan); err == nil {
		summary.Approach = doc.IntentText("APPROACH")
	}

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
