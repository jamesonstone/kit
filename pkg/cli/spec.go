package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/git"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

var specCopy bool

var specCmd = &cobra.Command{
	Use:   "spec <feature>",
	Short: "Create or open a feature specification",
	Long: `Create a new feature specification or open an existing one.

Creates:
  - Feature directory (e.g., docs/specs/0001-my-feature/)
  - SPEC.md with required sections and placeholder comments
  - Git branch matching the feature directory name (with --create-branch)

Updates PROJECT_PROGRESS_SUMMARY.md after creation.

Modes:
  Default:       Interactive prompts to gather spec details, then outputs a ready-to-use prompt
  --template:    Output the empty SPEC.md template and agent prompt (no interactive questions)
  --interactive: Force interactive mode even when stdin is not a terminal`,
	Args: cobra.ExactArgs(1),
	RunE: runSpec,
}

func init() {
	specCmd.Flags().Bool("create-branch", false, "create and switch to a git branch matching the feature name")
	specCmd.Flags().Bool("template", false, "output empty template and prompt without interactive questions")
	specCmd.Flags().Bool("interactive", false, "force interactive mode even when stdin is not a terminal")
	specCmd.Flags().BoolVar(&specCopy, "copy", false, "copy agent prompt to clipboard (suppresses stdout)")
	rootCmd.AddCommand(specCmd)
}

func runSpec(cmd *cobra.Command, args []string) error {
	createBranch, _ := cmd.Flags().GetBool("create-branch")
	specTemplateOnly, _ := cmd.Flags().GetBool("template")
	specInteractive, _ := cmd.Flags().GetBool("interactive")

	featureRef := args[0]

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

	// create or find feature
	feat, created, err := feature.EnsureExists(cfg, specsDir, featureRef)
	if err != nil {
		return err
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
	isInteractive := !specTemplateOnly && (specInteractive || isTerminal())

	// create git branch if --create-branch flag is set
	if createBranch && git.IsRepo(projectRoot) {
		createBranchForFeature(projectRoot, feat, cfg)
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
		return runSpecInteractive(specPath, feat, projectRoot, cfg, createBranch)
	}

	// template mode: output the template and instructions
	return runSpecTemplate(specPath, feat.Slug, projectRoot, cfg)
}

// createBranchForFeature creates and switches to a git branch for the feature.
func createBranchForFeature(projectRoot string, feat *feature.Feature, cfg *config.Config) {
	branchName := feat.DirName
	branchCreated, err := git.EnsureBranch(projectRoot, branchName, cfg.Branching.BaseBranch)
	if err != nil {
		fmt.Printf("  ⚠ Could not create branch: %v\n", err)
	} else if branchCreated {
		fmt.Printf("  ✓ Created and switched to branch: %s\n", branchName)
	} else {
		fmt.Printf("  ✓ Switched to existing branch: %s\n", branchName)
	}
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
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

// readLineRL reads a single line using the readline instance, returning empty string on EOF/interrupt.
func readLineRL(rl *readline.Instance) string {
	line, err := rl.Readline()
	if err != nil {
		if err == readline.ErrInterrupt || err == io.EOF {
			return ""
		}
		return ""
	}
	return strings.TrimSpace(line)
}

// runSpecInteractive prompts the user for each SPEC section and compiles a ready-to-use prompt
func runSpecInteractive(specPath string, feat *feature.Feature, projectRoot string, cfg *config.Config, branchAlreadyCreated bool) error {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          whiteBold + "   > " + reset,
		InterruptPrompt: "^C",
		EOFPrompt:       "",
		Stdin:           os.Stdin,
		Stdout:          os.Stdout,
		Stderr:          os.Stderr,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer rl.Close()

	fmt.Println("\n" + dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(whiteBold + "📝 Interactive Spec Builder" + reset)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println()

	// prompt for branch creation if in a git repo and not already created via flag
	if !branchAlreadyCreated && git.IsRepo(projectRoot) {
		rl.SetPrompt(whiteBold + "[y/N]: " + reset)
		fmt.Printf(dim+"Create feature branch '%s'?"+reset+" ", feat.DirName)
		branchAnswer := strings.ToLower(readLineRL(rl))
		if branchAnswer == "y" || branchAnswer == "yes" {
			createBranchForFeature(projectRoot, feat, cfg)
		}
		fmt.Println()
	}

	fmt.Println(dim + "Answer the following questions to generate a complete prompt for your coding agent." + reset)
	fmt.Println(dim + "Use ←/→ arrow keys to move through your text and correct mistakes." + reset)
	fmt.Println(dim + "Press Enter to skip a question (you can refine details with the agent later)." + reset)
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
	return outputCompiledPrompt(specPath, feat.Slug, projectRoot, cfg, &answers)
}

// outputCompiledPrompt generates the final agent prompt and either copies to clipboard or prints
func outputCompiledPrompt(specPath, featureSlug, projectRoot string, cfg *config.Config, answers *specAnswers) error {
	goalPct := cfg.GoalPercentage
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`You MUST update the specification file at: %s
This is the source-of-truth document for feature: %s

## Context Provided by User

`, specPath, featureSlug))

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

	sb.WriteString(fmt.Sprintf(`## Context Docs (read first)
- CONSTITUTION: %s — project-wide constraints, principles, priors

## Your Task

1. Read CONSTITUTION.md to understand project constraints and principles
2. Read the current SPEC.md at %s and understand the required sections
3. Analyze the codebase at %s to understand existing patterns
`, constitutionPath, specPath, projectRoot))

	if hasContext {
		sb.WriteString(fmt.Sprintf(`4. **IMMEDIATELY write all context above into the SPEC.md file at %s** — do NOT ask questions before doing this
5. Ask clarifying questions in batches of 10 until you reach >= %d%% understanding
6. After each round of questions, **save your updates to %s** before continuing
7. Continue refining each section of SPEC.md as you learn more:
`, specPath, goalPct, specPath))
	} else {
		sb.WriteString(fmt.Sprintf(`4. Ask clarifying questions in batches of 10 until you reach >= %d%% understanding
5. **Write your findings directly to %s** as you fill in each section:
`, goalPct, specPath))
	}

	sb.WriteString(fmt.Sprintf(`   - PROBLEM: What problem does this feature solve?
   - GOALS: What are the measurable outcomes?
   - NON-GOALS: What is explicitly out of scope?
   - USERS: Who will use this feature?
   - REQUIREMENTS: What must be true for this feature to be complete?
   - ACCEPTANCE: How do we verify the feature works?
   - EDGE-CASES: What unusual scenarios must be handled?

After each batch of questions, state your current understanding percentage.
Do NOT proceed to writing the spec until understanding >= %d%%.

## SUMMARY Section (MANDATORY)
Once you reach >= %d%% understanding, write a SUMMARY section at the top of SPEC.md:
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
- Ensure the spec respects constraints defined in CONSTITUTION.md
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times
`, goalPct, goalPct, specPath))

	prompt := sb.String()

	// copy to clipboard if requested
	if specCopy {
		if err := copyToClipboard(prompt); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		fmt.Println("✓ Copied agent prompt to clipboard")
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. Paste the prompt to your coding agent\n")
		fmt.Printf("  2. Work with the agent to refine the specification\n")
		fmt.Printf("  3. Run 'kit plan %s' to create the implementation plan\n", featureSlug)
		return nil
	}

	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(whiteBold + "✅ Copy this prompt to your coding agent:" + reset)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Print(prompt)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)

	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Copy the prompt above and paste it to your coding agent\n")
	fmt.Printf("  2. Work with the agent to refine the specification\n")
	fmt.Printf("  3. Run 'kit plan %s' to create the implementation plan\n", featureSlug)

	return nil
}

