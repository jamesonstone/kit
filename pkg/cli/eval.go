package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	kiteval "github.com/jamesonstone/kit/internal/eval"
)

var evalJSON bool

var evalCmd = &cobra.Command{
	Use:   "eval",
	Short: "Run local Kit harness evals",
	Long:  "Run small local regression checks for Kit's harness behavior.",
	Args:  cobra.NoArgs,
	RunE:  runEval,
}

func init() {
	evalCmd.Flags().BoolVar(&evalJSON, "json", false, "output eval report as JSON")
	rootCmd.AddCommand(evalCmd)
}

func runEval(cmd *cobra.Command, args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}
	report := kiteval.Run(projectRoot, cfg)
	if evalJSON {
		if err := outputJSON(cmd.OutOrStdout(), report); err != nil {
			return err
		}
		if report.Failed() {
			return fmt.Errorf("kit eval failed")
		}
		return nil
	}
	for _, result := range report.Cases {
		fmt.Fprintf(cmd.OutOrStdout(), "%s: %s", result.Name, result.Status)
		if result.Message != "" {
			fmt.Fprintf(cmd.OutOrStdout(), " - %s", result.Message)
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}
	if report.Failed() {
		return fmt.Errorf("kit eval failed")
	}
	return nil
}
