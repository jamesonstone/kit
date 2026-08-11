package releaseprompt

import (
	"strings"

	"github.com/jamesonstone/kit/internal/promptdoc"
)

func addCompletionAndReport(document *promptdoc.Document, config Config) {
	document.Heading(2, "Completion Gate")
	document.Paragraph("Complete only when all applicable statements are true:")
	document.BulletList(
		"all required repositories and relevant permitted PRs were inspected and represented in the final graph",
		"the direct request or accepted plan, exact authorized PR set, actor, heads/bases, merge method, policy, and final reconciled frontier are recorded",
		"all intended authorized safe PRs and authorized corrective PRs are merged, deployed to their required environment, and verified",
		"all actionable feedback and conflicts are resolved, and integrated diffs and required validation pass",
		"runtime, infrastructure, schema, migration, contract, security, and compatibility prerequisites are satisfied",
		"no prohibited or unapproved mutation occurred and every destructive proposal is approved, rejected, or explicitly blocked",
		"expected revisions are running, rollout and dependencies are healthy, and feature-specific verification passes",
		"the configured integrated-system verification passes or is explicitly and correctly not applicable",
		"no relevant PR remains open without an explicit reason, owner, impact, and unblock condition",
		"program and repository checkpoints are current when the cross-repository coordination trigger applies",
	)

	document.Heading(2, "Final Report")
	document.Paragraph("Describe what actually happened, not merely the initial plan. Use exact identities and evidence timestamps where they affect readiness or completion.")
	document.Heading(3, "Executive Summary")
	document.BulletList(
		"overall release, production verification, and integrated verification result",
		"authorization source and bounded PR set; repositories affected; PRs evaluated, `MERGE_READY`, merged, corrected, blocked, unknown, intentionally left open, or newly discovered",
		"infrastructure inspected and changed; destructive changes proposed, approved, rejected, or blocked",
		"major deviations, hidden dependencies, graph reorderings, compatibility findings, and residual risks",
		"total files changed and baseline-to-final lines added, removed, and changed without double-counting rebases or corrective iterations",
	)

	document.Heading(3, "Final Global Release Graph")
	document.Paragraph("Show the graph as executed: original, discovered, and corrective nodes; authorization state; dependency reasons; feature streams; reconciled merge frontiers; partial, complete, and blocked merge/deployment waves; convergence points; critical path; and changes from the initial graph.")

	document.Heading(3, "PR Results")
	document.Paragraph("For every relevant PR report repository, identifier, authorization state, readiness state, expected and observed head/base, authenticated merge actor, merge method, purpose, feature stream, final state, material remediation, merge revision, deployed artifact or revision, target environment, verification result, and exact remaining blocker when open.")

	document.Heading(3, "Infrastructure")
	document.Paragraph("Report exact infrastructure inspected, target identity, approved mutations performed, rationale, resulting state and verification, rollback readiness, and every destructive or replacing change proposed or blocked.")

	document.Heading(3, "Deviations and Findings")
	document.Paragraph("Include hidden dependencies, conflicts, additional or corrective PRs, infrastructure discrepancies, runtime dependencies, production regressions, compatibility or migration findings, approval changes, and graph reorderings.")

	document.Heading(3, "Correctness Overview")
	document.Paragraph("Assess every important subsystem, graph gate, environment, and integration concern using only:")
	document.CodeBlock("text", `VERIFIED
INFERRED
NOT_APPLICABLE
UNRESOLVED`)
	correctness := "Do not claim correctness beyond observed evidence."
	if requirements := strings.TrimSpace(config.FinalReportRequirements); requirements != "" {
		correctness += " " + inline(requirements)
	}
	document.Paragraph(correctness)

	document.Heading(3, "Remaining Risks")
	document.Paragraph("List only genuine unresolved risks with the issue, impact, reason it remains unresolved, accountable owner when known, exact unblock condition, and recommended follow-up.")

	document.Heading(2, "Expected Lifecycle")
	document.CodeBlock("text", `discover
→ resolve repository-local evidence
→ construct Global Release Graph
→ prepare and remediate independent nodes
→ satisfy readiness gates
→ merge by graph
→ deploy by separate graph waves
→ verify actual runtime state
→ checkpoint and update graph
→ run final integrated verification
→ process corrective PR loop when needed
→ reconcile final evidence
→ issue final report`)
	document.Paragraph("A partially completed safe release is preferable to a fully completed unsafe release.")
}
