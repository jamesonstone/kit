package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/git"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

var (
	specNoBranch     bool
	specTemplateOnly bool
	specInteractive  bool
)

var specCmd = &cobra.Command{
	Use:   "spec <feature>",
	Short: "Create or open a feature specification",
	Long: `Create a new feature specification or open an existing one.

Creates:
  - Feature directory (e.g., docs/specs/0001-my-feature/)
  - SPEC.md with required sections and placeholder comments
  - Git branch matching the feature directory name (unless --no-branch)

Updates PROJECT_PROGRESS_SUMMARY.md after creation.

Modes:
  Default:       Interactive prompts to gather spec details, then outputs a ready-to-use prompt
  --template:    Output the empty SPEC.md template and agent prompt (no interactive questions)
  --interactive: Force interactive mode even when stdin is not a terminal`,
	Args: cobra.ExactArgs(1),
	RunE: runSpec,
}

func init() {
	specCmd.Flags().BoolVar(&specNoBranch, "no-branch", false, "skip git branch creation")
	specCmd.Flags().BoolVar(&specTemplateOnly, "template", false, "output empty template and prompt without interactive questions")
	specCmd.Flags().BoolVar(&specInteractive, "interactive", false, "force interactive mode even when stdin is not a terminal")
	rootCmd.AddCommand(specCmd)
}

func runSpec(cmd *cobra.Command, args []string) error {
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

	// create git branch if enabled and not --no-branch
	if !specNoBranch && cfg.Branching.Enabled && git.IsRepo(projectRoot) {
		branchName := feat.DirName
		branchCreated, err := git.EnsureBranch(projectRoot, branchName, cfg.Branching.BaseBranch)
		if err != nil {
			fmt.Printf("  ⚠ Could not create branch: %v\n", err)
		} else if branchCreated {
			fmt.Printf("  ✓ Created and switched to branch: %s\n", branchName)
		} else {
			fmt.Printf("  ✓ Switched to existing branch: %s\n", branchName)
		}
	} else if specNoBranch {
		fmt.Println("  ℹ Skipped branch creation (--no-branch)")
	}

	// update PROJECT_PROGRESS_SUMMARY.md
	if err := rollup.Update(projectRoot, cfg); err != nil {
		fmt.Printf("  ⚠ Could not update PROJECT_PROGRESS_SUMMARY.md: %v\n", err)
	} else {
		fmt.Println("  ✓ Updated PROJECT_PROGRESS_SUMMARY.md")
	}

	fmt.Printf("\n✅ Feature '%s' ready!\n", feat.Slug)

	// determine if we should run interactive mode
	runInteractive := !specTemplateOnly && (specInteractive || isTerminal())

	if runInteractive {
		// interactive mode: gather details and compile prompt
		return runSpecInteractive(specPath, feat.Slug, projectRoot, cfg)
	}

	// template mode: output the template and instructions
	return runSpecTemplate(specPath, feat.Slug, projectRoot, cfg)
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

// runSpecInteractive prompts the user for each SPEC section and compiles a ready-to-use prompt
func runSpecInteractive(specPath, featureSlug, projectRoot string, cfg *config.Config) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n" + dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(whiteBold + "📝 Interactive Spec Builder" + reset)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println()
	fmt.Println(dim + "Answer the following questions to generate a complete prompt for your coding agent." + reset)
	fmt.Println(dim + "Press Enter to skip a question (you can refine details with the agent later)." + reset)
	fmt.Println()

	answers := specAnswers{}

	// PROBLEM
	fmt.Println(spec + "1. PROBLEM" + reset + " - What problem does this feature solve?")
	fmt.Println(dim + "   Example: Users cannot export their data in CSV format" + reset)
	fmt.Print(whiteBold + "   > " + reset)
	answers.Problem = readLine(reader)

	// GOALS
	fmt.Println()
	fmt.Println(spec + "2. GOALS" + reset + " - What are the measurable outcomes? (comma-separated)")
	fmt.Println(dim + "   Example: Export completes in <5s, supports 100k+ rows, CSV is RFC-compliant" + reset)
	fmt.Print(whiteBold + "   > " + reset)
	answers.Goals = readLine(reader)

	// NON-GOALS
	fmt.Println()
	fmt.Println(spec + "3. NON-GOALS" + reset + " - What is explicitly out of scope?")
	fmt.Println(dim + "   Example: Excel format, scheduled exports, email delivery" + reset)
	fmt.Print(whiteBold + "   > " + reset)
	answers.NonGoals = readLine(reader)

	// USERS
	fmt.Println()
	fmt.Println(spec + "4. USERS" + reset + " - Who will use this feature?")
	fmt.Println(dim + "   Example: Admin users, API consumers, data analysts" + reset)
	fmt.Print(whiteBold + "   > " + reset)
	answers.Users = readLine(reader)

	// REQUIREMENTS
	fmt.Println()
	fmt.Println(spec + "5. REQUIREMENTS" + reset + " - What must be true for this feature to be complete?")
	fmt.Println(dim + "   Example: Must handle Unicode, must include headers, must stream large files" + reset)
	fmt.Print(whiteBold + "   > " + reset)
	answers.Requirements = readLine(reader)

	// ACCEPTANCE
	fmt.Println()
	fmt.Println(spec + "6. ACCEPTANCE" + reset + " - How do we verify the feature works?")
	fmt.Println(dim + "   Example: Unit tests pass, integration tests cover edge cases, manual QA sign-off" + reset)
	fmt.Print(whiteBold + "   > " + reset)
	answers.Acceptance = readLine(reader)

	// EDGE-CASES
	fmt.Println()
	fmt.Println(spec + "7. EDGE-CASES" + reset + " - What unusual scenarios must be handled?")
	fmt.Println(dim + "   Example: Empty dataset, special characters in data, network timeout during export" + reset)
	fmt.Print(whiteBold + "   > " + reset)
	answers.EdgeCases = readLine(reader)

	fmt.Println()

	// generate the compiled prompt
	return outputCompiledPrompt(specPath, featureSlug, projectRoot, cfg, &answers)
}

