package prfix

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func ValidateMaxSubagents(value int) error {
	if value < 1 || value > HardMaxSubagents {
		return fmt.Errorf("--max-subagents must be between 1 and %d", HardMaxSubagents)
	}
	return nil
}

func RenderFeedback(items []Feedback) string {
	var output strings.Builder
	for index, item := range items {
		fmt.Fprintf(&output, "### Finding %d\n\n", index+1)
		fmt.Fprintf(&output, "- Kind: %s\n", item.Kind)
		if item.ThreadID != "" {
			fmt.Fprintf(&output, "- Review thread: %s\n", item.ThreadID)
		}
		location := item.Path
		if item.Line > 0 {
			location = fmt.Sprintf("%s:%d", item.Path, item.Line)
		}
		if location != "" {
			fmt.Fprintf(&output, "- Source: %s\n", location)
		}
		fmt.Fprintf(&output, "- Author: %s\n- URL: %s\n- Fingerprint: %s\n\n",
			item.Author, item.URL, item.Fingerprint)
		output.WriteString("```text\n")
		output.WriteString(strings.TrimSpace(item.Task))
		output.WriteString("\n```\n\n")
	}
	return strings.TrimSpace(output.String())
}

func BuildPrompt(target Target, lane Lane, feedback []Feedback, tasks string, maxSubagents int) (string, error) {
	if err := ValidateMaxSubagents(maxSubagents); err != nil {
		return "", err
	}
	if strings.TrimSpace(tasks) == "" {
		return "", fmt.Errorf("PR feedback tasks cannot be empty")
	}
	var prompt strings.Builder
	prompt.WriteString("Act as the one accountable supervisor for this pull-request feedback repair.\n\n")
	prompt.WriteString("## Resolve the repository contract\n\nRun this local-only, read-only command before planning or editing:\n\n")
	prompt.WriteString("```bash\nkit contract resolve --workflow pr-feedback-repair --json\n```\n\n")
	prompt.WriteString("Read the returned workflow, all dependencies, and repository-local RLM routing. Kit generated this prompt and prepared the lane; it did not launch agents or perform repairs.\n\n")
	prompt.WriteString("## Pinned repair lane\n\n")
	writePromptTable(&prompt, lane)
	prompt.WriteString("\n- Run every filesystem and Git operation in the recorded repair worktree.\n")
	prompt.WriteString("- Before editing, recheck the PR head, local HEAD, branch, worktree registration, push target, and dirty status. Stop if any pinned identity changed.\n")
	prompt.WriteString("- Never repair from a detached `PR-<number>` lane or create a second repair branch or pull request.\n")
	writeDirtyInstructions(&prompt, lane)
	prompt.WriteString("\n## Current active feedback\n\n")
	prompt.WriteString(strings.TrimSpace(tasks))
	prompt.WriteString("\n\n## Agent Team Plan\n\n")
	fmt.Fprintf(&prompt, "- Publish an Agent Team Plan before spawning. Use at most %d independent low-overlap concurrent lanes; the hard ceiling is %d.\n", maxSubagents, HardMaxSubagents)
	prompt.WriteString("- One supervisor owns scope, finding disposition, lane assignment, integration, validation, delivery, reflection, and thread resolution.\n")
	prompt.WriteString("- Serialize shared or ambiguous files and queue excess work. Subagents may not create, switch, move, or remove worktrees or mutate Git/GitHub delivery state.\n")
	prompt.WriteString("- After nontrivial repair, use a separate read-only verification lane. If no agent is actually spawned, report `single supervisor lane; no specialist or verification agents spawned`.\n\n")
	prompt.WriteString("## Repair and delivery contract\n\n")
	prompt.WriteString("1. Verify every finding against current `HEAD`, its current path and line, and the integrated implementation. Fix only still-valid findings.\n")
	prompt.WriteString("2. Record an evidence-based disposition for every item: fixed, stale, false-positive, out-of-scope, or human-needed. Do not silently drop feedback.\n")
	prompt.WriteString("3. Keep changes minimal, integrate every repair in the supervisor lane, and validate the complete combined diff under repository rules.\n")
	prompt.WriteString("4. Explicitly stage intended files and push one coherent batch only to the recorded existing PR branch after the delivery gate allows it.\n")
	prompt.WriteString("5. Verify the exact pushed commit equals the remote PR head. Re-read the findings, review the full pushed diff, reflect, and rerun required validation.\n")
	prompt.WriteString("6. Resolve only current unresolved, non-outdated review threads that are verified addressed by that pushed head. Leave stale, partial, human-needed, and non-thread feedback visible.\n")
	prompt.WriteString("7. Stop after two head epochs or two repair passes and report remaining feedback instead of creating an infinite loop.\n\n")
	writeResolutionInstructions(&prompt, target, feedback)
	prompt.WriteString("\n## Completion report\n\nReport the Agent Team Plan and actual agents used, every finding disposition, files changed or preserved, validation and full-diff review, pushed commit and exact-head proof, reflection, explicitly resolved thread IDs, remaining feedback, and any blocked or human-needed action.\n")
	return prompt.String(), nil
}

