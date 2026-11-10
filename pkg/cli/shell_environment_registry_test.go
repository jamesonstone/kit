package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

// shellEnvironmentRoute is the contract line that routes shell-dependent commands to its rule.
const shellEnvironmentRoute = "- Shell-dependent commands or environment mismatches: `shell-environment.md`.\n"

// shellEnvironmentRule returns the embedded shell environment rule.
func shellEnvironmentRule(t *testing.T) registryRuleset {
	t.Helper()
	registry, err := embeddedRulesetRegistry(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, rule := range registry {
		if rule.Slug == "shell-environment" {
			return rule
		}
	}
	t.Fatal("shell-environment is not shipped")
	return registryRuleset{}
}

// assertShellEnvironmentInstalled verifies the managed rule and its routing in every instruction file.
func assertShellEnvironmentInstalled(t *testing.T, root string, rule registryRuleset) {
	t.Helper()
	if got := readFile(t, filepath.Join(root, rulesetTarget(rule.Slug))); got != rule.Content {
		t.Fatal("installed shell-environment differs from the embedded rule")
	}
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	artifact, ok := cfg.RegistryArtifact(rulesetKind, rule.Slug)
	if !ok || artifact.State != registryArtifactStateManaged || artifact.InstalledHash != rule.NormalizedHash {
		t.Fatalf("shell-environment registry state = %#v", artifact)
	}
	for _, path := range instructionFiles(cfg) {
		content := readFile(t, filepath.Join(root, path))
		if !strings.Contains(content, templates.UniversalContractBlock()) || !strings.Contains(content, shellEnvironmentRoute) {
			t.Errorf("%s lacks current shell-environment routing", path)
		}
	}
}

// TestShellEnvironmentInstallsByDefaultAcrossGenerations verifies kit init installs the rule and
// kit reconcile adds it to existing projects of every generation, idempotently.
func TestShellEnvironmentInstallsByDefaultAcrossGenerations(t *testing.T) {
	setupMigrationEnvironment(t)
	rule := shellEnvironmentRule(t)
	if rule.Metadata.RegistryScope != rulesetRegistryScopeDownstream {
		t.Fatal("shell-environment must be a default downstream rule")
	}
	t.Run("init", func(t *testing.T) {
		assertShellEnvironmentInstalled(t, freshInitProject(t), rule)
	})
	t.Run("reconcile current without newly shipped rule", func(t *testing.T) {
		root := freshInitProject(t)
		if err := os.Remove(filepath.Join(root, rulesetTarget(rule.Slug))); err != nil {
			t.Fatal(err)
		}
		cfg := loadMigratedConfig(t, root)
		var artifacts []config.RegistryArtifact
		for _, artifact := range cfg.Registry.Artifacts {
			if artifact.Slug != rule.Slug {
				artifacts = append(artifacts, artifact)
			}
		}
		cfg.Registry.Artifacts = artifacts
		if err := config.Save(root, cfg); err != nil {
			t.Fatal(err)
		}
		for _, path := range instructionFiles(cfg) {
			file := filepath.Join(root, path)
			writeFile(t, file, strings.ReplaceAll(readFile(t, file), shellEnvironmentRoute, ""))
		}
		commitMigrationFixture(t, root)
		migrateProject(t, root, false)
		assertShellEnvironmentInstalled(t, root, rule)
		if changes := plannedChanges(t, root); len(changes) != 0 {
			t.Fatalf("second reconcile plans changes: %v", changes)
		}
	})
	for _, generation := range migrationFixtures {
		t.Run("reconcile "+generation, func(t *testing.T) {
			root := copyMigrationFixture(t, generation, nil)
			migrateProject(t, root, false)
			assertShellEnvironmentInstalled(t, root, rule)
			if changes := plannedChanges(t, root); len(changes) != 0 {
				t.Fatalf("second reconcile plans changes: %v", changes)
			}
		})
	}
}

func TestShellEnvironmentPreservesProjectCustomizations(t *testing.T) {
	setupMigrationEnvironment(t)
	rule := shellEnvironmentRule(t)
	root := freshInitProject(t)
	custom := rule.Content + "\nProject extension: use the documented container for builds.\n"
	writeFile(t, filepath.Join(root, rulesetTarget(rule.Slug)), custom)
	reference := "# Project environment\n\nUse the existing container workflow.\n"
	writeFile(t, filepath.Join(root, templates.TestingReferencePath), reference)
	cfg := loadMigratedConfig(t, root)
	for _, path := range instructionFiles(cfg) {
		file := filepath.Join(root, path)
		writeFile(t, file, readFile(t, file)+"\nProject environment: see docs/environment.md.\n")
	}
	commitMigrationFixture(t, root)
	migrateProject(t, root, false)
	if got := readFile(t, filepath.Join(root, rulesetTarget(rule.Slug))); got != custom {
		t.Fatal("reconcile overwrote a custom shell-environment rule")
	}
	if got := readFile(t, filepath.Join(root, templates.TestingReferencePath)); got != reference {
		t.Fatal("reconcile overwrote the project environment reference")
	}
	for _, path := range instructionFiles(cfg) {
		if !strings.Contains(readFile(t, filepath.Join(root, path)), "Project environment: see docs/environment.md.") {
			t.Errorf("reconcile lost project-owned guidance in %s", path)
		}
	}
	if changes := plannedChanges(t, root); len(changes) != 0 {
		t.Fatalf("second reconcile plans changes: %v", changes)
	}
}

func TestShellEnvironmentGuidanceIsPortable(t *testing.T) {
	rule := shellEnvironmentRule(t)
	for _, privateMarker := range []string{"/Users/", "/home/", "jamesonstone", "/bin/zsh"} {
		if strings.Contains(rule.Content, privateMarker) {
			t.Errorf("shared shell rule contains machine-specific marker %q", privateMarker)
		}
	}
	for _, marker := range []string{"Bash login startup does not automatically read `.bashrc`", "fish, PowerShell", "ZDOTDIR", "noninteractive CI setup", "prompt hooks"} {
		if !strings.Contains(rule.Content, marker) {
			t.Errorf("shared shell rule lost compatibility guidance %q", marker)
		}
	}
	if !strings.Contains(templates.TestingReference, "Record the project command environment here") {
		t.Fatal("new projects lack a place to record their command environment")
	}
}
