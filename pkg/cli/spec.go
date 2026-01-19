package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/git"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

var specNoBranch bool

var specCmd = &cobra.Command{
	Use:   "spec <feature>",
	Short: "Create or open a feature specification",
	Long: `Create a new feature specification or open an existing one.

Creates:
  - Feature directory (e.g., docs/specs/0001-my-feature/)
  - SPEC.md with required sections and placeholder comments
  - Git branch matching the feature directory name (unless --no-branch)

Updates PROJECT_PROGRESS_SUMMARY.md after creation.`,
	Args: cobra.ExactArgs(1),
	RunE: runSpec,
}

func init() {
	specCmd.Flags().BoolVar(&specNoBranch, "no-branch", false, "skip git branch creation")
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
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Edit %s to define the specification\n", specPath)
	fmt.Printf("  2. Run 'kit plan %s' to create the implementation plan\n", feat.Slug)

	// output easy-to-copy instruction for coding agents
	fmt.Println("\n" + dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println(whiteBold + "Copy this prompt to your coding agent:" + reset)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	goalPct := cfg.GoalPercentage
	fmt.Printf(`
Please review and complete the specification at %s.

This is a new feature: %s

Your task:
1. Read the SPEC.md template and understand the required sections
2. Analyze the codebase at %s to understand existing patterns
3. Ask clarifying questions in batches of 10 until you reach >= %d%% understanding
4. Fill in each section with clear, specific requirements:
   - PROBLEM: What problem does this feature solve?
   - GOALS: What are the measurable outcomes?
   - NON-GOALS: What is explicitly out of scope?
   - REQUIREMENTS: What must be true for this feature to be complete?
   - ACCEPTANCE: How do we verify the feature works?
   - EDGE-CASES: What unusual scenarios must be handled?

After each batch of questions, state your current understanding percentage.
Do NOT proceed to writing the spec until understanding >= %d%%.

Rules:
- keep language precise
- avoid implementation details
- focus on WHAT, not HOW
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times

`, specPath, feat.Slug, projectRoot, goalPct, goalPct)
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)

	return nil
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}
