package cli

func setupCapabilityRecords() []capabilityRecord {
	return []capabilityRecord{
		capability(
			"init",
			"Setup",
			"Initialize or refresh Kit project scaffolding in a repository.",
			mutationWritesFiles,
			withNetwork(
				"fetches the Kit ruleset registry from GitHub during default initialization and refresh",
				"--refresh and --dry-run --refresh read the registry for refresh planning; README badge planning may query gh repo visibility when a GitHub repository is configured or discoverable",
			),
			withFileWrites(
				"writes .kit.yaml, Kit project files, README.md managed status badges and final Maintainers section, docs, instruction files, registry rulesets, and .github/workflows/auto-assign.yml",
				"--dry-run previews refresh changes without writing files; auto-assignment workflow assignees come from local github.default_assignees, then global ~/.config/kit/.kit.yaml, and no-op safely when none are configured; README badge URLs come from local github.repository or the origin GitHub remote; private repositories skip public Shields GitHub metadata badges and keep only native GitHub Actions workflow badges when a conventional workflow exists; README Maintainers uses the managed Jameson/jamesonstone attribution",
			),
			withFlags(
				flag("--refresh", "refresh Kit-managed files and backfill or upgrade generated .kit.yaml settings such as loop.agent.command, README.md managed badges and Maintainers section, and managed GitHub auto-assignment workflow content"),
				flag("--dry-run", "preview --refresh without writing files", "read-only"),
				flag("--diff", "print planned --refresh changes as a unified diff; requires --dry-run", "read-only"),
				flag("--file", "limit --refresh to one Kit-managed file; repeat for multiple files"),
				flag("--force", "replace existing generated files when supported", "review local edits first"),
			),
			withRelated(
				related("scaffold", "generates individual workflow artifacts"),
				related("reconcile", "preferred reviewed surface for managed-file refreshes and coding-agent follow-up prompts"),
				related("loop review", "uses the loop agent config seeded by managed-file refresh"),
				related("project refresh", "semantic Constitution refresh after the project matures"),
			),
			withWhenToUse(
				"Use once to initialize Kit scaffolding, instruction docs, registry rulesets, managed README badges, the final README Maintainers section, and the optional GitHub auto-assignment workflow in a repository.",
				"Use `kit reconcile` in an existing Kit project when you want the interactive include-files, force, and coding-agent prompt choices around managed-file refreshes.",
				"Use `kit init --refresh` as the lower-level compatibility path for direct managed-file refreshes.",
			),
			withWhenNotToUse(
				"Do not use for semantic project documentation refresh alone; use `kit project refresh`.",
				"Do not use `--force` until local generated-doc or workflow changes have been reviewed.",
				"Prefer `kit reconcile` when a human should choose whether to include files, force changes, or output the follow-up coding-agent prompt.",
			),
			withCaveats(
				"The generated Constitution starter is a valid bootstrap state; the init prompt leaves project-specific sections unchanged until repository evidence demonstrates durable project-wide truth.",
				"After writes, human output or the generated prompt instructs the agent to move version-control-eligible command-created files from a protected root checkout into the exact issue worktree and ready pull request without disturbing unrelated changes.",
			),
			withExamples("kit init", "kit reconcile", "kit reconcile --include-files --dry-run --diff", "kit init --refresh", "kit init --refresh --file=README.md", "kit init --refresh --file=.github/workflows/auto-assign.yml"),
		),
		capability("scaffold", "Setup", "Generate workflow artifacts for Kit features.", mutationWritesFiles, withFileWrites("writes generated docs under selected project paths"), withRelated(related("scaffold agents", "writes repo agent instructions"))),
		capability("scaffold agents", "Setup", "Generate or refresh repo-local agent instruction files.", mutationWritesFiles, withFileWrites("writes AGENTS.md and docs/agents guidance"), withFlags(flag("--force", "replace existing agent guidance files", "review local edits first")), withRelated(related("init", "creates the broader project structure")), withCaveats("After a write, completion guidance requires moving version-control-eligible command-created files from a protected root checkout into the exact issue worktree and ready pull request without disturbing unrelated changes.")),
	}
}
