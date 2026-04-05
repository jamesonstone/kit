package rollup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jamesonstone/kit/internal/feature"
)

func TestFormatSummaryTableCell_PreservesFullMeaningWithoutTruncation(t *testing.T) {
	summary := "- This summary should remain fully visible even when it is long enough to have been truncated before because the semantic meaning matters."

	got := formatSummaryTableCell(summary)
	if got != summary {
		t.Fatalf("formatSummaryTableCell() = %q, want %q", got, summary)
	}
}

func TestFormatSummaryTableCell_NormalizesWhitespaceAndEscapesPipes(t *testing.T) {
	summary := "first line\nsecond line | third line"

	got := formatSummaryTableCell(summary)
	want := `first line second line \| third line`
	if got != want {
		t.Fatalf("formatSummaryTableCell() = %q, want %q", got, want)
	}
}

func TestExtractFeatureSummary_PrefersSpecSummaryForTableAndProblemForIntent(t *testing.T) {
	t.Helper()

	tempDir := t.TempDir()
	featureDir := filepath.Join(tempDir, "0001-example-feature")
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	spec := `# SPEC

## SUMMARY

- Concise feature summary for the progress table.

## PROBLEM

- Detailed intent text that is longer and should stay in the feature summary section.

## OPEN-QUESTIONS

- None.
`
	if err := os.WriteFile(filepath.Join(featureDir, "SPEC.md"), []byte(spec), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got := extractFeatureSummary(feature.Feature{
		Number:    1,
		Slug:      "example-feature",
		DirName:   "0001-example-feature",
		Path:      featureDir,
		CreatedAt: time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC),
	}, filepath.Dir(featureDir))

	if got.Summary != "- Concise feature summary for the progress table." {
		t.Fatalf("Summary = %q", got.Summary)
	}
	if got.Intent != "- Detailed intent text that is longer and should stay in the feature summary section." {
		t.Fatalf("Intent = %q", got.Intent)
	}
}
