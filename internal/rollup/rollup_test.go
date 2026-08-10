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

func TestExtractFeatureSummary_PrefersFrontMatterSummaryAndIntent(t *testing.T) {
	tempDir := t.TempDir()
	featureDir := filepath.Join(tempDir, "0001-example-feature")
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	spec := `---
kit_metadata_version: 1
artifact: spec
feature:
  id: "0001"
  slug: example-feature
  dir: 0001-example-feature
summary: Metadata summary.
intent: Metadata intent.
---
# SPEC

## SUMMARY

Body summary.

## PROBLEM

Body problem.
`
	if err := os.WriteFile(filepath.Join(featureDir, "SPEC.md"), []byte(spec), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got := extractFeatureSummary(feature.Feature{
		Number:  1,
		Slug:    "example-feature",
		DirName: "0001-example-feature",
		Path:    featureDir,
	}, filepath.Dir(featureDir))

	if got.Summary != "Metadata summary." {
		t.Fatalf("Summary = %q", got.Summary)
	}
	if got.Intent != "Metadata intent." {
		t.Fatalf("Intent = %q", got.Intent)
	}
}

func TestExtractFeatureSummaryUsesV3PurposeAndAcceptedPlan(t *testing.T) {
	featureDir := filepath.Join(t.TempDir(), "0001-memory")
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	spec := `---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: complete
feature:
  id: "0001"
  slug: memory
  dir: 0001-memory
---
# SPEC

## PURPOSE

Preserve consequential rationale.

## ACCEPTED PLAN

Use native planning, implement, validate, and curate repository memory.
`
	if err := os.WriteFile(filepath.Join(featureDir, "SPEC.md"), []byte(spec), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got := extractFeatureSummary(feature.Feature{
		Number:  1,
		Slug:    "memory",
		DirName: "0001-memory",
		Path:    featureDir,
		Phase:   feature.PhaseComplete,
	}, filepath.Dir(featureDir))
	if got.WorkflowVersion != 3 || got.Summary != "Preserve consequential rationale." || got.Intent != got.Summary {
		t.Fatalf("V3 summary = %#v", got)
	}
	if got.Approach != "Use native planning, implement, validate, and curate repository memory." || got.OpenItems != "none" {
		t.Fatalf("V3 approach/open items = %#v", got)
	}
	content := generateContent([]FeatureSummary{got}, config.Default())
	if strings.Contains(content, "PLAN.md") || strings.Contains(content, "TASKS.md") {
		t.Fatalf("V3 rollup contains legacy pointers:\n%s", content)
	}
}

func TestGenerateContentWritesConcreteProjectIntent(t *testing.T) {
	content := generateContent(nil, config.Default())

	if !strings.Contains(content, "## PROJECT INTENT\n\nKit is a coding-agent-first repository contract and evidence harness") {
		t.Fatalf("expected concrete project intent, got:\n%s", content)
	}
	if strings.Contains(content, "TODO") {
		t.Fatalf("expected generated content to avoid TODO placeholders, got:\n%s", content)
	}
}
