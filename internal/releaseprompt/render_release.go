package releaseprompt

import (
	"github.com/jamesonstone/kit/v3/internal/promptdoc"
)

func addReleaseAndInfrastructure(document *promptdoc.Document, config Config) {
	document.Heading(2, "Phase 3: Merge and Deploy by Graph")
	document.Paragraph("Before each wave, reconcile the standing-authority selector and pause state, resolved exact current PR set, expected heads and bases, authenticated actor, repository merge policy, readiness evidence, dependencies, deployment workflow and environment, and material effects. Merge only the matching `MERGE_READY` frontier, not a static PR list. Determine deployment waves separately from merge waves.")
	document.Paragraph("Default release progression:")
	document.CodeBlock("text", `release unit
→ merge through the repository's approved delivery path
→ post-merge source and artifact validation
→ verify runtime and infrastructure prerequisites
→ deploy the exact intended artifact
→ verify rollout, runtime health, and feature behavior
→ checkpoint evidence and update the graph
→ advance only now-ready dependent nodes`)
	document.Paragraph("A release unit may be one PR, one tightly coupled dependency wave, or multiple proven-independent changes. Independent authorized `MERGE_READY` nodes may merge concurrently when repository policy and same-base behavior permit it; serialize dependency chains and same-base sensitive operations. Do not expose an invalid intermediate production state merely to preserve PR-by-PR serialization.")
	document.Paragraph("Do not ask for renewed authorization solely because a later in-scope PR number or refreshed head was unknown when standing authority was granted. Stop for selector mismatch, pause or revocation, or material expansion. Default to one potentially interacting production release unit at a time and advance dependents only after observed verification.")
	document.Paragraph("Verify the authenticated GitHub actor before every repository boundary and confirm the post-merge actor. Accept only the expected human or explicitly authorized service identity; never substitute another profile silently. One repository's identity failure blocks that node and its dependents without erasing verified progress elsewhere.")

	document.Heading(3, "Infrastructure Policy")
	document.Paragraph("Mode: `" + inline(config.Infrastructure.Mode) + "`")
	document.Paragraph("Provider: `" + inline(config.Infrastructure.Provider) + "`")
	document.Paragraph("CLI or source of truth: `" + inline(config.Infrastructure.CLI) + "`")
	document.Paragraph("Identity verification:")
	document.CodeBlock("text", safeCode(config.Infrastructure.IdentityCheck))
	document.Paragraph("Policy:")
	document.CodeBlock("text", safeCode(config.Infrastructure.Policy))
	if config.Infrastructure.Mode == "none" {
		document.Paragraph("Infrastructure mutation is outside this release. Record a newly discovered prerequisite as a blocker or separately authorized graph node rather than changing infrastructure implicitly.")
	} else {
		document.Paragraph("Standing deployment authority covers only a named existing standard workflow for an authorized environment and exact merged artifact on already-provisioned resources. It includes rollout and runtime verification, not novel provider commands or workflow mutation.")
		document.Paragraph("IAM, network, KMS, secrets, persistent data-store or database-schema change, data loss, cluster control-plane change, infrastructure creation, replacement, or deletion, and nonstandard deployment effects require their own complete approval boundary. Unresolved classification fails closed.")
		document.Heading(4, "Covered Infrastructure Changes")
		document.OrderedList(1,
			"Verify provider identity and exact environment with the strongest repository-local mechanism.",
			"Inspect actual current state and produce the complete target, action, impact, recovery, and validation outline.",
			"Obtain the applicable explicit approval; standing merge/deploy authority never substitutes for it.",
			"Apply only the approved batch in one pass with bounded diagnosed retries.",
			"Re-read provider state, verify application integration, and checkpoint exact evidence.",
		)
		document.Heading(4, "Destructive Or Data-Loss Changes")
		document.Paragraph("Stop the affected mutation boundary. Report the exact resource and operation, reason, blast radius, availability and data-loss risk, security/IAM and cost impact, recovery implications, and post-change validation. Continue unrelated safe graph work while waiting when dependencies permit it.")
	}
}

