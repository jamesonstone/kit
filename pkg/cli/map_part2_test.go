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
)

func TestRunMap_FeatureScopedOutput(t *testing.T) {
	resetMapCommandState(t)

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
	resetMapCommandState(t)
	mapContext = true

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
	resetMapCommandState(t)
	mapContext = true
	mapJSON = true

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
	resetMapCommandState(t)
	mapAll = true

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
