package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

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

func TestRunRulesViewPrefersLocalRuleset(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)
	stubRulesetRegistry(t, registryRulesetForTest("safety-guardrails", []string{"git", "github"}))
	local := registryRulesetForTest("safety-guardrails", []string{"git", "github"})
	local.Content = strings.Replace(local.Content, "Description for safety-guardrails", "Local description", 1)
	writeFile(t, filepath.Join(projectRoot, "docs", "references", "rules", "safety-guardrails.md"), local.Content)

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	if err := runRulesView(cmd, []string{"safety-guardrails"}); err != nil {
		t.Fatalf("runRulesView() error = %v", err)
	}

	for _, check := range []string{
		"Source: docs/references/rules/safety-guardrails.md",
		"Local description",
	} {
		if !strings.Contains(out.String(), check) {
			t.Fatalf("expected local view output to contain %q, got:\n%s", check, out.String())
		}
	}
}

func TestProjectRulesetRegistryFiltersMaintainerOnlyRules(t *testing.T) {
	usage := registryRulesetForTest("kit-capabilities-usage", []string{"kit", "cli"})
	maintainer := registryRulesetForTest("command-capabilities", []string{"kit", "cli"})
	maintainer.Metadata.RegistryScope = rulesetRegistryScopeKitMaintainer

	filtered := projectRulesetRegistry([]registryRuleset{usage, maintainer})
	if len(filtered) != 1 || filtered[0].Slug != "kit-capabilities-usage" {
		t.Fatalf("filtered registry = %#v, want only downstream usage rule", filtered)
	}
}

func TestRunRulesAddRegistrySelectorDeactivatesExistingRuleset(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)
	registry := registryRulesetForTest("work-lane-gating", []string{"workflow"})
	stubRulesetRegistry(t, registry)
	writeFile(t, filepath.Join(projectRoot, "docs", "references", "rules", "work-lane-gating.md"), registry.Content)

	output := withStdin(t, "1\n", func() string {
		return captureStdout(t, func() {
			if err := runRulesAdd(&cobra.Command{}, nil); err != nil {
				t.Fatalf("runRulesAdd() error = %v", err)
			}
		})
	})

	content, err := os.ReadFile(filepath.Join(projectRoot, "docs", "references", "rules", "work-lane-gating.md"))
	if err != nil {
		t.Fatalf("expected local ruleset file: %v", err)
	}
	if !strings.Contains(string(content), "status: optional") {
		t.Fatalf("expected deactivated optional status, got:\n%s", content)
	}
	if !strings.Contains(output, "Deactivated: 1") {
		t.Fatalf("expected deactivate summary, got:\n%s", output)
	}
}

func TestRunRulesAddRegistrySelectorReactivatesModifiedRulesetWithoutOverwriting(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)
	registry := registryRulesetForTest("github-pr-delivery", []string{"github"})
	stubRulesetRegistry(t, registry)

	local := strings.Replace(registry.Content, "status: active", "status: optional", 1)
	local = strings.Replace(local, "## Examples", "Custom local guidance.\n\n## Examples", 1)
	writeFile(t, filepath.Join(projectRoot, "docs", "references", "rules", "github-pr-delivery.md"), local)

	output := withStdin(t, "1\n", func() string {
		return captureStdout(t, func() {
			if err := runRulesAdd(&cobra.Command{}, nil); err != nil {
				t.Fatalf("runRulesAdd() error = %v", err)
			}
		})
	})

	content, err := os.ReadFile(filepath.Join(projectRoot, "docs", "references", "rules", "github-pr-delivery.md"))
	if err != nil {
		t.Fatalf("expected local ruleset file: %v", err)
	}
	for _, check := range []string{"status: active", "Custom local guidance."} {
		if !strings.Contains(string(content), check) {
			t.Fatalf("expected local content to contain %q, got:\n%s", check, content)
		}
	}
	if !strings.Contains(output, "LOCAL-CUSTOM") || !strings.Contains(output, "Activated: 1") {
		t.Fatalf("expected modified activation output, got:\n%s", output)
	}
}

func TestNormalizedRulesetHashIgnoresStatusOnlyChanges(t *testing.T) {
	registry := registryRulesetForTest("status-only", []string{"git"})
	local := strings.Replace(registry.Content, "status: active", "status: optional", 1)

	registryHash, err := normalizedRulesetContentHash(registry.Content, registry.Metadata.Status)
	if err != nil {
		t.Fatalf("registry hash error: %v", err)
	}
	localHash, err := normalizedRulesetContentHash(local, registry.Metadata.Status)
	if err != nil {
		t.Fatalf("local hash error: %v", err)
	}
	if registryHash != localHash {
		t.Fatalf("status-only hash drift: registry %s local %s", registryHash, localHash)
	}
}
