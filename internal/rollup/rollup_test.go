package rollup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jamesonstone/kit/internal/config"
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

func TestGenerateContentWritesConcreteProjectIntent(t *testing.T) {
	content := generateContent(nil, config.Default())

	if !strings.Contains(content, "## PROJECT INTENT\n\nKit is a document-first workflow harness") {
		t.Fatalf("expected concrete project intent, got:\n%s", content)
	}
	if strings.Contains(content, "TODO") {
		t.Fatalf("expected generated content to avoid TODO placeholders, got:\n%s", content)
	}
}

func TestGenerateIncludesRemovedFeatureTombstones(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	cfg.RemovedFeatures = []config.RemovedFeature{
		{
			Number:    1,
			Slug:      "alpha",
			DirName:   "0001-alpha",
			CreatedAt: "2026-04-05T00:00:00Z",
			RemovedAt: "2026-05-06T12:00:00Z",
		},
	}
	notesDir := filepath.Join(projectRoot, "docs", "notes", "0001-alpha")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if err := Generate(projectRoot, cfg); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, err := os.ReadFile(cfg.ProgressSummaryPath(projectRoot))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(content)
	checks := []string{
		"| 0001 | alpha | `docs/specs/0001-alpha` | removed | no | 2026-04-05 | Removed by kit rm on 2026-05-06. |",
		"- **STATUS**: removed",
		"- **REMOVED AT**: 2026-05-06",
		"- **POINTERS**: removed; original docs path was `docs/specs/0001-alpha`; retained notes at `docs/notes/0001-alpha`",
	}
	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Fatalf("expected rollup to contain %q, got:\n%s", check, text)
		}
	}
}
