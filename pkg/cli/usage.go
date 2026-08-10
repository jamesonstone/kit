package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/commandset"
	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/usage"
)

type usageReportOptions struct {
	since, command, project string
	jsonOutput              bool
}

type usageClearOptions struct {
	all, yes, jsonOutput bool
	command, project     string
}

func init() {
	rootCmd.AddCommand(newUsageCommand())
}

func newUsageCommand() *cobra.Command {
	reportOpts := &usageReportOptions{since: "90d"}
	cmd := &cobra.Command{
		Use:   "usage",
		Short: "Inspect and manage bounded local Kit usage data",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runUsageReport(cmd, reportOpts)
		},
	}
	addUsageReportFlags(cmd, reportOpts)
	cmd.AddCommand(newUsageReportCommand(), newUsageStatusCommand(), newUsageRefreshCommand())
	cmd.AddCommand(newUsageClearCommand(), newUsageToggleCommand(true), newUsageToggleCommand(false))
	return cmd
}

func newUsageReportCommand() *cobra.Command {
	opts := &usageReportOptions{since: "90d"}
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Report aggregated local Kit command usage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runUsageReport(cmd, opts)
		},
	}
	addUsageReportFlags(cmd, opts)
	return cmd
}

func addUsageReportFlags(cmd *cobra.Command, opts *usageReportOptions) {
	cmd.Flags().StringVar(&opts.since, "since", opts.since, "report window such as 24h, 30d, or 90d")
	cmd.Flags().StringVar(&opts.command, "command", "", "filter by normalized command path")
	cmd.Flags().StringVar(&opts.project, "project", "", "filter by anonymized project ID or current")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "emit machine-readable JSON")
}

func runUsageReport(cmd *cobra.Command, opts *usageReportOptions) error {
	duration, err := parseUsageDuration(opts.since)
	if err != nil {
		return err
	}
	filter := usage.Filter{Since: time.Now().UTC().Add(-duration), Command: strings.TrimSpace(opts.command)}
	if opts.project != "" {
		filter.ProjectID, err = resolveUsageProjectFilter(opts.project)
		if err != nil {
			return err
		}
		if filter.ProjectID == "" {
			filter.ProjectID = "no-recorded-project"
		}
	}
	report, err := usage.BuildReport(filter)
	if err != nil {
		return err
	}
	observed := map[string]bool{}
	for _, item := range report.Commands {
		observed[item.Command] = true
	}
	for _, path := range commandset.TelemetryPaths() {
		if !observed[path] {
			report.ZeroUseCommands = append(report.ZeroUseCommands, path)
		}
	}
	sort.Strings(report.ZeroUseCommands)
	if opts.jsonOutput {
		return writeUsageJSON(cmd, report)
	}
	return renderUsageReport(cmd, report)
}

func newUsageStatusCommand() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show usage collection and storage status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, _, _ := config.FindProjectRootOptional()
			status, err := usage.Status(root)
			if err != nil {
				return err
			}
			if jsonOutput {
				return writeUsageJSON(cmd, status)
			}
			return renderUsageStatus(cmd, status)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "emit machine-readable JSON")
	return cmd
}

func newUsageRefreshCommand() *cobra.Command {
	var dryRun, jsonOutput bool
	cmd := &cobra.Command{
		Use:   "refresh",
		Short: "Validate, rotate, and prune bounded usage storage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			result, err := usage.Refresh(dryRun)
			if err != nil {
				return err
			}
			if jsonOutput {
				return writeUsageJSON(cmd, result)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Usage refresh: changed=%t dry_run=%t pruned_events=%d pruned_shards=%d\n", result.Changed, result.DryRun, result.Status.PrunedEvents, result.Status.PrunedShards)
			return err
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview validation and pruning without writes")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "emit machine-readable JSON")
	return cmd
}

func newUsageClearCommand() *cobra.Command {
	opts := &usageClearOptions{}
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Remove selected local usage history",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return runUsageClear(cmd, opts) },
	}
	cmd.Flags().BoolVar(&opts.all, "all", false, "remove all usage history")
	cmd.Flags().StringVar(&opts.command, "command", "", "remove history for one normalized command path")
	cmd.Flags().StringVar(&opts.project, "project", "", "remove history for an anonymized project ID or current")
	cmd.Flags().BoolVar(&opts.yes, "yes", false, "confirm removal without prompting")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "emit machine-readable JSON")
	return cmd
}