func writePromptTable(output *strings.Builder, lane Lane) {
	rows := [][2]string{
		{"Pull request", lane.PRURL}, {"Repository", lane.Repository},
		{"Head branch", lane.HeadBranch}, {"Expected remote head", lane.ExpectedHead},
		{"Local head", lane.LocalHead}, {"Repair worktree", lane.WorktreePath},
		{"Push target", lane.PushTarget}, {"Dirty ownership", lane.DirtyOwnership},
	}
	output.WriteString("| Field | Pinned value |\n| --- | --- |\n")
	for _, row := range rows {
		fmt.Fprintf(output, "| %s | `%s` |\n", row[0], strings.ReplaceAll(row[1], "|", "\\|"))
	}
}

func writeDirtyInstructions(output *strings.Builder, lane Lane) {
	if len(lane.DirtyPaths) == 0 {
		output.WriteString("- The lane was clean when this prompt was generated.\n")
		return
	}
	fmt.Fprintf(output, "- Existing dirty changes were explicitly marked `%s` for this repair.\n", lane.DirtyOwnership)
	for _, path := range lane.DirtyPaths {
		fmt.Fprintf(output, "  - `%s`\n", path)
	}
	if lane.DirtyOwnership == "exclude" {
		output.WriteString("- Preserve excluded paths without editing or staging them; stop if repair scope overlaps them.\n")
	} else {
		output.WriteString("- Review included paths as part of the full PR diff and validate their ownership before staging.\n")
	}
}

func writeResolutionInstructions(output *strings.Builder, target Target, feedback []Feedback) {
	threadIDs := ResolutionThreadIDs(feedback)
	output.WriteString("## Explicit thread resolution\n\n")
	if len(threadIDs) == 0 {
		output.WriteString("No collected item has a resolvable review-thread identity. Do not resolve top-level reviews or comments.\n")
		return
	}
	output.WriteString("After exact pushed-head verification, select only verified addressed IDs from this list:\n")
	for _, id := range threadIDs {
		fmt.Fprintf(output, "- `%s`\n", id)
	}
	command := fmt.Sprintf("kit pr fix --pr %s --resolve --head PUSHED_HEAD_SHA --yes", strconv.Quote(target.URL()))
	for _, id := range threadIDs {
		command += " --thread " + strconv.Quote(id)
	}
	output.WriteString("\nRemove every unaddressed ID, replace `PUSHED_HEAD_SHA`, then run explicitly:\n\n```bash\n")
	output.WriteString(command)
	output.WriteString("\n```\n")
}

func ResolutionThreadIDs(feedback []Feedback) []string {
	seen := map[string]bool{}
	var result []string
	for _, item := range feedback {
		if item.ThreadID != "" && !seen[item.ThreadID] {
			seen[item.ThreadID] = true
			result = append(result, item.ThreadID)
		}
	}
	sort.Strings(result)
	return result
}
