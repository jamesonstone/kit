package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/runstore"
)

var traceJSON bool

var traceCmd = &cobra.Command{
	Use:   "trace <feature-or-run-id>",
	Short: "Inspect verification run traces",
	Long:  "List feature verification runs or show a compact verification run detail.",
	Args:  cobra.ExactArgs(1),
	RunE:  runTrace,
}

func init() {
	traceCmd.Flags().BoolVar(&traceJSON, "json", false, "output trace data as JSON")
	rootCmd.AddCommand(traceCmd)
}

func runTrace(cmd *cobra.Command, args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}
	if run, err := runstore.Load(projectRoot, args[0]); err == nil {
		if traceJSON {
			return outputJSON(cmd.OutOrStdout(), run)
		}
		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "Run: %s\nFeature: %s\nStatus: %s\nArtifacts: %s\n", run.RunID, run.Feature.DirName, run.Status, run.ArtifactDir)
		for _, result := range run.Results {
			fmt.Fprintf(out, "- [%s] %s: %s\n", result.TaskID, result.Raw, result.Status)
		}
		return nil
	}

	feat, err := loadFeatureWithState(cfg.SpecsPath(projectRoot), cfg, args[0])
	if err != nil {
		return err
	}
	entries, err := runstore.List(projectRoot)
	if err != nil {
		return err
	}
	var selected []runstore.IndexEntry
	for _, entry := range entries {
		if entry.FeatureDir == feat.DirName {
			selected = append(selected, entry)
		}
	}
	if traceJSON {
		return outputJSON(cmd.OutOrStdout(), selected)
	}
	out := cmd.OutOrStdout()
	if len(selected) == 0 {
		fmt.Fprintf(out, "No verification runs found for %s\n", feat.DirName)
		return nil
	}
	fmt.Fprintf(out, "Verification runs for %s:\n", feat.DirName)
	for _, entry := range selected {
		fmt.Fprintf(out, "- %s %s tasks=%v artifacts=%s\n", entry.RunID, entry.Status, entry.TaskIDs, entry.ArtifactDir)
	}
	return nil
}
