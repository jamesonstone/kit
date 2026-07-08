package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/promptdoc"
)

const projectRefreshCommand = "kit project refresh"

func buildProjectRefreshPrompt(projectRoot string, cfg *config.Config) string {
	status, _ := calculateProjectRefreshStatus(projectRoot, cfg, time.Now().UTC())
	return buildProjectRefreshPromptWithOptions(projectRoot, cfg, projectRefreshPromptOptions{Status: status})
}

type projectRefreshPromptOptions struct {
	ConstitutionOnly bool
	Status           projectRefreshStatus
}

func buildProjectRefreshPromptWithOptions(projectRoot string, cfg *config.Config, opts projectRefreshPromptOptions) string {
	constitutionPath := cfg.ConstitutionAbsPath(projectRoot)
	summaryPath := cfg.ProgressSummaryPath(projectRoot)
	specsPath := cfg.SpecsPath(projectRoot)
	agentsPath := repoKnowledgeEntrypointPath(projectRoot, cfg)
	referencesPath := repoReferencesEntrypointPath(projectRoot, cfg)

	contextDocs := []string{
		fmt.Sprintf("Project root: %s", projectRoot),
		fmt.Sprintf("Constitution: %s", constitutionPath),
		fmt.Sprintf("Project summary: %s", summaryPath),
		fmt.Sprintf("Feature docs: %s", specsPath),
	}
	if agentsPath != "" {
		contextDocs = append(contextDocs, fmt.Sprintf("Agent routing docs: %s", agentsPath))
	}
	if referencesPath != "" {
		contextDocs = append(contextDocs, fmt.Sprintf("References: %s", referencesPath))
	}

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Heading(2, "Project Refresh")
		doc.Paragraph(fmt.Sprintf("Refresh durable project-level documentation for the Kit project at %s.", projectRoot))
		doc.Paragraph("Rules:")
		doc.BulletList(
			docsOnlyWorkflowRule("project-level documentation"),
			"this is semantic project refresh, not re-initialization; do not rerun `kit init` as the fix",
			"use `kit reconcile --all` for structural Kit contract drift instead of duplicating that audit manually",
			"preserve existing project wording when it remains accurate",
			"update `docs/CONSTITUTION.md` only for durable project-wide rules, constraints, vocabulary, conventions, or long-term direction",
			"do not use this command for structural scaffold updates; use `kit reconcile --all` for that work",
		)
		doc.Paragraph("Current cadence state:")
		doc.BulletList(projectRefreshStatusBullets(opts.Status)...)
		doc.Paragraph("Context to inspect:")
		doc.BulletList(contextDocs...)
		doc.Paragraph("Discovery commands:")
		doc.BulletList(
			"`git status --short`",
			"`git diff --stat`",
			"`kit status --all`",
			"`kit check --project`",
			"`kit reconcile --all` if structural document drift is suspected",
			fmt.Sprintf(
				"`rg -n \"TODO|placeholder|stale|outdated|%s\" %s %s`",
				filepath.Base(constitutionPath),
				filepath.Join(projectRoot, "docs"),
				filepath.Join(projectRoot, "README.md"),
			),
		)
		doc.Paragraph("Analyze for durable changes:")
		doc.BulletList(
			"new package, command, config, or workflow boundaries that should be project-level rules",
			"recurring conventions that emerged across feature docs or implementation work",
			"vocabulary that future agents need to use consistently",
			"constraints that the initial constitution could not know before the repository had real contents",
			"places where implementation reality challenges or refines the current constitution",
		)
		doc.Paragraph("Update guidance:")
		if opts.ConstitutionOnly {
			doc.BulletList(
				"`docs/CONSTITUTION.md`: refresh durable project-wide principles, constraints, definitions, vocabulary, workflow rules, and codebase map entries",
				"do not update feature docs, repository instruction docs, progress summaries, code, tests, runtime config, or generated artifacts in this Constitution-only pass",
				"if you discover structural drift or stale feature-specific docs, report it in `Findings` instead of editing those files",
			)
		} else {
			doc.BulletList(
				"`docs/CONSTITUTION.md`: refresh durable project-wide principles, constraints, definitions, and codebase map entries",
				"`docs/agents/*`, `AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md`: update only if routing or workflow guidance is stale; prefer `kit scaffold agents --append-only` when the generated contract changed",
				"`docs/PROJECT_PROGRESS_SUMMARY.md`: update if feature summaries or project state change",
				"feature docs under `docs/specs/`: leave alone unless the project-level refresh reveals a direct inconsistency that belongs there",
			)
		}
		doc.Paragraph("Verification:")
		doc.BulletList(
			"`kit check --project`",
			"`kit check --all` if feature docs or repository instruction files were touched",
			"confirm `docs/PROJECT_PROGRESS_SUMMARY.md` reflects any feature-summary or project-state changes",
			"`git diff -- docs AGENTS.md CLAUDE.md .github/copilot-instructions.md README.md`",
		)
		doc.Paragraph("Reply with exactly these sections:")
		doc.BulletList(
			"`Findings`: what project-level truth was stale, or `none`",
			"`Updates`: files changed and why; include `no project refresh needed` if nothing durable changed",
			"`Verification`: commands run and whether they passed",
		)
	})
}

func projectRefreshAdvisoryStep() string {
	return projectRefreshAdvisoryStepForStatus(projectRefreshStatus{})
}

func projectRefreshAdvisoryStepForStatus(status projectRefreshStatus) string {
	statusLine := "- current due state: unknown; run `kit project refresh --output-only` to inspect cadence state"
	if status.FeatureInterval > 0 {
		statusLine = "- current due state: " + formatProjectRefreshDueSummary(status)
	}
	return "Project refresh advisory gate\n" +
		statusLine + "\n" +
		"- decide whether this work revealed durable project-wide rules, constraints, vocabulary, conventions, or workflow changes\n" +
		"- if the project refresh is due, or if durable project-level truth changed, run `" + projectRefreshCommand + "` and refresh project-level docs before final handoff\n" +
		"- after completing a reviewed semantic Constitution refresh, run `kit project refresh --now` to record the cadence state\n" +
		"- if no, state `no project refresh needed` in the reflection notes"
}

func printProjectRefreshAdvisory(out io.Writer, projectRoot string, cfg *config.Config) error {
	status, err := calculateProjectRefreshStatus(projectRoot, cfg, time.Now().UTC())
	if err != nil {
		_, writeErr := fmt.Fprintf(out, "  ℹ Project refresh advisory: run `%s` if this work changed durable project-level rules (due status unavailable: %v).\n", projectRefreshCommand, err)
		return writeErr
	}
	return printProjectRefreshStatusSummary(out, status)
}
