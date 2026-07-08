package cli

import (
	"fmt"
	"strings"

	"github.com/jamesonstone/kit/internal/feature"
)

type statusWorkflowItem struct {
	label string
	color string
}

func formatPhaseValue(style humanOutputStyle, phase feature.Phase) string {
	if !style.enabled {
		return string(phase)
	}
	return phaseColumnColor(phase) + string(phase) + reset
}

func formatCurrentStepValue(style humanOutputStyle, phase feature.Phase) string {
	label := string(phase)
	switch phase {
	case feature.PhaseImplement:
		label = "implementation"
	case feature.PhaseValidate:
		label = "validation"
	case feature.PhaseReflect:
		label = "reflection"
	case feature.PhaseDeliver:
		label = "delivery"
	}
	return colorizeStatusText(style, label, phaseColumnColor(phase))
}

func formatStateValue(style humanOutputStyle, status *feature.FeatureStatus) string {
	if status.Removed {
		return colorizeStatusText(style, "removed", dim)
	}
	if status.Paused {
		return colorizeStatusText(style, "paused", constitution)
	}
	if status.Phase == feature.PhaseComplete {
		return colorizeStatusText(style, "complete", plan)
	}
	if status.Phase == feature.PhaseBlocked {
		return colorizeStatusText(style, "blocked", reflect)
	}
	return colorizeStatusText(style, "active (not paused)", plan)
}

func formatPausedValue(style humanOutputStyle, paused bool) string {
	if !style.enabled {
		if paused {
			return "yes"
		}
		return "no"
	}
	if paused {
		return constitution + "yes" + reset
	}
	return plan + "no" + reset
}

func formatTaskProgressValue(style humanOutputStyle, status *feature.FeatureStatus) string {
	if status.Progress != nil && status.Progress.HasTasks() {
		progress := fmt.Sprintf("%d/%d complete", status.Progress.Complete, status.Progress.Total)
		if incomplete := status.Progress.Incomplete(); incomplete > 0 {
			progress = fmt.Sprintf("%s (%d left)", progress, incomplete)
			return colorizeStatusText(style, progress, implement)
		}
		return colorizeStatusText(style, progress, plan)
	}
	if status.Files["tasks"].Exists {
		return colorizeStatusText(style, "TASKS.md has no markdown checkboxes", tasks)
	}
	return style.muted("not started")
}

func formatRemainingWorkValue(style humanOutputStyle, status *feature.FeatureStatus) string {
	if status.Removed {
		return colorizeStatusText(style, "nothing; feature is removed", dim)
	}
	if status.Phase == feature.PhaseComplete {
		return colorizeStatusText(style, "nothing; feature is complete", plan)
	}

	items := remainingWorkflowItems(status)
	if len(items) == 0 {
		return colorizeStatusText(style, "inspect feature documents", whiteBold)
	}

	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, colorizeStatusText(style, item.label, item.color))
	}
	return strings.Join(parts, " -> ")
}

func remainingWorkflowItems(status *feature.FeatureStatus) []statusWorkflowItem {
	if status.Removed || status.Phase == feature.PhaseComplete {
		return nil
	}

	if statusUsesV2Workflow(status) {
		switch status.Phase {
		case feature.PhaseClarify:
			return []statusWorkflowItem{
				{label: "clarify requirements", color: brainstorm},
				{label: "ready gate", color: spec},
				{label: "implement", color: implement},
				{label: "validate", color: tasks},
				{label: "reflect", color: reflect},
				{label: "deliver", color: spec},
				{label: "complete", color: plan},
			}
		case feature.PhaseReady:
			return []statusWorkflowItem{
				{label: "implement", color: implement},
				{label: "validate", color: tasks},
				{label: "reflect", color: reflect},
				{label: "deliver", color: spec},
				{label: "complete", color: plan},
			}
		case feature.PhaseValidate:
			return []statusWorkflowItem{
				{label: "validate acceptance criteria", color: tasks},
				{label: "reflect", color: reflect},
				{label: "deliver", color: spec},
				{label: "complete", color: plan},
			}
		case feature.PhaseDeliver:
			return []statusWorkflowItem{
				{label: "resolve delivery gate", color: spec},
				{label: "complete", color: plan},
			}
		case feature.PhaseBlocked:
			return []statusWorkflowItem{
				{label: "resolve blocker in SPEC.md", color: reflect},
			}
		case feature.PhaseImplement:
			return []statusWorkflowItem{
				{label: "complete SPEC.md implementation checklist", color: implement},
				{label: "validate", color: tasks},
				{label: "reflect", color: reflect},
				{label: "deliver", color: spec},
				{label: "complete", color: plan},
			}
		case feature.PhaseReflect:
			return []statusWorkflowItem{
				{label: "record reflection notes", color: reflect},
				{label: "sync documentation updates", color: spec},
				{label: "deliver", color: spec},
				{label: "complete", color: plan},
			}
		}
	}

	switch status.Phase {

	case feature.PhaseBrainstorm:
		return []statusWorkflowItem{
			{label: "SPEC.md", color: spec},
			{label: "PLAN.md", color: plan},
			{label: "TASKS.md", color: tasks},
			{label: "implement tasks", color: implement},
			{label: "reflect", color: reflect},
			{label: "complete", color: plan},
		}
	case feature.PhaseSpec:
		return []statusWorkflowItem{
			{label: "PLAN.md", color: plan},
			{label: "TASKS.md", color: tasks},
			{label: "implement tasks", color: implement},
			{label: "reflect", color: reflect},
			{label: "complete", color: plan},
		}
	case feature.PhasePlan:
		return []statusWorkflowItem{
			{label: "TASKS.md", color: tasks},
			{label: "implement tasks", color: implement},
			{label: "reflect", color: reflect},
			{label: "complete", color: plan},
		}
	case feature.PhaseTasks:
		return []statusWorkflowItem{
			{label: "define TASKS.md work items", color: tasks},
			{label: "implement tasks", color: implement},
			{label: "reflect", color: reflect},
			{label: "complete", color: plan},
		}
	case feature.PhaseImplement:
		label := "complete implementation tasks"
		if status.Progress != nil && status.Progress.HasTasks() {
			label = fmt.Sprintf("complete %d remaining task(s)", status.Progress.Incomplete())
		}
		return []statusWorkflowItem{
			{label: label, color: implement},
			{label: "reflect", color: reflect},
			{label: "complete", color: plan},
		}
	case feature.PhaseReflect:
		return []statusWorkflowItem{
			{label: "run reflection/verification", color: reflect},
			{label: "complete", color: plan},
		}
	default:
		return []statusWorkflowItem{
			{label: "inspect feature documents", color: whiteBold},
		}
	}
}
