package cli

import (
	"bytes"
	"io"
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
	for _, check := range []string{"AGENTS.md", "docs/agents/README.md", "docs/references/README.md"} {
		if !strings.Contains(got, check) {
			t.Fatalf("expected project map to contain %q, got %q", check, got)
		}
	}
	if !strings.Contains(got, "Feature Doc Key") {
		t.Fatalf("expected feature doc key, got %q", got)
	}
	if !strings.Contains(got, "┌") || !strings.Contains(got, "docs: B○ S● P○ T○ A○") {
		t.Fatalf("expected graphical feature card, got %q", got)
	}
	if !strings.Contains(got, "SPEC.md builds on ▶ 0002-beta") {
		t.Fatalf("expected relationship edge, got %q", got)
	}
	if !strings.Contains(got, "SPEC.md dependency docs/agents/RLM.md") || !strings.Contains(got, "[stale]") {
		t.Fatalf("expected dependency links, got %q", got)
	}
	if !strings.Contains(got, "Warnings") || !strings.Contains(got, `skipped invalid RELATIONSHIPS line "- follows: 0003-gamma"`) {
		t.Fatalf("expected warning output, got %q", got)
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
	if !strings.Contains(got, "docs/agents/README.md") {
		t.Fatalf("expected feature map to contain docs/agents/README.md, got %q", got)
	}
	if !strings.Contains(got, "Feature Focus") || !strings.Contains(got, "Incoming Relationships") {
		t.Fatalf("expected incoming section, got %q", got)
	}
	if !strings.Contains(got, "0001-alpha SPEC.md builds on ▶ 0002-beta") {
		t.Fatalf("expected incoming relationship, got %q", got)
	}
	if !strings.Contains(got, "Dependency Links") {
		t.Fatalf("expected dependency links section, got %q", got)
	}
}

func TestRunMap_ProjectWideOutput_UsesANSIColorWhenTerminalEnabled(t *testing.T) {
	previousCheck := terminalWriterCheck
	terminalWriterCheck = func(_ io.Writer) bool { return true }
	defer func() { terminalWriterCheck = previousCheck }()

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
	if !strings.Contains(got, "\033[") {
		t.Fatalf("expected ANSI color sequences in terminal output, got %q", got)
	}
	if !strings.Contains(got, "Feature Graph") {
		t.Fatalf("expected section heading in colored output, got %q", got)
	}
	if !strings.Contains(got, "0001-alpha") || !strings.Contains(got, "SPEC.md") || !strings.Contains(got, "builds on") {
		t.Fatalf("expected colored feature card and edge content, got %q", got)
	}
}

func setupMapProject(t *testing.T) string {
	t.Helper()

	projectRoot := t.TempDir()
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	writeMapProjectFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), "# CONSTITUTION\n")
	writeMapProjectFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), "# PROJECT PROGRESS SUMMARY\n")
	writeMapProjectFile(t, filepath.Join(projectRoot, "AGENTS.md"), "# AGENTS\n")
	writeMapProjectFile(t, filepath.Join(projectRoot, "CLAUDE.md"), "# CLAUDE\n")
	writeMapProjectFile(t, filepath.Join(projectRoot, ".github", "copilot-instructions.md"), "# COPILOT\n")
	writeMapProjectFile(t, filepath.Join(projectRoot, "docs", "agents", "README.md"), "# Agents Docs\n")
	writeMapProjectFile(t, filepath.Join(projectRoot, "docs", "references", "README.md"), "# References\n")
	writeMapProjectFile(t, filepath.Join(projectRoot, "docs", "specs", "0001-alpha", "SPEC.md"), `# SPEC

## RELATIONSHIPS

- builds on: `+"`0002-beta`"+`
- follows: 0003-gamma

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| docs/agents/RLM.md | doc | docs/agents/RLM.md | context routing | active |
| old-context.md | doc | docs/old-context.md | legacy context | stale |
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
