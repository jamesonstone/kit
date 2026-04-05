package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/jamesonstone/kit/internal/feature"
)

func outputNoActiveFeature(w io.Writer, asJSON bool, version string, backlogCount int) error {
	message := "No active feature in progress"
	if backlogCount > 0 {
		message = fmt.Sprintf("No active feature in progress (%d backlog item(s) available)", backlogCount)
	}

	if asJSON {
		data, err := json.MarshalIndent(statusJSONPayload(nil, version, message), "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w, string(data))
		return err
	}

	style := styleForWriter(w)

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, style.title("🤷", "No active feature in progress")); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if backlogCount > 0 {
		if _, err := fmt.Fprintln(
			w,
			style.muted("Run `kit backlog` to review deferred items or `kit resume <feature>` to resume one."),
		); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintln(
			w,
			style.muted("Run `kit brainstorm` or `kit spec <feature-name>` to start a new feature."),
		); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, style.muted(formatKitVersionInfo(version)))
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

func outputStatusText(w io.Writer, status *feature.FeatureStatus, version string) error {
	style := styleForWriter(w)

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, style.title("📊", fmt.Sprintf("Active Feature: %s-%s", status.ID, status.Name))); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if status.Summary != "" {
		if _, err := fmt.Fprintln(w, style.title("📝", "Summary")); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, status.Summary); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(w, style.title("⏸️", "Lifecycle")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Phase: %s\n", formatPhaseValue(style, status.Phase)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Paused: %s\n", formatPausedValue(style, status.Paused)); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w, style.title("📁", "Files")); err != nil {
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

	if _, err := fmt.Fprintln(w, style.title("📈", "Progress")); err != nil {
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

	if _, err := fmt.Fprintln(w, style.title("🎯", "Next")); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, determineNextAction(status)); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, style.muted(formatKitVersionInfo(version)))
	return err
}

func outputAllFeaturesStatusJSON(
	w io.Writer,
	activeStatus *feature.FeatureStatus,
	entries []allFeatureStatusEntry,
	backlogCount int,
	version string,
) error {
	data, err := json.MarshalIndent(map[string]interface{}{
		"mode":           "all",
		"kit_version":    version,
		"active_feature": activeStatus,
		"backlog_count":  backlogCount,
		"features":       entries,
	}, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

func outputAllFeaturesStatusText(
	w io.Writer,
	activeStatus *feature.FeatureStatus,
	entries []allFeatureStatusEntry,
	backlogCount int,
	version string,
) error {
	style := styleForWriter(w)

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, style.title("📊", "Project Overview")); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	switch {
	case activeStatus != nil:
		if _, err := fmt.Fprintf(w, "Active feature: %s-%s\n", activeStatus.ID, activeStatus.Name); err != nil {
			return err
		}
	case len(entries) == 0:
		if _, err := fmt.Fprintln(w, "Active feature: none"); err != nil {
			return err
		}
	default:
		if _, err := fmt.Fprintln(w, "Active feature: none in progress"); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "Backlog items: %d\n", backlogCount); err != nil {
		return err
	}

	if len(entries) == 0 {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, style.muted("No features found. Run `kit brainstorm` or `kit spec <feature-name>` to start one.")); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		_, err := fmt.Fprintln(w, style.muted(formatKitVersionInfo(version)))
		return err
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, style.title("🗺️", "Features")); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if err := printAllFeaturesProgressMatrix(w, entries, activeStatus); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, style.muted("Legend: ● complete, ◐ current phase, ○ not reached")); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, style.muted(formatKitVersionInfo(version)))
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
	return fmt.Sprintf("ℹ️ Kit version: %s", version)
}
