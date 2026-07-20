package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/templates"
)

func TestRunRulesAddCreatesRuleset(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	if err := runRulesAdd(cmd, []string{"frontend-ui"}); err != nil {
		t.Fatalf("runRulesAdd() error = %v", err)
	}

	path := filepath.Join(projectRoot, "docs", "references", "rules", "frontend-ui.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected ruleset file: %v", err)
	}
	for _, check := range []string{
		"kind: ruleset",
		"slug: frontend-ui",
		"- frontend",
		"## Purpose",
		"## Applies When",
		"## Rules",
		"## Anti-Patterns",
		"## Verification",
		"## Examples",
	} {
		if !strings.Contains(string(content), check) {
			t.Fatalf("expected ruleset content to contain %q, got:\n%s", check, content)
		}
	}
	if !strings.Contains(out.String(), "Created ruleset frontend-ui") {
		t.Fatalf("expected create output, got %q", out.String())
	}
}

func TestSafetyGuardrailsRegistryRulesetRequiresAutonomousRecovery(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "references", "rules", "safety-guardrails.md")
	ruleset, err := parseRulesetFile(path)
	if err != nil {
		t.Fatalf("parseRulesetFile() error = %v", err)
	}
	if issues := validateRulesetDocument(ruleset, "safety-guardrails"); len(issues) > 0 {
		t.Fatalf("safety-guardrails ruleset issues = %#v", issues)
	}
	if ruleset.Metadata.ReadPolicyDefault != document.ReferenceReadPolicyMust {
		t.Fatalf("read_policy_default = %q, want must", ruleset.Metadata.ReadPolicyDefault)
	}
	for _, check := range []string{
		"retry autonomously",
		"including `gh`",
		"Ask permission only before large-scale deletion or deleting sensitive files",
		"do not frame this as permission for a routine retry",
	} {
		if !strings.Contains(ruleset.Body, check) {
			t.Fatalf("expected safety-guardrails ruleset to contain %q", check)
		}
	}
	for _, forbidden := range []string{
		"Do not retry with mutation",
		"Surface the failure to the user and await instruction",
	} {
		if strings.Contains(ruleset.Body, forbidden) {
			t.Fatalf("expected safety-guardrails ruleset to omit blanket stop behavior %q", forbidden)
		}
	}
}

func TestConstitutionCurationRegistryRulesetIsValid(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "references", "rules", "constitution-curation.md")
	ruleset, err := parseRulesetFile(path)
	if err != nil {
		t.Fatalf("parseRulesetFile() error = %v", err)
	}
	if issues := validateRulesetDocument(ruleset, "constitution-curation"); len(issues) > 0 {
		t.Fatalf("constitution-curation ruleset issues = %#v", issues)
	}
	if ruleset.Metadata.RegistryScope != rulesetRegistryScopeDownstream {
		t.Fatalf("registry_scope = %q, want downstream", ruleset.Metadata.RegistryScope)
	}
	if ruleset.Metadata.ReadPolicyDefault != document.ReferenceReadPolicyMust {
		t.Fatalf("read_policy_default = %q, want must", ruleset.Metadata.ReadPolicyDefault)
	}
	for _, check := range []string{
		"Treat the exact generated Constitution starter as a valid bootstrap state",
		"When no project-wide truth changed, leave the Constitution unchanged",
		"Treat project-refresh cadence as a trigger for reviewed semantic analysis",
	} {
		if !strings.Contains(ruleset.Body, check) {
			t.Fatalf("expected constitution-curation ruleset to contain %q", check)
		}
	}
}

