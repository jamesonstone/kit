package feature

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetFeatureStatusPrefersFrontMatterSummary(t *testing.T) {
	featureDir := filepath.Join(t.TempDir(), "0001-alpha")
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	spec := `---
kit_metadata_version: 1
artifact: spec
feature:
  id: "0001"
  slug: alpha
  dir: 0001-alpha
summary: Metadata summary.
---
# SPEC

## SUMMARY

Body summary.
`
	if err := os.WriteFile(filepath.Join(featureDir, "SPEC.md"), []byte(spec), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	status, err := GetFeatureStatus(&Feature{
		Number:  1,
		Slug:    "alpha",
		DirName: "0001-alpha",
		Path:    featureDir,
	})
	if err != nil {
		t.Fatalf("GetFeatureStatus() error = %v", err)
	}
	if status.Summary != "Metadata summary." {
		t.Fatalf("Summary = %q", status.Summary)
	}
}
