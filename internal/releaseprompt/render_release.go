package releaseprompt

import (
	"github.com/jamesonstone/kit/v3/internal/promptdoc"
)

func addReleaseAndInfrastructure(document *promptdoc.Document, config Config) {
	document.Heading(2, "Phase 3: Merge and Deploy by Graph")
	document.Paragraph("Before each wave, reconcile the accepted authority, exact authorized PR set, expected heads and bases, authenticated actor, repository merge policy, readiness evidence, dependencies, and indirect effects. Merge only the authorized `MERGE_READY` frontier, not a static PR list. Determine deployment waves separately from merge waves.")
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
	document.Paragraph("Do not ask for per-PR reconfirmation after the bounded plan is accepted. Reauthorize only a material expansion of the PR set, target, effects, or recovery boundary. Default to one potentially interacting production release unit at a time. Deploy concurrently only when independence is demonstrated by separate runtime, data, dependency, mutation, and rollback boundaries. Never leave the outcome of an interacting unit unknown while advancing a dependent node.")
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
		document.Paragraph("Before any public-cloud, Kubernetes, or infrastructure-as-code mutation, produce one consolidated batch outline containing the exact target identity, resource actions, execution boundary, material availability/data/security/cost impact, rollback or recovery, and validation evidence. A merge known to trigger such mutation belongs inside this boundary before merge. Obtain one explicit confirmation for the complete bounded batch unless the accepted merge plan already contains and authorizes the same complete non-deletion outline.")
		document.Paragraph("Deletion, destruction, removal, or replacement always requires explicit confirmation after the complete outline, even when general release authorization exists. A material change to identity, environment, region, cluster, resources, action type, impact, recovery, or validation is a follow-up batch requiring a new consolidated outline and confirmation.")
		document.Heading(4, "Additive Infrastructure Changes")
		document.OrderedList(1,
			"Verify provider identity and exact environment with the strongest repository-local mechanism.",
			"Inspect actual current state and confirm the resource or configuration does not already satisfy the requirement.",
			"Prove the operation is within the approved batch, additive, backward compatible, and the smallest safe change.",
			"Apply the approved batch in one pass with bounded diagnosed retries only.",
			"Re-read provider state, verify application integration, and checkpoint exact evidence.",
		)
		document.Heading(4, "Destructive or Replacing Changes")
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

	document.Heading(2, "Failure and Corrective PR Loop")
	document.Paragraph("Stop advancement of an affected dependency chain whenever correctness, compatibility, data integrity, target identity, or runtime health is uncertain. Diagnose before retrying; do not assume a later PR will repair an earlier failure.")
	document.Paragraph("When one node in a wave fails, preserve and report successful independent nodes, classify the failed node and its dependents as `BLOCKED` or `UNKNOWN`, reconcile the remaining authorized frontier, and continue only work whose authority and independence remain proven. Report partial waves literally.")
	document.Paragraph("If source changes are required:")
	document.OrderedList(1,
		"Implement the smallest correction in the repository's owned delivery lane.",
		"Create a normal reviewed corrective PR.",
		"Add it as a first-class node, determine whether the accepted bounded plan already authorizes it, and obtain follow-up authorization when it does not.",
		"Recompute dependencies, the reconciled authorized frontier, merge waves, deployment waves, and the critical path.",
		"Pass the same readiness, review, compatibility, infrastructure, and validation gates.",
		"Merge, deploy, verify, checkpoint evidence, and update the graph.",
	)
	document.Paragraph("Prefer roll-forward when safer. Roll back only when the previous state is demonstrably compatible and lower risk. Infrastructure rollback remains subject to the same infrastructure approval and identity boundaries.")

	document.Heading(2, "Final Integrated-System Verification")
	document.Paragraph("Configured integration suite:")
	document.CodeBlock("text", safeCode(config.IntegrationSuite))
	document.Paragraph("When configured as `auto`, discover and use the established cross-system suite or operator process. When `none`, record integrated verification as explicitly not applicable and explain why individual release-unit evidence is sufficient. Never replace an established suite with a weaker custom smoke test.")
	document.Paragraph("Run the suite after all intended release units pass their individual environment verification. If it fails, diagnose the root cause, create corrective PR nodes where required, recompute the graph, process corrections through every gate, and rerun until it passes or a genuine manual blocker remains.")
}
