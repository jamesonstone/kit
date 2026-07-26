package cli

import (
	"fmt"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

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
		return fmt.Sprintf("Use native planning, then scaffold durable feature memory with `kit spec %s` when material rationale exists", status.Name)
	}

	workflowVersion := statusWorkflowVersion(status)
	if workflowVersion == document.WorkflowVersionV3 {
		switch status.Phase {
		case feature.PhaseClarify:
			return "Use native planning, then capture purpose, context, requirements, and the accepted plan in SPEC.md before implementation"
		case feature.PhaseReady:
			return "Begin implementation from the accepted native plan and keep material decisions and discoveries current"
		case feature.PhaseImplement:
			return "Continue implementation and update SPEC.md when consequential decisions or discoveries occur"
		case feature.PhaseValidate:
			return "Run relevant validation and record exact outcomes in SPEC.md"
		case feature.PhaseReflect:
			return "Curate the actual outcome and route durable knowledge into the correct repository memory"
		case feature.PhaseDeliver:
			return fmt.Sprintf("Repository memory is curated. Complete the feature with `kit complete %s` after any requested delivery mutation is resolved", status.Name)
		case feature.PhaseBlocked:
			return "Resolve the blocker recorded in SPEC.md or ask for the missing material decision"
		case feature.PhaseComplete:
			return "Feature is complete"
		}
	}

	if workflowVersion == document.WorkflowVersionV2 {
		switch status.Phase {
		case feature.PhaseClarify:
			return "Resolve remaining material non-discoverable ambiguity in SPEC.md until unresolved questions are 0 and acceptance criteria are binary-verifiable"
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

func statusWorkflowVersion(status *feature.FeatureStatus) int {
	if status == nil || status.Files == nil {
		return 0
	}
	spec, ok := status.Files["spec"]
	if !ok || !spec.Exists {
		return 0
	}
	doc, err := document.ParseFile(spec.Path, document.TypeSpec)
	if err != nil {
		return 0
	}
	if doc.Metadata == nil {
		return 0
	}
	return doc.Metadata.WorkflowVersion
}

func statusUsesV2Workflow(status *feature.FeatureStatus) bool {
	return statusWorkflowVersion(status) == document.WorkflowVersionV2
}
