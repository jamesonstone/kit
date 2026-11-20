package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUserConfigMissingIsRepoOnly(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
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
	if got, err := UserKitConfigPath(); err != nil || got != userPath {
		t.Fatalf("path %q %v", got, err)
	}
}

func writeTestConfig(t *testing.T, root, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, ConfigFileName), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
