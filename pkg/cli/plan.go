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
	planCmd.Flags().BoolVar(&planCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	planCmd.Flags().BoolVar(&planOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	addPromptOnlyFlag(planCmd)
	rootCmd.AddCommand(planCmd)
}

func runPlan(cmd *cobra.Command, args []string) error {
	planForce, _ := cmd.Flags().GetBool("force")
	warpMode, _ := cmd.Flags().GetBool("warp")
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

	if promptOnly {
		if planForce {
			return fmt.Errorf("--prompt-only cannot be used with --force")
		}
		return runPlanPromptOnly(args, projectRoot, cfg, warpMode, outputOnly)
	}

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

	if !outputOnly {
		fmt.Printf("📋 Creating plan for feature: %s\n", feat.DirName)
	}

	// check prerequisites
	specPath := filepath.Join(feat.Path, "SPEC.md")
	if !document.Exists(specPath) {
		if planForce || cfg.AllowOutOfOrder {
			// create empty SPEC.md
			if err := document.Write(specPath, templates.Spec); err != nil {
				return fmt.Errorf("failed to create SPEC.md: %w", err)
			}
			if !outputOnly {
				fmt.Println("  ✓ Created SPEC.md (--force)")
			}
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
		if !outputOnly {
			fmt.Println("  ✓ Created PLAN.md")
		}
	} else if !outputOnly {
		fmt.Println("  ✓ PLAN.md already exists")
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
		fmt.Printf("\n✅ Plan for '%s' ready!\n", feat.Slug)
		printNumberedNextSteps([]string{
			fmt.Sprintf("Edit %s to define the implementation approach", planPath),
			fmt.Sprintf("Run 'kit tasks %s' to create executable tasks", feat.Slug),
		})
	}
	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")

	if warpMode {
		return outputWarpPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, outputOnly)
	}

	return outputStandardPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, outputOnly)
}

func runPlanPromptOnly(args []string, projectRoot string, cfg *config.Config, warpMode, outputOnly bool) error {
	specsDir := cfg.SpecsPath(projectRoot)

	var (
		feat *feature.Feature
		err  error
	)

	if len(args) == 0 {
		feat, err = selectFeatureForPlanPromptOnly(specsDir)
		if err != nil {
			return err
		}
	} else {
		feat, err = feature.Resolve(specsDir, args[0])
		if err != nil {
			return fmt.Errorf("feature '%s' not found. Run 'kit spec %s' first to create it", args[0], args[0])
		}
	}

	specPath := filepath.Join(feat.Path, "SPEC.md")
	planPath := filepath.Join(feat.Path, "PLAN.md")
	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")

	if !document.Exists(specPath) {
		return fmt.Errorf("SPEC.md not found. Run 'kit spec %s' first", feat.Slug)
	}
	if !document.Exists(planPath) {
		return fmt.Errorf("PLAN.md not found. Run 'kit plan %s' first", feat.Slug)
	}

	if warpMode {
		return outputWarpPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, outputOnly)
	}

	return outputStandardPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, outputOnly)
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

	printSelectionHeader("Select a feature to plan:")
	for i, f := range candidates {
		fmt.Printf("  [%d] %s\n", i+1, f.DirName)
	}
	fmt.Println()
	fmt.Print(selectionPrompt(os.Stdout))

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

