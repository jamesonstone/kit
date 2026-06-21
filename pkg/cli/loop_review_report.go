package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func writeLoopReviewRunArtifact(projectRoot string, report loopReviewReport) error {
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
	if err := os.WriteFile(filepath.Join(dir, "run.json"), append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "summary.md"), []byte(loopReviewSummaryMarkdown(report)), 0o644)
}

func loopReviewSummaryMarkdown(report loopReviewReport) string {
	var builder strings.Builder
	builder.WriteString("# Kit Loop Review Run\n\n")
	builder.WriteString(fmt.Sprintf("- Run: `%s`\n", report.RunID))
	builder.WriteString(fmt.Sprintf("- Status: `%s`\n", report.Status))
	if report.Correctness > 0 {
		builder.WriteString(fmt.Sprintf("- Correctness: `%d%%`\n", report.Correctness))
	}
	if report.BaseRef != "" {
		builder.WriteString(fmt.Sprintf("- Base ref: `%s`\n", report.BaseRef))
	}
	if report.PRRef != "" {
		builder.WriteString(fmt.Sprintf("- PR: `%s`\n", report.PRRef))
	}
	if report.PRStatus != "" {
		builder.WriteString(fmt.Sprintf("- PR status: %s\n", report.PRStatus))
	}
	if report.StopReason != "" {
		builder.WriteString(fmt.Sprintf("- Stop reason: %s\n", report.StopReason))
	}
	builder.WriteString("\n## Iterations\n\n")
	for _, iteration := range report.Iterations {
		builder.WriteString(fmt.Sprintf("- %03d", iteration.Index))
		if iteration.Result != nil {
			builder.WriteString(fmt.Sprintf(" done=%t correctness=%d", iteration.Result.Done, iteration.Result.Correctness))
		}
		if iteration.Error != "" {
			builder.WriteString(fmt.Sprintf(" error=%q", iteration.Error))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

func outputLoopReviewReport(cmd *cobra.Command, report loopReviewReport, jsonOutput bool) error {
	if jsonOutput {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	out := cmd.OutOrStdout()
	if report.Status == "dry_run" {
		_, err := fmt.Fprintf(out, "Dry run: %s\n", report.StopReason)
		return err
	}

	result := lastLoopReviewResult(report)
	if result != nil && result.Done {
		fmt.Fprintf(out, "Correctness: %d%%\n", result.Correctness)
		if report.PRStatus != "" {
			fmt.Fprintf(out, "Status: %s\n", report.PRStatus)
		} else {
			fmt.Fprintln(out, "Status: done")
		}
		fmt.Fprintln(out)
		bullets := result.Bullets
		if len(bullets) == 0 {
			bullets = []string{"- No high, medium, or correctness-impacting issues found."}
		}
		for _, bullet := range bullets {
			fmt.Fprintln(out, bullet)
		}
		if strings.Contains(report.PRStatus, "pending") && report.StopReason != "" {
			fmt.Fprintln(out)
			fmt.Fprintln(out, report.StopReason)
		}
		fmt.Fprintln(out, "done")
		return nil
	}

	fmt.Fprintf(out, "Loop review run: %s\n", report.RunID)
	fmt.Fprintf(out, "Status: %s\n", report.Status)
	if report.StopReason != "" {
		fmt.Fprintf(out, "Stop reason: %s\n", report.StopReason)
	}
	if report.ArtifactDir != "" {
		fmt.Fprintf(out, "Artifacts: %s\n", report.ArtifactDir)
	}
	return nil
}

func lastLoopReviewResult(report loopReviewReport) *loopReviewAgentResult {
	for i := len(report.Iterations) - 1; i >= 0; i-- {
		if report.Iterations[i].Result != nil {
			return report.Iterations[i].Result
		}
	}
	return nil
}

func shouldPromptLoopReviewRerun(cmd *cobra.Command, opts loopReviewOptions) bool {
	if opts.JSON || opts.DryRun {
		return false
	}
	inFile, ok := cmd.InOrStdin().(*os.File)
	if !ok {
		return false
	}
	return terminalWriterCheck(cmd.OutOrStdout()) && term.IsTerminal(int(inFile.Fd()))
}

func promptLoopReviewRerun(in io.Reader, out io.Writer, message string, report loopReviewReport) (bool, error) {
	if _, err := fmt.Fprintln(out, message); err != nil {
		return false, err
	}
	if report.RunID != "" {
		if _, err := fmt.Fprintf(out, "Run: %s\n", report.RunID); err != nil {
			return false, err
		}
	}
	if report.Status != "" {
		if _, err := fmt.Fprintf(out, "Status: %s\n", report.Status); err != nil {
			return false, err
		}
	}
	if report.StopReason != "" {
		if _, err := fmt.Fprintf(out, "Stop reason: %s\n", report.StopReason); err != nil {
			return false, err
		}
	}
	if _, err := fmt.Fprint(out, "Run review loop again? [y/N]: "); err != nil {
		return false, err
	}
	return readLoopReviewConfirmation(in)
}

func readLoopReviewConfirmation(in io.Reader) (bool, error) {
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("failed to read review loop confirmation: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	case "", "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid review loop confirmation %q; enter y or n", strings.TrimSpace(line))
	}
}

func latestLoopReviewReport(projectRoot string) (loopReviewReport, bool, error) {
	loopDir := filepath.Join(projectRoot, ".kit", "loops")
	entries, err := os.ReadDir(loopDir)
	if errors.Is(err, os.ErrNotExist) {
		return loopReviewReport{}, false, nil
	}
	if err != nil {
		return loopReviewReport{}, false, fmt.Errorf("failed to inspect loop review artifacts: %w", err)
	}

	var latest loopReviewReport
	found := false
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		reportPath := filepath.Join(loopDir, entry.Name(), "run.json")
		data, err := os.ReadFile(reportPath)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return loopReviewReport{}, false, fmt.Errorf("failed to read %s: %w", reportPath, err)
		}
		var report loopReviewReport
		if err := json.Unmarshal(data, &report); err != nil {
			continue
		}
		if !isLoopReviewReport(report) {
			continue
		}
		if !found || report.StartedAt.After(latest.StartedAt) || (report.StartedAt.Equal(latest.StartedAt) && report.RunID > latest.RunID) {
			latest = report
			found = true
		}
	}
	return latest, found, nil
}

func isLoopReviewReport(report loopReviewReport) bool {
	return report.BaseRef != "" || report.PRRef != "" || report.PRStatus != ""
}

func isLoopReviewMaxIterations(report loopReviewReport) bool {
	return report.Status == "stopped" && strings.HasPrefix(report.StopReason, "max iterations reached:")
}

func stopLoopReview(report loopReviewReport, err error) (loopReviewReport, error) {
	report.Status = "stopped"
	report.StopReason = err.Error()
	report.EndedAt = time.Now().UTC()
	return report, err
}

func stopLoopReviewWithIteration(
	projectRoot string,
	report loopReviewReport,
	iteration loopReviewIteration,
	err error,
) (loopReviewReport, error) {
	iteration.Error = err.Error()
	iteration.EndedAt = time.Now().UTC()
	iteration.DurationMS = iteration.EndedAt.Sub(iteration.StartedAt).Milliseconds()
	report.Iterations = append(report.Iterations, iteration)
	return stopLoopReviewAfterWrite(projectRoot, report, err)
}

func stopLoopReviewAfterWrite(projectRoot string, report loopReviewReport, err error) (loopReviewReport, error) {
	report.Status = "stopped"
	report.StopReason = err.Error()
	report.EndedAt = time.Now().UTC()
	_ = writeLoopReviewRunArtifact(projectRoot, report)
	return report, err
}

func stopLoopReviewAgentFailure(
	projectRoot string,
	report loopReviewReport,
	iteration int,
	execResult loopAgentExecution,
) (loopReviewReport, error) {
	reason := fmt.Sprintf("agent command failed at iteration %d: %v", iteration, execResult.Err)
	if stderrLine := agentFailureDetailLine(execResult.Stderr); stderrLine != "" {
		reason = fmt.Sprintf("%s: %s", reason, stderrLine)
	}
	err := errors.New(reason)
	report.Status = "stopped"
	report.StopReason = reason
	report.EndedAt = time.Now().UTC()
	_ = writeLoopReviewRunArtifact(projectRoot, report)
	return report, err
}

func agentFailureDetailLine(content string) string {
	first := ""
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if first == "" {
			first = trimmed
		}
		lower := strings.ToLower(trimmed)
		if strings.Contains(lower, "error") || strings.Contains(lower, "failed") {
			return trimmed
		}
	}
	return first
}

func firstPromptLine(prompt string) string {
	for _, line := range strings.Split(prompt, "\n") {
		if strings.TrimSpace(line) != "" {
			return strings.TrimSpace(line)
		}
	}
	return "review prompt"
}
