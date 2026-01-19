package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

var planForce bool

var planCmd = &cobra.Command{
	Use:   "plan <feature>",
	Short: "Create or open a feature implementation plan",
	Long: `Create a new implementation plan for a feature.

Creates:
  - PLAN.md with required sections and placeholder comments

Prerequisites:
  - SPEC.md must exist (unless --force)

Updates PROJECT_PROGRESS_SUMMARY.md after creation.`,
	Args: cobra.ExactArgs(1),
	RunE: runPlan,
}

func init() {
	planCmd.Flags().BoolVar(&planForce, "force", false, "create missing SPEC.md with headers if it doesn't exist")
	rootCmd.AddCommand(planCmd)
}

func runPlan(cmd *cobra.Command, args []string) error {
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

	// resolve feature
	feat, err := feature.Resolve(specsDir, featureRef)
	if err != nil {
		return fmt.Errorf("feature '%s' not found. Run 'kit spec %s' first to create it", featureRef, featureRef)
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

	// output easy-to-copy instruction for coding agents
	goalPct := cfg.GoalPercentage
	fmt.Println("\n" + dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(whiteBold + "Copy this prompt to your coding agent:" + reset)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Printf(`
Please review and complete the implementation plan at %s.

This plan corresponds to the feature defined in:
- SPEC: %s

This is a new feature: %s

Your task:
1. Read SPEC.md fully and treat it as a fixed contract
2. Review the PLAN.md template and required sections
3. Identify any missing design decisions required for execution
4. Ask focused clarification questions until decisions can be made
5. Commit to concrete design decisions that make execution unambiguous

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
- do not introduce new scope
- do not write tasks
- avoid code unless strictly necessary
- keep language dense and factual
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times

The output of PLAN.md must make TASKS.md obvious and deterministic.

`, planPath, specPath, feat.Slug, goalPct)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)

	return nil
}
