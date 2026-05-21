package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/templates"
)

const (
	rulesetKind                    = "ruleset"
	rulesetDirRelPath              = "docs/references/rules"
	rulesetReferenceType           = "ruleset"
	rulesetReferenceIDPrefix       = "ruleset-"
	defaultRulesetReadPolicy       = document.ReferenceReadPolicyConditional
	rulesetReferenceStatus         = document.ReferenceStatusActive
	frontendRulesetAppliesTo       = "frontend"
	frontendProfileReferenceMarker = "frontend-profile"
)

var (
	rulesAddCopy        bool
	rulesAddEditor      string
	rulesAddEvidence    bool
	rulesAddForce       bool
	rulesAddInline      bool
	rulesAddMust        bool
	rulesAddOutputOnly  bool
	rulesAddSkip        bool
	rulesAddConditional bool
	rulesAddUseVim      bool
	rulesLinkReadPolicy string
	rulesetSlugPattern  = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)
	rulesetSectionRe    = regexp.MustCompile(`(?m)^##\s+(.+)$`)
)

var rulesCmd = &cobra.Command{
	Use:     "rules",
	Aliases: []string{"rule"},
	Short:   "Manage durable repo-local rulesets",
	Long: `Create, list, and link durable repo-local rulesets.

Rulesets live under docs/references/rules/ and are loaded through feature
front matter references. They are not inlined into always-loaded instruction
files or prompt bodies by default.`,
}

