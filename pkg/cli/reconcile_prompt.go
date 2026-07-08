package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/internal/promptdoc"
)

type reconcileFileSummary struct {
	Path        string
	ErrorCount  int
	WarnCount   int
	Issues      []string
	Actions     []string
	SearchHints []string
}

type reconcileCategorySummary struct {
	Name        string
	SearchHints []string
}

func buildReconcilePrompt(report *reconcileReport) string {
	scope := "whole project"
	verifyCmd := "`kit check --all`"
	if report.Feature != nil {
		scope = fmt.Sprintf("feature %s", report.Feature.Slug)
		verifyCmd = fmt.Sprintf("`kit check %s`", report.Feature.Slug)
	}

	fileSummaries := summarizeReconcileFiles(report.Findings)
	categorySummaries := summarizeReconcileCategories(report.Findings)
	errorCount, warningCount := reconcileSeverityCounts(report.Findings)

	rows := make([][]string, 0, len(fileSummaries))
	for _, summary := range fileSummaries {
		rows = append(rows, []string{
			reconcileSeverityBadge(summary.ErrorCount, summary.WarnCount),
			fmt.Sprintf("`%s`", summary.Path),
			fmt.Sprintf("%d", len(summary.Issues)),
			strings.Join(limitStrings(summary.Actions, 2), "; "),
		})
	}

	rules := []string{
		docsOnlyWorkflowRule("Kit-managed docs and scaffold files"),
		"preserve project wording when it already satisfies the current contract",
		fmt.Sprintf(
			"contract order: %s -> %s -> %s",
			templateSource(report.ProjectRoot),
			constitutionSource(report.ProjectRoot),
			initProjectSource(report.ProjectRoot),
		),
	}
	if !singleAgent {
		rules = append(
			rules[:2],
			append(
				[]string{"use subagents and queue work according to overlapping file changes; keep overlapping files in the same lane"},
				rules[2:]...,
			)...,
		)
	}
	if report.ReferenceMigration {
		rules = append(rules,
			"migrate deprecated front matter `dependencies` to canonical `references`; do not preserve the old front matter field",
			"for each migrated entry, map `location` to `target`, add a stable `id` when the reference may be updated later, add a graph `relation`, add `read_policy`, and keep `used_for` plus `status`",
			"use stable selectors such as headings, symbols, artifact IDs, command flags, URLs, or node IDs; set `selector_type` to `artifact`, `heading`, `symbol`, `command`, `url`, or `node_id`; replace unpinned line ranges when practical",
			"treat `relation` as the referenced target's role relative to the source artifact",
			"prefer `read_policy: must` for constraints, `conditional` for supporting inputs, `evidence` for verification material, and `skip` for stale references",
		)
	}
	if report.VerificationMigration {
		rules = append(rules,
			"verification migration is advisory; do not mark legacy docs invalid only because they predate executable task fields",
			"inspect active `TASKS.md` only; for completed or historical features, leave missing executable fields alone",
			"add `VERIFY`, `EXPECTED FILES`, `RISK`, and `ROLLBACK` only to tasks currently being implemented, verified, or reflected",
			"do not guess verification commands from prose; infer commands only when already documented or obvious from repo tooling",
			"if acceptance criteria are prose-only, propose runnable checks separately from confirmed checks and leave uncertain commands as `not yet declared`",
			"after migration, run `kit legacy verify <feature> --dry-run`, refresh `.kit/state.json`, then rerun `kit check <feature>` and `kit check --project`",
		)
	}

	snapshot := []string{
		fmt.Sprintf("findings: %d (%d errors, %d warnings)", len(report.Findings), errorCount, warningCount),
		fmt.Sprintf("files to touch: %d", len(fileSummaries)),
		fmt.Sprintf("verify after edits: %s", verifyCmd),
	}
	if report.ReferenceMigration {
		snapshot = append(snapshot, "reference migration: enabled")
	}
	if report.VerificationMigration {
		snapshot = append(snapshot, "verification migration: enabled")
	}
	if report.NeedsRollup {
		snapshot = append(snapshot, "also refresh `PROJECT_PROGRESS_SUMMARY.md`")
	} else {
		snapshot = append(snapshot, "refresh `PROJECT_PROGRESS_SUMMARY.md` only if it changes")
	}

	issueBullets := make([]string, 0, len(fileSummaries))
	for _, summary := range fileSummaries {
		issueBullets = append(issueBullets, fmt.Sprintf(
			"`%s`: %s",
			filepath.Base(summary.Path),
			strings.Join(limitStrings(summary.Issues, issueLimitForScope(report)), "; "),
		))
	}

	searchBullets := make([]string, 0, len(categorySummaries)+1)
	for _, category := range categorySummaries {
		searchBullets = append(
			searchBullets,
			fmt.Sprintf("%s: %s", category.Name, strings.Join(wrapCode(limitStrings(category.SearchHints, 2)), "; ")),
		)
	}
	if hasInstructionFileFinding(report.Findings) {
		searchBullets = append(searchBullets, fmt.Sprintf(
			"instruction files: `%s`",
			reconcileInstructionShortcut(report.ProjectRoot),
		))
	}

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf("Reconcile Kit-managed docs for the %s.", scope))
		if report.ReferenceMigration {
			doc.Paragraph("Migration target: replace deprecated front matter `dependencies` with canonical graph-like `references` and keep the prompt/context surface pointer-only.")
		}
		if report.VerificationMigration {
			doc.Paragraph("Migration target: add executable verification fields to active task details where checks are known, while keeping legacy feature docs compatible.")
		}
		doc.Paragraph("Rules:")
		doc.BulletList(rules...)
		doc.Paragraph("Audit snapshot:")
		doc.BulletList(snapshot...)
		doc.Paragraph("Files to fix:")
		doc.Table([]string{"Severity", "File", "Issues", "Focus"}, rows)
		doc.Paragraph("Notable issues:")
		doc.BulletList(issueBullets...)
		doc.Paragraph("Search shortcuts:")
		doc.BulletList(searchBullets...)
		doc.Paragraph("Reply with exactly these sections:")
		doc.BulletList(
			"`Findings`: bullets for what was stale",
			"`Updates`: bullets for what changed; include unresolved questions only if any remain",
			"`Verification`: bullets for commands run and whether they passed",
		)
	})
}