func selectFeatureForPlanPromptOnly(specsDir string) (*feature.Feature, error) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	var candidates []feature.Feature
	for _, f := range features {
		if document.Exists(filepath.Join(f.Path, "SPEC.md")) &&
			document.Exists(filepath.Join(f.Path, "PLAN.md")) {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no plans available to regenerate prompts for\n\nRun 'kit plan <feature>' first")
	}

	printSelectionHeader("Select a feature to regenerate the plan prompt for:")
	for i, f := range candidates {
		fmt.Printf("  [%d] %s (%s)\n", i+1, f.DirName, f.Phase)
	}
	fmt.Println()
	fmt.Print(selectionPrompt(os.Stdout))

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

func appendPlanDependencyInventoryStep(
	sb *strings.Builder,
	step int,
	planPath string,
	specPath string,
	brainstormPath string,
	hasBrainstorm bool,
) int {
	sb.WriteString(fmt.Sprintf("%d. Populate or refresh the `## DEPENDENCIES` table in `%s` before sign-off:\n", step, planPath))
	sb.WriteString(fmt.Sprintf("   - carry forward still-relevant dependencies from `%s`\n", specPath))
	if hasBrainstorm {
		sb.WriteString(fmt.Sprintf("   - carry forward still-relevant dependencies from `%s`\n", brainstormPath))
	}
	sb.WriteString("   - include skills, MCP tools, repo docs, design refs, APIs, libraries, datasets, assets, and other resources that shape the implementation strategy\n")
	sb.WriteString("   - use the columns `Dependency`, `Type`, `Location`, `Used For`, and `Status`\n")
	sb.WriteString("   - `Status` must be one of `active`, `optional`, or `stale`\n")
	sb.WriteString("   - for Figma or MCP-driven design dependencies, store the exact design URL or file/node reference in `Location`\n")
	sb.WriteString("   - if a dependency influenced the implementation strategy but is no longer current, keep it in the table with `Status` = `stale`\n")
	sb.WriteString("   - if no additional dependencies apply, keep the default `none` row\n")
	return step + 1
}

// outputStandardPlanPrompt outputs the standard coding agent prompt.
func outputStandardPlanPrompt(planPath, specPath, brainstormPath string, feat *feature.Feature, cfg *config.Config, outputOnly bool) error {
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
	step = appendPlanDependencyInventoryStep(&sb, step, planPath, specPath, brainstormPath, hasBrainstorm)
	step = appendNumberedSteps(
		&sb,
		step,
		clarificationLoopSteps(
			goalPct,
			"Reassess and continue with additional batches of up to 10 questions "+
				"until the plan is precise enough to produce a correct, "+
				"production-quality implementation",
		),
	)
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

- DEPENDENCIES
  - the docs, tools, design refs, APIs, libraries, datasets, assets, and other resources shaping the implementation strategy
  - keep exact URLs or file/node refs in the Location column
  - use Status = active, optional, or stale

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
- the ## DEPENDENCIES section must be current before sign-off and must keep exact locations for external design inputs
- do not write tasks
- avoid code unless strictly necessary
- keep language dense and factual
- Plan gate: acceptance criteria must be testable and mapped to explicit evidence in PLAN.md before sign-off
- ensure plan respects constraints defined in CONSTITUTION.md
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times

The output of PLAN.md must make TASKS.md obvious and deterministic.
`, goalPct))
	appendNonEmptySectionRules(&sb, "`PLAN.md`")

	prompt := sb.String()

	if err := outputPromptWithClipboardDefault(prompt, outputOnly, planCopy); err != nil {
		return fmt.Errorf("failed to output prompt: %w", err)
	}

	return nil
}

// outputWarpPlanPrompt outputs a prompt for Warp coding agent to fill PLAN.md from Warp plan.
func outputWarpPlanPrompt(planPath, specPath, brainstormPath string, feat *feature.Feature, cfg *config.Config, outputOnly bool) error {
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
	step = appendPlanDependencyInventoryStep(&sb, step, planPath, specPath, brainstormPath, hasBrainstorm)
	sb.WriteString(fmt.Sprintf("%d. Fill out each section of PLAN.md (file: %s), adding implementation details beyond what's in the Warp plan:\n\n", step, planPath))
	sb.WriteString("   - SUMMARY: one-paragraph overview (expand from Warp plan's high-level description)\n")
	sb.WriteString("   - APPROACH: detailed strategy and tradeoff decisions\n")
	sb.WriteString("   - COMPONENTS: logical modules with clear responsibility boundaries\n")
	sb.WriteString("   - DATA: data shapes, structures, and storage decisions\n")
	sb.WriteString("   - INTERFACES: commands, inputs, outputs, side effects\n")
	sb.WriteString("   - DEPENDENCIES: the resources that shape the implementation strategy, with exact URLs or file/node refs in `Location`\n")
	sb.WriteString("   - RISKS: technical risks with mitigation strategies\n")
	sb.WriteString("   - TESTING: validation strategy and test types\n\n")
	sb.WriteString(fmt.Sprintf("%d. Ensure PLAN.md has MORE detail than the Warp plan — it should make task breakdown obvious\n", step+1))
	appendNumberedSteps(
		&sb,
		step+2,
		clarificationLoopSteps(
			goalPct,
			"Reassess and continue with additional batches of up to 10 questions "+
				"until PLAN.md is precise enough to produce a correct, "+
				"production-quality implementation",
		),
	)
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(`After completing PLAN.md:
- state your confidence level that TASKS.md can be derived unambiguously
- do NOT treat PLAN.md as complete until confidence reaches ≥%d%%

Rules:
- focus on HOW, not WHAT (SPEC covers WHAT)
- use BRAINSTORM.md as research context only; SPEC.md remains the binding contract
- do not restate requirements verbatim
- do not introduce new scope beyond the Warp plan and SPEC.md
- the ## DEPENDENCIES section must be current before sign-off and must keep exact locations for external design inputs
- keep language dense and factual
- Plan gate: acceptance criteria must be testable and mapped to explicit evidence in PLAN.md before sign-off
- ensure plan respects constraints defined in CONSTITUTION.md
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature

The output of PLAN.md must make TASKS.md obvious and deterministic.
`, goalPct))
	appendNonEmptySectionRules(&sb, "`PLAN.md`")

	prompt := sb.String()

	if !outputOnly {
		fmt.Println()
		fmt.Println(whiteBold + "Warp Plan Integration" + reset)
		fmt.Println(dim + "The following files have been created:" + reset)
		fmt.Printf("  • PLAN.md: %s\n", planPath)
		fmt.Printf("  • SPEC.md: %s\n\n", specPath)
	}

	if err := outputPromptWithClipboardDefault(prompt, outputOnly, planCopy); err != nil {
		return fmt.Errorf("failed to output prompt: %w", err)
	}

	return nil
}
