package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/jamesonstone/kit/internal/feature"
)

func outputNoActiveFeature(w io.Writer, asJSON bool, version string, backlogCount int) error {
	return outputNoActiveFeatureWithManagedStatus(w, asJSON, version, backlogCount, nil)
}

func outputNoActiveFeatureWithManagedStatus(
	w io.Writer,
	asJSON bool,
	version string,
	backlogCount int,
	kitManaged *statusKitManagedSummary,
) error {
	message := "No active feature in progress"
	if backlogCount > 0 {
		message = fmt.Sprintf("No active feature in progress (%d backlog item(s) available)", backlogCount)
	}

	if asJSON {
		data, err := json.MarshalIndent(statusJSONPayload(nil, version, message, kitManaged), "", "  ")
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
			style.muted("Run `kit spec <feature-name>` to start a Kit v2 feature."),
		); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, style.muted(formatKitVersionInfo(version)))
	if err != nil {
		return err
	}
	return outputStatusKitManagedSummaryForHuman(w, kitManaged)
}

func outputStatusJSON(w io.Writer, status *feature.FeatureStatus, version string) error {
	return outputStatusJSONWithManagedStatus(w, status, version, nil)
}

func outputStatusJSONWithManagedStatus(
	w io.Writer,
	status *feature.FeatureStatus,
	version string,
	kitManaged *statusKitManagedSummary,
) error {
	data, err := json.MarshalIndent(statusJSONPayload(status, version, "", kitManaged), "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

func outputStatusText(w io.Writer, status *feature.FeatureStatus, version string) error {
	style := styleForWriter(w)
	featureName := fmt.Sprintf("%s-%s", status.ID, status.Name)

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, style.title("📊", fmt.Sprintf("Active Feature: %s", featureName))); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w, style.title("📍", "At a glance")); err != nil {
		return err
	}
	if err := printStatusField(w, style, "Feature", featureName); err != nil {
		return err
	}
	if err := printStatusField(w, style, "State", formatStateValue(style, status)); err != nil {
		return err
	}
	if err := printStatusField(w, style, "Paused", formatPausedValue(style, status.Paused)); err != nil {
		return err
	}
	if err := printStatusField(w, style, "Current step", formatCurrentStepValue(style, status.Phase)); err != nil {
		return err
	}
	if err := printStatusField(w, style, "Tasks", formatTaskProgressValue(style, status)); err != nil {
		return err
	}
	if err := printStatusField(w, style, "Left", formatRemainingWorkValue(style, status)); err != nil {
		return err
	}
	if status.Paused {
		if err := printStatusField(w, style, "Next", fmt.Sprintf("Run `kit resume %s` when ready", status.Name)); err != nil {
			return err
		}
		if err := printStatusField(w, style, "After resume", determineUnpausedNextAction(status)); err != nil {
			return err
		}
	} else if err := printStatusField(w, style, "Next", determineNextAction(status)); err != nil {
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

	progressTitle := "Artifact progress"
	if statusUsesV2Workflow(status) {
		progressTitle = "V2 phase progress"
	}
	if _, err := fmt.Fprintln(w, style.title("📈", progressTitle)); err != nil {
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

	if _, err := fmt.Fprintln(w, style.title("📁", "Files")); err != nil {
		return err
	}
	for _, file := range []struct {
		label string
		key   string
	}{
		{label: "SPEC.md", key: "spec"},
		{label: "BRAINSTORM.md", key: "brainstorm"},
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

	_, err := fmt.Fprintln(w, style.muted(formatKitVersionInfo(version)))
	return err
}

func outputAllFeaturesStatusJSON(
	w io.Writer,
	activeStatus *feature.FeatureStatus,
	entries []allFeatureStatusEntry,
	backlogCount int,
	version string,
	kitManaged ...*statusKitManagedSummary,
) error {
	payload := map[string]interface{}{
		"mode":           "all",
		"kit_version":    version,
		"active_feature": activeStatus,
		"backlog_count":  backlogCount,
		"features":       entries,
	}
	if len(kitManaged) > 0 && kitManaged[0] != nil {
		payload["kit_managed"] = kitManaged[0]
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}
