package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/templates"
)

func TestWorkLaneGatingRulesetRequiresExplicitPullRequestLane(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "references", "rules", "work-lane-gating.md")
	ruleset, err := parseRulesetFile(path)
	if err != nil {
		t.Fatalf("parseRulesetFile() error = %v", err)
	}
	if issues := validateRulesetDocument(ruleset, "work-lane-gating"); len(issues) > 0 {
		t.Fatalf("work-lane-gating ruleset issues = %#v", issues)
	}
	normalizedBody := strings.Join(strings.Fields(ruleset.Body), " ")
	for _, check := range []string{
		"Before I make any repository changes",
		"canonical worktree, and pull request for this work",
		"`c` means continue existing",
		"`n` or `y` means new lane",
		"shorthand leads a longer response",
		"remaining text is supplemental lane instructions",
		"Ambiguous or contradictory responses",
		"Pull-Request Landing Plan",
		"source, tests, documentation, specs, plans, notes, generated",
		"Do not infer the choice from a clean default branch",
		"Treat that exact checkout as read-only",
		"Never use the primary checkout as a temporary edit location",
		"Do not repeatedly ask",
		"Do not stage, commit, push",
	} {
		if !strings.Contains(normalizedBody, check) {
			t.Fatalf("expected work-lane-gating ruleset to contain %q", check)
		}
	}
	for _, forbidden := range []string{
		"automatic clean-preflight decision",
		"Do not ask whether to create a new issue",
	} {
		if strings.Contains(normalizedBody, forbidden) {
			t.Fatalf("expected work-lane-gating ruleset to omit %q", forbidden)
		}
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
