package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"

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
  - Current v2 SPEC.md phase and paused state
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

func attachFeatureNotesStatus(projectRoot string, status *feature.FeatureStatus, dirName string) {
	if status == nil || dirName == "" {
		return
	}

	status.Notes = &feature.FileStatus{
		Exists: featureNotesPathExists(projectRoot, dirName),
		Path:   featureNotesPath(projectRoot, dirName),
	}
}

func removedFeatureStatus(projectRoot string, cfg *config.Config, removed config.RemovedFeature) *feature.FeatureStatus {
	number := removed.Number
	slug := removed.Slug
	if number == 0 || slug == "" {
		parsedNumber, parsedSlug, ok := feature.ParseDirName(removed.DirName)
		if ok {
			if number == 0 {
				number = parsedNumber
			}
			if slug == "" {
				slug = parsedSlug
			}
		}
	}

	featurePath := filepath.Join(projectRoot, cfg.SpecsDir, removed.DirName)
	status := &feature.FeatureStatus{
		ID:        formatStatusFeatureID(number),
		Name:      slug,
		Path:      featurePath,
		Phase:     feature.PhaseRemoved,
		Removed:   true,
		RemovedAt: removed.RemovedAt,
		Files: map[string]feature.FileStatus{
			"brainstorm": {Exists: false, Path: filepath.Join(featurePath, "BRAINSTORM.md")},
			"spec":       {Exists: false, Path: filepath.Join(featurePath, "SPEC.md")},
			"plan":       {Exists: false, Path: filepath.Join(featurePath, "PLAN.md")},
			"tasks":      {Exists: false, Path: filepath.Join(featurePath, "TASKS.md")},
		},
	}
	attachFeatureNotesStatus(projectRoot, status, removed.DirName)
	return status
}

func sortAllFeatureStatusEntries(entries []allFeatureStatusEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Status.ID != entries[j].Status.ID {
			return entries[i].Status.ID < entries[j].Status.ID
		}
		return entries[i].Status.Name < entries[j].Status.Name
	})
}

func formatStatusFeatureID(number int) string {
	return fmt.Sprintf("%04d", number)
}

func init() {
	statusCmd.Flags().Bool("json", false, "output status as JSON")
	statusCmd.Flags().Bool("all", false, "show all features instead of only the active feature")
	rootCmd.AddCommand(statusCmd)
}

type allFeatureStatusEntry struct {
	Status     *feature.FeatureStatus `json:"status"`
	IsBacklog  bool                   `json:"is_backlog"`
	IsRemoved  bool                   `json:"is_removed,omitempty"`
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

	feat, err := feature.FindActiveFeatureWithState(specsDir, cfg)
	if err != nil {
		return fmt.Errorf("failed to find active feature: %w", err)
	}

	if feat == nil {
		backlog, backlogErr := feature.ListBacklogFeatures(specsDir, cfg)
		if backlogErr != nil {
			return fmt.Errorf("failed to list backlog items: %w", backlogErr)
		}
		if err := outputNoActiveFeatureWithManagedStatus(cmd.OutOrStdout(), jsonOutput, version, len(backlog), kitManaged); err != nil {
			return err
		}
		return outputProjectStatusSummariesForHuman(cmd.OutOrStdout(), projectRoot, cfg, jsonOutput, nil)
	}

	status, err := feature.GetFeatureStatus(feat)
	if err != nil {
		return fmt.Errorf("failed to get feature status: %w", err)
	}
	attachFeatureNotesStatus(projectRoot, status, feat.DirName)

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
		attachFeatureNotesStatus(projectRoot, activeStatus, activeFeat.DirName)
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
	return outputProjectRefreshStatusForHuman(out, projectRoot, cfg, jsonOutput)
}
