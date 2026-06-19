package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func createLoopArtifactDir(projectRoot, runID string) (string, error) {
	abs := filepath.Join(projectRoot, filepath.FromSlash(loopRelArtifactDir(runID)))
	if err := os.MkdirAll(abs, 0755); err != nil {
		return "", err
	}
	return abs, nil
}

func loopRelArtifactDir(runID string) string {
	return filepath.ToSlash(filepath.Join(".kit", "loops", runID))
}

func writeLoopIterationFile(artifactDir, runID string, index int, name, content string) (string, error) {
	dir := filepath.Join(artifactDir, fmt.Sprintf("%03d", index))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return filepath.ToSlash(filepath.Join(loopRelArtifactDir(runID), fmt.Sprintf("%03d", index), name)), nil
}

func writeLoopRunArtifact(projectRoot string, report loopReport) error {
	if report.RunID == "" {
		return nil
	}
	dir, err := createLoopArtifactDir(projectRoot, report.RunID)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "run.json"), append(data, '\n'), 0644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "summary.md"), []byte(loopSummaryMarkdown(report)), 0644)
}

func loopSummaryMarkdown(report loopReport) string {
	var builder strings.Builder
	builder.WriteString("# Kit Loop Run\n\n")
	builder.WriteString(fmt.Sprintf("- Run: `%s`\n", report.RunID))
	builder.WriteString(fmt.Sprintf("- Feature: `%s`\n", report.Feature))
	builder.WriteString(fmt.Sprintf("- Status: `%s`\n", report.Status))
	if report.StopReason != "" {
		builder.WriteString(fmt.Sprintf("- Stop reason: %s\n", report.StopReason))
	}
	builder.WriteString("\n## Iterations\n\n")
	if len(report.Iterations) == 0 {
		builder.WriteString("none\n")
		return builder.String()
	}
	for _, iteration := range report.Iterations {
		builder.WriteString(fmt.Sprintf("- %03d `%s`", iteration.Index, iteration.Stage))
		if iteration.Result != nil {
			builder.WriteString(fmt.Sprintf(" status=%s confidence=%d", iteration.Result.Status, iteration.Result.Confidence))
		}
		if iteration.Error != "" {
			builder.WriteString(fmt.Sprintf(" error=%q", iteration.Error))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

func outputLoopReport(cmd *cobra.Command, report loopReport, jsonOutput bool) error {
	if jsonOutput {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	out := cmd.OutOrStdout()
	if report.Status == "dry_run" {
		if len(report.Iterations) > 0 {
			_, err := fmt.Fprintf(out, "Dry run: %s\n", report.Iterations[0].Description)
			return err
		}
		_, err := fmt.Fprintln(out, "Dry run: no action")
		return err
	}
	fmt.Fprintf(out, "Loop run: %s\n", report.RunID)
	fmt.Fprintf(out, "Feature: %s\n", report.Feature)
	fmt.Fprintf(out, "Status: %s\n", report.Status)
	if report.StopReason != "" {
		fmt.Fprintf(out, "Stop reason: %s\n", report.StopReason)
	}
	if report.ArtifactDir != "" {
		fmt.Fprintf(out, "Artifacts: %s\n", report.ArtifactDir)
	}
	if len(report.Iterations) > 0 {
		fmt.Fprintln(out, "Iterations:")
		for _, iteration := range report.Iterations {
			if iteration.Result != nil {
				fmt.Fprintf(out, "  - %03d %s: %s confidence=%d\n", iteration.Index, iteration.Stage, iteration.Result.Status, iteration.Result.Confidence)
			} else {
				fmt.Fprintf(out, "  - %03d %s\n", iteration.Index, iteration.Stage)
			}
		}
	}
	return nil
}

func loopDryRunDescription(opts loopOptions, state loopStageState) string {
	if loopTargetComplete(state.Stage, opts.Until) || state.Stage == loopStageComplete {
		return fmt.Sprintf("target stage %s is already complete for %s", opts.Until, opts.Feature.DirName)
	}
	command := opts.Agent.Command
	if command == "" {
		command = "<configured-agent>"
	}
	args := strings.Join(opts.Agent.Args, " ")
	if args != "" {
		command += " " + args
	}
	return fmt.Sprintf("would run %s stage for %s with `%s`", state.Stage, opts.Feature.DirName, command)
}
