package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current feature status for coding agents",
	Long: `Display the active feature's status, including:
  - Feature name and ID
  - Business summary from SPEC.md
  - File existence (SPEC, PLAN, TASKS)
  - Task completion progress (from markdown checkboxes)
  - Suggested next action

Output is optimized for coding agents to quickly understand
which files to investigate for the current feature.`,
	Args: cobra.NoArgs,
	RunE: runStatus,
}

func init() {
	statusCmd.Flags().Bool("json", false, "output status as JSON")
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	version := currentVersion()

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	specsDir := cfg.SpecsPath(projectRoot)

	// find active feature
	feat, err := feature.FindActiveFeature(specsDir)
	if err != nil {
		return fmt.Errorf("failed to find active feature: %w", err)
	}

	if feat == nil {
		return outputNoActiveFeature(cmd.OutOrStdout(), jsonOutput, version)
	}

	// get full status
	status, err := feature.GetFeatureStatus(feat)
	if err != nil {
		return fmt.Errorf("failed to get feature status: %w", err)
	}

	if jsonOutput {
		return outputStatusJSON(cmd.OutOrStdout(), status, version)
	}

	return outputStatusText(cmd.OutOrStdout(), status, specsDir, version)
}

func determineNextAction(status *feature.FeatureStatus) string {
	if status.Files["brainstorm"].Exists && !status.Files["spec"].Exists {
		return fmt.Sprintf("Create specification from brainstorm: run `kit spec %s`", status.Name)
	}

	// check files in order
	if !status.Files["spec"].Exists {
		return fmt.Sprintf("Start research with `kit brainstorm %s` or create specification directly with `kit spec %s`", status.Name, status.Name)
	}

	if !status.Files["plan"].Exists {
		return fmt.Sprintf("Create implementation plan: run `kit plan %s`", status.Name)
	}

	if !status.Files["tasks"].Exists {
		return fmt.Sprintf("Create task list: run `kit tasks %s`", status.Name)
	}

	// tasks exist, check progress
	if status.Progress != nil && status.Progress.HasTasks() {
		incomplete := status.Progress.Incomplete()
		if incomplete > 0 {
			return fmt.Sprintf("Complete %d remaining task(s) in %s", incomplete, status.Files["tasks"].Path)
		}
		return "All tasks complete! Review and verify implementation."
	}

	// tasks file exists but no checkboxes found
	return fmt.Sprintf("Define tasks with markdown checkboxes in %s", status.Files["tasks"].Path)
}
