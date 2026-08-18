package cli

import (
	"fmt"
	"io"
	"sort"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/feature"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current feature status for coding agents",
	Long: `Display the active feature's status, including:
  - Feature name and ID
  - Business summary from SPEC.md
  - Current living SPEC.md phase and paused state
  - Remaining workflow work
  - File existence (SPEC plus optional legacy artifacts)
  - Legacy task completion progress when TASKS.md is present
  - Suggested next action

Output is optimized for coding agents to quickly understand
what step is active, what remains, and which files to inspect.

Use --all for a project-wide overview.`,
	Args: cobra.NoArgs,
	RunE: runStatus,
}

func sortAllFeatureStatusEntries(entries []allFeatureStatusEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Status.ID != entries[j].Status.ID {
			return entries[i].Status.ID < entries[j].Status.ID
		}
		return entries[i].Status.Name < entries[j].Status.Name
	})
}

func init() {
	statusCmd.Flags().Bool("json", false, "output status as JSON")
	statusCmd.Flags().Bool("all", false, "show all features instead of only the active feature")
	rootCmd.AddCommand(statusCmd)
}

type allFeatureStatusEntry struct {
	Status     *feature.FeatureStatus `json:"status"`
	IsBacklog  bool                   `json:"is_backlog"`
	NextAction string                 `json:"next_action"`
}

func runStatus(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	allOutput, _ := cmd.Flags().GetBool("all")
	version := currentVersion()

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}
	kitManaged, err := buildStatusKitManagedSummary(projectRoot, cfg)
	if err != nil {
		return err
	}

	specsDir := cfg.SpecsPath(projectRoot)

	if allOutput {
		return runStatusAll(cmd, projectRoot, specsDir, cfg, jsonOutput, version, kitManaged)
	}

	// find active feature
	feat, err := feature.FindActiveFeatureWithState(specsDir, cfg)
	if err != nil {
		return fmt.Errorf("failed to find active feature: %w", err)
	}

	if feat == nil {
		if err := outputNoActiveFeatureWithManagedStatus(cmd.OutOrStdout(), jsonOutput, version, 0, kitManaged); err != nil {
			return err
		}
		return outputProjectStatusSummariesForHuman(cmd.OutOrStdout(), projectRoot, cfg, jsonOutput, nil)
	}

	// get full status
	status, err := feature.GetFeatureStatus(feat)
	if err != nil {
		return fmt.Errorf("failed to get feature status: %w", err)
	}
	if jsonOutput {
		return outputStatusJSONWithManagedStatus(cmd.OutOrStdout(), status, version, kitManaged)
	}

	if err := outputStatusText(cmd.OutOrStdout(), status, version); err != nil {
		return err
	}
	return outputProjectStatusSummariesForHuman(cmd.OutOrStdout(), projectRoot, cfg, jsonOutput, kitManaged)
}

func runStatusAll(
	cmd *cobra.Command,
	projectRoot string,
	specsDir string,
	cfg *config.Config,
	jsonOutput bool,
	version string,
	kitManaged *statusKitManagedSummary,
) error {
	activeFeat, err := feature.FindActiveFeatureWithState(specsDir, cfg)
	if err != nil {
		return fmt.Errorf("failed to find active feature: %w", err)
	}

	var activeStatus *feature.FeatureStatus
	if activeFeat != nil {
		activeStatus, err = feature.GetFeatureStatus(activeFeat)
		if err != nil {
			return fmt.Errorf("failed to get active feature status: %w", err)
		}
	}

	entries, backlogCount, err := buildAllFeatureStatusEntries(projectRoot, specsDir, cfg)
	if err != nil {
		return err
	}

	if jsonOutput {
		return outputAllFeaturesStatusJSON(cmd.OutOrStdout(), activeStatus, entries, backlogCount, version, kitManaged)
	}

	if err := outputAllFeaturesStatusText(cmd.OutOrStdout(), activeStatus, entries, backlogCount, version); err != nil {
		return err
	}
	return outputProjectStatusSummariesForHuman(cmd.OutOrStdout(), projectRoot, cfg, jsonOutput, kitManaged)
}

func outputProjectStatusSummariesForHuman(
	out io.Writer,
	projectRoot string,
	cfg *config.Config,
	jsonOutput bool,
	kitManaged *statusKitManagedSummary,
) error {
	if jsonOutput {
		return nil
	}
	if err := outputStatusKitManagedSummaryForHuman(out, kitManaged); err != nil {
		return err
	}
	return nil
}

func buildAllFeatureStatusEntries(_ string, specsDir string, cfg *config.Config) ([]allFeatureStatusEntry, int, error) {
	features, err := feature.ListFeaturesWithState(specsDir, cfg)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list features: %w", err)
	}

	entries := make([]allFeatureStatusEntry, 0, len(features))
	backlogCount := 0
	for i := range features {
		status, err := feature.GetFeatureStatus(&features[i])
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get feature status for %s: %w", features[i].DirName, err)
		}
		entries = append(entries, allFeatureStatusEntry{
			Status:     status,
			IsBacklog:  false,
			NextAction: determineNextAction(status),
		})
	}
	sortAllFeatureStatusEntries(entries)

	return entries, backlogCount, nil
}
