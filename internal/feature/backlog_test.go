package feature

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestIsBacklogItem(t *testing.T) {
	featurePath := t.TempDir()
	if err := os.WriteFile(filepath.Join(featurePath, "BRAINSTORM.md"), []byte("# BRAINSTORM\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	feat := Feature{
		Path:   featurePath,
		Phase:  PhaseBrainstorm,
		Paused: true,
	}

	if !IsBacklogItem(feat) {
		t.Fatal("expected paused brainstorm feature with BRAINSTORM.md to be backlog")
	}
}

func TestFindActiveFeatureWithState_SkipsBacklogItems(t *testing.T) {
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