func summarizeReconcileFiles(findings []reconcileFinding) []reconcileFileSummary {
	summaries := make(map[string]*reconcileFileSummary)
	order := make([]string, 0, len(findings))

	for _, finding := range findings {
		summary, ok := summaries[finding.FilePath]
		if !ok {
			summary = &reconcileFileSummary{Path: finding.FilePath}
			summaries[finding.FilePath] = summary
			order = append(order, finding.FilePath)
		}

		if finding.Severity == reconcileSeverityError {
			summary.ErrorCount++
		} else {
			summary.WarnCount++
		}
		summary.Issues = appendUniqueString(summary.Issues, finding.Issue)
		summary.Actions = appendUniqueString(summary.Actions, shortActionForFinding(finding))
		for _, hint := range finding.SearchHints {
			summary.SearchHints = appendUniqueString(summary.SearchHints, hint)
		}
	}

	result := make([]reconcileFileSummary, 0, len(summaries))
	for _, path := range order {
		result = append(result, *summaries[path])
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].ErrorCount != result[j].ErrorCount {
			return result[i].ErrorCount > result[j].ErrorCount
		}
		if result[i].WarnCount != result[j].WarnCount {
			return result[i].WarnCount > result[j].WarnCount
		}
		return result[i].Path < result[j].Path
	})

	return result
}

func summarizeReconcileCategories(findings []reconcileFinding) []reconcileCategorySummary {
	grouped := map[string][]string{}
	order := []string{}

	for _, finding := range findings {
		category := reconcileFindingCategory(finding)
		if _, ok := grouped[category]; !ok {
			order = append(order, category)
		}
		for _, hint := range finding.SearchHints {
			grouped[category] = appendUniqueString(grouped[category], hint)
		}
	}

	result := make([]reconcileCategorySummary, 0, len(order))
	for _, category := range order {
		result = append(result, reconcileCategorySummary{
			Name:        category,
			SearchHints: grouped[category],
		})
	}

	return result
}
