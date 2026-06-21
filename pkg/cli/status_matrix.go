package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/jamesonstone/kit/internal/feature"
)

func printAllFeaturesProgressMatrix(
	w io.Writer,
	entries []allFeatureStatusEntry,
	activeStatus *feature.FeatureStatus,
) error {
	style := styleForWriter(w)
	const (
		featureWidth  = 28
		stateWidth    = 9
		progressWidth = 5
		notesWidth    = 5
	)

	header := statusMatrixField(style, "Feature", featureWidth, whiteBold, false) + "  " +
		statusMatrixField(style, "CLRFY", 5, brainstorm, false) + " " +
		statusMatrixField(style, "READY", 5, spec, false) + " " +
		statusMatrixField(style, "IMPL", 4, implement, false) + " " +
		statusMatrixField(style, "VALD", 4, tasks, false) + " " +
		statusMatrixField(style, "REFL", 4, reflect, false) + " " +
		statusMatrixField(style, "DLVR", 4, spec, false) + " " +
		statusMatrixField(style, "DONE", 4, plan, false) + "  " +
		statusMatrixField(style, "State", stateWidth, whiteBold, false) + "  " +
		statusMatrixField(style, "Prog", progressWidth, whiteBold, true) + "  " +
		statusMatrixField(style, "Notes", notesWidth, whiteBold, true)
	if _, err := fmt.Fprintln(w, header); err != nil {
		return err
	}

	separator := strings.Repeat("-", featureWidth) + "  " +
		strings.Repeat("-", 5) + " " +
		strings.Repeat("-", 5) + " " +
		strings.Repeat("-", 4) + " " +
		strings.Repeat("-", 4) + " " +
		strings.Repeat("-", 4) + " " +
		strings.Repeat("-", 4) + "  " +
		strings.Repeat("-", stateWidth) + "  " +
		strings.Repeat("-", progressWidth) + "  " +
		strings.Repeat("-", notesWidth)
	if _, err := fmt.Fprintln(w, style.muted(separator)); err != nil {
		return err
	}

	for _, entry := range entries {
		line := statusMatrixField(
			style,
			truncateString(fmt.Sprintf("%s-%s", entry.Status.ID, entry.Status.Name), featureWidth),
			featureWidth,
			statusMatrixFeatureColor(entry, activeStatus),
			false,
		) + "  " +
			phaseProgressField(style, entry.Status.Phase, feature.PhaseClarify, 5) + " " +
			phaseProgressField(style, entry.Status.Phase, feature.PhaseReady, 5) + " " +
			phaseProgressField(style, entry.Status.Phase, feature.PhaseImplement, 4) + " " +
			phaseProgressField(style, entry.Status.Phase, feature.PhaseValidate, 4) + " " +
			phaseProgressField(style, entry.Status.Phase, feature.PhaseReflect, 4) + " " +
			phaseProgressField(style, entry.Status.Phase, feature.PhaseDeliver, 4) + " " +
			phaseProgressField(style, entry.Status.Phase, feature.PhaseComplete, 4) + "  " +
			statusMatrixField(style, allFeaturesStateLabel(entry, activeStatus), stateWidth, statusMatrixStateColor(entry, activeStatus), false) + "  " +
			statusMatrixField(style, allFeaturesProgressLabel(entry.Status), progressWidth, statusMatrixProgressColor(entry.Status), true) + "  " +
			statusMatrixField(style, allFeaturesNotesLabel(entry.Status), notesWidth, statusMatrixNotesColor(entry.Status), true)
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}

	return nil
}

func allFeaturesNotesLabel(status *feature.FeatureStatus) string {
	if status.Notes != nil && status.Notes.Exists {
		return "yes"
	}
	return "no"
}

func statusMatrixNotesColor(status *feature.FeatureStatus) string {
	if status.Notes != nil && status.Notes.Exists {
		return brainstorm
	}
	return dim
}

func allFeaturesStateLabel(entry allFeatureStatusEntry, activeStatus *feature.FeatureStatus) string {
	if entry.IsRemoved || entry.Status.Removed {
		return "REMOVED"
	}
	if entry.IsBacklog {
		return "BACKLOG"
	}
	if entry.Status.Paused {
		return "PAUSED"
	}
	if sameFeatureStatus(entry.Status, activeStatus) {
		return "ACTIVE"
	}
	if entry.Status.Phase == feature.PhaseComplete {
		return "COMPLETE"
	}
	return "INFLIGHT"
}

