package releaseprompt

import (
	"github.com/jamesonstone/kit/v3/internal/promptdoc"
)

func addDiscoveryAndGraph(document *promptdoc.Document, config Config) {
	document.Heading(2, "Phase 1: Discover the Release Set")
	document.Paragraph("For every repository in scope, inspect local status and ownership, remotes and target branches, current open PRs, relevant code and history, linked issues, CI and review state, unresolved discussions, conflicts, runtime and deployment relationships, and cross-repository dependencies. Establish the complete permitted release set before mutation.")
	document.Paragraph("Prefer local `git` for equivalent evidence. Use `" + inline(config.SourceControl.CLI) + "` for authoritative remote state. Treat API capacity as finite: batch repository-level JSON fields, cache stable results, avoid repeated unchanged metadata, avoid short polling, and do not enumerate an organization without evidence that scope expansion requires it.")

	document.Heading(3, "PR Inventory")
	document.Paragraph("For every relevant PR, determine:")
	document.BulletList(
		"repository, purpose, feature stream, source and target branches, commits, changed files, linked issues, and linked/dependent PRs",
		"actual implementation behavior verified against code rather than metadata alone",
		"CI/check state, review decision, unresolved human and automated feedback, and merge conflicts",
		"API, event, message, schema, database, configuration, infrastructure, and security impact",
		"runtime dependencies, deployment ownership, environment implications, compatibility window, and rollback constraints",
	)

	document.Heading(3, "Mandatory Global Release Graph")
	document.Paragraph("Construct a directed Global Release Graph. It is the release control plane and must be updated whenever evidence changes a dependency assumption.")
	document.Paragraph("Model:")
	document.BulletList(
		"PR and corrective-PR nodes grouped into feature streams and repository boundaries",
		"authorization source, exact authorized PR set, expected head and base, authenticated actor, merge method, repository policy, and post-merge gate",
		"infrastructure, schema, API, contract, producer/consumer, frontend/backend, deployment, and validation prerequisites",
		"conflicts, shared mutation surfaces, compatibility windows, convergence points, safe parallelism, and failure-containment boundaries",
		"merge waves, separate deployment waves, the critical path, and the first safe release unit",
	)
	document.Paragraph("Use explicit edge types where useful:")
	document.CodeBlock("text", `REQUIRES
MUST_PRECEDE
ENABLES
RESOURCE_BEFORE_APPLICATION
SCHEMA_BEFORE_APPLICATION
BACKEND_BEFORE_FRONTEND
CONTRACT_BEFORE_CONSUMER
CONSUMER_TOLERANCE_BEFORE_PRODUCER
COMPATIBILITY_WINDOW
CONFLICTS_WITH
SHARES_MUTATION_SURFACE_WITH
INDEPENDENT_OF`)
	document.Paragraph("Give every material dependency edge a reason. Do not force whole feature streams into a false linear order when only specific nodes depend on one another.")

	document.Heading(3, "Planning Gate")
	document.Paragraph("Do not mutate until the graph and evidence can answer:")
	document.OrderedList(1,
		"What are all relevant PRs and feature streams within the permitted scope?",
		"Which exact PRs are authorized to merge by the direct request or accepted bounded plan, and which discovered nodes require follow-up authorization?",
		"Which nodes depend on which others, and why?",
		"What runtime, infrastructure, schema, migration, contract, and approval prerequisites exist?",
		"For each authorized node, what are the expected head/base, authenticated actor, merge method and policy, current readiness evidence, indirect deployment effects, and post-merge gate?",
		"Which work can safely proceed in parallel, and which files, systems, or production units must be serialized?",
		"What are the safest merge waves and separate deployment waves?",
		"What is the critical path and where do independent streams converge?",
		"What release unit executes first, what exact evidence makes it ready, and why is the ordering safe?",
	)
	document.Paragraph("If evidence is insufficient, investigate further. Do not mutate merely to answer an uncertainty that inspection can resolve.")
}