func runUsageClear(cmd *cobra.Command, opts *usageClearOptions) error {
	if opts.all && (opts.command != "" || opts.project != "") {
		return fmt.Errorf("--all cannot be combined with --command or --project")
	}
	filter := usage.Filter{Command: strings.TrimSpace(opts.command)}
	var err error
	if opts.project != "" {
		filter.ProjectID, err = resolveUsageProjectFilter(opts.project)
		if err != nil {
			return err
		}
	}
	if !opts.all && filter.Command == "" && filter.ProjectID == "" {
		filter.ProjectID, err = resolveUsageProjectFilter("current")
		if err != nil {
			return err
		}
		if filter.ProjectID == "" {
			return fmt.Errorf("no usage identity exists for the current project")
		}
	}
	if !opts.yes {
		confirmed, err := confirmUsageClear(cmd.InOrStdin(), cmd.ErrOrStderr())
		if err != nil || !confirmed {
			return err
		}
	}
	removed, err := usage.Clear(filter, opts.all)
	if err != nil {
		return err
	}
	result := map[string]any{"schema_version": usage.SchemaVersion, "removed_events": removed}
	if opts.jsonOutput {
		return writeUsageJSON(cmd, result)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed %d usage events.\n", removed)
	return err
}

func newUsageToggleCommand(enabled bool) *cobra.Command {
	name := "disable"
	if enabled {
		name = "enable"
	}
	var global, project, jsonOutput bool
	cmd := &cobra.Command{
		Use:   name,
		Short: strings.ToUpper(name[:1]) + name[1:] + " local usage collection",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if global == project {
				return fmt.Errorf("choose exactly one of --global or --project")
			}
			scope := "global"
			root := ""
			if project {
				scope = "project"
				var found bool
				var err error
				root, found, err = config.FindProjectRootOptional()
				if err != nil || !found {
					return errors.New("project scope requires a Kit project")
				}
			}
			path, err := usage.SetEnabled(scope, root, enabled)
			if err != nil {
				return err
			}
			result := map[string]any{"schema_version": usage.SchemaVersion, "scope": scope, "enabled": enabled, "config_path": path}
			if jsonOutput {
				return writeUsageJSON(cmd, result)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Usage collection %s for %s scope in %s.\n", name+"d", scope, path)
			return err
		},
	}
	cmd.Flags().BoolVar(&global, "global", false, "change machine-wide usage collection")
	cmd.Flags().BoolVar(&project, "project", false, "change usage collection for the current project")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "emit machine-readable JSON")
	return cmd
}

func parseUsageDuration(value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if strings.HasSuffix(value, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(value, "d"))
		if err != nil || days <= 0 {
			return 0, fmt.Errorf("invalid --since value %q", value)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("invalid --since value %q", value)
	}
	return duration, nil
}

func resolveUsageProjectFilter(value string) (string, error) {
	if value != "current" {
		return strings.TrimSpace(value), nil
	}
	root, found, err := config.FindProjectRootOptional()
	if err != nil || !found {
		return "", fmt.Errorf("current project is not a Kit project")
	}
	return usage.CurrentProjectID(root)
}

func confirmUsageClear(in io.Reader, out io.Writer) (bool, error) {
	if _, err := fmt.Fprint(out, "Remove selected Kit usage history? [y/N]: "); err != nil {
		return false, err
	}
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	case "", "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("answer must be yes or no")
	}
}

func writeUsageJSON(cmd *cobra.Command, value any) error {
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