// readLine reads a single line from the reader, trimming whitespace
func readLine(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

// outputCompiledPrompt generates and prints the final agent prompt
func outputCompiledPrompt(specPath, featureSlug, projectRoot string, cfg *config.Config, answers *specAnswers) error {
	goalPct := cfg.GoalPercentage

	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(whiteBold + "✅ Copy this prompt to your coding agent:" + reset)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)

	fmt.Printf(`
Please review and complete the specification at %s.

This is a new feature: %s

## Context Provided by User

`, specPath, featureSlug)

	// output user-provided context
	if answers.Problem != "" {
		fmt.Printf("**PROBLEM**: %s\n\n", answers.Problem)
	}
	if answers.Goals != "" {
		fmt.Printf("**GOALS**: %s\n\n", answers.Goals)
	}
	if answers.NonGoals != "" {
		fmt.Printf("**NON-GOALS**: %s\n\n", answers.NonGoals)
	}
	if answers.Users != "" {
		fmt.Printf("**USERS**: %s\n\n", answers.Users)
	}
	if answers.Requirements != "" {
		fmt.Printf("**REQUIREMENTS**: %s\n\n", answers.Requirements)
	}
	if answers.Acceptance != "" {
		fmt.Printf("**ACCEPTANCE**: %s\n\n", answers.Acceptance)
	}
	if answers.EdgeCases != "" {
		fmt.Printf("**EDGE-CASES**: %s\n\n", answers.EdgeCases)
	}

	// check if any answers were provided
	hasContext := answers.Problem != "" || answers.Goals != "" || answers.NonGoals != "" ||
		answers.Users != "" || answers.Requirements != "" || answers.Acceptance != "" ||
		answers.EdgeCases != ""

	fmt.Print(`## Your Task

1. Read the SPEC.md template and understand the required sections
2. Analyze the codebase at ` + projectRoot + ` to understand existing patterns
`)

	if hasContext {
		fmt.Print(`3. **IMMEDIATELY update SPEC.md** with the context provided above before asking any questions
4. Ask clarifying questions in batches of 10 until you reach >= ` + fmt.Sprintf("%d", goalPct) + `% understanding
5. Continue refining each section of SPEC.md as you learn more:
`)
	} else {
		fmt.Print(`3. Ask clarifying questions in batches of 10 until you reach >= ` + fmt.Sprintf("%d", goalPct) + `% understanding
4. Fill in each section with clear, specific requirements:
`)
	}

	fmt.Print(`   - PROBLEM: What problem does this feature solve?
   - GOALS: What are the measurable outcomes?
   - NON-GOALS: What is explicitly out of scope?
   - USERS: Who will use this feature?
   - REQUIREMENTS: What must be true for this feature to be complete?
   - ACCEPTANCE: How do we verify the feature works?
   - EDGE-CASES: What unusual scenarios must be handled?

After each batch of questions, state your current understanding percentage.
Do NOT proceed to writing the spec until understanding >= ` + fmt.Sprintf("%d", goalPct) + `%.

## SUMMARY Section (MANDATORY)
Once you reach >= ` + fmt.Sprintf("%d", goalPct) + `% understanding, write a SUMMARY section at the top of SPEC.md:
- 1-2 sentences maximum
- Information-dense: include the core problem, solution approach, and key constraint
- Written for a coding agent who needs to quickly understand the feature
- Example: "Adds CSV export for user data with streaming support for large datasets (>100k rows). Must complete in <5s and handle Unicode."
## Rules
- Keep language precise
- Avoid implementation details (focus on WHAT, not HOW)
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times

`)

	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)

	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Copy the prompt above and paste it to your coding agent\n")
	fmt.Printf("  2. Work with the agent to refine the specification\n")
	fmt.Printf("  3. Run 'kit plan %s' to create the implementation plan\n", featureSlug)

	return nil
}

// runSpecTemplate outputs the empty template and generic instructions (legacy behavior)
func runSpecTemplate(specPath, featureSlug, projectRoot string, cfg *config.Config) error {
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

	goalPct := cfg.GoalPercentage
	fmt.Printf(`Please review and complete the specification at %s.

This is a new feature: %s

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

1. Read the SPEC.md template and understand the required sections
2. Analyze the codebase at %s to understand existing patterns
3. **IMMEDIATELY update SPEC.md** with the context provided above before asking any questions
4. Ask clarifying questions in batches of 10 until you reach >= %d%% understanding
5. Continue refining each section of SPEC.md as you learn more:
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
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times

`, specPath, featureSlug, projectRoot, goalPct, goalPct, goalPct)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)

	return nil
}
