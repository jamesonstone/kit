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
	planCmd.Flags().BoolVar(&planCopy, "copy", false, "copy agent prompt to clipboard (suppresses stdout)")
	rootCmd.AddCommand(planCmd)
}

func runPlan(cmd *cobra.Command, args []string) error {
	planForce, _ := cmd.Flags().GetBool("force")
	warpMode, _ := cmd.Flags().GetBool("warp")

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

	if warpMode {
		outputWarpPlanPrompt(planPath, specPath, feat, cfg)
	} else {
		outputStandardPlanPrompt(planPath, specPath, feat, cfg)
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
func outputStandardPlanPrompt(planPath, specPath string, feat *feature.Feature, cfg *config.Config) {
	projectRoot, _ := config.FindProjectRoot()
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	goalPct := cfg.GoalPercentage

	prompt := fmt.Sprintf(`Please review and complete the implementation plan at %s.

This plan corresponds to the feature defined in:
- CONSTITUTION: %s (project-wide constraints)
- SPEC: %s

Feature: %s

Your task:
1. Read CONSTITUTION.md to understand project constraints and principles
2. Read SPEC.md fully and treat it as a fixed contract
3. Review the PLAN.md template and required sections
4. Identify any missing design decisions required for execution
5. Ask focused clarification questions until decisions can be made
6. Commit to concrete design decisions that make execution unambiguous

After each batch of questions:
- state your current understanding percentage of the implementation plan
- do NOT proceed to writing or updating PLAN.md until understanding >= %d%%

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
- do not restate requirements
- do not introduce new scope beyond SPEC.md
- do not write tasks
- avoid code unless strictly necessary
- keep language dense and factual
- ensure plan respects constraints defined in CONSTITUTION.md
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times

The output of PLAN.md must make TASKS.md obvious and deterministic.
`, planPath, constitutionPath, specPath, feat.Slug, goalPct)

	// copy to clipboard if requested
	if planCopy {
		if err := copyToClipboard(prompt); err != nil {
			fmt.Printf("failed to copy to clipboard: %v\n", err)
			return
		}
		fmt.Println("✓ Copied agent prompt to clipboard")
		return
	}

	fmt.Println("\n" + dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(whiteBold + "Copy this prompt to your coding agent:" + reset)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Print(prompt)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
}

// outputWarpPlanPrompt outputs a prompt for Warp coding agent to fill PLAN.md from Warp plan.
func outputWarpPlanPrompt(planPath, specPath string, feat *feature.Feature, cfg *config.Config) {
	projectRoot, _ := config.FindProjectRoot()
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	goalPct := cfg.GoalPercentage

	prompt := fmt.Sprintf(`I have created a Warp plan for the feature: %s

Please take the Warp plan you just generated and use it to fill out the PLAN.md document at:
%s

Context docs:
- CONSTITUTION: %s (project-wide constraints)
- SPEC: %s

Your task:
1. Read CONSTITUTION.md to understand project constraints and principles
2. Read the Warp plan you created and extract the key design decisions
3. Read SPEC.md to ensure alignment with requirements
4. Fill out each section of PLAN.md, adding implementation details beyond what's in the Warp plan:

   - SUMMARY: one-paragraph overview (expand from Warp plan's high-level description)
   - APPROACH: detailed strategy and tradeoff decisions
   - COMPONENTS: logical modules with clear responsibility boundaries
   - DATA: data shapes, structures, and storage decisions
   - INTERFACES: commands, inputs, outputs, side effects
   - RISKS: technical risks with mitigation strategies
   - TESTING: validation strategy and test types

5. Ensure PLAN.md has MORE detail than the Warp plan — it should make task breakdown obvious
6. Ask clarifying questions if any design decisions are ambiguous

After completing PLAN.md:
- state your confidence level that TASKS.md can be derived unambiguously
- aim for >= %d%% confidence before considering PLAN.md complete

Rules:
- focus on HOW, not WHAT (SPEC covers WHAT)
- do not restate requirements verbatim
- do not introduce new scope beyond the Warp plan and SPEC.md
- keep language dense and factual
- ensure plan respects constraints defined in CONSTITUTION.md
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature

The output of PLAN.md must make TASKS.md obvious and deterministic.
`, feat.Slug, planPath, constitutionPath, specPath, goalPct)

	// copy to clipboard if requested
	if planCopy {
		if err := copyToClipboard(prompt); err != nil {
			fmt.Printf("failed to copy to clipboard: %v\n", err)
			return
		}
		fmt.Println("✓ Copied agent prompt to clipboard")
		return
	}

	fmt.Println("\n" + dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(whiteBold + "📋 Warp Plan Integration" + reset)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println()
	fmt.Println(dim + "The following files have been created:" + reset)
	fmt.Printf("  • PLAN.md: %s\n", planPath)
	fmt.Printf("  • SPEC.md: %s\n", specPath)
	fmt.Println()
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(whiteBold + "Copy this prompt to continue with your Warp plan:" + reset)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Print(prompt)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
}
