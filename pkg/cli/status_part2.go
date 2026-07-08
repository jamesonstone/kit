package cli

import (
	"fmt"
	"io"
	"time"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

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

	if !status.Files["plan"].Exists {
		return fmt.Sprintf("Legacy staged feature: create implementation plan with `kit legacy plan %s`", status.Name)
	}

	if !status.Files["tasks"].Exists {
		return fmt.Sprintf("Legacy staged feature: create task list with `kit legacy tasks %s`", status.Name)
	}

	if status.Progress != nil && status.Progress.HasTasks() {
		incomplete := status.Progress.Incomplete()
		if incomplete > 0 {
			return fmt.Sprintf("Complete %d remaining task(s) in %s", incomplete, status.Files["tasks"].Path)
		}
		return fmt.Sprintf("All tasks are marked complete. If legacy staged coding has not started, run `kit legacy implement %s`; otherwise review and validate implementation.", status.Name)
	}

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