func addPreparationAndCompatibility(document *promptdoc.Document, config Config) {
	document.Heading(2, "Phase 2: Prepare PRs Safely")
	document.Paragraph("Prepare independent nodes concurrently only when the active host confirms separate execution, parallel scheduling, and current capacity. Keep one supervisor accountable, predict file overlap before delegation, let the host govern admission, and serialize overlapping implementation. Continue preparing independent future nodes while bounded checks or deployments run.")
	document.BulletList(
		"inspect code and dependencies",
		"classify and remediate review feedback",
		"resolve conflicts from intended integrated behavior",
		"implement required fixes and corrective tests",
		"investigate CI and environment failures",
		"analyze mixed-version, migration, infrastructure, and rollback safety",
		"prepare future graph nodes without advancing blocked dependencies",
	)

	document.Heading(3, "PR Readiness Gate")
	document.Paragraph("Classify every authorized candidate as `MERGE_READY`, `BLOCKED`, or `UNKNOWN`. Only `MERGE_READY` belongs in the reconciled merge frontier. A PR may reach `MERGE_READY` only when every applicable condition is satisfied:")
	document.BulletList(
		"implementation and intended behavior are correct against current code",
		"required tests, builds, lint, type checks, schemas, contracts, migrations, and hosted checks pass",
		"branch and target state are current; conflicts are intentionally resolved and the resulting integrated diff is reviewed",
		"all actionable human and automated feedback is addressed; incorrect and informational findings are recorded literally",
		"dependencies and runtime/infrastructure prerequisites exist or are safely scheduled in the graph",
		"mixed-version, backward/forward compatibility, data integrity, security, and rollback behavior are evaluated",
		"no unresolved correctness concern or ambiguous ownership remains",
	)
	document.Paragraph("Pending, missing, stale-head, policy-ineligible skipped, unattributed, or locally substituted required checks do not pass. Head drift, actor mismatch, unknown merge policy, or unknown indirect deployment/infrastructure effects produce `UNKNOWN` or `BLOCKED`, never `MERGE_READY`.")
	document.Paragraph("Configured review systems:")
	document.CodeBlock("text", safeCode(config.ReviewSystems))
	document.Paragraph("Configured required checks:")
	document.CodeBlock("text", safeCode(config.RequiredChecks))

	document.Heading(3, "Review Feedback")
	document.Paragraph("Classify each material finding as actionable, already addressed, incorrect, or informational/non-blocking. Verify it against current HEAD before changing code. Do not blindly apply automated suggestions. After remediation, rerun relevant validation, push through the repository's delivery contract, refresh hosted state efficiently, and reevaluate readiness.")

	document.Heading(3, "Merge Conflicts")
	document.Paragraph("Never mechanically choose one side. Understand both implementations and current target-branch behavior, preserve all required functionality and cross-project contracts, account for previously merged graph nodes, validate the result, and review the complete integrated diff. If an earlier merge invalidates a later PR assumption, correct the later implementation through its normal lane.")

	document.Heading(3, "Zero-Downtime and Compatibility")
	document.Paragraph("Assume rolling deployment and mixed-version operation unless the actual platform proves otherwise.")
	document.BulletList(
		"Prefer additive APIs and events; make new fields optional where appropriate and maintain compatibility windows for removals and renames.",
		"Deploy tolerant consumers before stricter producers when required; preserve idempotency, retry safety, duplicate-delivery safety, and compatible message evolution.",
		"Ensure old and new clients, producers, consumers, and service revisions can coexist throughout each deployment wave.",
		"Do not create an atomic multi-service upgrade requirement unless the platform provides and verifies that guarantee.",
	)
	document.Paragraph("Database and schema policy:")
	document.CodeBlock("text", safeCode(config.DatabaseMigrationPolicy))
	document.CodeBlock("text", `expand schema
→ deploy code compatible with old and new state
→ backfill or migrate
→ switch behavior
→ verify
→ remove legacy state in a later approved release`)
	document.Paragraph("Treat destructive or irreversible migrations as high risk and place their approval, recovery, and compatibility obligations explicitly in the graph.")
}