func TestGitHubPRDeliveryRulesetUsesAutonomousRecovery(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "references", "rules", "github-pr-delivery.md")
	ruleset, err := parseRulesetFile(path)
	if err != nil {
		t.Fatalf("parseRulesetFile() error = %v", err)
	}
	if issues := validateRulesetDocument(ruleset, "github-pr-delivery"); len(issues) > 0 {
		t.Fatalf("github-pr-delivery ruleset issues = %#v", issues)
	}
	for _, check := range []string{
		"retry autonomously",
		"another supported authenticated path such as `gh`",
		"without requesting routine retry permission",
		"Verify that no duplicate issue or PR was created",
	} {
		if !strings.Contains(ruleset.Body, check) {
			t.Fatalf("expected github-pr-delivery ruleset to contain %q", check)
		}
	}
	if strings.Contains(ruleset.Body, "stop and do not mutate to retry") {
		t.Fatal("expected github-pr-delivery ruleset to omit blanket mutation retry prohibition")
	}
}

func TestGitHubPRDeliveryRulesetPreservesAdditionalScopeLane(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "references", "rules", "github-pr-delivery.md")
	ruleset, err := parseRulesetFile(path)
	if err != nil {
		t.Fatalf("parseRulesetFile() error = %v", err)
	}
	if issues := validateRulesetDocument(ruleset, "github-pr-delivery"); len(issues) > 0 {
		t.Fatalf("github-pr-delivery ruleset issues = %#v", issues)
	}
	for _, check := range []string{
		"Create or reuse a separate GitHub issue for the additional scope",
		"Keep the existing pull request head branch. Do not create a second branch or pull request",
		"Scope every new commit for the additional work to its own issue number",
		"append a separate `Closes #123` line",
	} {
		if !strings.Contains(ruleset.Body, check) {
			t.Fatalf("expected github-pr-delivery ruleset to contain %q", check)
		}
	}
}

func TestWorkLaneGatingRulesetUsesAutonomousRecoveryAndCleanPreflight(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "references", "rules", "work-lane-gating.md")
	ruleset, err := parseRulesetFile(path)
	if err != nil {
		t.Fatalf("parseRulesetFile() error = %v", err)
	}
	if issues := validateRulesetDocument(ruleset, "work-lane-gating"); len(issues) > 0 {
		t.Fatalf("work-lane-gating ruleset issues = %#v", issues)
	}
	for _, check := range []string{
		"autonomous failure recovery",
		"when it can be proven safely",
		"request only the missing lane decision",
		"automatic clean-preflight decision",
		"Do not ask whether to create a new issue, branch, and pull request or continue existing work",
		"No existing issue, branch, or pull request covers the requested work",
		"create or reuse a separate human-assigned issue for the additional scope",
		"keep the existing branch and pull request",
	} {
		if !strings.Contains(ruleset.Body, check) {
			t.Fatalf("expected work-lane-gating ruleset to contain %q", check)
		}
	}
	if strings.Contains(ruleset.Body, "leave changes in the working tree, and await instruction") {
		t.Fatal("expected work-lane-gating ruleset to omit blanket await-instruction behavior")
	}
}

func TestRunRulesAddSupportsPolicyFlags(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)

	rulesAddMust = true
	cmd := &cobra.Command{}
	if err := runRulesAdd(cmd, []string{"security"}); err != nil {
		t.Fatalf("runRulesAdd() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(projectRoot, "docs", "references", "rules", "security.md"))
	if err != nil {
		t.Fatalf("expected ruleset file: %v", err)
	}
	if !strings.Contains(string(content), "read_policy_default: must") {
		t.Fatalf("expected must policy, got:\n%s", content)
	}
}

func TestRunRulesAddRejectsMultiplePolicyFlags(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)

	rulesAddMust = true
	rulesAddConditional = true
	err := runRulesAdd(&cobra.Command{}, []string{"security"})
	if err == nil || !strings.Contains(err.Error(), "choose only one") {
		t.Fatalf("expected multiple policy flag error, got %v", err)
	}
}

func TestRunRulesAddRejectsInvalidAndDuplicateSlug(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)

	cmd := &cobra.Command{}
	if err := runRulesAdd(cmd, []string{"Frontend UI"}); err == nil {
		t.Fatal("expected invalid slug to fail")
	}
	if err := runRulesAdd(cmd, []string{"frontend-ui"}); err != nil {
		t.Fatalf("initial runRulesAdd() error = %v", err)
	}
	if err := runRulesAdd(cmd, []string{"frontend-ui"}); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected duplicate ruleset error, got %v", err)
	}

	rulesAddForce = true
	if err := runRulesAdd(cmd, []string{"frontend-ui"}); err != nil {
		t.Fatalf("forced runRulesAdd() error = %v", err)
	}
}

