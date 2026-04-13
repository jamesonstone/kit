package cli

import (
	"fmt"
	"strings"

	"github.com/jamesonstone/kit/internal/promptdoc"
)

func buildDispatchPrompt(
	tasks []dispatchTask,
	maxSubagents int,
	workingDirectory string,
	inputSource dispatchInputSource,
) string {
	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Raw("/plan")
		doc.Paragraph("Prepare a subagent dispatch plan for the following task set.")
		doc.Heading(2, "Dispatch Context")
		doc.BulletList(
			fmt.Sprintf("Working directory: %s", workingDirectory),
			fmt.Sprintf("Input source: %s", inputSource),
			fmt.Sprintf("Effective max subagents: %d", maxSubagents),
		)
		doc.Heading(2, "Normalized Tasks")
		doc.Raw(renderDispatchTasks(tasks))
		doc.Heading(2, "Your Task")
		doc.OrderedList(1,
			"Stay in planning and discovery mode first",
			"Do NOT launch any subagents yet",
			"Inspect the repository and anticipate which files are likely to change for each normalized task before assigning work",
			"Build a predicted touched-file set for each task",
			"Cluster tasks by predicted file overlap so tasks that touch the same files or overlapping files stay together",
			"If file overlap is ambiguous or confidence is low, merge those tasks into the same cluster instead of parallelizing them",
			"Assign one subagent per cluster and preserve original task order within each cluster queue",
			fmt.Sprintf("Parallelize only disjoint clusters and never exceed %d concurrent subagents", maxSubagents),
			"If the number of clusters exceeds the concurrency cap, queue the remaining clusters and state their execution order explicitly",
			"Output a dry-run discovery report with the exact sections listed below before any subagent execution begins",
			"Wait for explicit user approval after the dry-run report and proposed queue before launching any subagents",
			fmt.Sprintf("After approval, launch at most %d concurrent subagents and keep queued clusters serialized behind them", maxSubagents),
			"When a subagent needs an isolated checkout, use `git worktree add ~/worktrees/<repo>-<branch> <branch>` or `git worktree add -b <branch> ~/worktrees/<repo>-<branch> <base-ref>` and keep all worktrees flat under `~/worktrees/`",
		)
		doc.Heading(2, "Required Dry-Run Report Sections")
		doc.BulletList(
			"normalized tasks",
			"predicted touched files per task",
			"overlap clusters",
			"dispatch queue",
			"subagent assignments",
			"risks and unknowns",
		)
		doc.Heading(2, "Rules")
		doc.BulletList(
			"discovery first, execution second",
			"do not invent parallelism where file overlap is unclear",
			"tasks with overlapping predicted file changes belong to the same subagent queue",
			"preserve the original task order inside each cluster",
			"do not create worktrees inside the repository tree; keep them flat under `~/worktrees/`",
			"keep the dry-run report reviewable and explicit before asking for approval",
		)
	})
}

func renderDispatchTasks(tasks []dispatchTask) string {
	return renderBuilderText(func(sb *strings.Builder) {
		appendDispatchTasks(sb, tasks)
	})
}

func appendDispatchTasks(sb *strings.Builder, tasks []dispatchTask) {
	for _, task := range tasks {
		sb.WriteString(fmt.Sprintf("### %s\n", task.ID))
		sb.WriteString("```text\n")
		sb.WriteString(task.Body)
		if !strings.HasSuffix(task.Body, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("```\n")
	}
	sb.WriteString("\n")
}
