package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/git"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold <feature>",
	Short: "Create a feature directory with all pipeline documents",
	Long: `Scaffold the full spec-driven development file structure for a feature.

Creates:
  - Feature directory (e.g., docs/specs/0001-my-feature/)
  - SPEC.md        — requirements (empty sections)
  - PLAN.md        — implementation plan (empty sections)
  - TASKS.md       — executable task list (empty sections)
  - ANALYSIS.md    — analysis scratchpad (empty sections)

No interactive prompts. No agent prompt output. Just files.

Updates PROJECT_PROGRESS_SUMMARY.md after creation.`,
	Args: cobra.ExactArgs(1),
	RunE: runScaffold,
}

func init() {
	scaffoldCmd.Flags().Bool("create-branch", false, "create and switch to a git branch matching the feature name")
	rootCmd.AddCommand(scaffoldCmd)
}

func runScaffold(cmd *cobra.Command, args []string) error {
	createBranch, _ := cmd.Flags().GetBool("create-branch")
	featureRef := args[0]

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	specsDir := cfg.SpecsPath(projectRoot)

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

	// scaffold all pipeline documents
	docs := []struct {
		name     string
		template string
	}{
		{"SPEC.md", templates.Spec},
		{"PLAN.md", templates.Plan},
		{"TASKS.md", templates.Tasks},
		{"ANALYSIS.md", templates.Analysis},
	}

	for _, d := range docs {
		path := filepath.Join(feat.Path, d.name)
		if document.Exists(path) {
			fmt.Printf("  ✓ %s already exists\n", d.name)
			continue
		}
		if err := document.Write(path, d.template); err != nil {
			return fmt.Errorf("failed to create %s: %w", d.name, err)
		}
		fmt.Printf("  ✓ Created %s\n", d.name)
	}

	// create git branch if requested
	if createBranch && git.IsRepo(projectRoot) {
		createBranchForFeature(projectRoot, feat, cfg)
	}

	// update rollup
	if err := rollup.Update(projectRoot, cfg); err != nil {
		fmt.Printf("  ⚠ Could not update PROJECT_PROGRESS_SUMMARY.md: %v\n", err)
	} else {
		fmt.Println("  ✓ Updated PROJECT_PROGRESS_SUMMARY.md")
	}

	fmt.Printf("\n✅ Feature '%s' scaffolded!\n", feat.Slug)
	fmt.Println("\nCreated structure:")
	fmt.Printf("  %s/\n", feat.DirName)
	fmt.Println("  ├── SPEC.md")
	fmt.Println("  ├── PLAN.md")
	fmt.Println("  ├── TASKS.md")
	fmt.Println("  └── ANALYSIS.md")
	fmt.Printf("\nNext: fill in SPEC.md with 'kit spec %s'\n", feat.Slug)

	return nil
}
