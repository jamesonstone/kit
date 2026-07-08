package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func resetMapCommandState(t *testing.T) {
	t.Helper()

	oldContext := mapContext
	oldJSON := mapJSON
	oldAll := mapAll

	mapContext = false
	mapJSON = false
	mapAll = false

	t.Cleanup(func() {
		mapContext = oldContext
		mapJSON = oldJSON
		mapAll = oldAll
	})
}

func TestRunMap_ProjectWideOutput(t *testing.T) {
	resetMapCommandState(t)
	mapAll = true

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

func TestRunMap_DefaultInteractiveSelector(t *testing.T) {
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

	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	if _, err := writePipe.WriteString("1\n"); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	if err := writePipe.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	oldStdin := os.Stdin
	os.Stdin = readPipe
	defer func() {
		os.Stdin = oldStdin
		_ = readPipe.Close()
	}()

	out := &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(out)

	if err := runMap(cmd, nil); err != nil {
		t.Fatalf("runMap() error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Select a feature to map:") {
		t.Fatalf("expected interactive selector prompt, got %q", got)
	}
	if !strings.Contains(got, "🗺️ Kit Map: 0001-alpha") {
		t.Fatalf("expected selected feature map output, got %q", got)
	}
}

func TestRunMap_DefaultInteractiveSelectorRejectsEmptySelection(t *testing.T) {
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

	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	if err := writePipe.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	oldStdin := os.Stdin
	os.Stdin = readPipe
	defer func() {
		os.Stdin = oldStdin
		_ = readPipe.Close()
	}()

	out := &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(out)

	err = runMap(cmd, nil)
	if err == nil {
		t.Fatalf("expected empty selection error")
	}
	if !strings.Contains(err.Error(), "no feature selected") || !strings.Contains(err.Error(), "kit map --all") {
		t.Fatalf("expected actionable error, got %q", err.Error())
	}
	if !strings.Contains(out.String(), "Select a feature to map:") {
		t.Fatalf("expected selector prompt, got %q", out.String())
	}
}

func TestRunMap_AllRejectsFeatureArgument(t *testing.T) {
	resetMapCommandState(t)
	mapAll = true

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

	cmd := &cobra.Command{}
	err = runMap(cmd, []string{"alpha"})
	if err == nil {
		t.Fatalf("expected --all with feature argument to fail")
	}
	if !strings.Contains(err.Error(), "--all cannot be used with a feature argument") {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}