func TestRunRulesAddInteractiveCreatesRulesetAndCopiesOptimizationPrompt(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)
	rulesAddCustom = true

	previousWait := awaitEditorLaunchConfirmation
	previousRunner := editorInputRunner
	t.Cleanup(func() {
		awaitEditorLaunchConfirmation = previousWait
		editorInputRunner = previousRunner
	})

	awaitEditorLaunchConfirmation = func(_ *os.File, _ io.Writer) error {
		return nil
	}
	editorInputRunner = func(_ freeTextInputConfig, fieldName, _ string) (string, bool, error) {
		if fieldName != "ruleset context" {
			t.Fatalf("unexpected field name %q", fieldName)
		}
		return "These rules guide frontend UI decisions with accessibility and responsive layout constraints.", true, nil
	}

	var copied string
	withClipboardCopy(t, func(text string) error {
		copied = text
		return nil
	})

	output := withStdin(t, "Frontend UI\n\n\n", func() string {
		return captureStdout(t, func() {
			if err := runRulesAdd(&cobra.Command{}, nil); err != nil {
				t.Fatalf("runRulesAdd() error = %v", err)
			}
		})
	})

	path := filepath.Join(projectRoot, "docs", "references", "rules", "frontend-ui.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected ruleset file: %v", err)
	}
	for _, check := range []string{
		"slug: frontend-ui",
		"read_policy_default: conditional",
		"- frontend",
		"These rules guide frontend UI decisions",
	} {
		if !strings.Contains(string(content), check) {
			t.Fatalf("expected ruleset content to contain %q, got:\n%s", check, content)
		}
	}
	for _, check := range []string{
		"Created ruleset frontend-ui",
		"Copied the prepared text to the clipboard",
	} {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q, got:\n%s", check, output)
		}
	}
	for _, check := range []string{
		"Optimize this Kit durable ruleset",
		path,
		"read_policy_default: conditional",
		"kit check --project",
	} {
		if !strings.Contains(copied, check) {
			t.Fatalf("expected copied prompt to contain %q, got:\n%s", check, copied)
		}
	}
}

func TestRunRulesAddInteractiveRejectsDuplicateBeforeEditor(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)
	rulesAddCustom = true

	writeFile(t, filepath.Join(projectRoot, "docs", "references", "rules", "frontend-ui.md"), templates.BuildRuleset("frontend-ui", []string{"frontend"}))

	previousRunner := editorInputRunner
	t.Cleanup(func() {
		editorInputRunner = previousRunner
	})
	editorInputRunner = func(_ freeTextInputConfig, _ string, _ string) (string, bool, error) {
		t.Fatal("editor should not open for duplicate ruleset")
		return "", false, nil
	}

	_ = withStdin(t, "Frontend UI\n", func() string {
		err := runRulesAdd(&cobra.Command{}, nil)
		if err == nil || !strings.Contains(err.Error(), "already exists") {
			t.Fatalf("expected duplicate ruleset error, got %v", err)
		}
		return ""
	})
}

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

