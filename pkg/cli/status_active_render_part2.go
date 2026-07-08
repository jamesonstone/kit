package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/jamesonstone/kit/internal/feature"
)

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
