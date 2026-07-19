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

	"github.com/jamesonstone/kit/internal/config"
)

const (
	healthStateUpdated   = "updated"
	healthStateUnhealthy = "unhealthy"
)

type healthChangeSummary struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
	Merged  int `json:"merged"`
	Skipped int `json:"skipped"`
}

type healthFileChange struct {
	Path   string `json:"path"`
	Action string `json:"action"`
}

type healthReport struct {
	State         string              `json:"state"`
	Managed       bool                `json:"managed"`
	DryRun        bool                `json:"dry_run"`
	Changes       healthChangeSummary `json:"changes"`
	Files         []healthFileChange  `json:"files,omitempty"`
	RegistryState string              `json:"registry_state,omitempty"`
	ProjectCheck  string              `json:"project_check"`
	Notes         []string            `json:"notes,omitempty"`
	NextActions   []string            `json:"next_actions,omitempty"`
	CheckError    string              `json:"check_error,omitempty"`
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Refresh and validate Kit-managed project health",
	Long: `Apply safe Kit-managed rules, instructions, configuration, and scaffold
updates, then validate the project contract. Local customizations and conflicts
are never force-overwritten; they remain explicit maintenance work for review.`,
	Args: cobra.NoArgs,
	RunE: runHealth,
}

func init() {
	healthCmd.Flags().Bool("dry-run", false, "preview Kit health changes without writing files")
	healthCmd.Flags().Bool("diff", false, "print the planned unified diff with --dry-run")
	healthCmd.Flags().Bool("json", false, "output the Kit health result as JSON")
	rootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, _ []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	diffOutput, _ := cmd.Flags().GetBool("diff")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	if diffOutput && !dryRun {
		return fmt.Errorf("--diff requires --dry-run")
	}
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
	report := healthReport{
		Managed:      cfg.IsHealthManaged(),
		DryRun:       dryRun,
		ProjectCheck: "not_run",
	}
	if !report.Managed {
		report.State = statusKitManagedStateDisabled
		return writeHealthReport(cmd.OutOrStdout(), report, jsonOutput)
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), statusKitManagedRefreshTimeout)
	defer cancel()
	plan, err := buildInitRefreshPlan(ctx, projectRoot, initRefreshOptions{dryRun: dryRun, diff: diffOutput, outputOnly: true})
	if err != nil {
		var registryErr *initRefreshRegistryError
		if errors.As(err, &registryErr) || errors.Is(err, context.DeadlineExceeded) {
			report.State = statusKitManagedStateUnknown
			report.RegistryState = statusKitManagedStateUnknown
			report.CheckError = err.Error()
			report.NextActions = []string{"rerun `kit health` when registry access is restored"}
			return writeHealthReport(cmd.OutOrStdout(), report, jsonOutput)
		}
		return err
	}
	report.Changes = healthChangeSummary{
		Created: plan.stats.created,
		Updated: plan.stats.updated,
		Merged:  plan.stats.merged,
		Skipped: plan.stats.skipped,
	}
	report.Files = healthChangedFiles(plan.changes)
	report.Notes = append(report.Notes, plan.notes...)

	if dryRun {
		report.State = healthPlannedState(report)
		report.RegistryState = report.State
		report.NextActions = healthNextActions(report)
		if diffOutput {
			if diff := renderInitRefreshDiff(plan.changes); strings.TrimSpace(diff) != "" {
				if _, err := fmt.Fprint(cmd.OutOrStdout(), diff); err != nil {
					return err
				}
			}
		}
		return writeHealthReport(cmd.OutOrStdout(), report, jsonOutput)
	}

	if err := applyInitRefreshFileChangesAtomically(plan.changes); err != nil {
		return err
	}
	updatedCfg, err := config.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to reload config after Kit health refresh: %w", err)
	}
	registryReport, err := buildRegistryStatusReport(projectRoot, updatedCfg)
	if err != nil {
		return err
	}
	report.RegistryState = registryReport.State
	if registryReport.CheckError != "" {
		report.CheckError = registryReport.CheckError
	}

	var checkOutput bytes.Buffer
	checkErr := checkProjectContractTo(&checkOutput, projectRoot, updatedCfg)
	if checkErr == nil {
		report.ProjectCheck = "passed"
	} else {
		report.ProjectCheck = "failed"
		report.CheckError = checkErr.Error()
	}
	report.State = healthFinalState(report, checkErr)
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

func healthChangedFiles(changes []initRefreshFileChange) []healthFileChange {
	files := make([]healthFileChange, 0, len(changes))
	for _, change := range changes {
		if change.result == instructionFileSkipped {
			continue
		}
		files = append(files, healthFileChange{Path: change.relativePath, Action: dryRunActionLabel(change.result)})
	}
	return files
}

func healthPlannedState(report healthReport) string {
	if healthNotesNeedAttention(report.Notes) {
		return statusKitManagedStateAttentionNeeded
	}
	if len(report.Files) > 0 {
		return statusKitManagedStateRefreshAvailable
	}
	return statusKitManagedStateCurrent
}

func healthFinalState(report healthReport, checkErr error) string {
	if checkErr != nil {
		return healthStateUnhealthy
	}
	if healthNotesNeedAttention(report.Notes) || report.RegistryState == statusKitManagedStateAttentionNeeded || report.RegistryState == statusKitManagedStateRefreshAvailable {
		return statusKitManagedStateAttentionNeeded
	}
	if report.RegistryState == statusKitManagedStateUnknown {
		return statusKitManagedStateUnknown
	}
	if len(report.Files) > 0 {
		return healthStateUpdated
	}
	return statusKitManagedStateCurrent
}

func healthNotesNeedAttention(notes []string) bool {
	for _, note := range notes {
		if strings.HasPrefix(note, "migrated exact generated V2 instruction artifacts") ||
			strings.HasPrefix(note, "exact legacy V1 instruction artifacts were refreshed") {
			continue
		}
		return true
	}
	return false
}

func healthNextActions(report healthReport) []string {
	var actions []string
	if report.State == statusKitManagedStateRefreshAvailable {
		actions = append(actions, "run `kit health` to apply the planned safe Kit-managed updates")
	}
	if report.State == statusKitManagedStateAttentionNeeded {
		actions = append(actions, "run `kit reconcile --output-only`, curate unresolved Kit-managed files, and rerun `kit health`")
	}
	if report.State == healthStateUpdated {
		actions = append(actions, "review and validate the complete Kit health diff before delivery")
	}
	if report.State == statusKitManagedStateUnknown {
		actions = append(actions, "rerun `kit health` when registry access is restored")
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
	if _, err := fmt.Fprintf(out, "Kit health: %s\n", report.State); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "Changes: %d created, %d updated, %d merged, %d skipped\n", report.Changes.Created, report.Changes.Updated, report.Changes.Merged, report.Changes.Skipped); err != nil {
		return err
	}
	for _, note := range report.Notes {
		if _, err := fmt.Fprintf(out, "Note: %s\n", note); err != nil {
			return err
		}
	}
	for _, action := range report.NextActions {
		if _, err := fmt.Fprintf(out, "Next: %s\n", action); err != nil {
			return err
		}
	}
	return nil
}
