package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	kitstate "github.com/jamesonstone/kit/internal/state"
)

var stateJSON bool

var stateCmd = &cobra.Command{
	Use:   "state [refresh]",
	Short: "Show or refresh generated Kit state",
	Long:  "Generate pointer-only .kit/state.json for agents and tools. Markdown remains authoritative.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runState,
}

func init() {
	stateCmd.Flags().BoolVar(&stateJSON, "json", false, "output generated state as JSON")
	rootCmd.AddCommand(stateCmd)
}

func runState(cmd *cobra.Command, args []string) error {
	if len(args) == 1 && args[0] != "refresh" {
		return fmt.Errorf("unknown state action %q", args[0])
	}
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}
	generated, err := kitstate.Generate(projectRoot, cfg)
	if err != nil {
		return err
	}
	if len(args) == 1 && args[0] == "refresh" {
		if err := kitstate.Write(projectRoot, generated); err != nil {
			return err
		}
	}
	if stateJSON {
		return outputJSON(cmd.OutOrStdout(), generated)
	}
	if len(args) == 1 && args[0] == "refresh" {
		fmt.Fprintf(cmd.OutOrStdout(), "Updated %s\n", kitstate.StatePath)
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Generated state for %d feature(s). Run `kit state refresh` to write %s.\n", len(generated.Features), kitstate.StatePath)
	return nil
}
