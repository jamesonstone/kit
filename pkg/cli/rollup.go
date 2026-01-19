package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
)

var rollupCmd = &cobra.Command{
	Use:   "rollup",
	Short: "Generate PROJECT_PROGRESS_SUMMARY.md",
	Long: `Analyze all feature specifications and generate PROJECT_PROGRESS_SUMMARY.md.

The summary includes:
  - Feature progress table with phase, created date, and summary
  - Project intent section
  - Global constraints reference
  - Feature summaries with status, intent, approach, and pointers

This command runs automatically after feature creation/refinement.`,
	RunE: runRollup,
}

func init() {
	rootCmd.AddCommand(rollupCmd)
}

func runRollup(cmd *cobra.Command, args []string) error {
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

	// list features
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return fmt.Errorf("failed to list features: %w", err)
	}

	fmt.Printf("📊 Generating PROJECT_PROGRESS_SUMMARY.md\n")
	fmt.Printf("   Found %d feature(s)\n", len(features))

	// generate rollup
	if err := rollup.Generate(projectRoot, cfg); err != nil {
		return fmt.Errorf("failed to generate rollup: %w", err)
	}

	summaryPath := cfg.ProgressSummaryPath(projectRoot)
	fmt.Printf("  ✓ Updated %s\n", summaryPath)

	fmt.Printf("\n✅ Rollup complete!\n")

	return nil
}
