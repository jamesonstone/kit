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
var specOutputOnly bool

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
  Default:        Output empty template and agent prompt (non-interactive)
  --interactive:  Prompt user for spec details, then output ready-to-use prompt
  --template:     Output empty template without interactive questions (deprecated, same as default)

Flags:
  --output-only:  Output prompt only, without status messages
  --copy:         Copy prompt to clipboard (combine with --output-only for prompt+copy)
  --interactive:  Force interactive prompts even when stdin is not a terminal`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSpec,
}

func init() {
	specCmd.Flags().Bool("template", false, "(deprecated) output empty template and prompt without interactive questions")
	specCmd.Flags().Bool("interactive", false, "prompt user for spec details interactively")
	specCmd.Flags().BoolVar(&specCopy, "copy", false, "copy agent prompt to clipboard")
	specCmd.Flags().BoolVar(&specOutputOnly, "output-only", false, "output prompt only, suppressing status messages")
	rootCmd.AddCommand(specCmd)
}

func runSpec(cmd *cobra.Command, args []string) error {
	specTemplateOnly, _ := cmd.Flags().GetBool("template")
	specInteractive, _ := cmd.Flags().GetBool("interactive")
	outputOnly, _ := cmd.Flags().GetBool("output-only")

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

	if created {
		fmt.Printf("📁 Created feature directory: %s\n", feat.DirName)
	} else {
		fmt.Printf("📁 Using existing feature: %s\n", feat.DirName)
	}

	// create SPEC.md if it doesn't exist
	specPath := filepath.Join(feat.Path, "SPEC.md")
	if !document.Exists(specPath) {
		if err := document.Write(specPath, templates.Spec); err != nil {
			return fmt.Errorf("failed to create SPEC.md: %w", err)
		}
		fmt.Println("  ✓ Created SPEC.md")
	} else {
		fmt.Println("  ✓ SPEC.md already exists")
	}

	// determine if we should run interactive mode
	// default is non-interactive (template mode), unless --interactive is explicitly set
	isInteractive := specInteractive && !specTemplateOnly

	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	if document.Exists(brainstormPath) {
		fmt.Println("  ✓ Found BRAINSTORM.md")
	}

	// update PROJECT_PROGRESS_SUMMARY.md
	if err := rollup.Update(projectRoot, cfg); err != nil {
		fmt.Printf("  ⚠ Could not update PROJECT_PROGRESS_SUMMARY.md: %v\n", err)
	} else {
		fmt.Println("  ✓ Updated PROJECT_PROGRESS_SUMMARY.md")
	}

	fmt.Printf("\n✅ Feature '%s' ready!\n", feat.Slug)

	if isInteractive {
		// interactive mode: gather details and compile prompt
		return runSpecInteractive(specPath, brainstormPath, feat, projectRoot, cfg, outputOnly)
	}

	// template mode: output the template and instructions
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

// specInputRuneFilter keeps Enter as submit, but converts Ctrl+J/Shift+Enter to newline input.
func specInputRuneFilter(r rune) (rune, bool) {
	if r == readline.CharCtrlJ {
		return '\n', true
	}
	return r, true
}

func normalizeSpecAnswer(raw string) string {
	return strings.TrimSpace(raw)
}

// readLineRL reads a single line using the readline instance, returning empty string on EOF/interrupt.
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
func runSpecInteractive(specPath, brainstormPath string, feat *feature.Feature, projectRoot string, cfg *config.Config, outputOnly bool) error {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:              whiteBold + "   > " + reset,
		InterruptPrompt:     "^C",
		EOFPrompt:           "",
		Stdin:               os.Stdin,
		Stdout:              os.Stdout,
		Stderr:              os.Stderr,
		FuncFilterInputRune: specInputRuneFilter,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer rl.Close()

	fmt.Println("\n" + dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(whiteBold + "📝 Interactive Spec Builder" + reset)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println()

	fmt.Println(dim + "Answer the following questions to generate a complete prompt for your coding agent." + reset)
	fmt.Println(dim + "Use ←/→ arrow keys to move through your text and correct mistakes." + reset)
	fmt.Println(dim + "Press Enter to continue; use Shift+Enter (or Ctrl+J) to add a newline." + reset)
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

// outputCompiledPrompt generates the final agent prompt and either copies to clipboard or prints.
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
	if hasBrainstorm {
		sb.WriteString(fmt.Sprintf("| BRAINSTORM | %s |\n", brainstormPath))
	}
	sb.WriteString(fmt.Sprintf("| SPEC | %s |\n", specPath))
	sb.WriteString(fmt.Sprintf("| Project Root | %s |\n\n", projectRoot))

	sb.WriteString("## Your Task\n\n")
	sb.WriteString(fmt.Sprintf("1. Read CONSTITUTION.md (file: %s) to understand project constraints and principles\n", constitutionPath))
	step := 2
	if hasBrainstorm {
		sb.WriteString(fmt.Sprintf("%d. Read BRAINSTORM.md (file: %s) and treat it as upstream research context\n", step, brainstormPath))
		step++
	}
	sb.WriteString(fmt.Sprintf("%d. Read the current SPEC.md (file: %s) and understand the required sections\n", step, specPath))
	step++
	sb.WriteString(fmt.Sprintf("%d. Analyze the codebase at %s to understand existing patterns\n", step, projectRoot))

	questionStart := step + 1

	if hasContext {
		sb.WriteString(fmt.Sprintf(`%d. **IMMEDIATELY write all context above into the SPEC.md file at %s** — do NOT ask questions before doing this
