package cli

import (
	"fmt"
	"strings"
)

func buildDispatchPrompt(
	tasks []dispatchTask,
	maxSubagents int,
	workingDirectory string,
	inputSource dispatchInputSource,
) string {
	var sb strings.Builder

	sb.WriteString("/plan\n\n")
	sb.WriteString("Prepare a subagent dispatch plan for the following task set.\n\n")
	sb.WriteString("## Dispatch Context\n")
	sb.WriteString(fmt.Sprintf("- Working directory: %s\n", workingDirectory))
	sb.WriteString(fmt.Sprintf("- Input source: %s\n", inputSource))
	sb.WriteString(fmt.Sprintf("- Effective max subagents: %d\n\n", maxSubagents))
	sb.WriteString("## Normalized Tasks\n")
	appendDispatchTasks(&sb, tasks)
	sb.WriteString("## Your Task\n")
	sb.WriteString("1. Stay in planning and discovery mode first\n")
	sb.WriteString("2. Do NOT launch any subagents yet\n")
	sb.WriteString("3. Inspect the repository and anticipate which files are likely to change for each normalized task before assigning work\n")
	sb.WriteString("4. Build a predicted touched-file set for each task\n")
	sb.WriteString("5. Cluster tasks by predicted file overlap so tasks that touch the same files or overlapping files stay together\n")
	sb.WriteString("6. If file overlap is ambiguous or confidence is low, merge those tasks into the same cluster instead of parallelizing them\n")
	sb.WriteString("7. Assign one subagent per cluster and preserve original task order within each cluster queue\n")
	sb.WriteString(fmt.Sprintf("8. Parallelize only disjoint clusters and never exceed %d concurrent subagents\n", maxSubagents))
	sb.WriteString("9. If the number of clusters exceeds the concurrency cap, queue the remaining clusters and state their execution order explicitly\n")
	sb.WriteString("10. Output a dry-run discovery report with the exact sections listed below before any subagent execution begins\n")
	sb.WriteString("11. Wait for explicit user approval after the dry-run report and proposed queue before launching any subagents\n")
	sb.WriteString(fmt.Sprintf("12. After approval, launch at most %d concurrent subagents and keep queued clusters serialized behind them\n\n", maxSubagents))
	sb.WriteString("## Required Dry-Run Report Sections\n")
	sb.WriteString("- normalized tasks\n")
	sb.WriteString("- predicted touched files per task\n")
	sb.WriteString("- overlap clusters\n")
	sb.WriteString("- dispatch queue\n")
	sb.WriteString("- subagent assignments\n")
	sb.WriteString("- risks and unknowns\n\n")
	sb.WriteString("## Rules\n")
	sb.WriteString("- discovery first, execution second\n")
	sb.WriteString("- do not invent parallelism where file overlap is unclear\n")
	sb.WriteString("- tasks with overlapping predicted file changes belong to the same subagent queue\n")
	sb.WriteString("- preserve the original task order inside each cluster\n")
	sb.WriteString("- keep the dry-run report reviewable and explicit before asking for approval\n")

	return sb.String()
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
