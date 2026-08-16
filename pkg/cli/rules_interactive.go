package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/document"
	"github.com/jamesonstone/kit/v3/internal/feature"
)

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

func promptRulesetAddInputs(projectRoot string, inputCfg freeTextInputConfig, readPolicyDefault string, policyExplicit bool) (rulesetAddInput, error) {
	reader := bufio.NewReader(os.Stdin)
	style := styleForStdout()
	printSectionBanner("📚", "Ruleset Builder")
	fmt.Println(style.muted("Create a durable repo-local ruleset under docs/references/rules/."))
	fmt.Println(style.muted("Keep rulesets pointer-loaded; do not copy them into always-loaded instruction files."))
	fmt.Println()

	name, slug, err := promptRulesetName(reader)
	if err != nil {
		return rulesetAddInput{}, err
	}
	if document.Exists(rulesetPath(projectRoot, slug)) && !rulesAddForce {
		return rulesetAddInput{}, fmt.Errorf("ruleset %q already exists at %s; use --force to overwrite", slug, rulesetTarget(slug))
	}

	if !policyExplicit {
		readPolicyDefault, err = promptRulesetReadPolicy(reader, readPolicyDefault)
		if err != nil {
			return rulesetAddInput{}, err
		}
	} else {
		fmt.Printf("%s\n", style.muted(fmt.Sprintf("Using read_policy_default from flag: %s", readPolicyDefault)))
	}

	appliesTo, err := promptRulesetAppliesTo(reader, defaultRulesetAppliesTo(slug))
	if err != nil {
		return rulesetAddInput{}, err
	}

	context, err := promptRulesetContext(inputCfg)
	if err != nil {
		return rulesetAddInput{}, err
	}

	return rulesetAddInput{
		Name:              name,
		Slug:              slug,
		AppliesTo:         appliesTo,
		ReadPolicyDefault: readPolicyDefault,
		Context:           context,
	}, nil
}

func promptRulesetName(reader *bufio.Reader) (string, string, error) {
	style := styleForStdout()
	fmt.Println(style.muted("Step 1 of 4: Enter a ruleset name."))
	fmt.Println(style.muted("It will be normalized to lowercase kebab-case."))
	fmt.Print(whiteBold + "   > " + reset)
	name, err := readRulesetLine(reader)
	if err != nil {
		return "", "", err
	}
	if name == "" {
		return "", "", fmt.Errorf("ruleset name cannot be empty")
	}
	slug := normalizeRulesetSlug(name)
	if err := validateRulesetSlug(slug); err != nil {
		return "", "", err
	}
	if slug != name {
		fmt.Printf(dim+"Using normalized ruleset slug: %s"+reset+"\n\n", slug)
	}
	return name, slug, nil
}

func promptRulesetReadPolicy(reader *bufio.Reader, fallback string) (string, error) {
	style := styleForStdout()
	fmt.Println(style.muted("Step 2 of 4: Choose how this ruleset should be loaded when referenced."))
	fmt.Println(style.muted("Use --must, --conditional, --evidence, or --skip to skip this prompt."))
	fmt.Println("  1. conditional (recommended) - load only when relevant to the current decision")
	fmt.Println("  2. must - load whenever a feature references this ruleset")
	fmt.Println("  3. evidence - load when checking or citing supporting guidance")
	fmt.Println("  4. skip - create as staged or inactive guidance")
	fmt.Printf("%s", whiteBold+"   > "+reset)
	answer, err := readRulesetLine(reader)
	if err != nil {
		return "", err
	}
	if answer == "" {
		return fallback, nil
	}
	switch strings.ToLower(answer) {
	case "1", "c", "conditional":
		return document.ReferenceReadPolicyConditional, nil
	case "2", "m", "must":
		return document.ReferenceReadPolicyMust, nil
	case "3", "e", "evidence":
		return document.ReferenceReadPolicyEvidence, nil
	case "4", "s", "skip":
		return document.ReferenceReadPolicySkip, nil
	default:
		return "", fmt.Errorf("invalid ruleset read policy %q", answer)
	}
}

func promptRulesetAppliesTo(reader *bufio.Reader, defaults []string) ([]string, error) {
	style := styleForStdout()
	defaultText := strings.Join(defaults, ",")
	fmt.Println()
	fmt.Println(style.muted("Step 3 of 4: Enter applies_to tags as a comma-separated list."))
	fmt.Println(style.muted(fmt.Sprintf("Press Enter to use: %s", defaultText)))
	fmt.Print(whiteBold + "   > " + reset)
	answer, err := readRulesetLine(reader)
	if err != nil {
		return nil, err
	}
	if answer == "" {
		return defaults, nil
	}
	return parseRulesetAppliesTo(answer)
}

func promptRulesetContext(inputCfg freeTextInputConfig) (string, error) {
	style := styleForStdout()
	fmt.Println()
	fmt.Println(style.muted("Step 4 of 4: Describe the durable rule context."))
	if inputCfg.usesEditor() {
		fmt.Printf("%s\n", style.muted(fmt.Sprintf("A %s will open for this response.", inputCfg.editorLabel())))
		return readEditorText(inputCfg, "ruleset context", false)
	}

	rl, err := newMultilineReadline()
	if err != nil {
		return "", fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer closeMultilineReadline(rl)

	fmt.Println(style.muted("Press Enter to submit. Use Shift+Enter or Ctrl+J to insert newlines."))
	context := readLineRL(rl)
	if context == "" {
		return "", fmt.Errorf("ruleset context cannot be empty")
	}
	return context, nil
}

func readRulesetLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
