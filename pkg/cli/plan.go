package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

var planCopy bool
var planOutputOnly bool

var planCmd = &cobra.Command{
	Use:   "plan [feature]",
	Short: "Create or open a feature implementation plan",
	Long: `Create a new implementation plan for a feature.

Creates:
  - PLAN.md with required sections and placeholder comments

Prerequisites:
  - SPEC.md must exist (unless --force)

If no feature is specified, shows an interactive selection of features
that have SPEC.md but no PLAN.md yet.

Updates PROJECT_PROGRESS_SUMMARY.md after creation.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPlan,
}

func init() {
	planCmd.Flags().Bool("force", false, "create missing SPEC.md with headers if it doesn't exist")
	planCmd.Flags().Bool("warp", false, "output prompt for Warp coding agent to fill PLAN.md from Warp plan")
	planCmd.Flags().BoolVar(&planCopy, "copy", false, "copy agent prompt to clipboard")
	planCmd.Flags().BoolVar(&planOutputOnly, "output-only", false, "output prompt only, suppressing status messages")
	rootCmd.AddCommand(planCmd)
}

func runPlan(cmd *cobra.Command, args []string) error {
	planForce, _ := cmd.Flags().GetBool("force")
	warpMode, _ := cmd.Flags().GetBool("warp")
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

	var feat *feature.Feature

	if len(args) == 0 {
		// interactive mode: select from features with SPEC but no PLAN
		feat, err = selectFeatureForPlan(specsDir)
		if err != nil {
			return err
		}
	} else {
		// direct mode: resolve feature by name
		featureRef := args[0]
		feat, err = feature.Resolve(specsDir, featureRef)
		if err != nil {
			return fmt.Errorf("feature '%s' not found. Run 'kit spec %s' first to create it", featureRef, featureRef)
		}
	}

	fmt.Printf("📋 Creating plan for feature: %s\n", feat.DirName)

	// check prerequisites
	specPath := filepath.Join(feat.Path, "SPEC.md")
	if !document.Exists(specPath) {
		if planForce || cfg.AllowOutOfOrder {
			// create empty SPEC.md
			if err := document.Write(specPath, templates.Spec); err != nil {
				return fmt.Errorf("failed to create SPEC.md: %w", err)
			}
			fmt.Println("  ✓ Created SPEC.md (--force)")
		} else {
			return fmt.Errorf("SPEC.md not found. Run 'kit spec %s' first or use --force", feat.Slug)
		}
	}

	// create PLAN.md if it doesn't exist
	planPath := filepath.Join(feat.Path, "PLAN.md")
	if !document.Exists(planPath) {
		if err := document.Write(planPath, templates.Plan); err != nil {
			return fmt.Errorf("failed to create PLAN.md: %w", err)
		}
		fmt.Println("  ✓ Created PLAN.md")
	} else {
		fmt.Println("  ✓ PLAN.md already exists")
	}

	// update PROJECT_PROGRESS_SUMMARY.md
	if err := rollup.Update(projectRoot, cfg); err != nil {
		fmt.Printf("  ⚠ Could not update PROJECT_PROGRESS_SUMMARY.md: %v\n", err)
	} else {
		fmt.Println("  ✓ Updated PROJECT_PROGRESS_SUMMARY.md")
	}

	fmt.Printf("\n✅ Plan for '%s' ready!\n", feat.Slug)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Edit %s to define the implementation approach\n", planPath)
	fmt.Printf("  2. Run 'kit tasks %s' to create executable tasks\n", feat.Slug)
	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")

	if warpMode {
		outputWarpPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, outputOnly)
	} else {
		outputStandardPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, outputOnly)
	}

	return nil
}

// selectFeatureForPlan shows an interactive numbered list of features
// that have SPEC.md but no PLAN.md yet.
func selectFeatureForPlan(specsDir string) (*feature.Feature, error) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	// filter to features with SPEC but no PLAN
	var candidates []feature.Feature
	for _, f := range features {
		specPath := filepath.Join(f.Path, "SPEC.md")
		planPath := filepath.Join(f.Path, "PLAN.md")
		if document.Exists(specPath) && !document.Exists(planPath) {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no features ready for planning (need SPEC.md without PLAN.md)\n\nRun 'kit spec <feature>' to create a new feature first")
	}

	fmt.Println()
	fmt.Println(whiteBold + "Select a feature to plan:" + reset)
	fmt.Println()
	for i, f := range candidates {
		fmt.Printf("  [%d] %s\n", i+1, f.DirName)
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

// outputStandardPlanPrompt outputs the standard coding agent prompt.
func outputStandardPlanPrompt(planPath, specPath, brainstormPath string, feat *feature.Feature, cfg *config.Config, outputOnly bool) {
	projectRoot, _ := config.FindProjectRoot()
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	goalPct := cfg.GoalPercentage
	hasBrainstorm := document.Exists(brainstormPath)

	var sb strings.Builder
	sb.WriteString("Please review and complete the implementation plan.\n\n")
	sb.WriteString("## File References\n")
	sb.WriteString("| Document | Path |\n")
	sb.WriteString("|----------|------|\n")
	sb.WriteString(fmt.Sprintf("| CONSTITUTION | %s |\n", constitutionPath))
	if hasBrainstorm {
		sb.WriteString(fmt.Sprintf("| BRAINSTORM | %s |\n", brainstormPath))
	}
	sb.WriteString(fmt.Sprintf("| SPEC | %s |\n", specPath))
	sb.WriteString(fmt.Sprintf("| PLAN | %s |\n", planPath))
	sb.WriteString(fmt.Sprintf("| Feature | %s |\n", feat.Slug))
	sb.WriteString(fmt.Sprintf("| Project Root | %s |\n\n", projectRoot))

	sb.WriteString("Your task:\n")
	sb.WriteString(fmt.Sprintf("1. Read CONSTITUTION.md (file: %s) to understand project constraints and principles\n", constitutionPath))
	step := 2
	if hasBrainstorm {
		sb.WriteString(fmt.Sprintf("%d. Read BRAINSTORM.md (file: %s) for upstream research context\n", step, brainstormPath))
		step++
	}
	sb.WriteString(fmt.Sprintf("%d. Read SPEC.md (file: %s) fully and treat it as a fixed contract\n", step, specPath))
	step++
	sb.WriteString(fmt.Sprintf("%d. Review the PLAN.md (file: %s) template and required sections\n", step, planPath))
	step++
	sb.WriteString(fmt.Sprintf("%d. Identify any missing design decisions required for execution\n", step))
	step++
	sb.WriteString(fmt.Sprintf("%d. Ask clarifying questions until you reach ≥%d%% confidence that you understand the problem and desired solution\n", step, goalPct))
	step++
	sb.WriteString(fmt.Sprintf("%d. Use numbered lists\n", step))
	step++
	sb.WriteString(fmt.Sprintf("%d. Ask questions in batches of up to 10\n", step))
	step++
	sb.WriteString(fmt.Sprintf("%d. For every question, include your current best proposed solution or assumption\n", step))
	step++
	sb.WriteString(fmt.Sprintf("%d. State uncertainties\n", step))
	step++
	sb.WriteString(fmt.Sprintf("%d. After each batch of up to 10 questions, output your current percentage understanding so the user can see progress\n", step))
	step++
	sb.WriteString(fmt.Sprintf("%d. Reassess and continue with additional batches of up to 10 questions until the plan is precise enough to produce a correct, production-quality implementation\n", step))
	step++
	sb.WriteString(fmt.Sprintf("%d. Commit to concrete design decisions that make execution unambiguous\n\n", step))

	sb.WriteString(fmt.Sprintf(`Before you write or update PLAN.md:
- after each batch of up to 10 questions, output your current percentage understanding of the implementation plan so the user can see progress
- do NOT treat PLAN.md as complete until confidence reaches ≥%d%%

For each section, write only what is required to enable clear task breakdown:

- SUMMARY
  - one-paragraph overview of the chosen approach

- APPROACH
  - high-level strategy
  - explicit tradeoff decisions
  - no code

- COMPONENTS
  - logical components/modules
  - clear responsibility boundaries

- DATA
  - data shapes, enums, tables, files
  - no schema or serialization code unless unavoidable

- INTERFACES
  - commands, inputs, outputs, side effects
  - files and artifacts touched

- RISKS
  - top technical or design risks
  - mitigation per risk

- TESTING
  - validation strategy
  - test types, not test code

Rules:
- focus on HOW, not WHAT
- use BRAINSTORM.md as research context only; SPEC.md remains the binding contract
- do not restate requirements
- do not introduce new scope beyond SPEC.md
- do not write tasks
- avoid code unless strictly necessary
- keep language dense and factual
- Plan gate: acceptance criteria must be testable and mapped to explicit evidence in PLAN.md before sign-off
- ensure plan respects constraints defined in CONSTITUTION.md
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times

The output of PLAN.md must make TASKS.md obvious and deterministic.
`, goalPct))

	prompt := sb.String()

	fmt.Println("\n" + dim + "────────────────────────────────────────────────────────────────────────" + reset)
	if planCopy {
		fmt.Println(whiteBold + "Agent prompt copied to clipboard" + reset)
	} else {
		fmt.Println(whiteBold + "Copy this prompt to your coding agent:" + reset)
	}
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)

	if err := outputPrompt(prompt, outputOnly, planCopy); err != nil {
		fmt.Printf("failed to output prompt: %v\n", err)
		return
	}

	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
}

// outputWarpPlanPrompt outputs a prompt for Warp coding agent to fill PLAN.md from Warp plan.
func outputWarpPlanPrompt(planPath, specPath, brainstormPath string, feat *feature.Feature, cfg *config.Config, outputOnly bool) {
	projectRoot, _ := config.FindProjectRoot()
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	goalPct := cfg.GoalPercentage
	hasBrainstorm := document.Exists(brainstormPath)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("I have created a Warp plan for the feature: %s\n\n", feat.Slug))
	sb.WriteString("## File References\n")
	sb.WriteString("| Document | Path |\n")
	sb.WriteString("|----------|------|\n")
	sb.WriteString(fmt.Sprintf("| CONSTITUTION | %s |\n", constitutionPath))
	if hasBrainstorm {
		sb.WriteString(fmt.Sprintf("| BRAINSTORM | %s |\n", brainstormPath))
	}
	sb.WriteString(fmt.Sprintf("| SPEC | %s |\n", specPath))
	sb.WriteString(fmt.Sprintf("| PLAN | %s |\n", planPath))
	sb.WriteString(fmt.Sprintf("| Project Root | %s |\n\n", projectRoot))

	sb.WriteString(fmt.Sprintf("Please take the Warp plan you just generated and use it to fill out the PLAN.md document at:\n%s\n\n", planPath))
	sb.WriteString("Your task:\n")
	sb.WriteString(fmt.Sprintf("1. Read CONSTITUTION.md (file: %s) to understand project constraints and principles\n", constitutionPath))
	step := 2
	if hasBrainstorm {
		sb.WriteString(fmt.Sprintf("%d. Read BRAINSTORM.md (file: %s) for upstream research context\n", step, brainstormPath))
		step++
	}
	sb.WriteString(fmt.Sprintf("%d. Read the Warp plan you created and extract the key design decisions\n", step))
	step++
	sb.WriteString(fmt.Sprintf("%d. Read SPEC.md (file: %s) to ensure alignment with requirements\n", step, specPath))
	step++
	sb.WriteString(fmt.Sprintf("%d. Fill out each section of PLAN.md (file: %s), adding implementation details beyond what's in the Warp plan:\n\n", step, planPath))
	sb.WriteString("   - SUMMARY: one-paragraph overview (expand from Warp plan's high-level description)\n")
	sb.WriteString("   - APPROACH: detailed strategy and tradeoff decisions\n")
	sb.WriteString("   - COMPONENTS: logical modules with clear responsibility boundaries\n")
	sb.WriteString("   - DATA: data shapes, structures, and storage decisions\n")
	sb.WriteString("   - INTERFACES: commands, inputs, outputs, side effects\n")
	sb.WriteString("   - RISKS: technical risks with mitigation strategies\n")
	sb.WriteString("   - TESTING: validation strategy and test types\n\n")
	sb.WriteString(fmt.Sprintf("%d. Ensure PLAN.md has MORE detail than the Warp plan — it should make task breakdown obvious\n", step+1))
	sb.WriteString(fmt.Sprintf("%d. Ask clarifying questions until you reach ≥%d%% confidence that you understand any remaining ambiguities in the problem and desired solution\n", step+2, goalPct))
	sb.WriteString(fmt.Sprintf("%d. Use numbered lists\n", step+3))
	sb.WriteString(fmt.Sprintf("%d. Ask questions in batches of up to 10\n", step+4))
	sb.WriteString(fmt.Sprintf("%d. For every question, include your current best proposed solution or assumption\n", step+5))
	sb.WriteString(fmt.Sprintf("%d. State uncertainties\n", step+6))
	sb.WriteString(fmt.Sprintf("%d. After each batch of up to 10 questions, output your current percentage understanding so the user can see progress\n", step+7))
	sb.WriteString(fmt.Sprintf("%d. Reassess and continue with additional batches of up to 10 questions until PLAN.md is precise enough to produce a correct, production-quality implementation\n\n", step+8))
	sb.WriteString(fmt.Sprintf(`After completing PLAN.md:
- state your confidence level that TASKS.md can be derived unambiguously
- do NOT treat PLAN.md as complete until confidence reaches ≥%d%%

Rules:
- focus on HOW, not WHAT (SPEC covers WHAT)
- use BRAINSTORM.md as research context only; SPEC.md remains the binding contract
- do not restate requirements verbatim
- do not introduce new scope beyond the Warp plan and SPEC.md
- keep language dense and factual
- Plan gate: acceptance criteria must be testable and mapped to explicit evidence in PLAN.md before sign-off
- ensure plan respects constraints defined in CONSTITUTION.md
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature

The output of PLAN.md must make TASKS.md obvious and deterministic.
`, goalPct))

	prompt := sb.String()

	fmt.Println("\n" + dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(whiteBold + "📋 Warp Plan Integration" + reset)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println()
	fmt.Println(dim + "The following files have been created:" + reset)
	fmt.Printf("  • PLAN.md: %s\n", planPath)
	fmt.Printf("  • SPEC.md: %s\n", specPath)
	fmt.Println()
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	if planCopy {
		fmt.Println(whiteBold + "Warp plan prompt copied to clipboard" + reset)
	} else {
		fmt.Println(whiteBold + "Copy this prompt to continue with your Warp plan:" + reset)
	}
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)

	if err := outputPrompt(prompt, outputOnly, planCopy); err != nil {
		fmt.Printf("failed to output prompt: %v\n", err)
		return
	}

	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
}
