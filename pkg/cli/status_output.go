package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/jamesonstone/kit/internal/feature"
)

func outputNoActiveFeature(w io.Writer, asJSON bool, version string) error {
	if asJSON {
		data, err := json.MarshalIndent(statusJSONPayload(nil, version, "No active feature in progress"), "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w, string(data))
		return err
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "🤷 No active feature in progress 📭"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "Run `kit brainstorm` or `kit spec <feature-name>` to start a new feature."); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, formatKitVersionInfo(version))
	return err
}

func outputStatusJSON(w io.Writer, status *feature.FeatureStatus, version string) error {
	data, err := json.MarshalIndent(statusJSONPayload(status, version, ""), "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

func outputStatusText(w io.Writer, status *feature.FeatureStatus, specsDir, version string) error {
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "📊 "+whiteBold+"Active Feature: "+reset+"%s-%s\n", status.ID, status.Name); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if status.Summary != "" {
		if _, err := fmt.Fprintf(w, "📝 "+whiteBold+"Summary: "+reset+"%s\n", status.Summary); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(w, "📁 "+whiteBold+"Files:"+reset); err != nil {
		return err
	}
	for _, file := range []struct {
		label string
		key   string
	}{
		{label: "BRAINSTORM.md", key: "brainstorm"},
		{label: "SPEC.md", key: "spec"},
		{label: "PLAN.md", key: "plan"},
		{label: "TASKS.md", key: "tasks"},
	} {
		if err := printFileStatus(w, file.label, status.Files[file.key]); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	if _, err := fmt.Fprint(w, "📈 "+whiteBold+"Progress: "+reset); err != nil {
		return err
	}
	if err := printProgressLine(w, status); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	nextAction := determineNextAction(status)
	if _, err := fmt.Fprintf(w, "🎯 "+whiteBold+"Next: "+reset+"%s\n", nextAction); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if err := printAllFeaturesProgress(w, specsDir); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, formatKitVersionInfo(version))
	return err
}

func statusJSONPayload(
	status *feature.FeatureStatus,
	version string,
	message string,
) map[string]interface{} {
	output := map[string]interface{}{
		"active_feature": status,
		"kit_version":    version,
	}
	if message != "" {
		output["message"] = message
	}
	return output
}

func formatKitVersionInfo(version string) string {
	return fmt.Sprintf("ℹ️  Kit version: %s", version)
}

func printFileStatus(w io.Writer, name string, fs feature.FileStatus) error {
	if fs.Exists {
		_, err := fmt.Fprintf(w, "   %s   ✓  %s\n", padRight(name, 10), fs.Path)
		return err
	}
	_, err := fmt.Fprintf(w, "   %s   ✗  "+dim+"(not created)"+reset+"\n", padRight(name, 10))
	return err
}

func printProgressLine(w io.Writer, status *feature.FeatureStatus) error {
	brainstormMark := marker(status.Files["brainstorm"].Exists)
	specMark := marker(status.Files["spec"].Exists)
	planMark := marker(status.Files["plan"].Exists)
	tasksMark := marker(status.Files["tasks"].Exists)
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
		_, err := fmt.Fprintf(w, " (%d/%d complete)", status.Progress.Complete, status.Progress.Total)
		return err
	}
	return nil
}

func marker(exists bool) string {
	if exists {
		return "✓"
	}
	return "✗"
}

func printAllFeaturesProgress(w io.Writer, specsDir string) error {
	features, err := feature.ListFeatures(specsDir)
	if err != nil || len(features) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(w, "🗺️  "+whiteBold+"All Features:"+reset); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, dim+"| Feature              | BRN  | SPEC | PLAN | TASK | IMPL | REFL | DONE |"+reset); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, dim+"|----------------------|------|------|------|------|------|------|------|"+reset); err != nil {
		return err
	}
	for _, feat := range features {
		if err := printFeatureProgressRow(w, &feat); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(w)
	return err
}

func printFeatureProgressRow(w io.Writer, feat *feature.Feature) error {
	name := padRight(truncateString(feat.DirName, 20), 20)
	_, err := fmt.Fprintf(
		w,
		"| %s | %s | %s | %s | %s | %s | %s | %s |\n",
		name,
		phaseMarker(feat.Phase, feature.PhaseBrainstorm),
		phaseMarker(feat.Phase, feature.PhaseSpec),
		phaseMarker(feat.Phase, feature.PhasePlan),
		phaseMarker(feat.Phase, feature.PhaseTasks),
		phaseMarker(feat.Phase, feature.PhaseImplement),
		phaseMarker(feat.Phase, feature.PhaseReflect),
		phaseMarker(feat.Phase, feature.PhaseComplete),
	)
	return err
}

func padRight(s string, width int) string {
	runeCount := utf8.RuneCountInString(s)
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

func phaseMarker(current feature.Phase, target feature.Phase) string {
	order := map[feature.Phase]int{
		feature.PhaseBrainstorm: 1,
		feature.PhaseSpec:       2,
		feature.PhasePlan:       3,
		feature.PhaseTasks:      4,
		feature.PhaseImplement:  5,
		feature.PhaseReflect:    6,
		feature.PhaseComplete:   7,
	}
	currentIdx := order[current]
	targetIdx := order[target]
	if targetIdx < currentIdx {
		return plan + " ●  " + reset
	}
	if targetIdx == currentIdx {
		if current == feature.PhaseComplete {
			return plan + " ●  " + reset
		}
		return implement + " ◐  " + reset
	}
	return dim + " ○  " + reset
}

func truncateString(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLen-1]) + "…"
}
