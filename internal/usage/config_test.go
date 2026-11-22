package usage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestUsageWritesCanonicalConfigAndKeepsExistingStorage(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	legacyDir := filepath.Join(home, ".config", "kit")
	legacyPath := filepath.Join(legacyDir, config.ConfigFileName)
	if err := config.UpdateUsageEnabled(legacyPath, false); err != nil {
		t.Fatal(err)
	}
	projectRoot := t.TempDir()
	if err := config.Save(projectRoot, config.Default()); err != nil {
		t.Fatal(err)
	}

	settings, err := EffectiveSettings(projectRoot)
	if err != nil || settings.Enabled || settings.ProjectState != "suppressed-by-global" {
		t.Fatalf("legacy global opt-out: %#v, %v", settings, err)
	}
	path, err := SetEnabled("global", "", true)
	wantPath := filepath.Join(xdg, "kit", config.ConfigFileName)
	if err != nil || path != wantPath {
		t.Fatalf("SetEnabled global wrote %q, want %q; err = %v", path, wantPath, err)
	}
	settings, err = EffectiveSettings(projectRoot)
	if err != nil || !settings.Enabled || settings.ProjectState != "default-enabled" {
		t.Fatalf("global enable is not a project setting: %#v, %v", settings, err)
	}
	if _, err := SetEnabled("project", projectRoot, false); err != nil {
		t.Fatal(err)
	}
	settings, err = EffectiveSettings(projectRoot)
	if err != nil || settings.Enabled || settings.ProjectState != "disabled" {
		t.Fatalf("project opt-out: %#v, %v", settings, err)
	}
	if old, err := os.ReadFile(legacyPath); err != nil || string(old) != "usage:\n  enabled: false\n" {
		t.Fatalf("legacy usage settings changed: %v", err)
	}
	if got, err := Directory(); err != nil || got != filepath.Join(legacyDir, "usage") {
		t.Fatalf("usage records moved to %q, err = %v", got, err)
	}
}
