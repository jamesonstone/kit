package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindProjectRootOptionalReturnsRootWhenConfigExists(t *testing.T) {
	projectRoot := t.TempDir()
	nested := filepath.Join(projectRoot, "a", "b")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, ConfigFileName), []byte("goal_percentage: 95\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Chdir(nested)

	root, found, err := FindProjectRootOptional()
	if err != nil {
		t.Fatalf("FindProjectRootOptional() error = %v", err)
	}
	if !found {
		t.Fatal("FindProjectRootOptional() found = false, want true")
	}
	if root != projectRoot {
		t.Fatalf("FindProjectRootOptional() root = %q, want %q", root, projectRoot)
	}
}

func TestFindProjectRootOptionalReturnsNotFoundWithoutError(t *testing.T) {
	t.Chdir(t.TempDir())

	root, found, err := FindProjectRootOptional()
	if err != nil {
		t.Fatalf("FindProjectRootOptional() error = %v", err)
	}
	if found {
		t.Fatal("FindProjectRootOptional() found = true, want false")
	}
	if root != "" {
		t.Fatalf("FindProjectRootOptional() root = %q, want empty", root)
	}
}

func TestGlobalConfigPathUsesDotConfigKit(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := GlobalConfigPath()
	if err != nil {
		t.Fatalf("GlobalConfigPath() error = %v", err)
	}

	want := filepath.Join(home, ".config", "kit", ConfigFileName)
	if got != want {
		t.Fatalf("GlobalConfigPath() = %q, want %q", got, want)
	}
}

func TestLoadGlobalReturnsDefaultWhenAbsent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, found, err := LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error = %v", err)
	}
	if found {
		t.Fatal("LoadGlobal() found = true, want false")
	}
	if cfg == nil {
		t.Fatal("LoadGlobal() cfg = nil")
	}
}

func TestPopulateGlobalConfigCreatesDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	defaults := Default()
	defaults.InstructionScaffoldVersion = CurrentInstructionScaffoldVersion

	path, changed, err := PopulateGlobalConfig(defaults)
	if err != nil {
		t.Fatalf("PopulateGlobalConfig() error = %v", err)
	}
	if !changed {
		t.Fatal("PopulateGlobalConfig() changed = false, want true")
	}
	wantPath := filepath.Join(home, ".config", "kit", ConfigFileName)
	if path != wantPath {
		t.Fatalf("path = %q, want %q", path, wantPath)
	}

	cfg, found, err := LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error = %v", err)
	}
	if !found {
		t.Fatal("LoadGlobal() found = false, want true")
	}
	if cfg.InstructionScaffoldVersion != CurrentInstructionScaffoldVersion {
		t.Fatalf("InstructionScaffoldVersion = %d, want %d", cfg.InstructionScaffoldVersion, CurrentInstructionScaffoldVersion)
	}
}
