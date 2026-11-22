package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUserConfigMissingIsRepoOnly(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	root := t.TempDir()
	writeTestConfig(t, root, "schema_version: 2\nspecs_dir: docs/specs\nconstitution_path: docs/CONSTITUTION.md\nfeature_naming:\n  numeric_width: 4\n  separator: \"-\"\n")
	cfg, inspection, err := LoadWithInspection(root)
	if err != nil || inspection.HasErrors() || cfg.WorkLedger.Enabled() {
		t.Fatalf("missing user file: %v %v %#v", err, inspection, cfg)
	}
}

func TestUserConfigMergePrecedence(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	userPath := filepath.Join(home, ".kit", ".kit.yaml")
	if err := os.MkdirAll(filepath.Dir(userPath), 0o755); err != nil {
		t.Fatal(err)
	}
	user := "schema_version: 2\nspecs_dir: from-user\nshell:\n  command:\n    - zsh\nwork_ledger:\n  provider: plane\n  owner_label: agent:local\n  plane:\n    workspace: from-user\n    project: USER\n"
	if err := os.WriteFile(userPath, []byte(user), 0o644); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	writeTestConfig(t, root, "schema_version: 2\nspecs_dir: from-repo\nconstitution_path: docs/CONSTITUTION.md\nfeature_naming:\n  numeric_width: 4\n  separator: \"-\"\nwork_ledger:\n  plane:\n    project: REPO\n")
	cfg, inspection, err := LoadWithInspection(root)
	if err != nil || inspection.HasErrors() {
		t.Fatalf("merge: %v %v", err, inspection)
	}
	if cfg.SpecsDir != "from-repo" || len(cfg.Shell.Command) != 1 || cfg.Shell.Command[0] != "zsh" {
		t.Fatalf("scalar merge: %#v %#v", cfg.SpecsDir, cfg.Shell)
	}
	if cfg.WorkLedger.Provider != "plane" || cfg.WorkLedger.OwnerLabel != "agent:local" {
		t.Fatalf("user ledger kept: %#v", cfg.WorkLedger)
	}
	if cfg.WorkLedger.Plane.Workspace != "from-user" || cfg.WorkLedger.Plane.Project != "REPO" {
		t.Fatalf("nested override: %#v", cfg.WorkLedger.Plane)
	}
	repoOnly, err := RepositoryWorkLedger(root)
	if err != nil || repoOnly.Provider != "" || repoOnly.Plane.Project != "REPO" || repoOnly.Plane.Workspace != "" {
		t.Fatalf("repository ledger: %v %#v", err, repoOnly)
	}
}

func TestWorkLedgerFromUserFileOnly(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	userPath := filepath.Join(home, ".kit", ".kit.yaml")
	if err := os.MkdirAll(filepath.Dir(userPath), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "work_ledger:\n  provider: plane\n  plane:\n    workspace: example-workspace\n    project: PROJ\n"
	if err := os.WriteFile(userPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	writeTestConfig(t, root, "schema_version: 2\nspecs_dir: docs/specs\nconstitution_path: docs/CONSTITUTION.md\nfeature_naming:\n  numeric_width: 4\n  separator: \"-\"\n")
	cfg, inspection, err := LoadWithInspection(root)
	if err != nil || inspection.HasErrors() || !cfg.WorkLedger.Enabled() || cfg.WorkLedger.Plane.Workspace != "example-workspace" {
		t.Fatalf("user ledger: %v %v %#v", err, inspection, cfg)
	}
	if got, err := GlobalConfigPath(); err != nil || got != filepath.Join(home, ".config", "kit", ConfigFileName) {
		t.Fatalf("path %q %v", got, err)
	}
}

func TestGlobalVersionsDoNotMarkRepositoryAsMigrated(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	global, err := GlobalConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(global), 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestConfig(t, filepath.Dir(global), "schema_version: 2\ninstruction_scaffold_version: 4\nspecs_dir: docs/custom\n")
	root := t.TempDir()
	writeTestConfig(t, root, "{}\n")
	cfg, inspection, err := LoadWithInspection(root)
	if err != nil || inspection.SchemaState != SchemaStateMissing || cfg.InstructionScaffoldVersion != 0 || cfg.SpecsDir != "docs/custom" {
		t.Fatalf("global defaults hid repository migration state: cfg=%#v inspection=%+v err=%v", cfg, inspection, err)
	}
}

func TestXDGGlobalConfigAndLegacyPathsMergeBeforeRepository(t *testing.T) {
	home := t.TempDir()
	xdg := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", xdg)

	writeGlobalTestConfig(t, filepath.Join(home, ".config", "kit"), `shell:
  command: [bash, -lc]
work_ledger:
  provider: plane
  owner_label: agent:home-config
  plane:
    workspace: config-workspace
    project: OLDEST
`)
	writeGlobalTestConfig(t, filepath.Join(home, ".kit"), `shell:
  command: [zsh, -lic]
work_ledger:
  plane:
    workspace: legacy-workspace
    project: LEGACY
`)
	writeGlobalTestConfig(t, filepath.Join(xdg, "kit"), `work_ledger:
  plane:
    project: CANONICAL
`)
	global, found, err := LoadGlobal()
	if err != nil || !found {
		t.Fatalf("global config: found %v, err %v", found, err)
	}
	if got := global.WorkLedger.Plane.Project; got != "CANONICAL" {
		t.Fatalf("canonical project = %q", got)
	}
	root := t.TempDir()
	writeTestConfig(t, root, "schema_version: 2\nshell: null\nwork_ledger:\n  plane:\n    project: REPO\n")
	cfg, inspection, err := LoadWithInspection(root)
	if err != nil || inspection.HasErrors() {
		t.Fatalf("project config: %v, %v", err, inspection)
	}
	if cfg.WorkLedger.Provider != "plane" || cfg.WorkLedger.OwnerLabel != "agent:home-config" ||
		cfg.WorkLedger.Plane.Workspace != "legacy-workspace" || cfg.WorkLedger.Plane.Project != "REPO" ||
		len(cfg.Shell.Command) != 0 {
		t.Fatalf("merged project config: ledger %#v, shell %#v", cfg.WorkLedger, cfg.Shell)
	}
}

func writeTestConfig(t *testing.T, root, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, ConfigFileName), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestProjectYAMLMergeOverridesGlobalDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	writeGlobalTestConfig(t, filepath.Join(xdg, "kit"), "shell:\n  command: [global-shell]\n")
	for name, body := range map[string]string{
		"root merge":   "local: &local\n  shell:\n    command: [repo-shell]\n<<: *local\n",
		"nested merge": "local: &local\n  command: [repo-shell]\nshell:\n  <<: *local\n",
	} {
		t.Run(name, func(t *testing.T) {
			cfg, inspection, err := ParseWithGlobalDefaults([]byte("schema_version: 2\n" + body))
			if err != nil || inspection.HasErrors() || len(cfg.Shell.Command) != 1 || cfg.Shell.Command[0] != "repo-shell" {
				t.Fatalf("expected repository command: cfg=%#v, inspection=%#v, error=%v", cfg, inspection, err)
			}
		})
	}
}
