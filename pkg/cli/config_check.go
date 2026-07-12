package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/jamesonstone/kit/internal/config"
)

var configCheckJSON bool

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Inspect and repair Kit project configuration",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var configCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate .kit.yaml and offer safe inline repairs",
	Args:  cobra.NoArgs,
	RunE:  runConfigCheck,
}

type configCheckReport struct {
	SchemaVersion        int                `json:"schema_version"`
	CurrentSchemaVersion int                `json:"current_schema_version"`
	SchemaState          config.SchemaState `json:"schema_state"`
	Valid                bool               `json:"valid"`
	AWS                  configCheckAWS     `json:"aws"`
	Findings             []config.Finding   `json:"findings"`
}

type configCheckAWS struct {
	Configured bool   `json:"configured"`
	Enabled    bool   `json:"enabled"`
	Profile    string `json:"profile,omitempty"`
	AccountID  string `json:"account_id,omitempty"`
}

type configRemediationOptions struct {
	Interactive bool
	Input       io.Reader
	Output      io.Writer
}

func init() {
	configCheckCmd.Flags().BoolVar(&configCheckJSON, "json", false, "emit a machine-readable validation report without prompting or writing")
	configCmd.AddCommand(configCheckCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigCheck(cmd *cobra.Command, args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	cfg, inspection, err := config.LoadWithInspection(projectRoot)
	if err != nil {
		return err
	}

	if !configCheckJSON && commandHasInteractiveTerminal(cmd) && inspection.SchemaState != config.SchemaStateNewer {
		changed, err := remediateProjectConfig(projectRoot, cfg, inspection, configRemediationOptions{
			Interactive: true,
			Input:       cmd.InOrStdin(),
			Output:      cmd.OutOrStdout(),
		})
		if err != nil {
			return err
		}
		if changed {
			cfg, inspection, err = config.LoadWithInspection(projectRoot)
			if err != nil {
				return err
			}
		}
	}

	report := buildConfigCheckReport(cfg, inspection)
	if configCheckJSON {
		if err := outputJSON(cmd.OutOrStdout(), report); err != nil {
			return err
		}
	} else {
		printConfigCheckReport(cmd.OutOrStdout(), report)
	}
	if !report.Valid {
		return newCLIExitError(errors.New("configuration validation failed"), 1, true)
	}
	return nil
}

func buildConfigCheckReport(cfg *config.Config, inspection config.Inspection) configCheckReport {
	aws := configCheckAWS{}
	findings := inspection.Findings
	if findings == nil {
		findings = []config.Finding{}
	}
	if cfg != nil && cfg.AWS != nil {
		aws.Configured = true
		aws.Enabled = cfg.AWS.IsEnabled()
		aws.Profile = cfg.AWS.Profile
		aws.AccountID = cfg.AWS.AccountID
	}
	return configCheckReport{
		SchemaVersion:        inspection.SchemaVersion,
		CurrentSchemaVersion: inspection.CurrentSchemaVersion,
		SchemaState:          inspection.SchemaState,
		Valid:                inspection.SchemaState == config.SchemaStateCurrent && !inspection.HasErrors(),
		AWS:                  aws,
		Findings:             findings,
	}
}

func printConfigCheckReport(out io.Writer, report configCheckReport) {
	fmt.Fprintln(out, "🔎 Checking .kit.yaml...")
	fmt.Fprintf(out, "  Schema: %s (configured=%d current=%d)\n", report.SchemaState, report.SchemaVersion, report.CurrentSchemaVersion)
	switch {
	case !report.AWS.Configured:
		fmt.Fprintln(out, "  AWS: not configured")
	case !report.AWS.Enabled:
		fmt.Fprintln(out, "  AWS: disabled")
	default:
		fmt.Fprintf(out, "  AWS: profile=%s account=%s\n", report.AWS.Profile, report.AWS.AccountID)
	}
	for _, finding := range report.Findings {
		fmt.Fprintf(out, "  - [%s] %s: %s\n", finding.Severity, finding.Field, finding.Message)
	}
	if report.Valid {
		fmt.Fprintln(out, "  ✅ Configuration is current and valid.")
	}
}

func commandHasInteractiveTerminal(cmd *cobra.Command) bool {
	return streamsHaveInteractiveTerminal(cmd.InOrStdin(), cmd.OutOrStdout())
}

func streamsHaveInteractiveTerminal(in io.Reader, out io.Writer) bool {
	inFile, ok := in.(*os.File)
	return ok && term.IsTerminal(int(inFile.Fd())) && terminalWriterCheck(out)
}
