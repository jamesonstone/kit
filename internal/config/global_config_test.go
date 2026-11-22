package config

import (
	"os"
	"path/filepath"
	"strings"
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

func TestGlobalConfigPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	xdg := t.TempDir()
	for _, tc := range []struct{ name, value, dir string }{
		{"empty", "", filepath.Join(home, ".config")},
		{"relative", "relative/config", filepath.Join(home, ".config")},
		{"absolute", xdg, xdg},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", tc.value)
			want := filepath.Join(tc.dir, "kit", ConfigFileName)
			if got, err := GlobalConfigPath(); err != nil || got != want {
				t.Fatalf("GlobalConfigPath() = %q, %v, want %q", got, err, want)
			}
		})
	}
}

func TestGlobalConfigPathsDeduplicateDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	paths, err := globalConfigPaths()
	if err != nil || len(paths) != 2 ||
		paths[0] != filepath.Join(home, ".config", "kit", ConfigFileName) ||
		paths[1] != filepath.Join(home, ".kit", ConfigFileName) {
		t.Fatalf("globalConfigPaths() = %v, %v", paths, err)
	}
}

func TestLoadGlobalReturnsDefaultWhenAbsent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

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
	t.Setenv("HOME", t.TempDir())
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)

	defaults := Default()
	defaults.InstructionScaffoldVersion = CurrentInstructionScaffoldVersion

	path, changed, err := PopulateGlobalConfig(defaults)
	if err != nil {
		t.Fatalf("PopulateGlobalConfig() error = %v", err)
	}
	if !changed {
		t.Fatal("PopulateGlobalConfig() changed = false, want true")
	}
	wantPath := filepath.Join(xdg, "kit", ConfigFileName)
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

func TestGlobalPreferencesUseLegacyFallbacks(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)

	const legacy = "usage:\n  enabled: false\ngithub:\n  default_assignees: [legacy-owner]\n"
	legacyDir := filepath.Join(home, ".config", "kit")
	writeGlobalTestConfig(t, legacyDir, legacy)

	x := filepath.Join(xdg, "kit")
	writeGlobalTestConfig(t, x, "usage: null\n")
	enabled, err := LoadGlobalUsageEnabled()
	if err != nil || enabled {
		t.Fatalf("legacy opt-out = %v, err = %v", enabled, err)
	}
	assignees, err := LoadGlobalDefaultAssignees()
	if err != nil || assignees == nil || len(*assignees) != 1 || (*assignees)[0] != "legacy-owner" {
		t.Fatalf("absent canonical assignees lost legacy settings: %v, %v", assignees, err)
	}

	writeGlobalTestConfig(t, x, "usage:\n  enabled: true\ngithub:\n  default_assignees: []\n")
	enabled, err = LoadGlobalUsageEnabled()
	if err != nil || !enabled {
		t.Fatalf("explicit canonical usage override = %v, err = %v", enabled, err)
	}
	assignees, err = LoadGlobalDefaultAssignees()
	if err != nil || assignees == nil || len(*assignees) != 0 {
		t.Fatalf("empty list should mask legacy assignees: %v, %v", assignees, err)
	}

	writeGlobalTestConfig(t, x, "github:\n  default_assignees: null\n")
	assignees, err = LoadGlobalDefaultAssignees()
	if err != nil || assignees != nil {
		t.Fatalf("explicit null should mask legacy assignees: %v, %v", assignees, err)
	}
	got, err := os.ReadFile(filepath.Join(legacyDir, ConfigFileName))
	if err != nil || string(got) != legacy {
		t.Fatalf("legacy config changed: %v", err)
	}
}

