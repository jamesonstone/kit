package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func runReviewLoopPrompt(
	cmdOut io.Writer,
	opts reviewLoopOptions,
	ctx reviewLoopPRContext,
	classified []reviewLoopClassifiedFinding,
	commonInstruction string,
) error {
	renderReviewLoopSummary(cmdOut, ctx, classified)

	fixTasks := reviewLoopFixTasks(classified)
	if len(fixTasks) == 0 {
		_, err := fmt.Fprintln(cmdOut, "No actionable current review feedback found.")
		return err
	}

	inputCfg := opts.InputConfig
	if !inputCfg.usesEditor() {
		inputCfg = newFreeTextInputConfig(opts.UseVim, opts.Editor, false, true)
	}

	initialContent := renderDispatchPRInputForEditor(dispatchPRInput{
		CommonReviewInstruction: commonInstruction,
		RawTasks:                renderDispatchReviewTasks(fixTasks),
	})
	edited, err := readEditorTextWithInitialContent(
		inputCfg,
		"review-loop dispatch tasks",
		initialContent,
		false,
		false,
	)
	if err != nil {
		return err
	}

	rawTasks, commonInstruction := splitDispatchPRInputFromEditor(edited, commonInstruction)
	if strings.TrimSpace(rawTasks) == "" {
		return fmt.Errorf("review-loop dispatch tasks cannot be empty")
	}

	tasks, err := normalizeDispatchTasks(rawTasks)
	if err != nil {
		return err
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	prompt := buildDispatchPrompt(tasks, opts.MaxSubagents, workingDirectory, dispatchInputSourcePR, dispatchPromptOptions{
		CommonReviewInstruction: commonInstruction,
	})
	if err := outputPromptWithoutSubagentsWithClipboardDefault(prompt, opts.OutputOnly, opts.Copy); err != nil {
		return err
	}

	if !opts.OutputOnly {
		printWorkflowInstructions("review-loop (supporting step)", []string{
			"review the generated dispatch prompt before launching any agent work",
			"do not mutate git or GitHub until repo-local delivery rules are loaded",
		})
	}

	return nil
}

func renderReviewLoopSummary(
	out io.Writer,
	ctx reviewLoopPRContext,
	classified []reviewLoopClassifiedFinding,
) {
	counts := map[reviewLoopClassification]int{}
	for _, finding := range classified {
		counts[finding.Kind]++
	}

	fmt.Fprintf(out, "Review loop summary for PR #%d", ctx.Target.Number)
	if strings.TrimSpace(ctx.URL) != "" {
		fmt.Fprintf(out, " (%s)", ctx.URL)
	}
	fmt.Fprintln(out)
	fmt.Fprintf(out, "Head: %s\n", ctx.HeadRefOID)
	fmt.Fprintf(out, "FIX: %d | VALID_OUT_OF_SCOPE: %d | FALSE_POSITIVE: %d | STALE: %d | NEEDS_HUMAN: %d\n",
		counts[reviewLoopFix],
		counts[reviewLoopValidOutOfScope],
		counts[reviewLoopFalsePositive],
		counts[reviewLoopStale],
		counts[reviewLoopNeedsHuman],
	)

	for _, finding := range classified {
		task := finding.Finding.Task
		fmt.Fprintf(out, "- [%s] %s\n", finding.Kind, reviewLoopSourceLabel(task))
		if strings.TrimSpace(finding.Reason) != "" {
			fmt.Fprintf(out, "  Reason: %s\n", finding.Reason)
		}
		if strings.TrimSpace(task.Author) != "" {
			fmt.Fprintf(out, "  Author: %s\n", task.Author)
		}
		if strings.TrimSpace(task.URL) != "" {
			fmt.Fprintf(out, "  URL: %s\n", task.URL)
		}
	}
}

func reviewLoopFixTasks(classified []reviewLoopClassifiedFinding) []dispatchReviewTask {
	tasks := make([]dispatchReviewTask, 0, len(classified))
	for _, finding := range classified {
		if finding.Kind == reviewLoopFix {
			tasks = append(tasks, finding.Finding.Task)
		}
	}
	return tasks
}

func reviewLoopSourceLabel(task dispatchReviewTask) string {
	path := strings.TrimSpace(task.Path)
	if path == "" {
		path = "(no path)"
	}
	if task.Line > 0 {
		return fmt.Sprintf("%s:%d", path, task.Line)
	}
	return path
}