var rulesAddCmd = &cobra.Command{
	Use:   "add [slug]",
	Short: "Create a durable repo-local ruleset",
	Long: `Create a durable repo-local ruleset.

With a slug argument, creates the ruleset non-interactively for scripts.
Without a slug, asks for the ruleset name, loading policy, applicability, and
rule context, then opens $EDITOR for the context by default, falls back to a
vim-compatible editor when $EDITOR is unset, and copies an agent optimization
prompt after the ruleset is saved.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRulesAdd,
}

var rulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List durable repo-local rulesets",
	Args:  cobra.NoArgs,
	RunE:  runRulesList,
}

var rulesLinkCmd = &cobra.Command{
	Use:   "link <feature> <slug>",
	Short: "Link a ruleset to a feature through canonical references",
	Args:  cobra.ExactArgs(2),
	RunE:  runRulesLink,
}

type rulesetMetadata struct {
	Kind              string   `yaml:"kind"`
	Slug              string   `yaml:"slug"`
	Status            string   `yaml:"status"`
	AppliesTo         []string `yaml:"applies_to"`
	ReadPolicyDefault string   `yaml:"read_policy_default"`
}

type rulesetDocument struct {
	Path      string
	Body      string
	Metadata  rulesetMetadata
	Sections  map[string]string
	ParseErr  error
	LineHints []string
}

type rulesetAddInput struct {
	Name              string
	Slug              string
	AppliesTo         []string
	ReadPolicyDefault string
	Context           string
}

func init() {
	addFreeTextInputFlags(rulesAddCmd, &rulesAddUseVim, &rulesAddEditor)
	addInlineTextInputFlag(rulesAddCmd, &rulesAddInline)
	rulesAddCmd.Flags().BoolVar(&rulesAddCopy, "copy", false, "copy optimization prompt to clipboard even with --output-only")
	rulesAddCmd.Flags().BoolVar(&rulesAddEvidence, "evidence", false, "set read_policy_default to evidence")
	rulesAddCmd.Flags().BoolVar(&rulesAddForce, "force", false, "overwrite an existing ruleset file")
	rulesAddCmd.Flags().BoolVar(&rulesAddMust, "must", false, "set read_policy_default to must")
	rulesAddCmd.Flags().BoolVar(&rulesAddOutputOnly, "output-only", false, "output optimization prompt text instead of copying it to the clipboard")
	rulesAddCmd.Flags().BoolVar(&rulesAddSkip, "skip", false, "set read_policy_default to skip")
	rulesAddCmd.Flags().BoolVar(&rulesAddConditional, "conditional", false, "set read_policy_default to conditional")
	rulesLinkCmd.Flags().StringVar(&rulesLinkReadPolicy, "read-policy", defaultRulesetReadPolicy, "ruleset read policy for this feature reference (must or conditional)")

	rulesCmd.AddCommand(rulesAddCmd)
	rulesCmd.AddCommand(rulesListCmd)
	rulesCmd.AddCommand(rulesLinkCmd)
	rootCmd.AddCommand(rulesCmd)
}

func runRulesAdd(cmd *cobra.Command, args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	readPolicyDefault, policyExplicit, err := selectedRulesetReadPolicy()
	if err != nil {
		return err
	}
	if rulesAddInline && (rulesAddUseVim || rulesAddEditor != "") {
		return fmt.Errorf("--inline cannot be used with --vim or --editor")
	}

	if len(args) == 0 {
		return runRulesAddInteractive(cmd, projectRoot, readPolicyDefault, policyExplicit)
	}

	if rulesAddCopy || rulesAddOutputOnly || rulesAddUseVim || rulesAddEditor != "" || rulesAddInline {
		return fmt.Errorf("--copy, --output-only, --vim, --editor, and --inline require interactive `kit rules add` with no slug")
	}

	slug := strings.TrimSpace(args[0])
	if err := validateRulesetSlug(slug); err != nil {
		return err
	}

	input := rulesetAddInput{
		Name:              slug,
		Slug:              slug,
		AppliesTo:         defaultRulesetAppliesTo(slug),
		ReadPolicyDefault: readPolicyDefault,
	}
	_, err = createRuleset(projectRoot, input)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Created ruleset %s at %s\n", slug, rulesetTarget(slug))
	return err
}

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
	return printRulesetList(cmd.OutOrStdout(), projectRoot, rulesets)
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

func validateRulesetSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("ruleset slug cannot be empty")
	}
	if !rulesetSlugPattern.MatchString(slug) {
		return fmt.Errorf("ruleset slug must be lowercase kebab-case")
	}
	return nil
}

func rulesetPath(projectRoot, slug string) string {
	return filepath.Join(projectRoot, filepath.FromSlash(rulesetTarget(slug)))
}

func rulesetTarget(slug string) string {
	return filepath.ToSlash(filepath.Join(rulesetDirRelPath, slug+".md"))
}

func defaultRulesetAppliesTo(slug string) []string {
	first, _, ok := strings.Cut(slug, "-")
	if ok && first != "" {
		return []string{first}
	}
	return []string{slug}
}

func listRulesets(projectRoot string) ([]rulesetDocument, error) {
	dir := filepath.Join(projectRoot, filepath.FromSlash(rulesetDirRelPath))
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read rulesets directory: %w", err)
	}

	var rulesets []rulesetDocument
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		ruleset, err := parseRulesetFile(path)
		if err != nil {
			return nil, err
		}
		if issues := validateRulesetDocument(ruleset, strings.TrimSuffix(entry.Name(), ".md")); len(issues) > 0 {
			return nil, fmt.Errorf("invalid ruleset %s: %s", filepath.ToSlash(path), strings.Join(issues, "; "))
		}
		rulesets = append(rulesets, ruleset)
	}

	sort.SliceStable(rulesets, func(i, j int) bool {
		return rulesets[i].Metadata.Slug < rulesets[j].Metadata.Slug
	})
	return rulesets, nil
}

func printRulesetList(w io.Writer, projectRoot string, rulesets []rulesetDocument) error {
	writer := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(writer, "SLUG\tPATH\tSTATUS\tAPPLIES_TO"); err != nil {
		return err
	}
	for _, ruleset := range rulesets {
		relPath, err := filepath.Rel(projectRoot, ruleset.Path)
		if err != nil {
			relPath = ruleset.Path
		}
		if _, err := fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%s\n",
			ruleset.Metadata.Slug,
			filepath.ToSlash(relPath),
			ruleset.Metadata.Status,
			strings.Join(ruleset.Metadata.AppliesTo, ","),
		); err != nil {
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to render ruleset list: %w", err)
	}
	return nil
}

func loadRuleset(projectRoot, slug string) (rulesetDocument, error) {
	path := rulesetPath(projectRoot, slug)
	if !document.Exists(path) {
		return rulesetDocument{}, fmt.Errorf("ruleset %q not found at %s", slug, rulesetTarget(slug))
	}
	return parseRulesetFile(path)
}

func parseRulesetFile(path string) (rulesetDocument, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return rulesetDocument{}, fmt.Errorf("failed to read %s: %w", path, err)
	}
	return parseRuleset(string(content), path), nil
}

func parseRuleset(content, path string) rulesetDocument {
	raw, body, err := splitRulesetFrontMatter(content)
	ruleset := rulesetDocument{
		Path:     path,
		Body:     body,
		Sections: rulesetSections(body),
		ParseErr: err,
	}
	if err != nil {
		return ruleset
	}
	if strings.TrimSpace(raw) == "" {
		ruleset.ParseErr = fmt.Errorf("front matter is empty")
		return ruleset
	}
	if err := yaml.Unmarshal([]byte(raw), &ruleset.Metadata); err != nil {
		ruleset.ParseErr = fmt.Errorf("failed to parse front matter: %w", err)
	}
	return ruleset
}

func splitRulesetFrontMatter(content string) (string, string, error) {
	lines := strings.SplitAfter(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", content, fmt.Errorf("missing YAML front matter")
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return strings.Join(lines[1:i], ""), strings.Join(lines[i+1:], ""), nil
		}
	}
	return "", content, fmt.Errorf("missing closing front matter delimiter")
}

func rulesetSections(body string) map[string]string {
	sections := make(map[string]string)
	matches := rulesetSectionRe.FindAllStringSubmatchIndex(body, -1)
	for i, match := range matches {
		name := strings.ToUpper(strings.TrimSpace(body[match[2]:match[3]]))
		start := match[1]
		end := len(body)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		sections[name] = strings.TrimSpace(body[start:end])
	}
	return sections
}

func validateRulesetDocument(ruleset rulesetDocument, expectedSlug string) []string {
	var issues []string
	if ruleset.ParseErr != nil {
		issues = append(issues, ruleset.ParseErr.Error())
		return issues
	}
	if ruleset.Metadata.Kind != rulesetKind {
		issues = append(issues, "front matter kind must be ruleset")
	}
	if ruleset.Metadata.Slug == "" {
		issues = append(issues, "front matter slug cannot be empty")
	} else if err := validateRulesetSlug(ruleset.Metadata.Slug); err != nil {
		issues = append(issues, err.Error())
	} else if expectedSlug != "" && ruleset.Metadata.Slug != expectedSlug {
		issues = append(issues, fmt.Sprintf("front matter slug %q does not match file slug %q", ruleset.Metadata.Slug, expectedSlug))
	}
	if ruleset.Metadata.Status == "" || !validRulesetStatus(ruleset.Metadata.Status) {
		issues = append(issues, "front matter status must be active, optional, or stale")
	}
	if len(ruleset.Metadata.AppliesTo) == 0 {
		issues = append(issues, "front matter applies_to must contain at least one entry")
	}
	for _, appliesTo := range ruleset.Metadata.AppliesTo {
		if err := validateRulesetSlug(appliesTo); err != nil {
			issues = append(issues, fmt.Sprintf("front matter applies_to entry %q is invalid", appliesTo))
		}
	}
	if ruleset.Metadata.ReadPolicyDefault == "" || !validRulesetReadPolicy(ruleset.Metadata.ReadPolicyDefault) {
		issues = append(issues, "front matter read_policy_default must be must, conditional, evidence, or skip")
	}
	for _, section := range requiredRulesetSections() {
		content, ok := ruleset.Sections[strings.ToUpper(section)]
		if !ok {
			issues = append(issues, fmt.Sprintf("missing required section ## %s", section))
			continue
		}
		if !meaningfulSectionContent(content) {
			issues = append(issues, fmt.Sprintf("required section ## %s is empty or placeholder-only", section))
		}
	}
	return issues
}

func requiredRulesetSections() []string {
	return []string{"Purpose", "Applies When", "Rules", "Anti-Patterns", "Verification", "Examples"}
}

func validRulesetStatus(value string) bool {
	switch value {
	case document.ReferenceStatusActive, document.ReferenceStatusOptional, document.ReferenceStatusStale:
		return true
	default:
		return false
	}
}

func validRulesetReadPolicy(value string) bool {
	switch value {
	case document.ReferenceReadPolicyMust,
		document.ReferenceReadPolicyConditional,
		document.ReferenceReadPolicyEvidence,
		document.ReferenceReadPolicySkip:
		return true
	default:
		return false
	}
}

func rulesetReference(slug, readPolicy string) document.MetadataReference {
	return document.MetadataReference{
		ID:         rulesetReferenceIDPrefix + slug,
		Name:       "Ruleset: " + slug,
		Type:       rulesetReferenceType,
		Target:     rulesetTarget(slug),
		Relation:   document.ReferenceRelationGuides,
		ReadPolicy: readPolicy,
		UsedFor:    "load durable " + slug + " rules only when relevant to the current decision",
		Status:     rulesetReferenceStatus,
	}
}

func rulesetLinkTargetDoc(feat *feature.Feature) (string, document.DocumentType, error) {
	candidates := []struct {
		name    string
		docType document.DocumentType
	}{
		{name: "SPEC.md", docType: document.TypeSpec},
		{name: "PLAN.md", docType: document.TypePlan},
		{name: "BRAINSTORM.md", docType: document.TypeBrainstorm},
		{name: "TASKS.md", docType: document.TypeTasks},
	}
	for _, candidate := range candidates {
		path := filepath.Join(feat.Path, candidate.name)
		if document.Exists(path) {
			return path, candidate.docType, nil
		}
	}
	return "", "", fmt.Errorf("feature %q has no document that can hold ruleset references", feat.Slug)
}

func featureRulesetReferenceErrors(projectRoot string, doc *document.Document) []string {
	var errors []string
	for _, reference := range doc.References() {
		if !isRulesetReference(reference) {
			continue
		}
		path, ok := rulesetReferencePath(projectRoot, reference.Target)
		if !ok {
			errors = append(errors, fmt.Sprintf("%s: ruleset reference %q must target docs/references/rules/<slug>.md", doc.Path, reference.Name))
			continue
		}
		if !document.Exists(path) {
			errors = append(errors, fmt.Sprintf("%s: ruleset reference %q points to missing file %s", doc.Path, reference.Name, filepath.ToSlash(strings.TrimSpace(reference.Target))))
			continue
		}
		ruleset, err := parseRulesetFile(path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to read ruleset reference %q: %v", doc.Path, reference.Name, err))
			continue
		}
		expectedSlug := strings.TrimSuffix(filepath.Base(path), ".md")
		if issues := validateRulesetDocument(ruleset, expectedSlug); len(issues) > 0 {
			errors = append(errors, fmt.Sprintf("%s: ruleset reference %q points to invalid ruleset: %s", doc.Path, reference.Name, strings.Join(issues, "; ")))
		}
	}
	return errors
}

func rulesetReferencePath(projectRoot, target string) (string, bool) {
	cleanTarget := filepath.Clean(filepath.FromSlash(strings.TrimSpace(target)))
	if cleanTarget == "." || strings.TrimSpace(target) == "" {
		return "", false
	}
	var absPath string
	var relPath string
	if filepath.IsAbs(cleanTarget) {
		absPath = cleanTarget
		rel, err := filepath.Rel(projectRoot, absPath)
		if err != nil {
			return "", false
		}
		relPath = rel
	} else {
		relPath = cleanTarget
		absPath = filepath.Join(projectRoot, relPath)
	}
	relSlash := filepath.ToSlash(filepath.Clean(relPath))
	if !strings.HasPrefix(relSlash, rulesetDirRelPath+"/") || !strings.HasSuffix(relSlash, ".md") {
		return "", false
	}
	return absPath, true
}

func isRulesetReference(reference document.MetadataReference) bool {
	return strings.EqualFold(strings.TrimSpace(reference.Type), rulesetReferenceType) ||
		strings.HasPrefix(filepath.ToSlash(strings.TrimSpace(reference.Target)), rulesetDirRelPath+"/")
}

func auditRulesets(projectRoot string) []reconcileFinding {
	dir := filepath.Join(projectRoot, filepath.FromSlash(rulesetDirRelPath))
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return []reconcileFinding{newFinding(
			reconcileSeverityError,
			dir,
			"failed to read ruleset directory",
			templateSource(projectRoot),
			"fix docs/references/rules/ permissions before validating rulesets",
			[]string{fmt.Sprintf("ls -la %s", dir)},
		)}
	}

	var findings []reconcileFinding
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		ruleset, err := parseRulesetFile(path)
		if err != nil {
			findings = append(findings, newFinding(
				reconcileSeverityError,
				path,
				"failed to read ruleset document",
				templateSource(projectRoot),
				"make the ruleset readable and retry validation",
				[]string{fmt.Sprintf("sed -n '1,220p' %s", path)},
			))
			continue
		}
		expectedSlug := strings.TrimSuffix(entry.Name(), ".md")
		for _, issue := range validateRulesetDocument(ruleset, expectedSlug) {
			findings = append(findings, newFinding(
				reconcileSeverityError,
				path,
				"ruleset document issue: "+issue,
				templateSource(projectRoot),
				"update the ruleset front matter and required sections to match the Kit ruleset contract",
				[]string{fmt.Sprintf("sed -n '1,220p' %s", path)},
			))
		}
	}
	return findings
}

func auditRulesetReferences(projectRoot string, path string, doc *document.Document) []reconcileFinding {
	var findings []reconcileFinding
	for _, issue := range featureRulesetReferenceErrors(projectRoot, doc) {
		findings = append(findings, newFinding(
			reconcileSeverityError,
			path,
			issue,
			templateSource(projectRoot),
			"create the referenced ruleset with `kit rules add <slug>` or update the feature reference target",
			[]string{
				fmt.Sprintf("sed -n '1,90p' %s", path),
				fmt.Sprintf("ls %s", filepath.Join(projectRoot, filepath.FromSlash(rulesetDirRelPath))),
			},
		))
	}
	return findings
}

func auditActiveFrontendRulesetAdvisory(projectRoot string, feat *feature.Feature) []reconcileFinding {
	if feat == nil || feat.Paused || feat.Phase == feature.PhaseComplete {
		return nil
	}
	if !featureHasActiveFrontendProfileDependency(feat.Path) {
		return nil
	}
	if featureHasActiveRulesetForApplicability(projectRoot, feat, frontendRulesetAppliesTo) {
		return nil
	}

	return []reconcileFinding{newFinding(
		reconcileSeverityWarning,
		feat.Path,
		"active frontend feature has no active frontend ruleset reference",
		templateSource(projectRoot),
		"create or link a frontend ruleset with `kit rules add frontend-ui` and `kit rules link "+feat.Slug+" frontend-ui --read-policy conditional` if durable frontend rules apply",
		[]string{
			fmt.Sprintf("rg -n \"type: %s|%s|%s\" %s", rulesetReferenceType, rulesetDirRelPath, frontendProfileReferenceMarker, feat.Path),
			fmt.Sprintf("find %s -maxdepth 1 -type f -name '*.md' -print", filepath.Join(projectRoot, filepath.FromSlash(rulesetDirRelPath))),
		},
	)}
}

func featureHasActiveRulesetForApplicability(projectRoot string, feat *feature.Feature, appliesTo string) bool {
	for _, source := range rulesetReferenceSources(feat.Path) {
		if !document.Exists(source.path) {
			continue
		}
		doc, err := document.ParseFile(source.path, source.docType)
		if err != nil {
			continue
		}
		for _, reference := range doc.References() {
			if !activeRulesetReference(reference) {
				continue
			}
			path, ok := rulesetReferencePath(projectRoot, reference.Target)
			if !ok || !document.Exists(path) {
				continue
			}
			ruleset, err := parseRulesetFile(path)
			if err != nil || len(validateRulesetDocument(ruleset, strings.TrimSuffix(filepath.Base(path), ".md"))) > 0 {
				continue
			}
			if ruleset.Metadata.Status == document.ReferenceStatusActive && slices.Contains(ruleset.Metadata.AppliesTo, appliesTo) {
				return true
			}
		}
	}
	return false
}

func activeRulesetReference(reference document.MetadataReference) bool {
	return isRulesetReference(reference) &&
		reference.Status == document.ReferenceStatusActive &&
		reference.ReadPolicy != document.ReferenceReadPolicySkip
}

func rulesetReferenceSources(featurePath string) []struct {
	path    string
	docType document.DocumentType
} {
	return []struct {
		path    string
		docType document.DocumentType
	}{
		{path: filepath.Join(featurePath, "BRAINSTORM.md"), docType: document.TypeBrainstorm},
		{path: filepath.Join(featurePath, "SPEC.md"), docType: document.TypeSpec},
		{path: filepath.Join(featurePath, "PLAN.md"), docType: document.TypePlan},
		{path: filepath.Join(featurePath, "TASKS.md"), docType: document.TypeTasks},
	}
}
