package cli

import (
	"fmt"
	"io"

	"github.com/jamesonstone/kit/internal/feature"
)

func formatPhaseValue(style humanOutputStyle, phase feature.Phase) string {
	if !style.enabled {
		return string(phase)
	}
	return phaseColumnColor(phase) + string(phase) + reset
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
	brainstormMark := fileProgressMarker(style, status.Files["brainstorm"].Exists, brainstorm)
	specMark := fileProgressMarker(style, status.Files["spec"].Exists, spec)
	planMark := fileProgressMarker(style, status.Files["plan"].Exists, plan)
	tasksMark := fileProgressMarker(style, status.Files["tasks"].Exists, tasks)

	if _, err := fmt.Fprintf(
		w,
		"BRAINSTORM %s → SPEC %s → PLAN %s → TASKS %s",
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
