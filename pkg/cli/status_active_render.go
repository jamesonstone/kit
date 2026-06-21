package cli

import (
	"fmt"
	"io"
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
	// Legacy staged fallback.
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

func colorizeStatusText(style humanOutputStyle, text string, color string) string {
	if !style.enabled || color == "" {
		return text
	}
	return color + text + reset
}

func printStatusField(w io.Writer, style humanOutputStyle, label string, value string) error {
	_, err := fmt.Fprintf(w, "%s %s\n", style.label(label+":"), value)
	return err
}

func printFileStatus(w io.Writer, name string, fs feature.FileStatus) error {
	style := styleForWriter(w)
	if fs.Exists {
		mark := "✓"
		if style.enabled {
			mark = plan + "✓" + reset
		}
		_, err := fmt.Fprintf(w, "   %s   %s  %s\n", padRight(name, 10), mark, fs.Path)
		return err
	}

	mark := "✗"
	if style.enabled {
		mark = implement + "✗" + reset
	}
	_, err := fmt.Fprintf(w, "   %s   %s  %s\n", padRight(name, 10), mark, style.muted("(not created)"))
	return err
}

func printProgressLine(w io.Writer, status *feature.FeatureStatus) error {
	style := styleForWriter(w)
	if !statusUsesV2Workflow(status) {
		brainstormMark := fileProgressMarker(style, status.Files["brainstorm"].Exists, brainstorm)
		specMark := fileProgressMarker(style, status.Files["spec"].Exists, spec)
		planMark := fileProgressMarker(style, status.Files["plan"].Exists, plan)
		tasksMark := fileProgressMarker(style, status.Files["tasks"].Exists, tasks)

		if _, err := fmt.Fprintf(
			w,
			"BRAINSTORM %s -> SPEC %s -> PLAN %s -> TASKS %s",
			brainstormMark,
			specMark,
			planMark,
			tasksMark,
		); err != nil {
			return err
		}

		if status.Progress != nil && status.Progress.HasTasks() {
			progress := fmt.Sprintf("%d/%d complete", status.Progress.Complete, status.Progress.Total)
			if style.enabled {
				color := implement
				if status.Progress.Complete == status.Progress.Total {
					color = plan
				}
				progress = color + progress + reset
			}
			_, err := fmt.Fprintf(w, " (%s)", progress)
			return err
		}

		return nil
	}

	phases := []feature.Phase{
		feature.PhaseClarify,
		feature.PhaseReady,
		feature.PhaseImplement,
		feature.PhaseValidate,
		feature.PhaseReflect,
		feature.PhaseDeliver,
		feature.PhaseComplete,
	}
	labels := []string{"CLARIFY", "READY", "IMPLEMENT", "VALIDATE", "REFLECT", "DELIVER", "COMPLETE"}
	currentRank := v2StatusPhaseRank(status.Phase)
	parts := make([]string, 0, len(phases))
	for i, phase := range phases {
		marker := activeStatusPhaseProgressMarker(style, currentRank, v2StatusPhaseRank(phase), phaseColumnColor(phase))
		parts = append(parts, fmt.Sprintf("%s %s", labels[i], marker))
	}
	if _, err := fmt.Fprint(w, strings.Join(parts, " -> ")); err != nil {
		return err
	}

	if status.Progress != nil && status.Progress.HasTasks() {
		progress := fmt.Sprintf("%d/%d complete", status.Progress.Complete, status.Progress.Total)
		if style.enabled {
			color := implement
			if status.Progress.Complete == status.Progress.Total {
				color = plan
			}
			progress = color + progress + reset
		}
		_, err := fmt.Fprintf(w, " (%s)", progress)
		return err
	}

	return nil
}

func v2StatusPhaseRank(phase feature.Phase) int {
	switch phase {
	case feature.PhaseClarify:
		return 1
	case feature.PhaseReady:
		return 2
	case feature.PhaseImplement:
		return 3
	case feature.PhaseValidate:
		return 4
	case feature.PhaseReflect:
		return 5
	case feature.PhaseDeliver:
		return 6
	case feature.PhaseComplete:
		return 7
	default:
		return 0
	}
}

func activeStatusPhaseProgressMarker(style humanOutputStyle, currentRank int, phaseRank int, color string) string {
	marker := "○"
	if currentRank >= phaseRank && phaseRank > 0 {
		marker = "●"
	}
	if currentRank == phaseRank && currentRank != v2StatusPhaseRank(feature.PhaseComplete) {
		marker = "◐"
	}
	if !style.enabled {
		return marker
	}
	return color + marker + reset
}

func fileProgressMarker(style humanOutputStyle, exists bool, color string) string {
	if exists {
		if style.enabled {
			return color + "✓" + reset
		}
		return "✓"
	}
	if style.enabled {
		return dim + "✗" + reset
	}
	return "✗"
}