func TestRulesetRegistrySectionArtifactsUseHeadingPathsAndSkipFences(t *testing.T) {
	registry := registryRulesetForTest("section-keys", []string{"git"})
	content := registry.Content + "\n## Parent One\n\n### Duplicate\n\none\n\n## Parent Two\n\n### Duplicate\n\ntwo\n\n```bash\n# not a heading\n## also not a heading\n```\n"

	sections := rulesetRegistrySectionArtifacts(content, registry.Metadata.Status)
	var keys []string
	for _, section := range sections {
		keys = append(keys, section.Key)
	}
	for _, want := range []string{
		"# ruleset: section-keys",
		"# ruleset: section-keys > ## purpose",
		"# ruleset: section-keys > ## rules",
		"# ruleset: section-keys > ## parent one > ### duplicate",
		"# ruleset: section-keys > ## parent two > ### duplicate",
	} {
		if !slices.Contains(keys, want) {
			t.Fatalf("expected section key %q in %#v", want, keys)
		}
	}
	for _, unwanted := range []string{"# not a heading", "## also not a heading"} {
		if slices.Contains(keys, unwanted) {
			t.Fatalf("unexpected fenced-code section key %q in %#v", unwanted, keys)
		}
	}
}

func TestFormatRulesetStateTokenUsesColorWhenEnabled(t *testing.T) {
	rendered := formatRulesetStateToken(humanOutputStyle{enabled: true}, "ACTIVE", plan)
	if !strings.Contains(rendered, "\033[") || !strings.Contains(rendered, "ACTIVE") {
		t.Fatalf("expected colored ACTIVE token, got %q", rendered)
	}
}

func TestRulesCommandSupportsRuleAlias(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"rule", "list"})
	if err != nil {
		t.Fatalf("rootCmd.Find(rule list) error = %v", err)
	}
	if cmd != rulesListCmd {
		t.Fatalf("expected rule list to resolve to rules list command, got %q", cmd.Name())
	}
}

func TestRunRulesLinkPreservesFrontMatterAndAvoidsDuplicates(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-alpha")
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), withFeatureFrontMatter(validSpecWithRelationships("none\n"), "spec", "0001-alpha"))
	writeFile(t, filepath.Join(projectRoot, "docs", "references", "rules", "api-conventions.md"), templates.BuildRuleset("api-conventions", []string{"api"}))

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	rulesLinkReadPolicy = document.ReferenceReadPolicyMust
	if err := runRulesLink(cmd, []string{"alpha", "api-conventions"}); err != nil {
		t.Fatalf("runRulesLink() error = %v", err)
	}
	if err := runRulesLink(cmd, []string{"alpha", "api-conventions"}); err != nil {
		t.Fatalf("second runRulesLink() error = %v", err)
	}

	doc, err := document.ParseFile(filepath.Join(featurePath, "SPEC.md"), document.TypeSpec)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}
	if doc.Metadata == nil || doc.Metadata.Feature.Slug != "alpha" {
		t.Fatalf("expected feature front matter to be preserved, got %#v", doc.Metadata)
	}
	var count int
	for _, reference := range doc.References() {
		if reference.ID == "ruleset-api-conventions" {
			count++
			if reference.ReadPolicy != document.ReferenceReadPolicyMust {
				t.Fatalf("ReadPolicy = %q, want must", reference.ReadPolicy)
			}
			if reference.Target != "docs/references/rules/api-conventions.md" {
				t.Fatalf("Target = %q", reference.Target)
			}
		}
	}
	if count != 1 {
		t.Fatalf("expected one ruleset reference, got %d in %#v", count, doc.References())
	}
}

func TestCheckFeatureFailsForMissingRulesetReference(t *testing.T) {
	projectRoot := setupRulesProject(t)
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	featurePath := filepath.Join(specsDir, "0001-alpha")
	spec := withFeatureFrontMatter(validSpecWithRelationships("none\n"), "spec", "0001-alpha")
	spec, _, err := document.UpsertMetadata(spec, document.TypeSpec, document.MetadataUpsert{
		References: []document.MetadataReference{rulesetReference("missing-rules", document.ReferenceReadPolicyConditional)},
	})
	if err != nil {
		t.Fatalf("UpsertMetadata() error = %v", err)
	}
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), spec)

	err = checkFeature(projectRoot, specsDir, "alpha")
	if err == nil || !strings.Contains(err.Error(), "validation failed") {
		t.Fatalf("expected checkFeature validation failure, got %v", err)
	}
}