%d. Ask clarifying questions until you reach ≥%d%% confidence that you understand the problem and desired solution
%d. Use numbered lists
%d. Ask questions in batches of up to 10
%d. For every question, include your current best proposed solution or assumption
%d. State uncertainties
%d. After each batch of up to 10 questions, output your current percentage understanding so the user can see progress
%d. Reassess, save your updates to %s, and continue with additional batches of up to 10 questions until the specification is precise enough to produce a correct, production-quality solution
%d. Continue refining each section of SPEC.md as you learn more:
`, questionStart, specPath, questionStart+1, goalPct, questionStart+2, questionStart+3, questionStart+4, questionStart+5, questionStart+6, questionStart+7, specPath, questionStart+8))
	} else {
		sb.WriteString(fmt.Sprintf(`%d. Ask clarifying questions until you reach ≥%d%% confidence that you understand the problem and desired solution
%d. Use numbered lists
%d. Ask questions in batches of up to 10
%d. For every question, include your current best proposed solution or assumption
%d. State uncertainties
%d. After each batch of up to 10 questions, output your current percentage understanding so the user can see progress
%d. Reassess, save your updates to %s, and continue with additional batches of up to 10 questions until the specification is precise enough to produce a correct, production-quality solution
%d. **Write your findings directly to %s** as you fill in each section:
`, questionStart, goalPct, questionStart+1, questionStart+2, questionStart+3, questionStart+4, questionStart+5, questionStart+6, specPath, questionStart+7, specPath))
	}

	if hasBrainstorm {
		sb.WriteString("   - Carry forward validated findings from BRAINSTORM.md into SPEC.md\n")
	}

	sb.WriteString(fmt.Sprintf(`   - PROBLEM: What problem does this feature solve?
   - GOALS: What are the measurable outcomes?
   - NON-GOALS: What is explicitly out of scope?
   - USERS: Who will use this feature?
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
- Spec gate: unresolved assumptions = 0 before sign-off; if unresolved assumptions remain, stop and resolve before marking SPEC complete
- Ensure the spec respects constraints defined in CONSTITUTION.md
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times
`, goalPct, goalPct, specPath))

	prompt := sb.String()

	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	if specCopy {
		fmt.Println(whiteBold + "Agent prompt copied to clipboard" + reset)
	} else {
		fmt.Println(whiteBold + "✅ Copy this prompt to your coding agent:" + reset)
	}
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)

	if err := outputPrompt(prompt, outputOnly, specCopy); err != nil {
		return err
	}

	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)

	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Copy the prompt above and paste it to your coding agent\n")
	fmt.Printf("  2. Work with the agent to refine the specification\n")
	fmt.Printf("  3. Run 'kit plan %s' to create the implementation plan\n", featureSlug)

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
	if hasBrainstorm {
		sb.WriteString(fmt.Sprintf("|- BRAINSTORM: %s — upstream research context and codebase findings\n", brainstormPath))
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
		sb.WriteString("2. Read BRAINSTORM.md and carry forward validated findings\n")
		sb.WriteString("3. Read the SPEC.md template and understand the required sections\n")
		sb.WriteString(fmt.Sprintf("4. Analyze the codebase at %s to understand existing patterns\n", projectRoot))
		sb.WriteString("5. **IMMEDIATELY update SPEC.md** with the context provided above before asking any questions\n")
		sb.WriteString(fmt.Sprintf("6. Ask clarifying questions until you reach ≥%d%% confidence that you understand the problem and desired solution\n", goalPct))
		sb.WriteString("7. Use numbered lists\n")
		sb.WriteString("8. Ask questions in batches of up to 10\n")
		sb.WriteString("9. For every question, include your current best proposed solution or assumption\n")
		sb.WriteString("10. State uncertainties\n")
		sb.WriteString("11. After each batch of up to 10 questions, output your current percentage understanding so the user can see progress\n")
		sb.WriteString(fmt.Sprintf("12. Reassess, save your updates to %s, and continue with additional batches of up to 10 questions until the specification is precise enough to produce a correct, production-quality solution\n", specPath))
		sb.WriteString("13. Continue refining each section of SPEC.md as you learn more:\n")
	} else {
		sb.WriteString("2. Read the SPEC.md template and understand the required sections\n")
		sb.WriteString(fmt.Sprintf("3. Analyze the codebase at %s to understand existing patterns\n", projectRoot))
		sb.WriteString("4. **IMMEDIATELY update SPEC.md** with the context provided above before asking any questions\n")
		sb.WriteString(fmt.Sprintf("5. Ask clarifying questions until you reach ≥%d%% confidence that you understand the problem and desired solution\n", goalPct))
		sb.WriteString("6. Use numbered lists\n")
		sb.WriteString("7. Ask questions in batches of up to 10\n")
		sb.WriteString("8. For every question, include your current best proposed solution or assumption\n")
		sb.WriteString("9. State uncertainties\n")
		sb.WriteString("10. After each batch of up to 10 questions, output your current percentage understanding so the user can see progress\n")
		sb.WriteString(fmt.Sprintf("11. Reassess, save your updates to %s, and continue with additional batches of up to 10 questions until the specification is precise enough to produce a correct, production-quality solution\n", specPath))
		sb.WriteString("12. Continue refining each section of SPEC.md as you learn more:\n")
	}

	sb.WriteString(fmt.Sprintf(`   - PROBLEM: What problem does this feature solve?
   - GOALS: What are the measurable outcomes?
   - NON-GOALS: What is explicitly out of scope?
   - USERS: Who will use this feature?
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
- Spec gate: unresolved assumptions = 0 before sign-off; if unresolved assumptions remain, stop and resolve before marking SPEC complete
- Ensure the spec respects constraints defined in CONSTITUTION.md
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times
`, goalPct, goalPct))

	prompt := sb.String()

	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Edit %s to define the specification\n", specPath)
	fmt.Printf("  2. Run 'kit plan %s' to create the implementation plan\n", featureSlug)

	fmt.Println("\n" + dim + "────────────────────────────────────────────────────────────────────────" + reset)
	if specCopy {
		fmt.Println(whiteBold + "Agent prompt copied to clipboard" + reset)
	} else {
		fmt.Println(whiteBold + "Copy this prompt to your coding agent:" + reset)
	}
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println()
	fmt.Println(dim + "⚠️  IMPORTANT: Before submitting this prompt, fill in the context section" + reset)
	fmt.Println(dim + "   with details about your feature. The more context you provide, the" + reset)
	fmt.Println(dim + "   better the agent can help you write the specification." + reset)
	fmt.Println()
	fmt.Println(dim + "   Tip: Run 'kit spec <feature> --interactive' for an interactive" + reset)
	fmt.Println(dim + "   experience that guides you through each section." + reset)
	fmt.Println()

	if err := outputPrompt(prompt, outputOnly, specCopy); err != nil {
		return err
	}

	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)

	return nil
}
