package cli

func setupCapabilityRecords() []capabilityRecord {
	return []capabilityRecord{
		capability("init", "Agent Workflow", "Initialize or refresh Kit project scaffolding and coding-agent routing.", mutationWritesFiles,
			withNetwork("fetches the configured rules registry and may query GitHub repository visibility"),
			withFileWrites("creates project scaffolding, local registry rules, agent instructions, and repository-local workflows", "--dry-run previews refreshes; existing project-owned content is preserved unless an explicit managed replacement is selected"),
			withFlags(flag("--refresh", "refresh Kit-managed artifacts"), flag("--dry-run", "preview refreshes without writes", "read-only"), flag("--diff", "show a dry-run diff", "read-only"), flag("--file", "limit refresh to selected managed paths"), flag("--force", "replace selected managed content", "review local changes first")),
			withRelated(related("context resolve", "loads the materialized repository contract"), related("reconcile", "audits and curates drift")),
			withExamples("kit init", "kit init --refresh --dry-run --diff")),
	}
}
