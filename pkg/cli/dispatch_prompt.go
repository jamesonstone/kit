package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jamesonstone/kit/internal/promptdoc"
)

func buildDispatchPrompt(
	tasks []dispatchTask,
	maxSubagents int,
	workingDirectory string,
	inputSource dispatchInputSource,
	options dispatchPromptOptions,
) string {
	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph("Prepare an Agent Team Plan for the following task set.")
		doc.Heading(2, "Dispatch Context")
		doc.BulletList(
			fmt.Sprintf("Working directory: %s", workingDirectory),
			fmt.Sprintf("Input source: %s", inputSource),
			fmt.Sprintf("Effective max subagents: %d", maxSubagents),
			fmt.Sprintf("Default max concurrent lanes: %d", defaultDispatchMaxSubagents),
			fmt.Sprintf("Hard ceiling: %d", hardDispatchMaxSubagents),
		)
		if strings.TrimSpace(options.CommonReviewInstruction) != "" {
			doc.Heading(2, "Common Review Instruction")
			doc.CodeBlock("text", options.CommonReviewInstruction)
		}
		doc.Heading(2, "Normalized Tasks")
		doc.Raw(renderDispatchTasks(tasks))
		doc.Heading(2, "Your Task")
		doc.OrderedList(1,
			"Act as the one accountable supervisor for scope, integration, validation, evidence, delivery gating, and final reporting",
			"Stay in discovery and assignment-design workflow first",
			"When present and relevant, load `docs/references/rules/agent-team-orchestration.md` before finalizing the team shape",
			"Inspect the repository and anticipate which files are likely to change for each normalized task before assigning work",
			"Build a predicted touched-file set for each task",
			"Cluster tasks by predicted file overlap so tasks that touch the same files or overlapping files stay together",
			"If file overlap is ambiguous, confidence is low, or the work requires continuous design judgment, merge those tasks into the same supervisor lane instead of parallelizing them",
			"Use one supervisor lane only for trivial, tightly coupled, high-overlap, high-ambiguity, no-subagent-runtime, or explicitly single-agent work; record the exception when used",
			"Assign one implementation subagent per low-overlap cluster only when the split improves correctness or throughput, and preserve original task order within each cluster queue",
			fmt.Sprintf("Parallelize only disjoint clusters and never exceed %d concurrent subagents", maxSubagents),
			"If the number of clusters exceeds the concurrency cap, queue the remaining clusters and state their execution order explicitly",
			"Include at least one read-only verification lane by default after implementation unless the change is documentation-only, trivial, tightly coupled, the runtime cannot spawn subagents, or the user requested single-agent execution",
			"Output an Agent Team Plan with the exact sections listed below before any subagent execution begins",
			fmt.Sprintf("After recording the Agent Team Plan, self-direct execution by launching at most %d concurrent subagents and keeping queued clusters serialized behind them", maxSubagents),
			"Keep all subagent work in the existing project directory; do not create or use git worktrees",
			"If the current branch or dirty state is unsuitable for a subagent, stop and ask the user how to proceed instead of creating an alternate checkout",
		)
		doc.Heading(2, "Required Agent Team Plan Sections")
		doc.BulletList(
			"supervisor responsibilities",
			"normalized tasks",
			"proposed lanes",
			"subagents that will actually be spawned",
			"logical-only lanes that will not be spawned",
			"intentionally omitted implementation or verification lanes with reasons",
			"predicted touched files per lane",
			"overlap risks",
			"max concurrency",
			"serialized work",
			"validation/review lanes",
			"risks and unknowns",
		)
		doc.Heading(2, "Rules")
		doc.BulletList(
			"one accountable supervisor; do not parallelize accountability",
			"discovery first, execution second",
			"do not invent parallelism where file overlap is unclear",
			"tasks with overlapping predicted file changes belong in the same implementation queue or a serial execution plan",
			"preserve the original task order inside each cluster",
			"do not describe a logical lane as a spawned agent unless a separate agent actually ran",
			"subagents must not independently expand scope",
			"subagents must not create branches, stage, commit, push, open PRs, resolve review threads, or mark the whole workflow complete unless explicitly assigned and allowed by the supervisor",
			"verification agents are read-only by default and must not edit files, stage changes, commit, push, resolve threads, or close their own findings",
			"do not create or use git worktrees for agent work",
			"keep the Agent Team Plan reviewable and explicit before self-directed execution",
			"final reporting must state actual subagents spawned, logical lanes not spawned, and any single-lane exception; if no separate agents ran, state exactly: `single supervisor lane; no specialist or verification agents spawned`",
		)
		if inputSource == dispatchInputSourcePR {
			appendDispatchPRReflectionCycle(doc, options)
		}
	})
}

type dispatchPromptOptions struct {
	CommonReviewInstruction string
	CodeRabbitOnly          bool
	PRTarget                string
}

func appendDispatchPRReflectionCycle(doc *promptdoc.Document, options dispatchPromptOptions) {
	targetArg := dispatchPromptPRTargetArg(options.PRTarget)
	resolveFlag := ""
	resolveScope := "all active PR review conversations"
	if options.CodeRabbitOnly {
		resolveFlag = " --coderabbit"
		resolveScope = "all active CodeRabbit-authored PR review conversations"
	}
	resolveCommand := fmt.Sprintf("kit dispatch --pr %s%s --resolve --yes", targetArg, resolveFlag)

	doc.Heading(2, "PR Reflection and Resolution Cycle")
	doc.Paragraph("Because this prompt came from PR review feedback, complete this post-implementation cycle after validation and push-to-PR. `kit pr fix` itself performs no Git or GitHub mutation; these instructions are for the coding agent after the repository delivery gate allows staging, commit, push, and review-thread resolution.")
	doc.OrderedList(1,
		fmt.Sprintf("Before editing, record the current PR head with `gh pr view %s --json headRefOid -q .headRefOid`.", targetArg),
		"Address every normalized review task, or document an explicit stale/no-op decision with code-context evidence.",
		"Run validation, then review the full diff in repository context instead of checking only the commented lines.",
		"After repo-local delivery rules allow it, commit and push the repair to the PR branch.",
		fmt.Sprintf("After push-to-PR, verify the remote PR head still equals the commit you pushed with `gh pr view %s --json headRefOid -q .headRefOid` and `git rev-parse HEAD`.", targetArg),
		"Run a reflection cycle against the pushed diff: re-read the review tasks, inspect the changed code in context, confirm each addressed conversation is fixed or intentionally no-op, and rerun any validation required by the reflection.",
		fmt.Sprintf("If no code has been pushed to the PR after your push and %s from this prompt are addressed, resolve them with the gh-backed Kit resolver: `%s`.", resolveScope, resolveCommand),
		"If the PR head changed after your push, if the editor task list omitted any active conversation, if any conversation is only partially addressed, or if validation/reflection is uncertain, do not run broad resolution; re-fetch active conversations with gh and resolve only the verified addressed conversations, or ask the user.",
		"Final response must include the pushed commit, reflection evidence, resolved conversation count or reason resolution was skipped, and any remaining PR conversations.",
	)
}

func dispatchPromptPRTargetArg(target string) string {
	target = strings.TrimSpace(target)
	if target == "" {
		return "<target>"
	}
	return strconv.Quote(target)
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
