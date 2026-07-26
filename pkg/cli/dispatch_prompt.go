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
		doc.Paragraph("Prepare an Agent Team Plan for the following task set, then execute it conservatively.")
		doc.Heading(2, "Dispatch Context")
		doc.BulletList(
			fmt.Sprintf("Working directory: `%s`", workingDirectory),
			fmt.Sprintf("Input source: %s", inputSource),
			fmt.Sprintf("Max concurrent subagents: %d (hard ceiling %d)", maxSubagents, hardDispatchMaxSubagents),
		)
		if options.RepairContext != nil {
			appendDispatchRepairContext(doc, *options.RepairContext)
		}
		if strings.TrimSpace(options.CommonReviewInstruction) != "" {
			doc.Heading(2, "Common Review Instruction")
			doc.CodeBlock("text", options.CommonReviewInstruction)
		}
		doc.Heading(2, "Normalized Tasks")
		doc.Raw(renderDispatchTasks(tasks))
		doc.Heading(2, "Routing Task")
		doc.BulletList(
			"Act as the one accountable supervisor for scope, integration, validation, evidence, delivery gates, and reporting.",
			"Inspect just enough repository structure to predict touched files/interfaces for each task. Cluster by file overlap; keep shared, ambiguous, or continuously coupled work in one serialized lane.",
			fmt.Sprintf("Spawn only independent low-overlap clusters, at most %d concurrently. Queue excess clusters in original task order; use one supervisor lane when splitting would not improve correctness or throughput.", maxSubagents),
			"Plan at least one read-only verification lane after nontrivial implementation unless the work is documentation-only, tightly coupled, explicitly single-agent, or the runtime cannot spawn agents.",
			"Keep each task in its assigned checkout or prepared worktree. Subagents may not independently create, switch, move, or remove worktrees, expand scope, or mutate Git/GitHub delivery state unless explicitly assigned and authorized.",
		)
		doc.Heading(2, "Agent Team Plan Output")
		doc.BulletList(
			"Supervisor responsibility and normalized task-to-lane mapping.",
			"For each lane: actual or logical-only status, predicted files/interfaces, dependencies, overlap risk, and validation owner.",
			"Max concurrency, serialized queue, omitted lanes with reasons, read-only verification plan, and remaining unknowns.",
			"Publish the plan before spawning, then self-direct execution within it. Final reporting distinguishes actual agents from logical lanes; if none ran, state `single supervisor lane; no specialist or verification agents spawned`.",
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
	RepairContext           *repairContext
}

func appendDispatchRepairContext(doc *promptdoc.Document, repair repairContext) {
	preparation := "reused"
	if repair.WorktreeCreated {
		preparation = "created"
	}
	rows := [][]string{
		{"Repository", repair.Repository},
		{"Head branch", repair.HeadBranch},
		{"Expected remote head", firstNonEmpty(repair.ExpectedHeadOID, "not supplied")},
		{"Local head", repair.LocalHeadOID},
		{"Repair worktree", repair.WorktreePath},
		{"Worktree preparation", preparation},
		{"Existing changes", string(repair.ExistingChanges)},
		{"Push target", repair.PushTarget},
	}
	if repair.PRURL != "" {
		rows = append([][]string{
			{"Pull request", repair.PRURL},
		}, rows...)
	}

	doc.Heading(2, "Repair Lane")
	doc.Table([]string{"Field", "Resolved Value"}, rows)
	if repair.DirtyStatus != "" {
		doc.Paragraph("Existing worktree status recorded when this prompt was generated:")
		doc.CodeBlock("text", repair.DirtyStatus)
	}

	pathArg := strconv.Quote(repair.WorktreePath)
	instructions := []string{
		fmt.Sprintf("Run every filesystem and Git command in `%s`; use an explicit command working directory or `git -C %s ...` instead of the invoking checkout.", repair.WorktreePath, pathArg),
		fmt.Sprintf("Before editing, verify `git -C %s symbolic-ref --quiet --short HEAD` is exactly `%s` and preserve any state that no longer matches this prompt.", pathArg, repair.HeadBranch),
	}
	if repair.ExpectedHeadOID != "" {
		instructions = append(instructions,
			fmt.Sprintf("The prompt was prepared for remote head `%s`. If local `HEAD` or the current PR head differs, fetch and reconcile conservatively without stash, reset, clean, rebase, force operations, or discarding user work.", repair.ExpectedHeadOID),
		)
	}
	switch repair.ExistingChanges {
	case repairChangesInclude:
		instructions = append(instructions,
			"The user explicitly included the recorded existing changes in this repair. Review them as part of the full PR diff, keep only work that belongs to the target, and validate it before delivery.",
		)
	case repairChangesExclude:
		instructions = append(instructions,
			"The user explicitly excluded the recorded existing changes. Preserve them, do not stage or modify their paths, and stop for user direction if the repair overlaps them.",
		)
	}
	instructions = append(instructions,
		fmt.Sprintf("Any authorized push must update `%s` only; never create a second repair branch or pull request for this target.", repair.PushTarget),
	)
	doc.BulletList(instructions...)
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