func TestGlobalConfigPopulationPreservesLegacyKeys(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	legacyDir := filepath.Join(home, ".kit")
	const legacy = "specs_dir: inherited-specs\n"
	writeGlobalTestConfig(t, legacyDir, legacy)

	canonical := filepath.Join(xdg, "kit", ConfigFileName)
	path, create, update, err := PlanGlobalConfig(nil)
	if err != nil || path != canonical || !create || update {
		t.Fatalf("planned path %q, create %t, update %t, error %v", path, create, update, err)
	}
	if _, err := os.Stat(canonical); !os.IsNotExist(err) {
		t.Fatalf("plan wrote canonical config: %v", err)
	}
	path, changed, err := PopulateGlobalConfig(nil)
	if err != nil || path != canonical || !changed {
		t.Fatalf("populated %q, changed %t, error %v", path, changed, err)
	}
	body, err := os.ReadFile(canonical)
	if err != nil || strings.Contains(string(body), "specs_dir:") {
		t.Fatalf("canonical file masks inherited keys: %q, %v", body, err)
	}
	cfg, found, err := LoadGlobal()
	if err != nil || !found || cfg.SpecsDir != "inherited-specs" {
		t.Fatalf("lost inherited setting, found %t, err %v", found, err)
	}
	if got, err := os.ReadFile(filepath.Join(legacyDir, ConfigFileName)); err != nil || string(got) != legacy {
		t.Fatalf("legacy config changed: %v", err)
	}
	if _, _, update, err := PlanGlobalConfig(nil); err != nil || update {
		t.Fatalf("planned redundant write, update %t, err %v", update, err)
	}
}

func TestGlobalConfigPopulationPreservesMergedKeys(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	dir := filepath.Join(xdg, "kit")
	const body = "config_defaults: &defaults\n  specs_dir: inherited-custom\n<<: *defaults\n"
	writeGlobalTestConfig(t, dir, body)
	path, changed, err := PopulateGlobalConfig(nil)
	if err != nil || !changed {
		t.Fatalf("populate changed=%v err=%v", changed, err)
	}
	cfg, found, err := LoadGlobal()
	if err != nil || !found || cfg.SpecsDir != "inherited-custom" {
		t.Fatalf("expected YAML-merged settings to survive population: cfg=%#v, err=%v", cfg, err)
	}
	got, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(got), "&defaults") || !strings.Contains(string(got), "<<: *defaults") {
		t.Fatalf("expected original YAML merge to stay in place: %s, err=%v", got, err)
	}
}

func TestGlobalConfigWriteRefusesAnAliasOfFallback(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	fallback := filepath.Join(home, ".kit", ConfigFileName)
	const body = "usage:\n  enabled: false\n"
	writeGlobalTestConfig(t, filepath.Dir(fallback), body)
	path, err := GlobalConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(fallback, path); err != nil {
		t.Fatal(err)
	}
	if _, err := GlobalConfigWritePath(); err == nil || !strings.Contains(err.Error(), "aliases fallback") {
		t.Fatalf("expected an actionable alias error, got %v", err)
	}
	if _, _, err := PopulateGlobalConfig(nil); err == nil {
		t.Fatal("init would write through the alias")
	}
	got, err := os.ReadFile(fallback)
	if err != nil || string(got) != body {
		t.Fatalf("fallback changed: %s, err=%v", got, err)
	}
}

func TestGlobalConfigWriteRefusesDirectoryAliasBeforeCreation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	legacyDir := filepath.Join(home, ".kit")
	if err := os.MkdirAll(legacyDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(legacyDir, filepath.Join(xdg, "kit")); err != nil {
		t.Fatal(err)
	}
	if _, err := GlobalConfigWritePath(); err == nil || !strings.Contains(err.Error(), "aliases fallback") {
		t.Fatalf("expected an actionable directory alias error, got %v", err)
	}
	if _, _, err := PopulateGlobalConfig(nil); err == nil {
		t.Fatal("init would create a fallback file through the alias")
	}
	if _, err := os.Stat(filepath.Join(legacyDir, ConfigFileName)); !os.IsNotExist(err) {
		t.Fatalf("fallback should still be absent: %v", err)
	}
}

func writeGlobalTestConfig(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestConfig(t, dir, content)
}