func TestRunCheckProjectFailsForInvalidRuleset(t *testing.T) {
	projectRoot := setupCoherentProjectForCheck(t)
	writeFile(t, filepath.Join(projectRoot, "docs", "references", "rules", "frontend-ui.md"), `---
kind: ruleset
slug: frontend-ui
status: active
applies_to:
  - frontend
read_policy_default: conditional
---

# Ruleset: frontend-ui

## Purpose

purpose
`)
	setWorkingDirectory(t, projectRoot)

	checkProject = true
	checkAll = false
	t.Cleanup(func() {
		checkProject = false
		checkAll = false
	})

	cmd := &cobra.Command{}
	err := runCheck(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "project validation failed") {
		t.Fatalf("expected invalid ruleset project failure, got %v", err)
	}
}

func TestRunReconcileWarnsForActiveFrontendFeatureMissingRuleset(t *testing.T) {
	projectRoot := setupRulesProjectWithFrontendFeatures(t)
	setWorkingDirectory(t, projectRoot)
	resetReconcileFlags(t)

	cmd := &cobra.Command{}
	cmd.Flags().Bool("output-only", true, "")
	output := captureStdout(t, func() {
		if err := runReconcile(cmd, nil); err != nil {
			t.Fatalf("runReconcile() error = %v", err)
		}
	})
	if !strings.Contains(output, "active frontend feature has no active frontend ruleset reference") {
		t.Fatalf("expected frontend ruleset advisory, got:\n%s", output)
	}
}

func TestRunReconcileSkipsHistoricalFrontendRulesetAdvisory(t *testing.T) {
	projectRoot := setupRulesProjectWithFrontendFeatures(t)
	setWorkingDirectory(t, projectRoot)
	resetReconcileFlags(t)

	cmd := &cobra.Command{}
	cmd.Flags().Bool("output-only", true, "")
	output := captureStdout(t, func() {
		if err := runReconcile(cmd, []string{"historical-frontend"}); err != nil {
			t.Fatalf("runReconcile() error = %v", err)
		}
	})
	if strings.Contains(output, "active frontend feature has no active frontend ruleset reference") {
		t.Fatalf("expected historical feature to avoid ruleset advisory, got:\n%s", output)
	}
}

func setupRulesProject(t *testing.T) string {
	t.Helper()
	projectRoot := t.TempDir()
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	return projectRoot
}

func setupRulesProjectWithFrontendFeatures(t *testing.T) string {
	t.Helper()
	projectRoot := setupCoherentProjectForCheck(t)
	writeFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), validProgressSummaryForFeatures(
		[]string{"0001-historical-frontend", "0002-active-frontend"},
	))

	historicalPath := filepath.Join(projectRoot, "docs", "specs", "0001-historical-frontend")
	writeRulesFeatureDocs(t, historicalPath, "0001-historical-frontend", true)
	activePath := filepath.Join(projectRoot, "docs", "specs", "0002-active-frontend")
	writeRulesFeatureDocs(t, activePath, "0002-active-frontend", false)
	return projectRoot
}

func writeRulesFeatureDocs(t *testing.T, featurePath, dirName string, complete bool) {
	t.Helper()
	spec := withFeatureFrontMatter(validSpecWithRelationships("none\n"), "spec", dirName)
	spec, _, err := document.UpsertMetadata(spec, document.TypeSpec, document.MetadataUpsert{
		References: canonicalFrontendProfileReferences(dirName),
	})
	if err != nil {
		t.Fatalf("UpsertMetadata() error = %v", err)
	}
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), spec)
	writeFile(t, filepath.Join(featurePath, "PLAN.md"), withFeatureFrontMatter(validPlan(), "plan", dirName))
	writeFile(t, filepath.Join(featurePath, "TASKS.md"), withFeatureFrontMatter(legacyTasksWithoutExecutableFields(complete), "tasks", dirName))
}

