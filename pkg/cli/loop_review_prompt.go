package cli

import (
	"fmt"
	"strings"
)

func buildLoopReviewPrompt(
	opts loopReviewOptions,
	target loopReviewTarget,
	prCtx *reviewLoopPRContext,
	prFeedback string,
) string {
	var builder strings.Builder
	builder.WriteString("# Kit Loop Review\n\n")
	builder.WriteString("Run a local correctness review and repair pass over changed code.\n\n")
	builder.WriteString("## Target\n\n")
	builder.WriteString(fmt.Sprintf("- Base ref: `%s`\n", target.BaseRef))
	builder.WriteString("- Scope: changes on the current branch relative to the remote mainline, plus staged and unstaged working-tree changes.\n")
	if opts.Feature != nil {
		builder.WriteString(fmt.Sprintf("- Feature docs: `%s`\n", opts.Feature.DirName))
	}
	if prCtx != nil {
		builder.WriteString(fmt.Sprintf("- Pull request: `%s`\n", reviewLoopTargetRef(prCtx.Target)))
		if prCtx.URL != "" {
			builder.WriteString(fmt.Sprintf(" (%s)", prCtx.URL))
		}
		builder.WriteString("\n")
	}
	if target.NoLocalChanges {
		builder.WriteString("- Changed files: none detected.\n")
	} else {
		builder.WriteString("- Changed files:\n")
		for _, path := range target.ChangedFiles {
			builder.WriteString(fmt.Sprintf("  - `%s`\n", path))
		}
	}
	builder.WriteString("\n## Diff Evidence\n\n")
	appendStatBlock(&builder, "Branch diff", target.DiffStat)
	appendStatBlock(&builder, "Unstaged diff", target.WorkingStat)
	appendStatBlock(&builder, "Staged diff", target.StagedStat)
	if strings.TrimSpace(prFeedback) != "" {
		builder.WriteString("\n## CodeRabbit Feedback To Ingest\n\n")
		builder.WriteString(strings.TrimSpace(prFeedback))
		builder.WriteString("\n")
	}
	builder.WriteString("\n## Instructions\n\n")
	builder.WriteString("- Inspect the actual diff and surrounding code before changing anything.\n")
	builder.WriteString("- Use Kit RLM: load repo-local docs just in time, prefer the smallest relevant sections, and stop loading once the repair decision is supported.\n")
	builder.WriteString("- When repository invariants, progress history, or workflow rules affect the fix, consult `docs/CONSTITUTION.md`, `docs/PROJECT_PROGRESS_SUMMARY.md`, active feature docs, and relevant `docs/references/rules/*` files.\n")
	builder.WriteString("- Fix high, medium, and correctness-impacting issues; do not churn on low-risk style unless it affects correctness.\n")
	builder.WriteString("- For PR feedback, verify every finding against current code; skip stale, resolved, or no-op feedback with a brief reason.\n")
	builder.WriteString("- Run the smallest relevant validation commands and add or update focused tests when needed.\n")
	builder.WriteString("- Emit concise progress updates for long-running work, command failures, blockers, and any pending user input.\n")
	builder.WriteString("- Share brief rationale summaries for decisions; do not expose private chain-of-thought.\n")
	if prCtx != nil && opts.ResolvePRFeedback {
		builder.WriteString("- Do not stage, commit, push, post PR comments, or mutate GitHub except for the review-thread resolution step below.\n")
		builder.WriteString(fmt.Sprintf("- After fixes and no-op decisions are validated, resolve all matching current unresolved review threads for `%s`, including human and CodeRabbit feedback, with `kit dispatch --pr %d --resolve --yes` without `--coderabbit`.\n", reviewLoopTargetRef(prCtx.Target), prCtx.Target.Number))
		builder.WriteString("- Resolve only feedback you verified as fixed or intentionally no-op; do not resolve unfixed, uncertain, stale, or unrelated feedback.\n")
		builder.WriteString("- Report the review-thread resolution command and result in the final response.\n")
	} else {
		builder.WriteString("- Do not stage, commit, push, post PR comments, resolve review threads, or mutate GitHub.\n")
	}
	builder.WriteString("- If no blocking issues remain, report `done`; otherwise make the next minimal fix and report what changed.\n")
	builder.WriteString(fmt.Sprintf("- Do not report `done` unless correctness is at least %d%% and there are no high, medium, or correctness-impacting issues.\n", opts.MinConfidence))
	prompt := builder.String()
	if opts.Feature != nil {
		prompt = preparePromptForFeature(prompt, opts.UseSubagents && !singleAgent, opts.Feature.Path)
	} else {
		prompt = preparePrompt(prompt, opts.UseSubagents && !singleAgent)
	}
	var final strings.Builder
	final.WriteString(strings.TrimRight(prompt, "\n"))
	final.WriteString("\n\n")
	if opts.UseSubagents && !singleAgent {
		appendLoopReviewSubagentPreAnalysis(&final)
		final.WriteString("\n")
	}
	appendLoopReviewFinalOutput(&final, prCtx != nil, prCtx != nil && opts.ResolvePRFeedback)
	return final.String()
}

func appendLoopReviewSubagentPreAnalysis(builder *strings.Builder) {
	builder.WriteString("## Review Subagent Pre-Analysis\n\n")
	builder.WriteString("- Before launching any subagents, inspect the diff at a high level and identify independent review lanes.\n")
	builder.WriteString("- Emit a concise progress update naming the planned subagent count and lane boundaries.\n")
	builder.WriteString("- Use zero subagents when the changed files are tightly coupled or the split is unclear.\n")
	builder.WriteString("- Keep the parent agent responsible for final synthesis, correctness scoring, validation, and the required final output.\n")
}

func appendLoopReviewFinalOutput(builder *strings.Builder, includePRPending bool, includeResolutionStatus bool) {
	builder.WriteString("## Required Final Output\n\n")
	builder.WriteString("Keep the final response information dense and short:\n\n")
	builder.WriteString("```text\n")
	builder.WriteString("Correctness: 97%\n")
	builder.WriteString("Status: <short status>\n\n")
	builder.WriteString("- Issue: <short finding>; Fix: <short action>.\n")
	builder.WriteString("- Issue: <short finding>; Fix: <short action>.\n")
	if includeResolutionStatus {
		builder.WriteString("\nReview threads: <resolved count and skipped/remaining reason>.\n")
	}
	if includePRPending {
		builder.WriteString("\nCodeRabbit has not completed for PR #<number> yet.\n")
		builder.WriteString("Rerun: kit loop review --pr <number>\n")
	}
	builder.WriteString("done\n")
	builder.WriteString("```\n")
}

func appendStatBlock(builder *strings.Builder, title, content string) {
	builder.WriteString(fmt.Sprintf("### %s\n\n", title))
	if strings.TrimSpace(content) == "" {
		builder.WriteString("none\n\n")
		return
	}
	builder.WriteString("```text\n")
	builder.WriteString(strings.TrimSpace(content))
	builder.WriteString("\n```\n\n")
}
