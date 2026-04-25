package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/promptdoc"
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
	cfg, _ := loadRepoInstructionContext(projectRoot)
	repoAgentsPath := repoKnowledgeEntrypointPath(projectRoot, cfg)
	repoReferencesPath := repoReferencesEntrypointPath(projectRoot, cfg)

	rows := [][]string{
		{"CONSTITUTION", constitutionPath, "project-wide constraints"},
	}
	if repoAgentsPath != "" {
		rows = append(rows, []string{"AGENTS DOCS", repoAgentsPath, "repo-local entrypoint and read-order guide"})
	}
	if repoReferencesPath != "" {
		rows = append(rows, []string{"REFERENCES", repoReferencesPath, "repo-wide references only when relevant"})
	}
	rows = append(rows, []string{"PROJECT_PROGRESS_SUMMARY", summaryPath, "cross-feature context"})
	if status.Files["brainstorm"].Exists {
		rows = append(rows, []string{"BRAINSTORM", brainstormPath, "upstream research and framing"})
	}
	if status.Files["spec"].Exists {
		rows = append(rows, []string{"SPEC", specPath, "requirements and acceptance"})
	}
	if status.Files["plan"].Exists {
		rows = append(rows, []string{"PLAN", planPath, "approach and design decisions"})
	}
	if status.Files["tasks"].Exists {
		rows = append(rows, []string{"TASKS", tasksPath, "execution status and remaining work"})
	}
	rows = append(rows, []string{"Project Root", projectRoot, "repository context"})

	featureDocs := []string{}
	if status.Files["brainstorm"].Exists {
		featureDocs = append(featureDocs, "`BRAINSTORM.md`")
	}
	if status.Files["spec"].Exists {
		featureDocs = append(featureDocs, "`SPEC.md`")
	}
	if status.Files["plan"].Exists {
		featureDocs = append(featureDocs, "`PLAN.md`")
	}
	if status.Files["tasks"].Exists {
		featureDocs = append(featureDocs, "`TASKS.md`")
	}

	steps := []string{
		"Stay in plan mode for this catch-up step",
		"Do NOT implement code, edit files, or begin execution yet",
	}
	if repoAgentsPath != "" {
		steps = append(steps, "Read `docs/agents/README.md` and only the linked docs relevant to this feature")
	}
	if repoReferencesPath != "" {
		steps = append(steps, "Read `docs/references/README.md` only if a repo-wide reference materially shapes this feature")
	}
	steps = append(steps,
		"Read `CONSTITUTION.md` first",
		"Read `PROJECT_PROGRESS_SUMMARY.md` for cross-feature context",
		"Read the feature docs in order:\n- "+strings.Join(featureDocs, "\n- "),
		"Reconstruct the current stage and state of the feature from the repository artifacts before making any recommendations",
		"Start by asking clarifying questions in a short numbered batch",
		"For each question, include your current best recommendation, assumption, or default",
		fmt.Sprintf("Use the standard batch-approval syntax for planning questions: %s", approvalSyntaxSummary),
		"Ask explicitly whether the user wants to continue planning, validate the current state, or move into implementation",
		"Do NOT switch from catch-up/planning into implementation until the user explicitly approves that move",
		fmt.Sprintf("If conversation context is missing, you may optionally ask the user to provide a prior summary or use `kit summarize %s`, but treat repository documents and current code as the primary source of truth", feat.Slug),
	)

	if status.Phase == feature.PhaseComplete {
		steps = append(steps,
			"This feature is already marked `complete`; treat this catch-up as review or reopen triage only",
			"Do not assume implementation should resume unless the user explicitly asks to reopen work on this feature",
		)
	} else {
		steps = append(steps,
			"After you have caught up, summarize what stage the feature is in, what is already decided, what is still open, and what the next sensible step would be",
			"Stop after the catch-up summary and questions unless the user explicitly approves moving to implementation",
		)
	}

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Raw("/plan")
		doc.Paragraph(fmt.Sprintf("Catch up on feature: %s", feat.Slug))
		doc.Heading(2, "Current Stage And State")
		doc.BulletList(
			fmt.Sprintf("Active feature: %s", feat.DirName),
			fmt.Sprintf("Current phase: %s", status.Phase),
			fmt.Sprintf("Current state: %s", catchupStateSummary(status)),
			fmt.Sprintf("Next canonical artifact: %s", resumeNextCanonicalArtifact(status)),
			fmt.Sprintf("Next recommended command: %s", resumeNextRecommendedCommand(status)),
			fmt.Sprintf("Known blockers: %s", resumeKnownBlockers(status)),
			fmt.Sprintf("Validation state: %s", resumeValidationState(status)),
		)
		if status.Summary != "" {
			doc.Heading(2, "Feature Summary")
			doc.Paragraph(status.Summary)
		}
		doc.Heading(2, "Context Docs")
		doc.Table([]string{"File", "Path", "Use"}, rows)
		doc.Heading(2, "Your Task")
		doc.OrderedList(1, steps...)
		doc.Heading(2, "Rules")
		doc.BulletList(
			"this command is feature-scoped; do not broaden into a project-wide handoff unless the user asks",
			"do not duplicate the full `kit handoff` workflow",
			"do not duplicate the full `kit summarize` workflow",
			"do not output implementation instructions like `kit implement` unless the user explicitly asks to proceed",
			"repository documents and current code are the source of truth when prior conversation context is incomplete",
			fmt.Sprintf("feature path: %s", feat.Path),
			fmt.Sprintf("project root: %s", projectRoot),
		)
	})
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

func resumeNextCanonicalArtifact(status *feature.FeatureStatus) string {
	if status == nil {
		return "unknown"
	}
	if status.Phase == feature.PhaseComplete {
		return "TASKS.md reflection marker"
	}
	if !status.Files["spec"].Exists {
		return "SPEC.md"
	}
	if !status.Files["plan"].Exists {
		return "PLAN.md"
	}
	if !status.Files["tasks"].Exists {
		return "TASKS.md"
	}
	return "TASKS.md"
}

func resumeNextRecommendedCommand(status *feature.FeatureStatus) string {
	if status == nil {
		return "inspect feature artifacts"
	}
	return catchupNextAction(status)
}

func resumeKnownBlockers(status *feature.FeatureStatus) string {
	if status == nil {
		return "unknown"
	}
	if status.Paused {
		return "feature is paused"
	}
	if !status.Files["spec"].Exists {
		return "SPEC.md is missing"
	}
	if !status.Files["plan"].Exists {
		return "PLAN.md is missing"
	}
	if !status.Files["tasks"].Exists {
		return "TASKS.md is missing"
	}
	if status.Progress == nil || !status.Progress.HasTasks() {
		return "TASKS.md has no markdown checkbox tasks"
	}
	return "none recorded in Kit artifacts"
}

func resumeValidationState(status *feature.FeatureStatus) string {
	if status == nil {
		return "unknown"
	}
	return fmt.Sprintf("unknown from current artifacts; run `kit check %s` when validation is needed", status.Name)
}

func presenceWord(exists bool) string {
	if exists {
		return "present"
	}
	return "absent"
}
