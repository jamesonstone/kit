package cli

func inspectionCapabilityRecords() []capabilityRecord {
	return []capabilityRecord{
		capability("status", "Inspect & Repair", "Show feature and Kit-managed project state.", mutationNetwork, withNetwork("fetches registry state with a bounded timeout; registry failure is reported as unknown"), withFlags(flag("--json", "emit machine-readable status"), flag("--all", "show all feature state")), withRelated(related("reconcile", "repairs managed drift"))),
		capability("registry", "Inspect & Repair", "Inspect the configured Kit registry.", mutationNetwork, withRelated(related("registry status", "reports freshness and actions"))),
		capability("registry status", "Inspect & Repair", "Report registry and managed-file freshness.", mutationNetwork, withNetwork("fetches the configured rules registry unless managed health is disabled"), withFlags(flag("--json", "emit machine-readable status", "read-only")), withRelated(related("health", "applies safe maintenance"))),
		capability("health", "Inspect & Repair", "Apply safe managed updates and validate project health.", mutationWritesFiles, withNetwork("fetches the configured rules registry"), withFileWrites("applies conflict-free managed updates", "--dry-run and --diff do not write; custom and conflicting content is preserved"), withFlags(flag("--dry-run", "preview without writes", "read-only"), flag("--diff", "show dry-run diff", "read-only"), flag("--json", "emit machine-readable results")), withRelated(related("usage report", "weekly maintenance reads aggregate usage once"), related("reconcile", "curates unresolved drift"))),
		capability("capabilities", "Inspect & Repair", "Describe exact supported commands, side effects, and safety behavior.", mutationNone, withFlags(flag("--json", "emit machine-readable records"), flag("--full", "include full supported records"), flag("--search", "search supported records")), withRelated(related("context resolve", "selects local evidence after command choice"))),
		capability("usage", "Inspect & Repair", "Inspect and manage bounded private local command usage data.", mutationWritesFiles, withNetwork("none"), withFileWrites("ordinary supported commands append minimal events under ~/.config/kit/usage", "report/status are read-only; refresh/clear/toggles mutate only local usage state"), withFlags(flag("--since", "bound report history"), flag("--command", "filter by normalized command"), flag("--project", "filter by anonymized project"), flag("--json", "emit machine-readable results")), withCaveats("Usage commands do not record themselves; no arguments, values, paths, output, prompts, environment, URLs, or secrets are collected.")),
		capability("usage report", "Inspect & Repair", "Aggregate bounded local command usage.", mutationNone, withFlags(flag("--since", "report window such as 90d"), flag("--command", "command filter"), flag("--project", "anonymized project filter"), flag("--json", "emit JSON"))),
		capability("usage status", "Inspect & Repair", "Show effective collection settings, coverage, bounds, and diagnostics.", mutationNone, withFlags(flag("--json", "emit JSON"))),
		capability("usage refresh", "Inspect & Repair", "Validate, rotate, and prune bounded usage storage.", mutationWritesFiles, withFlags(flag("--dry-run", "preview maintenance", "read-only"), flag("--json", "emit JSON"))),
		capability("usage clear", "Inspect & Repair", "Clear selected local usage events.", mutationDestructive, withFlags(flag("--all", "clear all history"), flag("--before", "clear older history"), flag("--command", "clear one command"), flag("--project", "clear one anonymized project"), flag("--yes", "confirm non-interactively"))),
		capability("usage enable", "Inspect & Repair", "Enable local usage collection at project or global scope.", mutationWritesFiles),
		capability("usage disable", "Inspect & Repair", "Disable local usage collection at project or global scope.", mutationWritesFiles),
		capability("config", "Inspect & Repair", "Inspect Kit project configuration.", mutationWritesFiles, withRelated(related("config check", "validates and offers bounded repairs"))),
		capability("config check", "Inspect & Repair", "Validate .kit.yaml and offer safe bounded repairs.", mutationWritesFiles, withNetwork("none on a complete fast path", "interactive AWS remediation may list profiles and verify STS identity"), withFileWrites("interactive repairs may update schema and AWS fields", "--json is read-only"), withFlags(flag("--json", "validate without prompts or writes", "read-only"))),
		capability("aws", "Inspect & Repair", "Verify the project-bound AWS context.", mutationNetwork, withRelated(related("aws verify", "checks exact identity"))),
		capability("aws verify", "Inspect & Repair", "Verify configured AWS profile and account through STS.", mutationNetwork, withNetwork("calls aws sts get-caller-identity using the configured profile"), withFlags(flag("--json", "emit verified identity"))),
		capability("check", "Inspect & Repair", "Validate feature or whole-project Kit contracts.", mutationNone, withFlags(flag("--all", "check all features"), flag("--project", "check the project contract"))),
		capability("pr", "Inspect & Repair", "Generate pull-request repair and release-orchestration prompts.", mutationGit, withRelated(related("pr fix", "collects active feedback and prepares the repair lane"), related("pr orchestrate", "renders a dependency-aware release prompt"))),
		capability("pr fix", "Inspect & Repair", "Collect active PR feedback and produce a coding-agent repair prompt.", mutationGit, withNetwork("lists/fetches GitHub PRs and paginated active review threads"), withFileWrites("prompt-only by default", "may prepare the exact writable same-repository PR-head worktree"), withGitMutation("may fetch and create/attach the exact PR-head worktree; never edits source, stages, commits, pushes, comments, resolves, or merges"), withFlags(flag("--pr", "target URL, Markdown link, owner/repo#number, or current-repo number"), flag("--coderabbit", "filter to CodeRabbit"), flag("--edit", "edit collected tasks"), flag("--output-only", "print prompt"), flag("--max-subagents", "default 3, hard ceiling 4")), withRelated(related("context resolve", "loads pr-feedback-repair evidence"), related("dispatch", "provides explicit thread resolution"))),
		capability("pr orchestrate", "Inspect & Repair", "Resolve bounded repository scope into a release-orchestration prompt.", mutationNetwork,
			withNetwork("none when local Git metadata is sufficient", "may run one cached targeted gh repo view per repository when identity or default-branch evidence is missing"),
			withFileWrites("none; dry-run and normal generation do not write repositories"),
			withGitMutation("none; Kit does not enumerate, edit, merge, deploy, or mutate release state"),
			withFlags(flag("--repos", "repeat exact repository paths"), flag("--root", "scan only the root and immediate child repositories"), flag("--scope", "strict, related, or organization expansion"), flag("--infra", "auto, none, direct, iac, mixed, or custom"), flag("--verify", "auto or a tagged verification mechanism"), flag("--integration-suite", "auto, tagged mechanism, or none"), flag("--dry-run", "emit resolved YAML provenance and prompt without clipboard access", "read-only"), flag("--output-only", "print raw prompt"), flag("--copy", "copy in addition to output")),
			withRelated(related("context resolve", "loads release-orchestration and pull-request-merge evidence"), related("pr fix", "repairs review feedback before readiness"), related("dispatch", "supports bounded execution topology")),
			withWhenToUse("Use when a coding agent should construct and execute an authorized dependency-aware multi-repository release plan."),
			withWhenNotToUse("Do not use it as a release executor; Kit generates instructions but does not enumerate PRs, merge, deploy, mutate infrastructure, or launch an agent."),
			withExamples("kit pr orchestrate --repos ./service-a --repos ./service-b --verify auto --dry-run"),
			withCaveats("Only filename-level clues and sanitized repository metadata are discovered; arguments, paths, and prompt contents are excluded from usage telemetry.")),
		capability("improve", "Inspect & Repair", "Run Kit's deterministic benchmark harness.", mutationExecutesCommands, withRelated(related("improve run", "runs a benchmark suite"))),
		capability("improve run", "Inspect & Repair", "Run a deterministic Kit benchmark suite in disposable fixtures.", mutationExecutesCommands, withFileWrites("writes .kit/improve run evidence", "--dry-run does not write"), withFlags(flag("--suite", "select suite"), flag("--kit-binary", "evaluate an exact binary"), flag("--dry-run", "plan without writes", "read-only"), flag("--json", "emit manifest"))),
		capability("rules", "Inspect & Repair", "Manage durable repository-local rulesets.", mutationWritesFiles),
		capability("rules add", "Inspect & Repair", "Import or create a ruleset.", mutationWritesFiles),
		capability("rules list", "Inspect & Repair", "List local and registry rulesets.", mutationNetwork),
		capability("rules view", "Inspect & Repair", "Inspect a local or registry ruleset.", mutationNetwork),
		capability("rules link", "Inspect & Repair", "Link a ruleset through canonical feature references.", mutationWritesFiles),
		reconcileCapabilityRecord(),
	}
}

