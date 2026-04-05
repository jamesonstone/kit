package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
)

func TestRunMap_ProjectWideOutput(t *testing.T) {
	projectRoot := setupMapProject(t)
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWD)
	}()
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	out := &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(out)

	if err := runMap(cmd, nil); err != nil {
		t.Fatalf("runMap() error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "🗺️ Kit Map") {
		t.Fatalf("expected heading, got %q", got)
	}
	if !strings.Contains(got, "0001-alpha [phase: spec] [paused: no]") {
		t.Fatalf("expected feature summary, got %q", got)
	}
	if !strings.Contains(got, "SPEC.md builds on -> 0002-beta [resolved]") {
		t.Fatalf("expected relationship edge, got %q", got)
	}
}

func TestRunMap_FeatureScopedOutput(t *testing.T) {
	projectRoot := setupMapProject(t)
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWD)
	}()
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	out := &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(out)

	if err := runMap(cmd, []string{"beta"}); err != nil {
		t.Fatalf("runMap() error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "🗺️ Kit Map: 0002-beta") {
		t.Fatalf("expected feature heading, got %q", got)
	}
	if !strings.Contains(got, "Incoming Relationships") {
		t.Fatalf("expected incoming section, got %q", got)
	}
	if !strings.Contains(got, "0001-alpha via SPEC.md builds on -> 0002-beta [resolved]") {
		t.Fatalf("expected incoming relationship, got %q", got)
	}
}

func setupMapProject(t *testing.T) string {
	t.Helper()

	projectRoot := t.TempDir()
	if err := config.Save(projectRoot, config.Default()); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	writeMapProjectFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), "# CONSTITUTION\n")
	writeMapProjectFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), "# PROJECT PROGRESS SUMMARY\n")
	writeMapProjectFile(t, filepath.Join(projectRoot, "docs", "specs", "0001-alpha", "SPEC.md"), `# SPEC

## RELATIONSHIPS

- builds on: 0002-beta
`)
	writeMapProjectFile(t, filepath.Join(projectRoot, "docs", "specs", "0002-beta", "SPEC.md"), `# SPEC

## RELATIONSHIPS

none
`)

	return projectRoot
}

func writeMapProjectFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
