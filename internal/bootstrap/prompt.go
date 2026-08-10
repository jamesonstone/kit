package bootstrap

func repositoryBootstrapPrompt() string {
	return `Bootstrap this repository from verified evidence as a coding agent.

1. Start by running:

   kit contract resolve --workflow repository-bootstrap --json

2. Read the returned routing entrypoints, workflow, mandatory rules, and
   dependencies. Follow docs/agents/RLM.md: begin with routing and indices,
   then inspect only relevant manifests, build scripts, CI, tests,
   documentation, code boundaries, current specifications, history, and
   external-system evidence.
3. Preserve all project-owned content and secrets. Never read or reproduce
   .env values, grant direnv trust, or infer credentials, endpoints, owners,
   environments, or production state.
4. Populate docs/CONSTITUTION.md only with demonstrated durable project-wide
   truth. Leave project semantic sections unchanged when evidence is
   insufficient.
5. Populate docs/PROJECT_PROGRESS_SUMMARY.md from current specifications and
   repository history. Do not invent features; retain its valid empty state
   when no current feature evidence exists.
6. Populate docs/references/testing.md with actual safe, verified test, build,
   lint, typecheck, integration, and end-to-end commands and their environments.
   Record unavailable or unsafe validation literally.
7. Populate docs/references/tooling.md from real package, toolchain, and
   configuration evidence. Populate docs/references/external-systems.md only
   from verified recurring integrations.
8. Replace the Makefile help-only starter only when repository-native commands
   are proven. Keep recipes as thin wrappers, add no guessed or placeholder
   targets, and run every safe target you add.
9. Review README project identity plus only the bounded Kit badge and
   Maintainers sections; do not overwrite surrounding project prose.
10. Validate the resulting repository bootstrap under the resolved testing,
    safety, delivery, source-size, and repository-memory rules.

Report files changed and preserved, validation performed, unresolved evidence
gaps, and the Repository Memory decision with its rationale and artifacts.
Kit supplied deterministic starters and this contract; it did not infer
project truth, launch an agent, or supervise this work.
`
}
