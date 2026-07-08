package cli

import (
	"fmt"
	"io"

	"github.com/jamesonstone/kit/internal/feature"
)

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
		if _, err := fmt.Fprintln(w, style.muted("No features found. Run `kit spec <feature-name>` to start a Kit v2 feature.")); err != nil {
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
	if _, err := fmt.Fprintln(w, style.muted("Legend: ● complete, ◐ current phase, ○ not reached; Notes=yes means docs/notes are retained")); err != nil {
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
	kitManaged ...*statusKitManagedSummary,
) map[string]interface{} {
	output := map[string]interface{}{
		"active_feature": status,
		"kit_version":    version,
	}
	if message != "" {
		output["message"] = message
	}
	if len(kitManaged) > 0 && kitManaged[0] != nil {
		output["kit_managed"] = kitManaged[0]
	}
	return output
}

func formatKitVersionInfo(version string) string {
	return fmt.Sprintf("ℹ️ Kit version: %s", version)
}
