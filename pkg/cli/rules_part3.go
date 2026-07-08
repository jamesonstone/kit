package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
)

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

func parseRulesetAppliesTo(raw string) ([]string, error) {
	parts := strings.Split(raw, ",")
	var appliesTo []string
	for _, part := range parts {
		normalized := normalizeRulesetSlug(part)
		if normalized == "" {
			continue
		}
		if err := validateRulesetSlug(normalized); err != nil {
			return nil, fmt.Errorf("invalid applies_to entry %q: %w", strings.TrimSpace(part), err)
		}
		appliesTo = append(appliesTo, normalized)
	}
	if len(appliesTo) == 0 {
		return nil, fmt.Errorf("applies_to must contain at least one entry")
	}
	return appliesTo, nil
}

func normalizeRulesetSlug(value string) string {
	slug := strings.ToLower(strings.TrimSpace(value))
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	var builder strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			builder.WriteRune(r)
		}
	}
	slug = builder.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return strings.Trim(slug, "-")
}

func buildRulesetOptimizationPrompt(projectRoot, path string, input rulesetAddInput) string {
	relPath, err := filepath.Rel(projectRoot, path)
	if err != nil {
		relPath = path
	}
	relPath = filepath.ToSlash(relPath)
	referencesReadme := filepath.Join(projectRoot, "docs", "references", "README.md")
	rlmPath := filepath.Join(projectRoot, "docs", "agents", "RLM.md")
	initContractPath := filepath.Join(projectRoot, "docs", "specs", "0000_INIT_PROJECT.md")

	var sb strings.Builder
	sb.WriteString("Optimize this Kit durable ruleset for correctness, semantic clarity, and RLM just-in-time loading.\n\n")
	sb.WriteString("Ruleset file:\n")
	sb.WriteString("- " + filepath.Join(projectRoot, filepath.FromSlash(relPath)) + "\n\n")
	sb.WriteString("Creation context:\n")
	sb.WriteString("- name: " + input.Name + "\n")
	sb.WriteString("- slug: " + input.Slug + "\n")
	sb.WriteString("- applies_to: " + strings.Join(input.AppliesTo, ", ") + "\n")
	sb.WriteString("- read_policy_default: " + input.ReadPolicyDefault + "\n\n")
	sb.WriteString("Task:\n")
	sb.WriteString("1. Read the ruleset file and treat the captured context as the human source of truth.\n")
	sb.WriteString("2. Load only the Kit contract sections needed for this decision, starting with:\n")
	sb.WriteString("   - " + referencesReadme + "\n")
	sb.WriteString("   - " + rlmPath + "\n")
	sb.WriteString("   - " + initContractPath + "\n")
	sb.WriteString("3. Rewrite or reorganize the ruleset so it is durable, concise, scan-friendly, and directly useful to coding agents.\n")
	sb.WriteString("4. Preserve valid YAML front matter with `kind: ruleset`, `slug`, `status`, `applies_to`, and `read_policy_default`.\n")
	sb.WriteString("5. Preserve these required sections exactly: `Purpose`, `Applies When`, `Rules`, `Anti-Patterns`, `Verification`, and `Examples`.\n")
	sb.WriteString("6. Make rules specific and testable; move vague advice into concrete acceptance or verification guidance.\n")
	sb.WriteString("7. Keep the artifact pointer-loaded: do not inline it into AGENTS.md, CLAUDE.md, copilot instructions, or generated prompt bodies by default.\n")
	sb.WriteString("8. Do not create a broad policy engine or unrelated docs churn.\n")
	sb.WriteString("9. Run `kit check --project` and `kit rules list` after editing.\n\n")
	sb.WriteString("Output expectation:\n")
	sb.WriteString("- Edit only the ruleset unless the validation evidence requires a small contract/doc fix.\n")
	sb.WriteString("- Summarize changed files and verification results.\n")
	return sb.String()
}
