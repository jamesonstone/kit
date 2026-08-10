package agentcli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/registry"
)

func TestInitJSONAndDryRunAreWriteFree(t *testing.T) {
	registryRoot := writeEmptyRegistry(t)
	for _, mode := range []string{"--json", "--dry-run"} {
		t.Run(strings.TrimPrefix(mode, "--"), func(t *testing.T) {
			project := t.TempDir()
			configHome := t.TempDir()
			t.Setenv("KIT_CONFIG_HOME", configHome)
			withWorkingDirectory(t, project)
			command := NewRoot()
			command.SetOut(&bytes.Buffer{})
			command.SetArgs([]string{"init", mode, "--registry-path", registryRoot})
			if err := command.Execute(); err != nil {
				t.Fatal(err)
			}
			for _, path := range []string{".kit.yaml", ".env", "Makefile", "docs"} {
				if _, err := os.Stat(filepath.Join(project, path)); !os.IsNotExist(err) {
					t.Errorf("%s init wrote %s: %v", mode, path, err)
				}
			}
			if _, err := os.Stat(filepath.Join(configHome, "kit", ".kit.yaml")); !os.IsNotExist(err) {
				t.Errorf("%s init wrote user config: %v", mode, err)
			}
		})
	}
}

func TestInitOutputOnlyAppliesAndExplicitCopyUsesSamePrompt(t *testing.T) {
	registryRoot := writeEmptyRegistry(t)
	project := t.TempDir()
	t.Setenv("KIT_CONFIG_HOME", t.TempDir())
	withWorkingDirectory(t, project)
	previous := clipboardCopyFunc
	var copied string
	clipboardCopyFunc = func(content string) error {
		copied = content
		return nil
	}
	t.Cleanup(func() { clipboardCopyFunc = previous })
	output := &bytes.Buffer{}
	command := NewRoot()
	command.SetOut(output)
	command.SetArgs([]string{"init", "--output-only", "--copy", "--registry-path", registryRoot})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if output.String() == "" || copied != output.String() {
		t.Fatalf("copied prompt differs from raw output\ncopy=%q\nout=%q", copied, output)
	}
	if !strings.Contains(output.String(), "kit contract resolve --workflow repository-bootstrap --json") {
		t.Fatal("raw prompt is missing repository-bootstrap resolution")
	}
	for _, path := range []string{".kit.yaml", ".env", "Makefile", "docs/PROJECT_PROGRESS_SUMMARY.md"} {
		if _, err := os.Stat(filepath.Join(project, path)); err != nil {
			t.Errorf("output-only init did not create %s: %v", path, err)
		}
	}
	config, _, err := registry.LoadProject(project)
	if err != nil {
		t.Fatal(err)
	}
	if config.Registry.Source.Path != registryRoot || config.Registry.Source.Repo != "" {
		t.Fatalf("local registry source = %#v", config.Registry.Source)
	}
}

func TestInitSchemaV1FailsClosedBeforeAnyBootstrapWrite(t *testing.T) {
	project := t.TempDir()
	configHome := t.TempDir()
	t.Setenv("KIT_CONFIG_HOME", configHome)
	withWorkingDirectory(t, project)
	if err := os.WriteFile(filepath.Join(project, ".kit.yaml"), []byte("schema_version: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	command := NewRoot()
	command.SetOut(&bytes.Buffer{})
	command.SetArgs([]string{"init", "--output-only"})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "kit reconcile --json --diff") {
		t.Fatalf("schema-v1 init error = %v", err)
	}
	for _, path := range []string{".env", "Makefile", "docs"} {
		if _, err := os.Stat(filepath.Join(project, path)); !os.IsNotExist(err) {
			t.Errorf("schema-v1 init wrote %s: %v", path, err)
		}
	}
	if _, err := os.Stat(filepath.Join(configHome, "kit", ".kit.yaml")); !os.IsNotExist(err) {
		t.Errorf("schema-v1 init wrote user config: %v", err)
	}
}

func writeEmptyRegistry(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, "registry")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "catalog.yaml"), []byte("schema_version: 1\nartifacts: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func withWorkingDirectory(t *testing.T, path string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(path); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
}
