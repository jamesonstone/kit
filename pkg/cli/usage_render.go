package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/v3/internal/usage"
)

func renderUsageReport(cmd *cobra.Command, report usage.Report) error {
	out := cmd.OutOrStdout()
	if _, err := fmt.Fprintf(out, "Kit usage since %s: %d calls, %d succeeded, %d failed\n", report.Since.Format("2006-01-02"), report.TotalCalls, report.Successes, report.Failures); err != nil {
		return err
	}
	for _, item := range report.Commands {
		if _, err := fmt.Fprintf(out, "  %-24s %5d calls  %4d failed\n", item.Command, item.Calls, item.Failures); err != nil {
			return err
		}
	}
	if len(report.ZeroUseCommands) > 0 {
		if _, err := fmt.Fprintf(out, "Zero observed use: %s\n", strings.Join(report.ZeroUseCommands, ", ")); err != nil {
			return err
		}
	}
	if len(report.Diagnostics) > 0 {
		_, err := fmt.Fprintf(out, "Diagnostics: %d; run `kit usage status` for storage details.\n", len(report.Diagnostics))
		return err
	}
	return nil
}

func renderUsageStatus(cmd *cobra.Command, status usage.StorageStatus) error {
	out := cmd.OutOrStdout()
	if _, err := fmt.Fprintf(out, "Usage collection: enabled=%t global=%t project=%s\n", status.Enabled, status.GlobalEnabled, status.ProjectState); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "Storage: %s (%d shards, %d/%d bytes, %d-day retention)\n", status.Directory, status.ShardCount, status.TotalBytes, status.MaxTotalBytes, status.RetentionDays); err != nil {
		return err
	}
	for _, diagnostic := range status.Diagnostics {
		if _, err := fmt.Fprintf(out, "  %s: %s\n", diagnostic.Level, diagnostic.Message); err != nil {
			return err
		}
	}
	return nil
}
