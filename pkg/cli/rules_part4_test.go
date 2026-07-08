package cli

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/templates"
	"github.com/spf13/cobra"
)

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
