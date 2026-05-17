package cli

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/promptdoc"
)

const projectRefreshCommand = "kit prompt project refresh"

func buildProjectRefreshPrompt(projectRoot string, cfg *config.Config) string {
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
		doc.Raw("/plan")
		doc.Heading(2, "Project Refresh")
		doc.Paragraph(fmt.Sprintf("Refresh durable project-level documentation for the Kit project at %s.", projectRoot))
		doc.Paragraph("Rules:")
		doc.BulletList(
			"docs only; do not change product code, tests, runtime config, or generated artifacts unless the user separately asks",
			"this is semantic project refresh, not re-initialization; do not rerun `kit init` as the fix",
			"use `kit reconcile --all` for structural Kit contract drift instead of duplicating that audit manually",
			"preserve existing project wording when it remains accurate",
			"update `docs/CONSTITUTION.md` only for durable project-wide rules, constraints, vocabulary, conventions, or long-term direction",
			"update repository instruction docs only when canonical routing or workflow guidance changed",
		)
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
		doc.BulletList(
			"`docs/CONSTITUTION.md`: refresh durable project-wide principles, constraints, definitions, and codebase map entries",
			"`docs/agents/*`, `AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md`: update only if routing or workflow guidance is stale; prefer `kit scaffold agents --append-only` when the generated contract changed",
			"`docs/PROJECT_PROGRESS_SUMMARY.md`: update with `kit rollup` if feature summaries or project state change",
			"feature docs under `docs/specs/`: leave alone unless the project-level refresh reveals a direct inconsistency that belongs there",
		)
		doc.Paragraph("Verification:")
		doc.BulletList(
			"`kit check --project`",
			"`kit check --all` if feature docs or repository instruction files were touched",
			"`kit rollup` if `docs/PROJECT_PROGRESS_SUMMARY.md` needs regeneration",
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
	return "Project refresh advisory gate\n" +
		"- decide whether this work revealed durable project-wide rules, constraints, vocabulary, conventions, or workflow changes\n" +
		"- if yes, run `" + projectRefreshCommand + "` and refresh project-level docs before final handoff\n" +
		"- if no, state `no project refresh needed` in the reflection notes"
}

func printProjectRefreshAdvisory(out io.Writer) error {
	_, err := fmt.Fprintf(out, "  ℹ Project refresh advisory: if this work changed durable project-level rules, run `%s`.\n", projectRefreshCommand)
	return err
}