// runSpecTemplate outputs the empty template and generic instructions (legacy behavior)
func runSpecTemplate(specPath, featureSlug, projectRoot string, cfg *config.Config) error {
	goalPct := cfg.GoalPercentage
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")

	prompt := fmt.Sprintf(`Please review and complete the specification at %s.

This is a new feature: %s

## Context Docs (read first)
- CONSTITUTION: %s — project-wide constraints, principles, priors

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
2. Read the SPEC.md template and understand the required sections
3. Analyze the codebase at %s to understand existing patterns
4. **IMMEDIATELY update SPEC.md** with the context provided above before asking any questions
5. Ask clarifying questions in batches of 10 until you reach >= %d%% understanding
6. Continue refining each section of SPEC.md as you learn more:
   - PROBLEM: What problem does this feature solve?
   - GOALS: What are the measurable outcomes?
   - NON-GOALS: What is explicitly out of scope?
   - USERS: Who will use this feature?
   - REQUIREMENTS: What must be true for this feature to be complete?
   - ACCEPTANCE: How do we verify the feature works?
   - EDGE-CASES: What unusual scenarios must be handled?

After each batch of questions, state your current understanding percentage.
Do NOT proceed to writing the spec until understanding >= %d%%.

## SUMMARY Section (MANDATORY)
Once you reach >= %d%% understanding, write a SUMMARY section at the top of SPEC.md:
- 1-2 sentences maximum
- Information-dense: include the core problem, solution approach, and key constraint
- Written for a coding agent who needs to quickly understand the feature
- Example: "Adds CSV export for user data with streaming support for large datasets (>100k rows). Must complete in <5s and handle Unicode."

## Rules
- Keep language precise
- Avoid implementation details (focus on WHAT, not HOW)
- Ensure the spec respects constraints defined in CONSTITUTION.md
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times
`, specPath, featureSlug, constitutionPath, projectRoot, goalPct, goalPct, goalPct)

	// copy to clipboard if requested
	if specCopy {
		if err := copyToClipboard(prompt); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		fmt.Println("✓ Copied agent prompt to clipboard")
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. Paste the prompt to your coding agent\n")
		fmt.Printf("  2. Fill in the context section before submitting\n")
		fmt.Printf("  3. Run 'kit plan %s' to create the implementation plan\n", featureSlug)
		return nil
	}

	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Edit %s to define the specification\n", specPath)
	fmt.Printf("  2. Run 'kit plan %s' to create the implementation plan\n", featureSlug)

	fmt.Println("\n" + dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(whiteBold + "Copy this prompt to your coding agent:" + reset)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println()
	fmt.Println(dim + "⚠️  IMPORTANT: Before submitting this prompt, fill in the context section" + reset)
	fmt.Println(dim + "   with details about your feature. The more context you provide, the" + reset)
	fmt.Println(dim + "   better the agent can help you write the specification." + reset)
	fmt.Println()
	fmt.Println(dim + "   Tip: Run 'kit spec <feature>' without --template for an interactive" + reset)
	fmt.Println(dim + "   experience that guides you through each section." + reset)
	fmt.Println()
	fmt.Print(prompt)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)

	return nil
}
