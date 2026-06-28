package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
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

	specsDir := cfg.SpecsPath(projectRoot)

	if allOutput {
		return runStatusAll(cmd, projectRoot, specsDir, cfg, jsonOutput, version)
	}

	// find active feature
	feat, err := feature.FindActiveFeatureWithState(specsDir, cfg)
	if err != nil {
		return fmt.Errorf("failed to find active feature: %w", err)
	}

	if feat == nil {
		backlog, backlogErr := feature.ListBacklogFeatures(specsDir, cfg)
		if backlogErr != nil {
			return fmt.Errorf("failed to list backlog items: %w", backlogErr)
		}
		if err := outputNoActiveFeature(cmd.OutOrStdout(), jsonOutput, version, len(backlog)); err != nil {
			return err
		}
		return outputProjectRefreshStatusForHuman(cmd.OutOrStdout(), projectRoot, cfg, jsonOutput)
	}

	// get full status
	status, err := feature.GetFeatureStatus(feat)
	if err != nil {
		return fmt.Errorf("failed to get feature status: %w", err)
	}
	attachFeatureNotesStatus(projectRoot, status, feat.DirName)

	if jsonOutput {
		return outputStatusJSON(cmd.OutOrStdout(), status, version)
	}

	if err := outputStatusText(cmd.OutOrStdout(), status, version); err != nil {
		return err
	}
	return outputProjectRefreshStatusForHuman(cmd.OutOrStdout(), projectRoot, cfg, jsonOutput)
}

func runStatusAll(
	cmd *cobra.Command,
	projectRoot string,
	specsDir string,
	cfg *config.Config,
	jsonOutput bool,
	version string,
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
		return outputAllFeaturesStatusJSON(cmd.OutOrStdout(), activeStatus, entries, backlogCount, version)
	}

	if err := outputAllFeaturesStatusText(cmd.OutOrStdout(), activeStatus, entries, backlogCount, version); err != nil {
		return err
	}
	return outputProjectRefreshStatusForHuman(cmd.OutOrStdout(), projectRoot, cfg, jsonOutput)
}

func outputProjectRefreshStatusForHuman(out io.Writer, projectRoot string, cfg *config.Config, jsonOutput bool) error {
	if jsonOutput {
		return nil
	}
	status, err := calculateProjectRefreshStatus(projectRoot, cfg, time.Now().UTC())
	if err != nil {
		_, writeErr := fmt.Fprintf(out, "  ⚠ Project refresh due status unavailable: %v\n", err)
		return writeErr
	}
	return printProjectRefreshStatusSummary(out, status)
}

func buildAllFeatureStatusEntries(projectRoot string, specsDir string, cfg *config.Config) ([]allFeatureStatusEntry, int, error) {
	features, err := feature.ListFeaturesWithState(specsDir, cfg)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list features: %w", err)
	}

	entries := make([]allFeatureStatusEntry, 0, len(features))
	backlogCount := 0
	liveFeatureDirs := make(map[string]struct{}, len(features))
	for i := range features {
		liveFeatureDirs[features[i].DirName] = struct{}{}
		status, err := feature.GetFeatureStatus(&features[i])
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get feature status for %s: %w", features[i].DirName, err)
		}
		attachFeatureNotesStatus(projectRoot, status, features[i].DirName)

		isBacklog := feature.IsBacklogItem(features[i])
		if isBacklog {
			backlogCount++
		}

		entries = append(entries, allFeatureStatusEntry{
			Status:     status,
			IsBacklog:  isBacklog,
			NextAction: determineNextAction(status),
		})
	}
	for _, removed := range cfg.RemovedFeatures {
		if removed.DirName == "" {
			continue
		}
		if _, exists := liveFeatureDirs[removed.DirName]; exists {
			continue
		}

		status := removedFeatureStatus(projectRoot, cfg, removed)
		entries = append(entries, allFeatureStatusEntry{
			Status:     status,
			IsRemoved:  true,
			NextAction: determineNextAction(status),
		})
	}
	sortAllFeatureStatusEntries(entries)

	return entries, backlogCount, nil
}

func determineNextAction(status *feature.FeatureStatus) string {
	if status.Removed {
		if status.Notes != nil && status.Notes.Exists {
			return fmt.Sprintf("Feature is removed. Retained notes are available at %s for follow-up work.", status.Notes.Path)
		}
		return "Feature is removed. No retained notes are available."
	}
	nextAction := determineUnpausedNextAction(status)
	if !status.Paused {
		return nextAction
	}

	return fmt.Sprintf(
		"Feature is paused. Run `kit resume %s` when ready. Suggested next step after resume: %s",
		status.Name,
		nextAction,
	)
}

func determineUnpausedNextAction(status *feature.FeatureStatus) string {
	if status.Files["brainstorm"].Exists && !status.Files["spec"].Exists {
		return fmt.Sprintf("Create specification from brainstorm: run `kit spec %s`", status.Name)
	}

	if !status.Files["spec"].Exists {
		return fmt.Sprintf("Start v2 workflow with `kit spec %s`; use `kit legacy brainstorm %s` only for staged migration research", status.Name, status.Name)
	}

	if statusUsesV2Workflow(status) {
		switch status.Phase {
		case feature.PhaseClarify:
			return "Continue v2 clarification in SPEC.md until unresolved questions are 0 and acceptance criteria are binary-verifiable"
		case feature.PhaseReady:
			return "Begin v2 implementation from the SPEC.md implementation plan and task checklist"
		case feature.PhaseImplement:
			return "Continue v2 implementation and keep SPEC.md task status current"
		case feature.PhaseValidate:
			return "Run validation mapped 1:1 to SPEC.md acceptance criteria and record evidence"
		case feature.PhaseReflect:
			return "Record reflection notes, documentation sync status, and remaining risks in SPEC.md"
		case feature.PhaseDeliver:
			return fmt.Sprintf("Delivery gate is ready. Complete the feature with `kit complete %s` after any requested delivery mutation is resolved", status.Name)
		case feature.PhaseBlocked:
			return "Feature is blocked. Resolve the blocker recorded in SPEC.md or ask the user for the missing decision"
		case feature.PhaseComplete:
			return "Feature is complete"
		}
	}

	// Legacy staged fallback.
	if !status.Files["plan"].Exists {
		return fmt.Sprintf("Legacy staged feature: create implementation plan with `kit legacy plan %s`", status.Name)
	}

	if !status.Files["tasks"].Exists {
		return fmt.Sprintf("Legacy staged feature: create task list with `kit legacy tasks %s`", status.Name)
	}

	// tasks exist, check progress
	if status.Progress != nil && status.Progress.HasTasks() {
		incomplete := status.Progress.Incomplete()
		if incomplete > 0 {
			return fmt.Sprintf("Complete %d remaining task(s) in %s", incomplete, status.Files["tasks"].Path)
		}
		return fmt.Sprintf("All tasks are marked complete. If legacy staged coding has not started, run `kit legacy implement %s`; otherwise review and validate implementation.", status.Name)
	}

	// tasks file exists but no checkboxes found
	return fmt.Sprintf("Define tasks with markdown checkboxes in %s", status.Files["tasks"].Path)
}

func statusUsesV2Workflow(status *feature.FeatureStatus) bool {
	if status == nil || status.Files == nil {
		return false
	}
	spec, ok := status.Files["spec"]
	if !ok || !spec.Exists {
		return false
	}
	doc, err := document.ParseFile(spec.Path, document.TypeSpec)
	if err != nil {
		return false
	}
	return doc.Metadata != nil && doc.Metadata.WorkflowVersion == 2
}