func allFeaturesProgressLabel(status *feature.FeatureStatus) string {
	if status.Removed {
		return "-"
	}
	if status.Progress != nil && status.Progress.HasTasks() {
		return fmt.Sprintf("%d/%d", status.Progress.Complete, status.Progress.Total)
	}
	return "-"
}

func phaseProgressField(
	style humanOutputStyle,
	current feature.Phase,
	target feature.Phase,
	width int,
) string {
	marker := phaseProgressMarker(current, target)
	color := dim
	switch marker {
	case "●":
		color = phaseColumnColor(target)
	case "◐":
		color = whiteBold
	}
	return statusMatrixField(style, marker, width, color, true)
}

func phaseProgressMarker(current feature.Phase, target feature.Phase) string {
	order := map[feature.Phase]int{
		feature.PhaseClarify:   1,
		feature.PhaseReady:     2,
		feature.PhaseImplement: 3,
		feature.PhaseValidate:  4,
		feature.PhaseReflect:   5,
		feature.PhaseDeliver:   6,
		feature.PhaseComplete:  7,
	}

	currentIndex, ok := order[current]
	if !ok {
		return "○"
	}
	targetIndex := order[target]

	if targetIndex < currentIndex {
		return "●"
	}
	if targetIndex == currentIndex {
		if current == feature.PhaseComplete {
			return "●"
		}
		return "◐"
	}
	return "○"
}

func phaseColumnColor(phase feature.Phase) string {
	switch phase {
	case feature.PhaseClarify, feature.PhaseBrainstorm:
		return brainstorm
	case feature.PhaseReady, feature.PhaseSpec:
		return spec
	case feature.PhasePlan:
		return plan
	case feature.PhaseTasks:
		return tasks
	case feature.PhaseImplement:
		return implement
	case feature.PhaseValidate:
		return tasks
	case feature.PhaseReflect:
		return reflect
	case feature.PhaseDeliver:
		return spec
	case feature.PhaseComplete:
		return plan
	default:
		return whiteBold
	}
}

func statusMatrixField(style humanOutputStyle, text string, width int, color string, alignRight bool) string {
	raw := text
	if alignRight {
		raw = padLeft(raw, width)
	} else {
		raw = padRight(raw, width)
	}
	if !style.enabled || color == "" {
		return raw
	}
	return color + raw + reset
}

func statusMatrixFeatureColor(entry allFeatureStatusEntry, activeStatus *feature.FeatureStatus) string {
	if entry.IsRemoved || entry.Status.Removed {
		return dim
	}
	if entry.IsBacklog {
		return brainstorm
	}
	if entry.Status.Paused {
		return constitution
	}
	if sameFeatureStatus(entry.Status, activeStatus) {
		return whiteBold
	}
	if entry.Status.Phase == feature.PhaseComplete {
		return plan
	}
	return ""
}

func statusMatrixStateColor(entry allFeatureStatusEntry, activeStatus *feature.FeatureStatus) string {
	if entry.IsRemoved || entry.Status.Removed {
		return dim
	}
	if entry.IsBacklog {
		return brainstorm
	}
	if entry.Status.Paused {
		return constitution
	}
	if sameFeatureStatus(entry.Status, activeStatus) {
		return spec
	}
	if entry.Status.Phase == feature.PhaseComplete {
		return plan
	}
	return dim
}

func statusMatrixProgressColor(status *feature.FeatureStatus) string {
	if status.Removed {
		return dim
	}
	if status.Progress == nil || !status.Progress.HasTasks() {
		return dim
	}
	if status.Progress.Complete == status.Progress.Total {
		return plan
	}
	return implement
}

func sameFeatureStatus(a, b *feature.FeatureStatus) bool {
	if a == nil || b == nil {
		return false
	}
	if a.Path != "" && b.Path != "" {
		return a.Path == b.Path
	}
	return a.ID == b.ID && a.Name == b.Name
}

func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-1]) + "…"
}

func padLeft(s string, width int) string {
	runeCount := len([]rune(s))
	if runeCount >= width {
		return s
	}
	return spaces(width-runeCount) + s
}

func padRight(s string, width int) string {
	runeCount := len([]rune(s))
	if runeCount >= width {
		return s
	}
	return s + spaces(width-runeCount)
}

func spaces(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}
