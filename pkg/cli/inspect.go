package cli

import (
	"encoding/json"
	"fmt"

	"github.com/jamesonstone/kit/v3/internal/guidance"
	"github.com/spf13/cobra"
)

func newInspectCommand() *cobra.Command {
	var path string
	var jsonOutput bool
	cmd := &cobra.Command{
		Use: "inspect", Short: "Inspect current local project guidance without changing files",
		Long:         "List local rulesets, SOPs, and key guidance in the nearest Kit project. Use --path to select another project or nested directory. Symlink guidance is reported as invalid. No instructions are executed and no files are repaired.",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			report, inspectErr := guidance.Inspect(path)
			if jsonOutput {
				if inspectErr != nil && len(report.Errors) == 0 {
					report.Errors = append(report.Errors, inspectErr.Error())
				}
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(report); err != nil {
					return err
				}
			} else if report.ProjectRoot != "" {
				out := cmd.OutOrStdout()
				if _, err := fmt.Fprintf(out, "Project: %s\n", report.ProjectRoot); err != nil {
					return err
				}
				for _, finding := range report.Config.Findings {
					if _, err := fmt.Fprintf(out, "Config %s: %s\n", finding.Severity, finding.Message); err != nil {
						return err
					}
				}
				for _, entry := range report.Entries {
					if _, err := fmt.Fprintf(out, "%s [%s] %s\n", entry.Kind, entry.State, entry.Path); err != nil {
						return err
					}
					if entry.Metadata != nil {
						if _, err := fmt.Fprintf(out, "  %s: %s\n", entry.Metadata.Slug, entry.Metadata.Description); err != nil {
							return err
						}
						if entry.Metadata.Owner != "" {
							if _, err := fmt.Fprintf(out, "  Owner: %s\n", entry.Metadata.Owner); err != nil {
								return err
							}
						}
					}
				}
				for _, problem := range report.Errors {
					if _, err := fmt.Fprintf(out, "Error: %s\n", problem); err != nil {
						return err
					}
				}
			}
			return inspectErr
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "project directory or a directory inside it (default: current directory)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "emit a deterministic JSON report")
	return cmd
}

func init() { rootCmd.AddCommand(newInspectCommand()) }
