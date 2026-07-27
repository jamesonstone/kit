package cli

import (
	"bytes"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/templates"
)

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
