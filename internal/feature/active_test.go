package feature

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestFindActiveFeatureWithStateIgnoresLegacyPausedState(t *testing.T) {
	specsDir := t.TempDir()
	cfg := config.Default()
	cfg.SetFeaturePaused("0002-backlog-item", true)

	createFeatureDir(t, specsDir, "0001-active-feature", map[string]string{
		"SPEC.md": "# SPEC\n",
	})
	createFeatureDir(t, specsDir, "0002-backlog-item", map[string]string{
		"BRAINSTORM.md": "# BRAINSTORM\n",
	})

	active, err := FindActiveFeatureWithState(specsDir, cfg)
	if err != nil {
		t.Fatalf("FindActiveFeatureWithState() error = %v", err)
	}
	if active == nil {
		t.Fatal("expected active feature")
	}
	if active.DirName != "0002-backlog-item" || active.Paused {
		t.Fatalf("active = %#v, want newest feature with inert paused state", active)
	}
}

func TestListFeaturesWithStateIgnoresLegacyRemovedTombstone(t *testing.T) {
	specsDir := t.TempDir()
	cfg := config.Default()
	cfg.RemovedFeatures = []config.RemovedFeature{{
		Number: 1, Slug: "active-feature", DirName: "0001-active-feature", RemovedAt: "2026-01-01T00:00:00Z",
	}}
	createFeatureDir(t, specsDir, "0001-active-feature", map[string]string{
		"SPEC.md": "# SPEC\n",
	})

	features, err := ListFeaturesWithState(specsDir, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(features) != 1 || features[0].DirName != "0001-active-feature" {
		t.Fatalf("features = %#v, want tombstone ignored", features)
	}
}

func TestFindActiveFeatureWithState_SkipsCompletedFeatures(t *testing.T) {
	specsDir := t.TempDir()
	cfg := config.Default()

	createFeatureDir(t, specsDir, "0001-active-feature", map[string]string{
		"SPEC.md": "# SPEC\n",
	})
	createFeatureDir(t, specsDir, "0002-completed-feature", map[string]string{
		"TASKS.md": "- [x] done\n\n" + ReflectionCompleteMarker + "\n",
	})

	active, err := FindActiveFeatureWithState(specsDir, cfg)
	if err != nil {
		t.Fatalf("FindActiveFeatureWithState() error = %v", err)
	}
	if active == nil {
		t.Fatal("expected active feature")
	}
	if active.DirName != "0001-active-feature" {
		t.Fatalf("active.DirName = %q, want %q", active.DirName, "0001-active-feature")
	}
}

func createFeatureDir(t *testing.T, specsDir, dirName string, files map[string]string) {
	t.Helper()

	featurePath := filepath.Join(specsDir, dirName)
	if err := os.MkdirAll(featurePath, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(featurePath, name), []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile(%q) error = %v", name, err)
		}
	}
}
