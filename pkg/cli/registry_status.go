package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
)

const statusKitManagedStateDisabled = "disabled"

type registryStatusReport struct {
	State          string                 `json:"state"`
	Managed        bool                   `json:"managed"`
	SourceRepo     string                 `json:"source_repo,omitempty"`
	SourceBranch   string                 `json:"source_branch,omitempty"`
	PlannedChanges int                    `json:"planned_changes"`
	Registry       statusRegistrySummary  `json:"registry"`
	Items          []statusKitManagedItem `json:"items,omitempty"`
	NextActions    []string               `json:"next_actions,omitempty"`
	CheckError     string                 `json:"check_error,omitempty"`
}

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Inspect Kit registry freshness",
	Args:  cobra.NoArgs,
}

var registryStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show compact Kit registry freshness",
	Args:  cobra.NoArgs,
	RunE:  runRegistryStatus,
}

func init() {
	registryStatusCmd.Flags().Bool("json", false, "output registry status as JSON")
	registryCmd.AddCommand(registryStatusCmd)
	rootCmd.AddCommand(registryCmd)
}

func runRegistryStatus(cmd *cobra.Command, _ []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}
	report, err := buildRegistryStatusReport(projectRoot, cfg)
	if err != nil {
		return err
	}
	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		return writeRegistryStatusJSON(cmd.OutOrStdout(), report)
	}
	return writeRegistryStatusHuman(cmd.OutOrStdout(), report)
}

func buildRegistryStatusReport(projectRoot string, cfg *config.Config) (registryStatusReport, error) {
	report := registryStatusReport{
		Managed:      cfg.IsHealthManaged(),
		SourceRepo:   cfg.Registry.Source.Repo,
		SourceBranch: cfg.Registry.Source.Branch,
	}
	if !report.Managed {
		report.State = statusKitManagedStateDisabled
		return report, nil
	}

	summary, err := buildStatusKitManagedSummary(projectRoot, cfg)
	if err != nil {
		return registryStatusReport{}, err
	}
	report.State = summary.State
	report.PlannedChanges = summary.ManagedFiles.Planned
	report.Registry = summary.Registry
	report.Items = summary.Items
	report.NextActions = summary.NextActions
	report.CheckError = summary.ManagedFiles.CheckError
	if report.SourceRepo == "" {
		report.SourceRepo = summary.Registry.SourceRepo
	}
	if report.SourceBranch == "" {
		report.SourceBranch = summary.Registry.SourceBranch
	}
	return report, nil
}

func writeRegistryStatusJSON(out io.Writer, report registryStatusReport) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func writeRegistryStatusHuman(out io.Writer, report registryStatusReport) error {
	switch report.State {
	case statusKitManagedStateDisabled:
		_, err := fmt.Fprintln(out, "disabled (health.managed=false)")
		return err
	case statusKitManagedStateRefreshAvailable:
		_, err := fmt.Fprintf(out, "refresh_available (%d planned change(s))\n", report.PlannedChanges)
		return err
	case statusKitManagedStateAttentionNeeded:
		parts := []string{}
		if report.Registry.Conflicts > 0 {
			parts = append(parts, fmt.Sprintf("%d conflict(s)", report.Registry.Conflicts))
		}
		if report.Registry.LocalCustom > 0 {
			parts = append(parts, fmt.Sprintf("%d local custom", report.Registry.LocalCustom))
		}
		if len(parts) == 0 {
			parts = append(parts, "review required")
		}
		_, err := fmt.Fprintf(out, "attention_needed (%s)\n", strings.Join(parts, ", "))
		return err
	case statusKitManagedStateUnknown:
		if report.CheckError != "" {
			_, err := fmt.Fprintf(out, "unknown (%s)\n", report.CheckError)
			return err
		}
	}
	_, err := fmt.Fprintln(out, report.State)
	return err
}
