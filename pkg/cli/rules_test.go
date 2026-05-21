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
	rulesAddConditional = false
	rulesAddUseVim = false
	rulesLinkReadPolicy = defaultRulesetReadPolicy
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