func reconcileCapabilityRecord() capabilityRecord {
	return capability("reconcile", "Inspect & Repair", "Reconcile managed files, rules, and documentation drift.", mutationWritesFiles,
		withNetwork("fetches the registry only when managed files are included"),
		withFileWrites("default and --prompt-only generate guidance without project writes", "--include-files applies existing conflict-aware refresh behavior; --dry-run previews"),
		withGitMutation("none"),
		withFlags(flag("--include-files", "include managed refresh"), flag("--all", "whole-project audit"), flag("--force", "force selected managed refresh"), flag("--dry-run", "preview without writes", "read-only"), flag("--diff", "show dry-run diff", "read-only"), flag("--file", "limit managed paths"), flag("--output-only", "print prompt"), flag("--copy", "copy prompt"), flag("--migrate-references", "include legacy reference migration guidance"), flag("--migrate-verification", "include legacy verification guidance"), flag("--prompt-only", "generate guidance without mutation", "read-only")),
		withRelated(related("health", "applies safe scheduled maintenance"), related("context resolve", "loads repository-maintenance evidence")),
		withWhenToUse("Use whole-project reconcile to audit every version-control-eligible handwritten source and test file against the exact 300-physical-line limit."),
		withCaveats(
			"The existing conflict-aware reconcile behavior and flags are preserved.",
			"V3 whole-project reconciliation checks `AGENTS.md` for the ordered Codex pre-response thread-title and thread-pin gate, including fail-visible first-commentary semantics.",
			"Whole-project output emits literal `source-file-size audit: complete` evidence with candidate, eligible-file, and violation counts; missing or incomplete evidence cannot support a clean result.",
		))
}
