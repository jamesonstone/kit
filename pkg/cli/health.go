package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/v3/internal/config"
)

const healthStateUnhealthy = "unhealthy"

type healthChangeSummary struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
	Merged  int `json:"merged"`
	Removed int `json:"removed"`
	Skipped int `json:"skipped"`
}

type healthReport struct {
	State        string                        `json:"state"`
	Managed      bool                          `json:"managed"`
	Changes      healthChangeSummary           `json:"pending_changes"`
	Files        []managedFileDeliverySnapshot `json:"pending_files,omitempty"`
	ProjectCheck string                        `json:"project_check"`
	Notes        []string                      `json:"notes,omitempty"`
	NextActions  []string                      `json:"next_actions,omitempty"`
	CheckError   string                        `json:"check_error,omitempty"`
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Report whether a Kit project is current and healthy (read-only)",
	Long: `Diagnose a Kit project without changing it: compute what ` + "`kit reconcile`" + `
would change, run the project check, and report actionable findings.

Health never writes files. Run ` + "`kit reconcile`" + ` to apply pending changes; it
keeps the primary checkout read-only by writing to a linked worktree.

Exits non-zero only when the project check fails. Pending reconcile changes are
reported as state refresh_available.`,
	Args: cobra.NoArgs,
	RunE: runHealth,
}

func init() {
	healthCmd.Flags().Bool("diff", false, "print the changes `kit reconcile` would make as a unified diff")
	healthCmd.Flags().Bool("json", false, "output the Kit health result as JSON")
	healthCmd.Flags().Bool("dry-run", false, "accepted for compatibility; health is always read-only")
	_ = healthCmd.Flags().MarkHidden("dry-run")
	rootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, _ []string) error {
	diffOutput, _ := cmd.Flags().GetBool("diff")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	if diffOutput && jsonOutput {
		return fmt.Errorf("--diff cannot be combined with --json")
	}

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}
	report := healthReport{Managed: cfg.IsHealthManaged(), ProjectCheck: "not_run"}
	if !report.Managed {
		report.State = statusKitManagedStateDisabled
		return writeHealthReport(cmd.OutOrStdout(), report, jsonOutput)
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), statusKitManagedRefreshTimeout)
	defer cancel()
	plan, err := buildInitRefreshPlan(ctx, projectRoot, initRefreshOptions{dryRun: true, outputOnly: true})
	if err != nil {
		var registryErr *initRefreshRegistryError
		if errors.As(err, &registryErr) || errors.Is(err, context.DeadlineExceeded) {
			report.State = statusKitManagedStateUnknown
			report.CheckError = err.Error()
			report.NextActions = []string{"rerun `kit health`"}
			return writeHealthReport(cmd.OutOrStdout(), report, jsonOutput)
		}
		return err
	}
	report.Changes = healthChangeSummary{
		Created: plan.stats.created,
		Updated: plan.stats.updated,
		Merged:  plan.stats.merged,
		Removed: plan.stats.removed,
		Skipped: plan.stats.skipped,
	}
	report.Files = managedFileDeliverySnapshotFromInitRefresh(projectRoot, plan.changes)
	report.Notes = append(report.Notes, plan.notes...)
	if diffOutput {
		if diff := renderInitRefreshDiff(plan.changes); strings.TrimSpace(diff) != "" {
			if _, err := fmt.Fprint(cmd.OutOrStdout(), diff); err != nil {
				return err
			}
		}
	}

	var checkOutput bytes.Buffer
	checkErr := checkProjectContractTo(&checkOutput, projectRoot, cfg)
	if checkErr == nil {
		report.ProjectCheck = "passed"
	} else {
		report.ProjectCheck = "failed"
		report.CheckError = checkErr.Error()
	}
	report.State = healthState(report, checkErr)
	report.NextActions = healthNextActions(report)

	if !jsonOutput {
		if _, err := io.Copy(cmd.OutOrStdout(), &checkOutput); err != nil {
			return err
		}
	}
	if err := writeHealthReport(cmd.OutOrStdout(), report, jsonOutput); err != nil {
		return err
	}
	if checkErr != nil {
		return &silentCLIError{err: checkErr}
	}
	return nil
}

func healthState(report healthReport, checkErr error) string {
	switch {
	case checkErr != nil:
		return healthStateUnhealthy
	case len(report.Notes) > 0:
		return statusKitManagedStateAttentionNeeded
	case len(report.Files) > 0:
		return statusKitManagedStateRefreshAvailable
	default:
		return statusKitManagedStateCurrent
	}
}

func healthNextActions(report healthReport) []string {
	var actions []string
	if len(report.Files) > 0 {
		actions = append(actions, "run `kit reconcile` to apply the pending changes (preview with `kit reconcile --dry-run --diff`)")
	}
	if len(report.Notes) > 0 {
		actions = append(actions, "review the notes; Kit keeps edited or ambiguous files until you resolve them")
	}
	if report.State == healthStateUnhealthy {
		actions = append(actions, "fix the project check findings, then rerun `kit health`")
	}
	return actions
}

func writeHealthReport(out io.Writer, report healthReport, jsonOutput bool) error {
	if jsonOutput {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	if report.State == statusKitManagedStateDisabled {
		_, err := fmt.Fprintln(out, "Kit health disabled (health.managed=false).")
		return err
	}
	lines := []string{
		fmt.Sprintf("Kit health: %s", report.State),
		fmt.Sprintf("Pending reconcile changes: %d created, %d updated, %d merged, %d removed", report.Changes.Created, report.Changes.Updated, report.Changes.Merged, report.Changes.Removed),
	}
	for _, note := range report.Notes {
		lines = append(lines, "Note: "+note)
	}
	for _, action := range report.NextActions {
		lines = append(lines, "Next: "+action)
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(out, line); err != nil {
			return err
		}
	}
	return nil
}
