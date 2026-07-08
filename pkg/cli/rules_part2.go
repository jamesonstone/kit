package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/templates"
	"github.com/spf13/cobra"
)

func runRulesAddInteractive(cmd *cobra.Command, projectRoot, readPolicyDefault string, policyExplicit bool) error {
	inputCfg := newFreeTextInputConfig(rulesAddUseVim, rulesAddEditor, rulesAddInline, true)
	input, err := promptRulesetAddInputs(projectRoot, inputCfg, readPolicyDefault, policyExplicit)
	if err != nil {
		return err
	}

	path, err := createRuleset(projectRoot, input)
	if err != nil {
		return err
	}

	if !rulesAddOutputOnly {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created ruleset %s at %s\n", input.Slug, rulesetTarget(input.Slug))
	}

	prompt := buildRulesetOptimizationPrompt(projectRoot, path, input)
	if err := outputPromptWithClipboardDefault(prompt, rulesAddOutputOnly, rulesAddCopy); err != nil {
		return err
	}

	if !rulesAddOutputOnly {
		printWorkflowInstructions("rules add", []string{
			fmt.Sprintf("review and refine %s", path),
			fmt.Sprintf("link the ruleset only where relevant with `kit rules link <feature> %s --read-policy conditional`", input.Slug),
			"run `kit check --project` after agent optimization",
		})
	}

	return nil
}

func createRuleset(projectRoot string, input rulesetAddInput) (string, error) {
	path := rulesetPath(projectRoot, input.Slug)
	if document.Exists(path) && !rulesAddForce {
		return "", fmt.Errorf("ruleset %q already exists at %s; use --force to overwrite", input.Slug, rulesetTarget(input.Slug))
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("failed to create ruleset directory: %w", err)
	}
	content := templates.BuildRulesetWithOptions(templates.RulesetOptions{
		Slug:              input.Slug,
		AppliesTo:         input.AppliesTo,
		ReadPolicyDefault: input.ReadPolicyDefault,
		Context:           input.Context,
	})
	if err := document.Write(path, content); err != nil {
		return "", fmt.Errorf("failed to write ruleset %q: %w", input.Slug, err)
	}
	return path, nil
}

func runRulesList(cmd *cobra.Command, args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	rulesets, err := listRulesets(projectRoot)
	if err != nil {
		return err
	}
	if len(rulesets) == 0 {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "No rulesets found.")
		return err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	return printRulesetList(cmd.OutOrStdout(), projectRoot, cfg, rulesets)
}

func runRulesView(cmd *cobra.Command, args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	slug := strings.TrimSpace(args[0])
	if err := validateRulesetSlug(slug); err != nil {
		return err
	}
	content, source, err := loadRulesetViewContent(cmd.Context(), projectRoot, slug)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Source: %s\n\n%s", source, ensureTrailingNewline(content))
	return err
}

func runRulesLink(cmd *cobra.Command, args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	feat, err := feature.Resolve(cfg.SpecsPath(projectRoot), args[0])
	if err != nil {
		return fmt.Errorf("feature %q not found", args[0])
	}

	slug := strings.TrimSpace(args[1])
	if err := validateRulesetSlug(slug); err != nil {
		return err
	}
	readPolicy := strings.TrimSpace(rulesLinkReadPolicy)
	if readPolicy != document.ReferenceReadPolicyMust && readPolicy != document.ReferenceReadPolicyConditional {
		return fmt.Errorf("--read-policy must be one of: must, conditional")
	}

	ruleset, err := loadRuleset(projectRoot, slug)
	if err != nil {
		return err
	}
	if issues := validateRulesetDocument(ruleset, slug); len(issues) > 0 {
		return fmt.Errorf("ruleset %q is invalid: %s", slug, strings.Join(issues, "; "))
	}

	targetPath, docType, err := rulesetLinkTargetDoc(feat)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(targetPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", targetPath, err)
	}

	reference := rulesetReference(slug, readPolicy)
	updated, changed, err := document.UpsertMetadata(string(content), docType, document.MetadataUpsert{
		Feature:    document.FeatureMetadataFromDir(feat.DirName),
		References: []document.MetadataReference{reference},
	})
	if err != nil {
		return fmt.Errorf("failed to update feature references in %s: %w", targetPath, err)
	}
	if changed {
		if err := document.Write(targetPath, updated); err != nil {
			return fmt.Errorf("failed to write feature references in %s: %w", targetPath, err)
		}
	}

	relPath, _ := filepath.Rel(projectRoot, targetPath)
	action := "Updated"
	if !changed {
		action = "Already linked"
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s ruleset %s in %s\n", action, slug, filepath.ToSlash(relPath))
	return err
}

func selectedRulesetReadPolicy() (string, bool, error) {
	selected := ""
	count := 0
	for _, option := range []struct {
		enabled bool
		policy  string
	}{
		{enabled: rulesAddMust, policy: document.ReferenceReadPolicyMust},
		{enabled: rulesAddConditional, policy: document.ReferenceReadPolicyConditional},
		{enabled: rulesAddEvidence, policy: document.ReferenceReadPolicyEvidence},
		{enabled: rulesAddSkip, policy: document.ReferenceReadPolicySkip},
	} {
		if !option.enabled {
			continue
		}
		selected = option.policy
		count++
	}
	if count > 1 {
		return "", false, fmt.Errorf("choose only one of --must, --conditional, --evidence, or --skip")
	}
	if selected != "" {
		return selected, true, nil
	}
	return defaultRulesetReadPolicy, false, nil
}
