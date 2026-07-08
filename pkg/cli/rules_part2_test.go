package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/templates"
	"github.com/spf13/cobra"
)

func TestRunRulesListStableOrdering(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)

	writeFile(t, filepath.Join(projectRoot, "docs", "references", "rules", "testing.md"), templates.BuildRuleset("testing", []string{"testing"}))
	writeFile(t, filepath.Join(projectRoot, "docs", "references", "rules", "api-conventions.md"), templates.BuildRuleset("api-conventions", []string{"api"}))

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	if err := runRulesList(cmd, nil); err != nil {
		t.Fatalf("runRulesList() error = %v", err)
	}

	rendered := out.String()
	apiIndex := strings.Index(rendered, "api-conventions")
	testingIndex := strings.Index(rendered, "testing")
	if apiIndex < 0 || testingIndex < 0 || apiIndex > testingIndex {
		t.Fatalf("expected stable slug ordering, got:\n%s", rendered)
	}
	for _, check := range []string{"SLUG", "PATH", "STATUS", "APPLIES_TO"} {
		if !strings.Contains(rendered, check) {
			t.Fatalf("expected list output to contain %q, got:\n%s", check, rendered)
		}
	}
}

func TestRunRulesAddRegistrySelectorImportsMissingRuleset(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)
	stubRulesetRegistry(t, registryRulesetForTest("safety-guardrails", []string{"git", "github"}))

	output := withStdin(t, "1\n", func() string {
		return captureStdout(t, func() {
			if err := runRulesAdd(&cobra.Command{}, nil); err != nil {
				t.Fatalf("runRulesAdd() error = %v", err)
			}
		})
	})

	path := filepath.Join(projectRoot, "docs", "references", "rules", "safety-guardrails.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected imported ruleset file: %v", err)
	}
	if !strings.Contains(string(content), "slug: safety-guardrails") || !strings.Contains(string(content), "status: active") {
		t.Fatalf("unexpected imported content:\n%s", content)
	}
	if !strings.Contains(output, "Imported: 1") {
		t.Fatalf("expected import summary, got:\n%s", output)
	}
}

func TestRunRulesAddRegistrySelectorShowsRulesetDescription(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)
	stubRulesetRegistry(t, registryRulesetForTest("safety-guardrails", []string{"git", "github"}))

	output := withStdin(t, "\n", func() string {
		return captureStdout(t, func() {
			if err := runRulesAdd(&cobra.Command{}, nil); err != nil {
				t.Fatalf("runRulesAdd() error = %v", err)
			}
		})
	})

	if !strings.Contains(output, "Description for safety-guardrails") {
		t.Fatalf("expected selector output to include ruleset description, got:\n%s", output)
	}
}

func TestRenderRegistryRulesetSelectorUsesStructuredTable(t *testing.T) {
	registry := registryRulesetForTest("safety-guardrails", []string{"git", "github"})
	entries := []registrySelectorEntry{
		{
			Registry:      registry,
			DesiredActive: true,
		},
	}

	var out bytes.Buffer
	renderRegistryRulesetSelector(&out, entries, -1)
	rendered := out.String()

	for _, check := range []string{
		"┌",
		"│ No ",
		"│ Use ",
		"│ Ruleset",
		"│ State",
		"│ Source",
		"│ Description",
		"[x]",
		"ACTIVE",
		"REGISTRY",
		"Description for safety-guardrails",
	} {
		if !strings.Contains(rendered, check) {
			t.Fatalf("expected selector table to contain %q, got:\n%s", check, rendered)
		}
	}
}

func TestRenderRegistryRulesetSelectorUsesColorWhenEnabled(t *testing.T) {
	previousCheck := terminalWriterCheck
	terminalWriterCheck = func(_ io.Writer) bool { return true }
	t.Cleanup(func() {
		terminalWriterCheck = previousCheck
	})

	registry := registryRulesetForTest("github-pr-delivery", []string{"github"})
	entries := []registrySelectorEntry{
		{
			Registry:      registry,
			Installed:     true,
			Modified:      true,
			DesiredActive: true,
		},
	}

	var out bytes.Buffer
	renderRegistryRulesetSelector(&out, entries, 0)
	rendered := out.String()

	for _, check := range []string{
		"\033[",
		"Tab/Down/j move",
		"Shift+Tab/Up/k move",
		">1",
		"ACTIVE",
		"MODIFIED",
		"github-pr-delivery",
	} {
		if !strings.Contains(rendered, check) {
			t.Fatalf("expected colored selector table to contain %q, got:\n%s", check, rendered)
		}
	}
}

func TestMoveRegistrySelectorCursorSupportsTabWrapping(t *testing.T) {
	tests := []struct {
		name   string
		cursor int
		count  int
		delta  int
		wrap   bool
		want   int
	}{
		{name: "tab moves down", cursor: 0, count: 3, delta: 1, wrap: true, want: 1},
		{name: "tab wraps to first", cursor: 2, count: 3, delta: 1, wrap: true, want: 0},
		{name: "shift tab moves up", cursor: 2, count: 3, delta: -1, wrap: true, want: 1},
		{name: "shift tab wraps to last", cursor: 0, count: 3, delta: -1, wrap: true, want: 2},
		{name: "down clamps", cursor: 2, count: 3, delta: 1, wrap: false, want: 2},
		{name: "up clamps", cursor: 0, count: 3, delta: -1, wrap: false, want: 0},
		{name: "empty list", cursor: 0, count: 0, delta: 1, wrap: true, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := moveRegistrySelectorCursor(tt.cursor, tt.count, tt.delta, tt.wrap)
			if got != tt.want {
				t.Fatalf("moveRegistrySelectorCursor() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestRawTerminalLineWriterTranslatesLFToCRLF(t *testing.T) {
	var out bytes.Buffer
	writer := &rawTerminalLineWriter{writer: &out, fd: 123}

	n, err := writer.Write([]byte("a\nb\r\nc\r"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if n != len("a\nb\r\nc\r") {
		t.Fatalf("Write() n = %d, want %d", n, len("a\nb\r\nc\r"))
	}
	if _, err := writer.Write([]byte("\nd\n")); err != nil {
		t.Fatalf("second Write() error = %v", err)
	}

	want := "a\r\nb\r\nc\r\nd\r\n"
	if out.String() != want {
		t.Fatalf("raw terminal output = %q, want %q", out.String(), want)
	}
	if writer.Fd() != 123 {
		t.Fatalf("Fd() = %d, want 123", writer.Fd())
	}
}

func TestRunRulesViewShowsRegistryRulesetBeforeImport(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)
	stubRulesetRegistry(t, registryRulesetForTest("safety-guardrails", []string{"git", "github"}))

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	if err := runRulesView(cmd, []string{"safety-guardrails"}); err != nil {
		t.Fatalf("runRulesView() error = %v", err)
	}

	for _, check := range []string{
		"Source: https://github.com/jamesonstone/kit/blob/main/docs/references/rules/safety-guardrails.md",
		"description: 'Description for safety-guardrails'",
		"# Ruleset: safety-guardrails",
	} {
		if !strings.Contains(out.String(), check) {
			t.Fatalf("expected view output to contain %q, got:\n%s", check, out.String())
		}
	}
}
