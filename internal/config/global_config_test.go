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

func TestGlobalConfigPathUsesDotKit(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := GlobalConfigPath()
	if err != nil {
		t.Fatalf("GlobalConfigPath() error = %v", err)
	}

	want := filepath.Join(home, ".kit", ConfigFileName)
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
	wantPath := filepath.Join(home, ".kit", ConfigFileName)
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

func TestLegacyGlobalConfigIsLimitedToExistingConsumers(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	legacyDir := filepath.Join(home, ".config", "kit")
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const legacy = `usage:
  enabled: false
github:
  default_assignees: [legacy-owner]
shell:
  command: [legacy-shell]
work_ledger:
  provider: plane
  plane:
    workspace: old-workspace
    project: OLD
`
	writeTestConfig(t, legacyDir, legacy)

	root := t.TempDir()
	writeTestConfig(t, root, "schema_version: 2\n")
	project, _, err := LoadWithInspection(root)
	if err != nil {
		t.Fatal(err)
	}
	if project.WorkLedger.Enabled() || len(project.Shell.Command) != 0 || !project.IsUsageEnabled() {
		t.Fatalf("legacy settings entered the project overlay: %#v", project)
	}
	enabled, err := LoadGlobalUsageEnabled()
	if err != nil || enabled {
		t.Fatalf("legacy opt-out = %v, err = %v", enabled, err)
	}
	assignees, err := LoadGlobalDefaultAssignees()
	if err != nil || assignees == nil || len(*assignees) != 1 || (*assignees)[0] != "legacy-owner" {
		t.Fatalf("legacy assignees = %v, err = %v", assignees, err)
	}

	globalDir := filepath.Join(home, ".kit")
	if err := os.MkdirAll(globalDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestConfig(t, globalDir, "usage:\n  enabled: true\n")
	assignees, err = LoadGlobalDefaultAssignees()
	if err != nil || assignees == nil || len(*assignees) != 1 || (*assignees)[0] != "legacy-owner" {
		t.Fatalf("missing canonical key lost legacy assignees: %v, %v", assignees, err)
	}
	enabled, err = LoadGlobalUsageEnabled()
	if err != nil || !enabled {
		t.Fatalf("canonical usage override = %v, err = %v", enabled, err)
	}
	writeTestConfig(t, globalDir, "github:\n  default_assignees: []\n")
	assignees, err = LoadGlobalDefaultAssignees()
	if err != nil || assignees == nil || len(*assignees) != 0 {
		t.Fatalf("empty canonical list should override the legacy fallback: %v, %v", assignees, err)
	}
	got, err := os.ReadFile(filepath.Join(legacyDir, ConfigFileName))
	if err != nil || string(got) != legacy {
		t.Fatalf("legacy config changed: %v", err)
	}
}
