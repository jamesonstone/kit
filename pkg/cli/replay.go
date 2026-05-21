package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/runstore"
	verifyengine "github.com/jamesonstone/kit/internal/verify"
)

var replayJSON bool
var replayNoWrite bool
var replayTimeout string

var replayCmd = &cobra.Command{
	Use:   "replay <run-id>",
	Short: "Replay a verification run's commands",
	Long:  "Rerun the recorded verification commands from a prior run. This does not reconstruct model reasoning.",
	Args:  cobra.ExactArgs(1),
	RunE:  runReplay,
}

func init() {
	replayCmd.Flags().BoolVar(&replayJSON, "json", false, "output replay result as JSON")
	replayCmd.Flags().BoolVar(&replayNoWrite, "no-write", false, "do not write .kit/runs artifacts")
	replayCmd.Flags().StringVar(&replayTimeout, "timeout", "", "per-command timeout such as 30s or 2m")
	rootCmd.AddCommand(replayCmd)
}

type replayReport struct {
	ParentRunID string             `json:"parent_run_id"`
	Run         verifyengine.Run   `json:"run"`
	Comparison  []replayComparison `json:"comparison"`
}

type replayComparison struct {
	CommandID      string `json:"command_id"`
	PreviousStatus string `json:"previous_status"`
	CurrentStatus  string `json:"current_status"`
}

func runReplay(cmd *cobra.Command, args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	parent, err := runstore.Load(projectRoot, args[0])
	if err != nil {
		return err
	}
	timeout, err := parseOptionalDuration(replayTimeout)
	if err != nil {
		return err
	}
	run := verifyengine.ExecuteRun(context.Background(), verifyengine.RunOptions{
		ProjectRoot:   projectRoot,
		Feature:       parent.Feature,
		TaskIDs:       parent.TaskIDs,
		ExpectedFiles: parent.ExpectedFiles,
		Commands:      parent.Commands,
		Timeout:       timeout,
		ParentRunID:   parent.RunID,
	})
	if !replayNoWrite {
		if err := runstore.Write(projectRoot, &run); err != nil {
			return err
		}
	}
	report := replayReport{
		ParentRunID: parent.RunID,
		Run:         run,
		Comparison:  compareRuns(parent, run),
	}
	if replayJSON {
		return outputJSON(cmd.OutOrStdout(), report)
	}
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Replayed run %s -> %s\n", parent.RunID, run.RunID)
	fmt.Fprintf(out, "Status: %s\n", run.Status)
	if run.ArtifactDir != "" {
		fmt.Fprintf(out, "Artifacts: %s\n", run.ArtifactDir)
	}
	for _, comparison := range report.Comparison {
		fmt.Fprintf(out, "- %s: %s -> %s\n", comparison.CommandID, comparison.PreviousStatus, comparison.CurrentStatus)
	}
	return nil
}

func compareRuns(parent, current verifyengine.Run) []replayComparison {
	previous := make(map[string]string)
	for _, result := range parent.Results {
		previous[result.CommandID] = result.Status
	}
	comparisons := make([]replayComparison, 0, len(current.Results))
	for _, result := range current.Results {
		comparisons = append(comparisons, replayComparison{
			CommandID:      result.CommandID,
			PreviousStatus: previous[result.CommandID],
			CurrentStatus:  result.Status,
		})
	}
	return comparisons
}