func addVerificationAndFailure(document *promptdoc.Document, config Config) {
	document.Heading(2, "Production and Environment Verification")
	document.Paragraph("Deployment context:")
	document.CodeBlock("text", safeCode(config.DeploymentContext))
	document.Paragraph("Target environment: `" + inline(config.Production.Environment) + "`")
	document.Paragraph("Verification mechanism:")
	document.CodeBlock("text", safeCode(config.Production.Verification))
	document.Paragraph("When verification is `auto`, discover the strongest established repository or operator mechanism before deployment. Prefer configured commands and scripts over invented probes. Repository inventory hints are candidates, not proof; inspect their documented ownership and prerequisites without exposing secrets.")
	document.Paragraph("For every release unit, establish actual runtime state using the strongest applicable signals:")
	document.BulletList(
		"the expected commit, artifact, image, revision, or version is the one running",
		"rollout, process, container, task, pod, dependency, and health-endpoint state is stable",
		"logs, errors, alarms, metrics, rate limits, and saturation show no material regression",
		"authentication, authorization, APIs, UI, workflows, events, webhooks, persistence, and feature-specific behavior are correct",
		"migration, data integrity, compatibility, rollback readiness, and cleanup obligations are satisfied",
	)
	document.Paragraph("Do not advance dependent nodes until the release unit is demonstrably healthy. Record unobservable signals as `UNRESOLVED`, not passing.")

	document.Heading(2, "Failure and PR Remediation Loop")
	document.Paragraph("Stop advancement of an affected dependency chain whenever correctness, compatibility, data integrity, target identity, or runtime health is uncertain. Diagnose before retrying; do not assume a later PR will repair an earlier failure.")
	document.Paragraph("When one node in a wave fails, preserve and report successful independent nodes, classify the failed node and its dependents as `BLOCKED` or `UNKNOWN`, reconcile the remaining authorized frontier, and continue only work whose authority and independence remain proven. Report partial waves literally.")
	document.Paragraph("If source changes are required:")
	document.OrderedList(1,
		"Implement the smallest correction in the existing PR's owned delivery lane when it remains within that issue and declared scope.",
		"Under standing authority that includes blocker repair, ordinarily merge the current base into the existing head branch, apply or regenerate the correction, commit, and push to the same branch without rebasing, force-pushing, or retargeting.",
		"Mark the changed PR `UNKNOWN`, invalidate old-head readiness, rerun every required current-head gate, then restore `MERGE_READY` without renewed authority when the selector still matches.",
		"Create a reviewed replacement PR and first-class graph node only when scope, architecture, head safety, repository policy, or the user requires it; revalidate a selector-matching replacement without a new authorization prompt.",
		"Recompute dependencies, the reconciled authorized frontier, merge waves, deployment waves, and the critical path.",
		"Pass the same readiness, review, compatibility, infrastructure, and validation gates.",
		"Merge, deploy, verify, checkpoint evidence, and update the graph.",
	)
	document.Paragraph("Prefer roll-forward when safer. Roll back only when the previous state is demonstrably compatible and lower risk. Infrastructure rollback remains subject to the same infrastructure approval and identity boundaries.")

	document.Heading(2, "Final Integrated-System Verification")
	document.Paragraph("Configured integration suite:")
	document.CodeBlock("text", safeCode(config.IntegrationSuite))
	document.Paragraph("When configured as `auto`, discover and use the established cross-system suite or operator process. When `none`, record integrated verification as explicitly not applicable and explain why individual release-unit evidence is sufficient. Never replace an established suite with a weaker custom smoke test.")
	document.Paragraph("Run the suite after all intended release units pass their individual environment verification. If it fails, diagnose the root cause, remediate existing PRs in place when routine and scope-preserving, create exceptional replacement nodes only when required, recompute the graph, process corrections through every gate, and rerun until it passes or a genuine manual blocker remains.")
}
