package cli

func githubDeliveryHardGateStep() string {
	return "**GitHub delivery hard gate**\n" +
		"- If the user asks to create or mutate an issue, branch, staging, commit, push, or pull request, stop before that mutation\n" +
		"- Treat the delivery request as a new mutation boundary even when implementation or reflection is already complete\n" +
		"- In a Kit-managed project, repo-local Kit delivery rules outrank global GitHub/plugin defaults\n" +
		"- Load `.kit.yaml`, `docs/agents/README.md`, `docs/agents/GUARDRAILS.md`, `docs/agents/TOOLING.md`, relevant `docs/references/rules/*`, and GitHub templates before delivery\n" +
		"- Re-run and report branch/status/staleness recon: `pwd`, `git status --short --branch`, `git remote -v`, current branch, default/base branch, active PRs for the current branch, matching issues, and git author/committer identity\n" +
		"- Resolve issue source, issue number/link, branch name, branch base, staging method, commit format, PR title format, PR template, draft/ready state, required checks, cross-repo dependencies, and unknowns/blockers\n" +
		"- Output the Delivery Contract before mutation; if any field is unknown, ambiguous, missing, or conflicts with generic defaults, stop and ask\n" +
		"- Do not use `codex/*` branches, ad hoc issue/PR bodies, draft PRs by default, bulk staging, generic commit messages, or PRs that omit the repo template unless the Kit contract or user explicitly overrides it"
}