func resetRulesFlags(t *testing.T) {
	t.Helper()
	previousCopy := rulesAddCopy
	previousEditor := rulesAddEditor
	previousEvidence := rulesAddEvidence
	previousForce := rulesAddForce
	previousInline := rulesAddInline
	previousMust := rulesAddMust
	previousOutputOnly := rulesAddOutputOnly
	previousSkip := rulesAddSkip
	previousCustom := rulesAddCustom
	previousConditional := rulesAddConditional
	previousUseVim := rulesAddUseVim
	previousReadPolicy := rulesLinkReadPolicy
	t.Cleanup(func() {
		rulesAddCopy = previousCopy
		rulesAddEditor = previousEditor
		rulesAddEvidence = previousEvidence
		rulesAddForce = previousForce
		rulesAddInline = previousInline
		rulesAddMust = previousMust
		rulesAddOutputOnly = previousOutputOnly
		rulesAddSkip = previousSkip
		rulesAddCustom = previousCustom
		rulesAddConditional = previousConditional
		rulesAddUseVim = previousUseVim
		rulesLinkReadPolicy = previousReadPolicy
	})
	rulesAddCopy = false
	rulesAddEditor = ""
	rulesAddEvidence = false
	rulesAddForce = false
	rulesAddInline = false
	rulesAddMust = false
	rulesAddOutputOnly = false
	rulesAddSkip = false
	rulesAddCustom = false
	rulesAddConditional = false
	rulesAddUseVim = false
	rulesLinkReadPolicy = defaultRulesetReadPolicy
}

func stubRulesetRegistry(t *testing.T, rulesets ...registryRuleset) {
	t.Helper()
	previous := rulesetRegistryFetcher
	t.Cleanup(func() {
		rulesetRegistryFetcher = previous
	})
	rulesetRegistryFetcher = func(_ context.Context) ([]registryRuleset, error) {
		return rulesets, nil
	}
}

func registryRulesetForTest(slug string, appliesTo []string) registryRuleset {
	content := templates.BuildRulesetWithOptions(templates.RulesetOptions{
		Slug:              slug,
		Description:       "Description for " + slug,
		AppliesTo:         appliesTo,
		ReadPolicyDefault: "conditional",
	})
	return registryRulesetWithContentForTest(slug, content, "test-"+slug+"-commit")
}

func registryRulesetWithContentForTest(slug, content, commit string) registryRuleset {
	parsed := parseRuleset(content, slug+".md")
	hash, err := normalizedRulesetContentHash(content, parsed.Metadata.Status)
	if err != nil {
		panic(err)
	}
	return registryRuleset{
		Slug:           slug,
		Content:        content,
		Metadata:       parsed.Metadata,
		SourceRepo:     rulesetRegistryRepoFullName(),
		SourceBranch:   rulesetRegistryBranch,
		SourceCommit:   commit,
		SourcePath:     rulesetTarget(slug),
		NormalizedHash: hash,
	}
}

func stubRulesetRegistryContent(t *testing.T, contentByCommit map[string]string) {
	t.Helper()
	previous := rulesetRegistryContentFetcher
	t.Cleanup(func() {
		rulesetRegistryContentFetcher = previous
	})
	rulesetRegistryContentFetcher = func(_ context.Context, _ string, commit string, _ string) (string, error) {
		content, ok := contentByCommit[commit]
		if !ok {
			return "", os.ErrNotExist
		}
		return content, nil
	}
}

func resetReconcileFlags(t *testing.T) {
	t.Helper()
	previousOutputOnly := reconcileOutputOnly
	previousAll := reconcileAll
	previousCopy := reconcileCopy
	previousMigrateReferences := reconcileMigrateReferences
	previousMigrateVerification := reconcileMigrateVerification
	t.Cleanup(func() {
		reconcileOutputOnly = previousOutputOnly
		reconcileAll = previousAll
		reconcileCopy = previousCopy
		reconcileMigrateReferences = previousMigrateReferences
		reconcileMigrateVerification = previousMigrateVerification
	})
	reconcileOutputOnly = false
	reconcileAll = false
	reconcileCopy = false
	reconcileMigrateReferences = false
	reconcileMigrateVerification = false
}
