package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

var specCopy bool
var specEditor string
var specOutputOnly bool
var specUseVim bool

var specCmd = &cobra.Command{
	Use:   "spec [feature]",
	Short: "Create or open a feature specification",
	Long: `Create a new feature specification or open an existing one.

Creates:
  - Feature directory (e.g., docs/specs/0001-my-feature/)
  - SPEC.md with required sections and placeholder comments

Updates PROJECT_PROGRESS_SUMMARY.md after creation.

If no feature is specified, shows an interactive selection of existing
features with BRAINSTORM.md or SPEC.md.

Modes:
  Default:        Copy the generated prompt to the clipboard and show status (non-interactive)
  --interactive:  Prompt user for spec details, then output ready-to-use prompt
  --template:     Output empty template without interactive questions (deprecated, same as default)

Flags:
  --output-only:  Output the raw prompt to stdout instead of copying it to the clipboard
  --copy:         Copy prompt to clipboard (mainly useful with --output-only)
  --interactive:  Force interactive prompts even when stdin is not a terminal
  --vim:          Open free-text responses in a vim-compatible editor`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSpec,
}

func init() {
	addFreeTextInputFlags(specCmd, &specUseVim, &specEditor)
	specCmd.Flags().Bool("template", false, "(deprecated) output empty template and prompt without interactive questions")
	specCmd.Flags().Bool("interactive", false, "prompt user for spec details interactively")
	specCmd.Flags().BoolVar(&specCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	specCmd.Flags().BoolVar(&specOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	addPromptOnlyFlag(specCmd)
	rootCmd.AddCommand(specCmd)
}

func runSpec(cmd *cobra.Command, args []string) error {
	specTemplateOnly, _ := cmd.Flags().GetBool("template")
	specInteractive, _ := cmd.Flags().GetBool("interactive")
	outputOnly, _ := cmd.Flags().GetBool("output-only")
	promptOnly := promptOnlyEnabled(cmd)

	// find project root
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	specsDir := cfg.SpecsPath(projectRoot)

	// ensure specs directory exists
	if err := ensureDir(specsDir); err != nil {
		return err
	}

	if promptOnly {
		if specTemplateOnly || specInteractive || specUseVim || specEditor != "" {
			return fmt.Errorf("--prompt-only cannot be used with --template, --interactive, --vim, or --editor")
		}
		return runSpecPromptOnly(args, projectRoot, cfg, outputOnly)
	}

	var (
		feat    *feature.Feature
		created bool
	)

	if len(args) == 0 {
		feat, err = selectFeatureForSpec(specsDir)
		if err != nil {
			return err
		}
	} else {
		featureRef := args[0]

		// create or find feature
		feat, created, err = feature.EnsureExists(cfg, specsDir, featureRef)
		if err != nil {
			return err
		}
	}

	if !outputOnly {
		if created {
			fmt.Printf("📁 Created feature directory: %s\n", feat.DirName)
		} else {
			fmt.Printf("📁 Using existing feature: %s\n", feat.DirName)
		}
	}

	// create SPEC.md if it doesn't exist
	specPath := filepath.Join(feat.Path, "SPEC.md")
	if !document.Exists(specPath) {
		if err := document.Write(specPath, templates.Spec); err != nil {
			return fmt.Errorf("failed to create SPEC.md: %w", err)
		}
		if !outputOnly {
			fmt.Println("  ✓ Created SPEC.md")
		}
	} else if !outputOnly {
		fmt.Println("  ✓ SPEC.md already exists")
	}

	// determine if we should run interactive mode
	// default is non-interactive (template mode), unless --interactive is explicitly set
	isInteractive := specInteractive && !specTemplateOnly
	inputCfg := newFreeTextInputConfig(specUseVim, specEditor)
	if inputCfg.usesEditor() && !isInteractive {
		return fmt.Errorf("--vim and --editor require --interactive")
	}

	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	if !outputOnly && document.Exists(brainstormPath) {
		fmt.Println("  ✓ Found BRAINSTORM.md")
	}

	// update PROJECT_PROGRESS_SUMMARY.md
	if err := rollup.Update(projectRoot, cfg); err != nil {
		if !outputOnly {
			fmt.Printf("  ⚠ Could not update PROJECT_PROGRESS_SUMMARY.md: %v\n", err)
		}
	} else if !outputOnly {
		fmt.Println("  ✓ Updated PROJECT_PROGRESS_SUMMARY.md")
	}

	if !outputOnly {
		fmt.Printf("\n✅ Feature '%s' ready!\n", feat.Slug)
	}

	if isInteractive {
		// interactive mode: gather details and compile prompt
		return runSpecInteractive(specPath, brainstormPath, feat, projectRoot, cfg, inputCfg, outputOnly)
	}

	// template mode: output the template and instructions
	return runSpecTemplate(specPath, brainstormPath, feat.Slug, projectRoot, cfg, outputOnly)
}

func runSpecPromptOnly(args []string, projectRoot string, cfg *config.Config, outputOnly bool) error {
	specsDir := cfg.SpecsPath(projectRoot)

	var (
		feat *feature.Feature
		err  error
	)

	if len(args) == 0 {
		feat, err = selectFeatureForSpecPromptOnly(specsDir)
		if err != nil {
			return err
		}
	} else {
		feat, err = feature.Resolve(specsDir, args[0])
		if err != nil {
			return fmt.Errorf("feature '%s' not found: %w", args[0], err)
		}
	}

	specPath := filepath.Join(feat.Path, "SPEC.md")
	if !document.Exists(specPath) {
		return fmt.Errorf("SPEC.md not found. Run 'kit spec %s' first", feat.Slug)
	}

	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	return runSpecTemplate(specPath, brainstormPath, feat.Slug, projectRoot, cfg, outputOnly)
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// selectFeatureForSpec shows an interactive numbered list of existing
// features that have BRAINSTORM.md or SPEC.md.
func selectFeatureForSpec(specsDir string) (*feature.Feature, error) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	var candidates []feature.Feature
	for _, f := range features {
		brainstormPath := filepath.Join(f.Path, "BRAINSTORM.md")
		specPath := filepath.Join(f.Path, "SPEC.md")
		if document.Exists(brainstormPath) || document.Exists(specPath) {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no brainstorms or specifications found\n\nRun 'kit brainstorm' or 'kit spec <feature>' to start a feature")
	}

	fmt.Println()
	fmt.Println(whiteBold + "Select a feature to continue into spec:" + reset)
	fmt.Println()
	for i, f := range candidates {
		label := f.DirName
		if document.Exists(filepath.Join(f.Path, "BRAINSTORM.md")) && !document.Exists(filepath.Join(f.Path, "SPEC.md")) {
			label += " (brainstorm)"
		}
		fmt.Printf("  [%d] %s\n", i+1, label)
	}
	fmt.Println()
	fmt.Print(whiteBold + "Enter number: " + reset)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	num, err := strconv.Atoi(input)
	if err != nil || num < 1 || num > len(candidates) {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	selected := candidates[num-1]
	return &selected, nil
}

func selectFeatureForSpecPromptOnly(specsDir string) (*feature.Feature, error) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	var candidates []feature.Feature
	for _, f := range features {
		if document.Exists(filepath.Join(f.Path, "SPEC.md")) {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no specifications available to regenerate prompts for\n\nRun 'kit spec <feature>' first")
	}

	fmt.Println()
	fmt.Println(whiteBold + "Select a feature to regenerate the spec prompt for:" + reset)
	fmt.Println()
	for i, f := range candidates {
		fmt.Printf("  [%d] %s (%s)\n", i+1, f.DirName, f.Phase)
	}
	fmt.Println()
	fmt.Print(whiteBold + "Enter number: " + reset)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	num, err := strconv.Atoi(input)
	if err != nil || num < 1 || num > len(candidates) {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	selected := candidates[num-1]
	return &selected, nil
}

// isTerminal checks if stdin is connected to a terminal
func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// specAnswers holds the user's responses to interactive prompts
type specAnswers struct {
	Problem      string
	Goals        string
	NonGoals     string
	Users        string
	Requirements string
	Acceptance   string
	EdgeCases    string
}

func normalizeSpecAnswer(raw string) string {
	return strings.TrimSpace(raw)
}

func appendRepoInstructionContextRows(sb *strings.Builder, projectRoot string, cfg *config.Config) {
	for _, path := range repoInstructionPaths(projectRoot, cfg) {
		label := filepath.Base(path)
		if strings.Contains(path, ".github/copilot-instructions.md") {
			label = "COPILOT"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s |\n", label, path))
	}
}

func appendSpecSkillDiscoveryContextRows(sb *strings.Builder, projectRoot string, cfg *config.Config) {
	skillsDir := cfg.SkillsPath(projectRoot)
	globalInputs := globalSkillDiscoveryInputs()

	sb.WriteString(fmt.Sprintf("| Canonical Skills Root | %s/*/SKILL.md |\n", skillsDir))
	sb.WriteString(fmt.Sprintf("| Claude Global | %s |\n", globalInputs[0]))
	sb.WriteString(fmt.Sprintf("| Codex Global AGENTS | %s |\n", globalInputs[1]))
	sb.WriteString(fmt.Sprintf("| Codex Global Instructions | %s |\n", globalInputs[2]))
	sb.WriteString(fmt.Sprintf("| Codex Global Skills | %s |\n", globalInputs[3]))
}

func appendRepoInstructionReadStep(
	sb *strings.Builder,
	step int,
	projectRoot string,
	cfg *config.Config,
) int {
	sb.WriteString(fmt.Sprintf("%d. Read the repository instruction files first:\n", step))
	for _, path := range repoInstructionPaths(projectRoot, cfg) {
		sb.WriteString(fmt.Sprintf("   - `%s`\n", path))
	}
	return step + 1
}

func appendSpecSkillDiscoveryStep(
	sb *strings.Builder,
	step int,
	projectRoot string,
	cfg *config.Config,
	specPath string,
) int {
	skillsDir := cfg.SkillsPath(projectRoot)
	globalInputs := globalSkillDiscoveryInputs()

	sb.WriteString(fmt.Sprintf("%d. Perform a skills discovery phase before treating SPEC.md as complete:\n", step))
	sb.WriteString(fmt.Sprintf("   - inspect repo-local canonical skills under `%s/*/SKILL.md`\n", skillsDir))
	sb.WriteString("   - inspect documented global inputs:\n")
	for _, path := range globalInputs {
		sb.WriteString(fmt.Sprintf("     - `%s`\n", path))
	}
	sb.WriteString("   - choose the minimal relevant set of skills for this feature\n")
	sb.WriteString(fmt.Sprintf("   - write the selected skills into the `## SKILLS` table in `%s`\n", specPath))
	sb.WriteString("   - if no additional skills apply, keep the required `none | n/a | n/a | no additional skills required | no` row\n")
	sb.WriteString("   - do not use `.claude/skills` as canonical discovery input\n")
	return step + 1
}

func appendSpecDependencyInventoryStep(
	sb *strings.Builder,
	step int,
	specPath string,
	brainstormPath string,
	hasBrainstorm bool,
) int {
	sb.WriteString(fmt.Sprintf("%d. Populate or refresh the `## DEPENDENCIES` table in `%s` before sign-off:\n", step, specPath))
	if hasBrainstorm {
		sb.WriteString(fmt.Sprintf("   - carry forward still-relevant dependencies from `%s`\n", brainstormPath))
	}
	sb.WriteString("   - keep `## SKILLS` focused on execution-time agent skills and track broader supporting inputs in `## DEPENDENCIES`\n")
	sb.WriteString("   - include skills, MCP tools, repo docs, design refs, APIs, libraries, datasets, assets, and other resources that shaped the feature definition\n")
	sb.WriteString("   - use the columns `Dependency`, `Type`, `Location`, `Used For`, and `Status`\n")
	sb.WriteString("   - `Status` must be one of `active`, `optional`, or `stale`\n")
	sb.WriteString("   - for Figma or MCP-driven design dependencies, store the exact design URL or file/node reference in `Location`\n")
	sb.WriteString("   - if a dependency influenced decisions but is no longer current, keep it in the table with `Status` = `stale`\n")
	sb.WriteString("   - if no additional dependencies apply, keep the default `none` row\n")
	return step + 1
}

// readLineRL reads from the readline instance, returning empty string on EOF/interrupt.
func readLineRL(rl *readline.Instance) string {
	line, err := rl.Readline()
	if err != nil {
		if err == readline.ErrInterrupt || err == io.EOF {
			return ""
		}
		return ""
	}
	return normalizeSpecAnswer(line)
}

// runSpecInteractive prompts the user for each SPEC section and compiles a ready-to-use prompt.
func runSpecInteractive(
	specPath, brainstormPath string,
	feat *feature.Feature,
	projectRoot string,
	cfg *config.Config,
	inputCfg freeTextInputConfig,
	outputOnly bool,
) error {
	if inputCfg.usesEditor() {
		return runSpecInteractiveWithEditor(specPath, brainstormPath, feat, projectRoot, cfg, inputCfg, outputOnly)
	}

	return runSpecInteractiveWithReadline(specPath, brainstormPath, feat, projectRoot, cfg, outputOnly)
}

func runSpecInteractiveWithReadline(specPath, brainstormPath string, feat *feature.Feature, projectRoot string, cfg *config.Config, outputOnly bool) error {
	rl, err := newMultilineReadline()
	if err != nil {
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer closeMultilineReadline(rl)

	fmt.Println("\n" + dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(whiteBold + "📝 Interactive Spec Builder" + reset)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println()

	fmt.Println(dim + "Answer the following questions to generate a complete prompt for your coding agent." + reset)
	fmt.Println(dim + "Use ←/→ arrow keys to move through your text and correct mistakes." + reset)
	fmt.Println(dim + "Press Enter to continue; use Shift+Enter or Ctrl+J to add newlines." + reset)
	fmt.Println(dim + "Consecutive blank lines are preserved." + reset)
	fmt.Println(dim + "Press Enter on an empty response to skip a question." + reset)
	if document.Exists(brainstormPath) {
		fmt.Println(dim + "Existing brainstorm research will also be referenced in the generated prompt." + reset)
	}
	fmt.Println()

	// reset prompt for question inputs
	rl.SetPrompt(whiteBold + "   > " + reset)

	answers := specAnswers{}

	// PROBLEM
	fmt.Println(spec + "1. PROBLEM" + reset + " - What problem does this feature solve?")
	fmt.Println(dim + "   Example: Users cannot export their data in CSV format" + reset)
	answers.Problem = readLineRL(rl)

	// GOALS
	fmt.Println()
	fmt.Println(spec + "2. GOALS" + reset + " - What are the measurable outcomes? (comma-separated)")
	fmt.Println(dim + "   Example: Export completes in <5s, supports 100k+ rows, CSV is RFC-compliant" + reset)
	answers.Goals = readLineRL(rl)

	// NON-GOALS
	fmt.Println()
	fmt.Println(spec + "3. NON-GOALS" + reset + " - What is explicitly out of scope?")
	fmt.Println(dim + "   Example: Excel format, scheduled exports, email delivery" + reset)
	answers.NonGoals = readLineRL(rl)

	// USERS
	fmt.Println()
	fmt.Println(spec + "4. USERS" + reset + " - Who will use this feature?")
	fmt.Println(dim + "   Example: Admin users, API consumers, data analysts" + reset)
	answers.Users = readLineRL(rl)

	// REQUIREMENTS
	fmt.Println()
	fmt.Println(spec + "5. REQUIREMENTS" + reset + " - What must be true for this feature to be complete?")
	fmt.Println(dim + "   Example: Must handle Unicode, must include headers, must stream large files" + reset)
	answers.Requirements = readLineRL(rl)

	// ACCEPTANCE
	fmt.Println()
	fmt.Println(spec + "6. ACCEPTANCE" + reset + " - How do we verify the feature works?")
	fmt.Println(dim + "   Example: Unit tests pass, integration tests cover edge cases, manual QA sign-off" + reset)
	answers.Acceptance = readLineRL(rl)

	// EDGE-CASES
	fmt.Println()
	fmt.Println(spec + "7. EDGE-CASES" + reset + " - What unusual scenarios must be handled?")
	fmt.Println(dim + "   Example: Empty dataset, special characters in data, network timeout during export" + reset)
	answers.EdgeCases = readLineRL(rl)

	fmt.Println()

	// generate the compiled prompt
	return outputCompiledPrompt(specPath, brainstormPath, feat.Slug, projectRoot, cfg, &answers, outputOnly)
}

func runSpecInteractiveWithEditor(
	specPath, brainstormPath string,
	feat *feature.Feature,
	projectRoot string,
	cfg *config.Config,
	inputCfg freeTextInputConfig,
	outputOnly bool,
) error {
	fmt.Println("\n" + dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(whiteBold + "📝 Interactive Spec Builder" + reset)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println()

	fmt.Println(dim + "Answer the following questions to generate a complete prompt for your coding agent." + reset)
	fmt.Println(dim + "A vim-compatible editor will open for each free-text response." + reset)
	fmt.Println(dim + "Save and quit to submit. Quit without save to skip that question." + reset)
	if document.Exists(brainstormPath) {
		fmt.Println(dim + "Existing brainstorm research will also be referenced in the generated prompt." + reset)
	}
	fmt.Println()

	answers := specAnswers{}
	var err error

	fmt.Println(spec + "1. PROBLEM" + reset + " - What problem does this feature solve?")
	fmt.Println(dim + "   Example: Users cannot export their data in CSV format" + reset)
	answers.Problem, err = readEditorText(inputCfg, "problem", true)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(spec + "2. GOALS" + reset + " - What are the measurable outcomes? (comma-separated)")
	fmt.Println(dim + "   Example: Export completes in <5s, supports 100k+ rows, CSV is RFC-compliant" + reset)
	answers.Goals, err = readEditorText(inputCfg, "goals", true)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(spec + "3. NON-GOALS" + reset + " - What is explicitly out of scope?")
	fmt.Println(dim + "   Example: Excel format, scheduled exports, email delivery" + reset)
	answers.NonGoals, err = readEditorText(inputCfg, "non-goals", true)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(spec + "4. USERS" + reset + " - Who will use this feature?")
	fmt.Println(dim + "   Example: Admin users, API consumers, data analysts" + reset)
	answers.Users, err = readEditorText(inputCfg, "users", true)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(spec + "5. REQUIREMENTS" + reset + " - What must be true for this feature to be complete?")
	fmt.Println(dim + "   Example: Must handle Unicode, must include headers, must stream large files" + reset)
	answers.Requirements, err = readEditorText(inputCfg, "requirements", true)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(spec + "6. ACCEPTANCE" + reset + " - How do we verify the feature works?")
	fmt.Println(dim + "   Example: Unit tests pass, integration tests cover edge cases, manual QA sign-off" + reset)
	answers.Acceptance, err = readEditorText(inputCfg, "acceptance", true)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(spec + "7. EDGE-CASES" + reset + " - What unusual scenarios must be handled?")
	fmt.Println(dim + "   Example: Empty dataset, special characters in data, network timeout during export" + reset)
	answers.EdgeCases, err = readEditorText(inputCfg, "edge-cases", true)
	if err != nil {
		return err
	}

	fmt.Println()

	return outputCompiledPrompt(specPath, brainstormPath, feat.Slug, projectRoot, cfg, &answers, outputOnly)
}

func outputCompiledPrompt(specPath, brainstormPath, featureSlug, projectRoot string, cfg *config.Config, answers *specAnswers, outputOnly bool) error {
	goalPct := cfg.GoalPercentage
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	hasBrainstorm := document.Exists(brainstormPath)

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`You MUST update the specification file at:
**File**: %s
**Feature**: %s
**Project Root**: %s

## Context Provided by User

`, specPath, featureSlug, projectRoot))

	// output user-provided context
	if answers.Problem != "" {
		sb.WriteString(fmt.Sprintf("**PROBLEM**: %s\n\n", answers.Problem))
	}
	if answers.Goals != "" {
		sb.WriteString(fmt.Sprintf("**GOALS**: %s\n\n", answers.Goals))
	}
	if answers.NonGoals != "" {
		sb.WriteString(fmt.Sprintf("**NON-GOALS**: %s\n\n", answers.NonGoals))
	}
	if answers.Users != "" {
		sb.WriteString(fmt.Sprintf("**USERS**: %s\n\n", answers.Users))
	}
	if answers.Requirements != "" {
		sb.WriteString(fmt.Sprintf("**REQUIREMENTS**: %s\n\n", answers.Requirements))
	}
	if answers.Acceptance != "" {
		sb.WriteString(fmt.Sprintf("**ACCEPTANCE**: %s\n\n", answers.Acceptance))
	}
	if answers.EdgeCases != "" {
		sb.WriteString(fmt.Sprintf("**EDGE-CASES**: %s\n\n", answers.EdgeCases))
	}

	// check if any answers were provided
	hasContext := answers.Problem != "" || answers.Goals != "" || answers.NonGoals != "" ||
		answers.Users != "" || answers.Requirements != "" || answers.Acceptance != "" ||
		answers.EdgeCases != ""

	sb.WriteString("## Context Docs (read first)\n")
	sb.WriteString("| File | Purpose |\n")
	sb.WriteString("|------|----------|\n")
	sb.WriteString(fmt.Sprintf("| CONSTITUTION | %s |\n", constitutionPath))
	appendRepoInstructionContextRows(&sb, projectRoot, cfg)
	if hasBrainstorm {
		sb.WriteString(fmt.Sprintf("| BRAINSTORM | %s |\n", brainstormPath))
	}
	sb.WriteString(fmt.Sprintf("| SPEC | %s |\n", specPath))
	appendSpecSkillDiscoveryContextRows(&sb, projectRoot, cfg)
	sb.WriteString(fmt.Sprintf("| Project Root | %s |\n\n", projectRoot))

	sb.WriteString("## Your Task\n\n")
	sb.WriteString(fmt.Sprintf("1. Read CONSTITUTION.md (file: %s) to understand project constraints and principles\n", constitutionPath))
	step := appendRepoInstructionReadStep(&sb, 2, projectRoot, cfg)
	if hasBrainstorm {
		sb.WriteString(fmt.Sprintf("%d. Read BRAINSTORM.md (file: %s) and treat it as upstream research context\n", step, brainstormPath))
		step++
	}
	sb.WriteString(fmt.Sprintf("%d. Read the current SPEC.md (file: %s) and understand the required sections\n", step, specPath))
	step++
	sb.WriteString(fmt.Sprintf("%d. Analyze the codebase at %s to understand existing patterns\n", step, projectRoot))

	questionStart := step + 1

	if hasContext {
		sb.WriteString(fmt.Sprintf(
			"%d. **IMMEDIATELY write all context above into the SPEC.md file at %s** — do NOT ask questions before doing this\n",
			questionStart,
			specPath,
		))
		questionStart++
	}

	questionStart = appendSpecSkillDiscoveryStep(&sb, questionStart, projectRoot, cfg, specPath)
	questionStart = appendSpecDependencyInventoryStep(&sb, questionStart, specPath, brainstormPath, hasBrainstorm)

	questionStart = appendNumberedSteps(
		&sb,
		questionStart,
		clarificationLoopSteps(
			goalPct,
			fmt.Sprintf(
				"Reassess, save your updates to %s, and continue with additional "+
					"batches of up to 10 questions until the specification is precise "+
					"enough to produce a correct, production-quality solution",
				specPath,
			),
		),
	)

	if hasContext {
		sb.WriteString(fmt.Sprintf(
			"%d. Continue refining each section of SPEC.md as you learn more:\n",
			questionStart,
		))
	} else {
		sb.WriteString(fmt.Sprintf(
			"%d. **Write your findings directly to %s** as you fill in each section:\n",
			questionStart,
			specPath,
		))
	}

	if hasBrainstorm {
		sb.WriteString("   - Carry forward validated findings from BRAINSTORM.md into SPEC.md\n")
	}

	sb.WriteString(fmt.Sprintf(`   - PROBLEM: What problem does this feature solve?
   - GOALS: What are the measurable outcomes?
   - NON-GOALS: What is explicitly out of scope?
   - USERS: Who will use this feature?
   - SKILLS: Which documented skills should the coding agent use for this feature, where do they live, and when should each one trigger?
   - DEPENDENCIES: Which supporting docs, tools, design refs, APIs, libraries, datasets, assets, and other inputs shaped this specification?
   - REQUIREMENTS: What must be true for this feature to be complete?
   - ACCEPTANCE: How do we verify the feature works?
   - EDGE-CASES: What unusual scenarios must be handled?

Do NOT treat SPEC.md as complete until confidence reaches ≥%d%% and unresolved assumptions = 0.

## SUMMARY Section (MANDATORY)
Once you reach ≥%d%% confidence, write a SUMMARY section at the top of SPEC.md:
- 1-2 sentences maximum
- Information-dense: include the core problem, solution approach, and key constraint
- Written for a coding agent who needs to quickly understand the feature
- Example: "Adds CSV export for user data with streaming support for large datasets (>100k rows). Must complete in <5s and handle Unicode."

## IMPORTANT: File Update Requirement
All specification content MUST be written to: %s
This file is the single source of truth for this feature. Do not leave content only in chat — persist everything to the file.

## Rules
- Keep language precise
- Avoid implementation details (focus on WHAT, not HOW)
- the ## SKILLS section is mandatory and must be populated before sign-off
- the ## DEPENDENCIES section must be current before sign-off and must keep exact locations for external design inputs
- use repo instruction files, repo-local canonical skills, and documented global inputs during the skills discovery phase
- keep the selected skill set minimal and actionable
- do not use .claude/skills as canonical discovery input
- Spec gate: unresolved assumptions = 0 before sign-off; if unresolved assumptions remain, stop and resolve before marking SPEC complete
- Ensure the spec respects constraints defined in CONSTITUTION.md
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times
`, goalPct, goalPct, specPath))

	prompt := sb.String()

	if err := outputPromptWithClipboardDefault(prompt, outputOnly, specCopy); err != nil {
		return err
	}
	if !outputOnly {
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. Copy the prompt above and paste it to your coding agent\n")
		fmt.Printf("  2. Work with the agent to refine the specification\n")
		fmt.Printf("  3. Run 'kit plan %s' to create the implementation plan\n", featureSlug)
	}

	return nil
}

// runSpecTemplate outputs the empty template and generic instructions (legacy behavior).
func runSpecTemplate(specPath, brainstormPath, featureSlug, projectRoot string, cfg *config.Config, outputOnly bool) error {
	goalPct := cfg.GoalPercentage
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	hasBrainstorm := document.Exists(brainstormPath)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Please review and complete the specification at %s.\n\n", specPath))
	sb.WriteString(fmt.Sprintf("This is a new feature: %s\n\n", featureSlug))
	sb.WriteString("## Context Docs (read first)\n")
	sb.WriteString(fmt.Sprintf("|- CONSTITUTION: %s — project-wide constraints, principles, priors\n", constitutionPath))
	for _, path := range repoInstructionPaths(projectRoot, cfg) {
		sb.WriteString(fmt.Sprintf("|- REPO INSTRUCTION: %s — active workflow and skill usage rules\n", path))
	}
	if hasBrainstorm {
		sb.WriteString(fmt.Sprintf("|- BRAINSTORM: %s — upstream research context and codebase findings\n", brainstormPath))
	}
	sb.WriteString(fmt.Sprintf("|- CANONICAL SKILLS: %s/*/SKILL.md — repo-local reusable skills\n", cfg.SkillsPath(projectRoot)))
	for _, path := range globalSkillDiscoveryInputs() {
		sb.WriteString(fmt.Sprintf("|- GLOBAL INPUT: %s — documented global skill or instruction input\n", path))
	}

	sb.WriteString(`

## Context Provided by User
<!-- ⚠️ FILL THIS OUT BEFORE SUBMITTING TO YOUR CODING AGENT -->

**PROBLEM**:
<!-- What problem does this feature solve? -->

**GOALS**:
<!-- What are the measurable outcomes? (comma-separated) -->

**NON-GOALS**:
<!-- What is explicitly out of scope? -->

**USERS**:
<!-- Who will use this feature? -->

**REQUIREMENTS**:
<!-- What must be true for this feature to be complete? -->

**ACCEPTANCE**:
<!-- How do we verify the feature works? -->

**EDGE-CASES**:
<!-- What unusual scenarios must be handled? -->

## Your Task

1. Read CONSTITUTION.md to understand project constraints and principles
`)

	if hasBrainstorm {
		sb.WriteString("2. Read the repository instruction files first\n")
		sb.WriteString("3. Read BRAINSTORM.md and carry forward validated findings\n")
		sb.WriteString("4. Read the SPEC.md template and understand the required sections\n")
		sb.WriteString(fmt.Sprintf("5. Analyze the codebase at %s to understand existing patterns\n", projectRoot))
		sb.WriteString("6. **IMMEDIATELY update SPEC.md** with the context provided above before asking any questions\n")
		sb.WriteString("7. Perform a skills discovery phase before asking sign-off questions:\n")
		sb.WriteString(fmt.Sprintf("   - inspect repo-local canonical skills under `%s/*/SKILL.md`\n", cfg.SkillsPath(projectRoot)))
		for _, path := range globalSkillDiscoveryInputs() {
			sb.WriteString(fmt.Sprintf("   - inspect `%s`\n", path))
		}
		sb.WriteString(fmt.Sprintf("   - populate the `## SKILLS` table in `%s`\n", specPath))
		sb.WriteString("   - keep the required `none | n/a | n/a | no additional skills required | no` row if nothing else applies\n")
		sb.WriteString("   - do not use `.claude/skills` as canonical discovery input\n")
		sb.WriteString(fmt.Sprintf("8. Populate or refresh the `## DEPENDENCIES` table in `%s` before sign-off:\n", specPath))
		sb.WriteString(fmt.Sprintf("   - carry forward still-relevant dependencies from `%s`\n", brainstormPath))
		sb.WriteString("   - keep `## SKILLS` focused on execution-time agent skills and track broader supporting inputs in `## DEPENDENCIES`\n")
		sb.WriteString("   - use the columns `Dependency`, `Type`, `Location`, `Used For`, and `Status`\n")
		sb.WriteString("   - `Status` must be one of `active`, `optional`, or `stale`\n")
		sb.WriteString("   - for Figma or MCP-driven design dependencies, store the exact design URL or file/node reference in `Location`\n")
		sb.WriteString("   - if a dependency influenced decisions but is no longer current, keep it in the table with `Status` = `stale`\n")
		sb.WriteString("   - keep the default `none` row only if no additional dependencies apply\n")
		nextStep := appendNumberedSteps(
			&sb,
			9,
			clarificationLoopSteps(
				goalPct,
				fmt.Sprintf(
					"Reassess, save your updates to %s, and continue with additional "+
						"batches of up to 10 questions until the specification is precise "+
						"enough to produce a correct, production-quality solution",
					specPath,
				),
			),
		)
		sb.WriteString(fmt.Sprintf(
			"%d. Continue refining each section of SPEC.md as you learn more:\n",
			nextStep,
		))
	} else {
		sb.WriteString("2. Read the repository instruction files first\n")
		sb.WriteString("3. Read the SPEC.md template and understand the required sections\n")
		sb.WriteString(fmt.Sprintf("4. Analyze the codebase at %s to understand existing patterns\n", projectRoot))
		sb.WriteString("5. **IMMEDIATELY update SPEC.md** with the context provided above before asking any questions\n")
		sb.WriteString("6. Perform a skills discovery phase before asking sign-off questions:\n")
		sb.WriteString(fmt.Sprintf("   - inspect repo-local canonical skills under `%s/*/SKILL.md`\n", cfg.SkillsPath(projectRoot)))
		for _, path := range globalSkillDiscoveryInputs() {
			sb.WriteString(fmt.Sprintf("   - inspect `%s`\n", path))
		}
		sb.WriteString(fmt.Sprintf("   - populate the `## SKILLS` table in `%s`\n", specPath))
		sb.WriteString("   - keep the required `none | n/a | n/a | no additional skills required | no` row if nothing else applies\n")
		sb.WriteString("   - do not use `.claude/skills` as canonical discovery input\n")
		sb.WriteString(fmt.Sprintf("7. Populate or refresh the `## DEPENDENCIES` table in `%s` before sign-off:\n", specPath))
		sb.WriteString("   - keep `## SKILLS` focused on execution-time agent skills and track broader supporting inputs in `## DEPENDENCIES`\n")
		sb.WriteString("   - use the columns `Dependency`, `Type`, `Location`, `Used For`, and `Status`\n")
		sb.WriteString("   - `Status` must be one of `active`, `optional`, or `stale`\n")
		sb.WriteString("   - for Figma or MCP-driven design dependencies, store the exact design URL or file/node reference in `Location`\n")
		sb.WriteString("   - if a dependency influenced decisions but is no longer current, keep it in the table with `Status` = `stale`\n")
		sb.WriteString("   - keep the default `none` row only if no additional dependencies apply\n")
		nextStep := appendNumberedSteps(
			&sb,
			8,
			clarificationLoopSteps(
				goalPct,
				fmt.Sprintf(
					"Reassess, save your updates to %s, and continue with additional "+
						"batches of up to 10 questions until the specification is precise "+
						"enough to produce a correct, production-quality solution",
					specPath,
				),
			),
		)
		sb.WriteString(fmt.Sprintf(
			"%d. Continue refining each section of SPEC.md as you learn more:\n",
			nextStep,
		))
	}

	sb.WriteString(fmt.Sprintf(`   - PROBLEM: What problem does this feature solve?
   - GOALS: What are the measurable outcomes?
   - NON-GOALS: What is explicitly out of scope?
   - USERS: Who will use this feature?
   - SKILLS: Which documented skills should the coding agent use for this feature, where do they live, and when should each one trigger?
   - DEPENDENCIES: Which supporting docs, tools, design refs, APIs, libraries, datasets, assets, and other inputs shaped this specification?
   - REQUIREMENTS: What must be true for this feature to be complete?
   - ACCEPTANCE: How do we verify the feature works?
   - EDGE-CASES: What unusual scenarios must be handled?

Do NOT treat SPEC.md as complete until confidence reaches ≥%d%% and unresolved assumptions = 0.

## SUMMARY Section (MANDATORY)
Once you reach ≥%d%% confidence, write a SUMMARY section at the top of SPEC.md:
- 1-2 sentences maximum
- Information-dense: include the core problem, solution approach, and key constraint
- Written for a coding agent who needs to quickly understand the feature
- Example: "Adds CSV export for user data with streaming support for large datasets (>100k rows). Must complete in <5s and handle Unicode."

## Rules
- Keep language precise
- Avoid implementation details (focus on WHAT, not HOW)
- the ## SKILLS section is mandatory and must be populated before sign-off
- the ## DEPENDENCIES section must be current before sign-off and must keep exact locations for external design inputs
- use repo instruction files, repo-local canonical skills, and documented global inputs during the skills discovery phase
- keep the selected skill set minimal and actionable
- do not use .claude/skills as canonical discovery input
- Spec gate: unresolved assumptions = 0 before sign-off; if unresolved assumptions remain, stop and resolve before marking SPEC complete
- Ensure the spec respects constraints defined in CONSTITUTION.md
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times
`, goalPct, goalPct))

	prompt := sb.String()

	if !outputOnly {
		fmt.Println()
		fmt.Println(dim + "⚠️  IMPORTANT: Before submitting this prompt, fill in the context section" + reset)
		fmt.Println(dim + "   with details about your feature. The more context you provide, the" + reset)
		fmt.Println(dim + "   better the agent can help you write the specification." + reset)
		fmt.Println()
		fmt.Println(dim + "   Tip: Run 'kit spec <feature> --interactive' for an interactive" + reset)
		fmt.Println(dim + "   experience that guides you through each section." + reset)
		fmt.Println()
	}

	if err := outputPromptWithClipboardDefault(prompt, outputOnly, specCopy); err != nil {
		return err
	}
	if !outputOnly {
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. Edit %s to define the specification\n", specPath)
		fmt.Printf("  2. Run 'kit plan %s' to create the implementation plan\n", featureSlug)
	}

	return nil
}
