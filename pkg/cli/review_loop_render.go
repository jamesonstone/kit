package cli

import (
	"fmt"
	"io"
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

	workingDirectory, err := resolvePromptWorktreeRoot(ctx.LocalRoot)
	if err != nil {
		return err
	}

	target := opts.PRRef
	if strings.TrimSpace(ctx.URL) != "" {
		target = ctx.URL
	}
	if ctx.Repair != nil {
		workingDirectory = ctx.Repair.WorktreePath
		target = ctx.Repair.PRURL
	}
	prompt := buildDispatchPrompt(tasks, opts.MaxSubagents, workingDirectory, dispatchInputSourcePR, dispatchPromptOptions{
		CodeRabbitOnly:          opts.CodeRabbitOnly,
		CommonReviewInstruction: commonInstruction,
		PRTarget:                target,
		RepairContext:           ctx.Repair,
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

	_, _ = fmt.Fprintf(out, "Review loop summary for PR #%d", ctx.Target.Number)
	if strings.TrimSpace(ctx.URL) != "" {
		_, _ = fmt.Fprintf(out, " (%s)", ctx.URL)
	}
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintf(out, "Head: %s\n", ctx.HeadRefOID)
	_, _ = fmt.Fprintf(out, "FIX: %d | VALID_OUT_OF_SCOPE: %d | FALSE_POSITIVE: %d | STALE: %d | NEEDS_HUMAN: %d\n",
		counts[reviewLoopFix],
		counts[reviewLoopValidOutOfScope],
		counts[reviewLoopFalsePositive],
		counts[reviewLoopStale],
		counts[reviewLoopNeedsHuman],
	)

	for _, finding := range classified {
		task := finding.Finding.Task
		_, _ = fmt.Fprintf(out, "- [%s] %s\n", finding.Kind, reviewLoopSourceLabel(task))
		if strings.TrimSpace(finding.Reason) != "" {
			_, _ = fmt.Fprintf(out, "  Reason: %s\n", finding.Reason)
		}
		if strings.TrimSpace(task.Author) != "" {
			_, _ = fmt.Fprintf(out, "  Author: %s\n", task.Author)
		}
		if strings.TrimSpace(task.URL) != "" {
			_, _ = fmt.Fprintf(out, "  URL: %s\n", task.URL)
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
