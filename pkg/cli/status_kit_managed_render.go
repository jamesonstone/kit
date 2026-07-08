package cli

import (
	"fmt"
	"io"
	"strings"
)

func outputStatusKitManagedSummaryForHuman(out io.Writer, summary *statusKitManagedSummary) error {
	if summary == nil {
		return nil
	}
	style := styleForWriter(out)
	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, style.title("🧩", "Kit-managed files")); err != nil {
		return err
	}
	if err := printStatusField(out, style, "State", formatKitManagedState(style, summary.State)); err != nil {
		return err
	}
	if err := printStatusField(out, style, "Managed files", formatManagedFileSummary(summary.ManagedFiles)); err != nil {
		return err
	}
	if err := printStatusField(out, style, "Registry", formatRegistrySummary(summary.Registry)); err != nil {
		return err
	}
	if err := printStatusKitManagedItems(out, summary.Items); err != nil {
		return err
	}
	for _, action := range summary.NextActions {
		if err := printStatusField(out, style, "Next", action); err != nil {
			return err
		}
	}
	return nil
}

func printStatusKitManagedItems(out io.Writer, items []statusKitManagedItem) error {
	if len(items) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(out, "  Items:"); err != nil {
		return err
	}
	for _, item := range firstStatusKitManagedItems(items, 8) {
		detail := ""
		if item.Detail != "" {
			detail = " - " + item.Detail
		}
		if _, err := fmt.Fprintf(out, "    - %s [%s]%s\n", item.Path, item.State, detail); err != nil {
			return err
		}
	}
	if extra := len(items) - 8; extra > 0 {
		if _, err := fmt.Fprintf(out, "    - ... %d more\n", extra); err != nil {
			return err
		}
	}
	return nil
}

func firstStatusKitManagedItems(items []statusKitManagedItem, limit int) []statusKitManagedItem {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func formatKitManagedState(style humanOutputStyle, state string) string {
	label := strings.ToUpper(state)
	if !style.enabled {
		return label
	}
	switch state {
	case statusKitManagedStateCurrent:
		return plan + label + reset
	case statusKitManagedStateRefreshAvailable:
		return implement + label + reset
	case statusKitManagedStateAttentionNeeded:
		return tasks + label + reset
	default:
		return dim + label + reset
	}
}

func formatManagedFileSummary(summary statusManagedFilesSummary) string {
	if summary.Unchecked {
		if summary.CheckError != "" {
			return fmt.Sprintf("unchecked (%s)", summary.CheckError)
		}
		return "unchecked"
	}
	if summary.Planned == 0 {
		return fmt.Sprintf("current (%d checked)", summary.Skipped)
	}
	parts := []string{}
	if summary.Created > 0 {
		parts = append(parts, fmt.Sprintf("%d created", summary.Created))
	}
	if summary.Updated > 0 {
		parts = append(parts, fmt.Sprintf("%d updated", summary.Updated))
	}
	if summary.Merged > 0 {
		parts = append(parts, fmt.Sprintf("%d merged", summary.Merged))
	}
	if len(parts) == 0 {
		return fmt.Sprintf("%d planned change(s)", summary.Planned)
	}
	return fmt.Sprintf("%d planned change(s): %s", summary.Planned, strings.Join(parts, ", "))
}

func formatRegistrySummary(summary statusRegistrySummary) string {
	var parts []string
	parts = append(parts, "local state only")
	parts = append(parts, fmt.Sprintf("%d managed", summary.Managed))
	if summary.Missing > 0 {
		parts = append(parts, fmt.Sprintf("%d missing", summary.Missing))
	}
	if summary.LocalCustom > 0 {
		parts = append(parts, fmt.Sprintf("%d local custom", summary.LocalCustom))
	}
	if summary.Conflicts > 0 {
		parts = append(parts, fmt.Sprintf("%d conflict", summary.Conflicts))
	}
	if summary.Unknown > 0 {
		parts = append(parts, fmt.Sprintf("%d unknown", summary.Unknown))
	}
	return strings.Join(parts, ", ")
}
