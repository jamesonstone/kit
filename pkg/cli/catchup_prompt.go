package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/feature"
)

func buildCatchupPrompt(
	feat *feature.Feature,
	status *feature.FeatureStatus,
	projectRoot string,
) string {
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	summaryPath := filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md")
	brainstormPath := status.Files["brainstorm"].Path
	specPath := status.Files["spec"].Path
	planPath := status.Files["plan"].Path
	tasksPath := status.Files["tasks"].Path

	var sb strings.Builder
	sb.WriteString("/plan\n\n")
	sb.WriteString(fmt.Sprintf("Catch up on feature: %s\n\n", feat.Slug))
	sb.WriteString("## Current Stage And State\n")
	sb.WriteString(fmt.Sprintf("- Feature: %s\n", feat.DirName))
	sb.WriteString(fmt.Sprintf("- Current stage: %s\n", status.Phase))
	sb.WriteString(fmt.Sprintf("- Current state: %s\n", catchupStateSummary(status)))
	sb.WriteString(fmt.Sprintf("- Next suggested action: %s\n\n", catchupNextAction(status)))

	if status.Summary != "" {
		sb.WriteString("## Feature Summary\n")
		sb.WriteString(fmt.Sprintf("%s\n\n", status.Summary))
	}

	sb.WriteString("## Context Docs\n")
	sb.WriteString("| File | Path | Use |\n")
	sb.WriteString("|------|------|-----|\n")
	sb.WriteString(fmt.Sprintf("| CONSTITUTION | %s | project-wide constraints |\n", constitutionPath))
	sb.WriteString(fmt.Sprintf("| PROJECT_PROGRESS_SUMMARY | %s | cross-feature context |\n", summaryPath))
	if status.Files["brainstorm"].Exists {
		sb.WriteString(fmt.Sprintf("| BRAINSTORM | %s | upstream research and framing |\n", brainstormPath))
	}
	if status.Files["spec"].Exists {
		sb.WriteString(fmt.Sprintf("| SPEC | %s | requirements and acceptance |\n", specPath))
	}
	if status.Files["plan"].Exists {
		sb.WriteString(fmt.Sprintf("| PLAN | %s | approach and design decisions |\n", planPath))
	}
	if status.Files["tasks"].Exists {
		sb.WriteString(fmt.Sprintf("| TASKS | %s | execution status and remaining work |\n", tasksPath))
	}
	sb.WriteString(fmt.Sprintf("| Project Root | %s | repository context |\n\n", projectRoot))

	sb.WriteString("## Your Task\n")
	sb.WriteString("1. Stay in plan mode for this catch-up step\n")
	sb.WriteString("2. Do NOT implement code, edit files, or begin execution yet\n")
	sb.WriteString("3. Read `CONSTITUTION.md` first\n")
	sb.WriteString("4. Read `PROJECT_PROGRESS_SUMMARY.md` for cross-feature context\n")
	sb.WriteString("5. Read the feature docs in order:\n")
	if status.Files["brainstorm"].Exists {
		sb.WriteString("   - `BRAINSTORM.md`\n")
	}
	if status.Files["spec"].Exists {
		sb.WriteString("   - `SPEC.md`\n")
	}
	if status.Files["plan"].Exists {
		sb.WriteString("   - `PLAN.md`\n")
	}
	if status.Files["tasks"].Exists {
		sb.WriteString("   - `TASKS.md`\n")
	}
	sb.WriteString("6. Reconstruct the current stage and state of the feature from the repository artifacts before making any recommendations\n")
	sb.WriteString("7. Start by asking clarifying questions in a short numbered batch\n")
	sb.WriteString("8. For each question, include your current best recommendation, assumption, or default\n")
	sb.WriteString("9. Use the standard batch-approval syntax for planning questions: " + approvalSyntaxSummary + "\n")
	sb.WriteString("10. Ask explicitly whether the user wants to continue planning, validate the current state, or move into implementation\n")
	sb.WriteString("11. Do NOT switch from catch-up/planning into implementation until the user explicitly approves that move\n")
	sb.WriteString(
		fmt.Sprintf(
			"12. If conversation context is missing, you may optionally ask the user to provide a prior summary or use `kit summarize %s`, but treat repository documents and current code as the primary source of truth\n",
			feat.Slug,
		),
	)

	if status.Phase == feature.PhaseComplete {
		sb.WriteString("13. This feature is already marked `complete`; treat this catch-up as review or reopen triage only\n")
		sb.WriteString("14. Do not assume implementation should resume unless the user explicitly asks to reopen work on this feature\n")
	} else {
		sb.WriteString("13. After you have caught up, summarize what stage the feature is in, what is already decided, what is still open, and what the next sensible step would be\n")
		sb.WriteString("14. Stop after the catch-up summary and questions unless the user explicitly approves moving to implementation\n")
	}

	sb.WriteString("\nRules:\n")
	sb.WriteString("- this command is feature-scoped; do not broaden into a project-wide handoff unless the user asks\n")
	sb.WriteString("- do not duplicate the full `kit handoff` workflow\n")
	sb.WriteString("- do not duplicate the full `kit summarize` workflow\n")
	sb.WriteString("- do not output implementation instructions like `kit implement` unless the user explicitly asks to proceed\n")
	sb.WriteString("- repository documents and current code are the source of truth when prior conversation context is incomplete\n")
	sb.WriteString(fmt.Sprintf("- feature path: %s\n", feat.Path))
	sb.WriteString(fmt.Sprintf("- project root: %s\n", projectRoot))

	return sb.String()
}

func catchupStateSummary(status *feature.FeatureStatus) string {
	var parts []string
	parts = append(parts, fmt.Sprintf(
		"artifacts - BRAINSTORM %s, SPEC %s, PLAN %s, TASKS %s",
		presenceWord(status.Files["brainstorm"].Exists),
		presenceWord(status.Files["spec"].Exists),
		presenceWord(status.Files["plan"].Exists),
		presenceWord(status.Files["tasks"].Exists),
	))
	if status.Progress != nil && status.Progress.HasTasks() {
		parts = append(parts, fmt.Sprintf(
			"task progress %d/%d complete",
			status.Progress.Complete,
			status.Progress.Total,
		))
	}
	return strings.Join(parts, "; ")
}

func catchupNextAction(status *feature.FeatureStatus) string {
	if status.Phase == feature.PhaseComplete {
		return "Feature is complete; confirm whether the user wants review only or to reopen work"
	}
	return determineNextAction(status)
}

func presenceWord(exists bool) string {
	if exists {
		return "present"
	}
	return "absent"
}
