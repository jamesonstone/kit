package cli

import (
	"bytes"
	"encoding/json"
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
	if !strings.Contains(got, "SPEC.md reference docs/agents/RLM.md") || !strings.Contains(got, "[informs, skip, stale]") {
		t.Fatalf("expected reference links, got %q", got)
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
	if !strings.Contains(got, "Reference Links") {
		t.Fatalf("expected reference links section, got %q", got)
	}
}

func TestRunMap_FeatureContextOutput(t *testing.T) {
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
	writeMapProjectFile(t, filepath.Join(projectRoot, "docs", "specs", "0001-alpha", "PLAN.md"), `---
kit_metadata_version: 1
artifact: plan
feature:
  id: "0001"
  slug: alpha
  dir: 0001-alpha
references:
  - name: RLM duplicate
    type: doc
    target: docs/agents/RLM.md
    relation: informs
    read_policy: must
    used_for: implementation context routing
    status: active
---
# PLAN
`)

	out := &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(out)
	restoreMapContext := mapContext
	restoreMapJSON := mapJSON
	mapContext = true
	mapJSON = false
	defer func() {
		mapContext = restoreMapContext
		mapJSON = restoreMapJSON
	}()

	if err := runMap(cmd, []string{"alpha"}); err != nil {
		t.Fatalf("runMap() error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Kit Context Plan: 0001-alpha") {
		t.Fatalf("expected context heading, got %q", got)
	}
	if !strings.Contains(got, "MUST") || !strings.Contains(got, "read `docs/agents/RLM.md`") {
		t.Fatalf("expected must-read plan after duplicate policy ranking, got %q", got)
	}
	if strings.Count(got, "read `docs/agents/RLM.md`") != 1 {
		t.Fatalf("expected duplicate target to be merged once, got %q", got)
	}
	if strings.Contains(got, "# Agents Docs") {
		t.Fatalf("expected pointer-only context plan without inlined file contents, got %q", got)
	}
}

func TestRunMap_JSONOutput(t *testing.T) {
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
	restoreMapContext := mapContext
	restoreMapJSON := mapJSON
	mapContext = true
	mapJSON = true
	defer func() {
		mapContext = restoreMapContext
		mapJSON = restoreMapJSON
	}()

	if err := runMap(cmd, []string{"alpha"}); err != nil {
		t.Fatalf("runMap() error = %v", err)
	}

	var payload struct {
		Feature string `json:"feature"`
		Groups  []struct {
			ReadPolicy string `json:"read_policy"`
			Entries    []struct {
				ReadTarget string `json:"read_target"`
				Resolved   bool   `json:"resolved"`
			} `json:"entries"`
		} `json:"groups"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; output %q", err, out.String())
	}
	if payload.Feature != "0001-alpha" || len(payload.Groups) == 0 || len(payload.Groups[0].Entries) == 0 {
		t.Fatalf("unexpected JSON context payload: %#v", payload)
	}
	if payload.Groups[0].Entries[0].ReadTarget == "" {
		t.Fatalf("expected read target in JSON payload: %#v", payload)
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
	writeMapProjectFile(t, filepath.Join(projectRoot, "docs", "specs", "0001-alpha", "SPEC.md"), `---
kit_metadata_version: 1
artifact: spec
feature:
  id: "0001"
  slug: alpha
  dir: 0001-alpha
references:
  - name: docs/agents/RLM.md
    type: doc
    target: docs/agents/RLM.md
    relation: informs
    read_policy: conditional
    used_for: context routing
    status: active
  - name: old-context.md
    type: doc
    target: docs/old-context.md
    relation: informs
    read_policy: skip
    used_for: legacy context
    status: stale
---
# SPEC

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
