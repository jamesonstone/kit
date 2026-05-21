package runstore

import (
	"testing"
	"time"

	"github.com/jamesonstone/kit/internal/verify"
)

func TestLatestForFeatureUsesCompletionTime(t *testing.T) {
	projectRoot := t.TempDir()
	featureRef := verify.FeatureRef{ID: "0001", Slug: "fixture", DirName: "0001-fixture"}
	start := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	full := verify.Run{
		SchemaVersion: verify.SchemaVersion,
		RunID:         "full",
		Feature:       featureRef,
		TaskIDs:       []string{"T001", "T002"},
		Status:        verify.RunStatusPass,
		StartedAt:     start,
		EndedAt:       start.Add(10 * time.Second),
	}
	if err := Write(projectRoot, &full); err != nil {
		t.Fatalf("Write(full) error = %v", err)
	}
	nested := verify.Run{
		SchemaVersion: verify.SchemaVersion,
		RunID:         "nested",
		Feature:       featureRef,
		TaskIDs:       []string{"T001"},
		Status:        verify.RunStatusPass,
		StartedAt:     start.Add(2 * time.Second),
		EndedAt:       start.Add(3 * time.Second),
	}
	if err := Write(projectRoot, &nested); err != nil {
		t.Fatalf("Write(nested) error = %v", err)
	}

	latest, ok, err := LatestForFeature(projectRoot, featureRef.DirName)
	if err != nil {
		t.Fatalf("LatestForFeature() error = %v", err)
	}
	if !ok {
		t.Fatal("LatestForFeature() ok = false")
	}
	if latest.RunID != "full" {
		t.Fatalf("latest RunID = %q, want full", latest.RunID)
	}
}
