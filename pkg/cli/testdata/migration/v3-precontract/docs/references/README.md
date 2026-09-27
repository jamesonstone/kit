# References

## Purpose

- This directory holds durable repo-local references that are broader than one feature
- Keep long-lived background context here instead of in injected top-level instruction files
- Link these files from feature front matter references when they materially shape work
- Store durable rulesets under `rules/<slug>.md` and link them with `kit rules link` instead of copying rules into agent instruction files
- Use `rules/agent-completion-output.md` before substantial terminal completion or handoff responses; it requires no response format and instead names the facts a response must not leave out
- Use `rules/work-lane-gating.md` before any coding-agent repository
  file or delivery mutation to default to a new worklane without asking,
  require the pull-request landing plan, preserve exact existing-PR lifecycle
  work in place, and keep the primary checkout read-only
- Use `rules/kit-capabilities-usage.md` in downstream projects for Kit command discovery guidance
- Use `rules/constitution-curation.md` after implementation and validation to keep the Constitution aligned with demonstrated project-wide truth
- Use `rules/cross-repository-program-coordination.md` before implementing or resuming accepted plans that span multiple repositories with dependent deliverables, staged deployment or activation, or expected handoff
- Use `rules/human-authorship.md` before any commit, pull request, issue, review comment, or other attribution text so only the human user is displayed as author
- Use `rules/github-pr-merge.md` and resolve `pull-request-merge` before any merge or merge-queue mutation; explicit bounded standing authority may bind later in-scope PRs and refreshed heads, every resolved exact current node must be `MERGE_READY`, and its SHA/head OID is evidence only, never an authorization identity
- Use `rules/deletion-safety.md` before designing deletion behavior or deleting persistent project, user, business, or external-system state to require recoverable soft delete by default and exact manual confirmation before hard delete
- Use `rules/slack-read-only.md` when Slack is in scope so access stays read-only by default and every Slack send or other mutation requires explicit, message-specific human approval
- Use `rules/infrastructure-change-approval.md` before mutating public-cloud resources, Kubernetes resources or cluster state, or infrastructure-as-code source, configuration, or state; standing authority covers only named standard deployments on already-provisioned targets, while IAM, network, KMS, secrets, database schema/data-loss, infrastructure create/replace/delete, destructive, and nonstandard deployment effects retain their own approval boundaries
- Use `rules/aws-agent-toolkit-guidance.md` before AWS-dependent work to select current Agent Toolkit skills, official documentation, the AWS MCP Server or CLI fallback, verified identity, infrastructure approval, and secret-safe handling
- Use `rules/testing-and-environment-validation.md` before implementation and validation, including browser automation and browser testing, to preserve code-level checks, browser lifecycle ownership, and environment evidence safely
- Use `rules/deadline-mode.md` only when the user explicitly signals a real deadline or time constraint, to narrow testing and implementation scope without weakening required approvals, security, or migration/compatibility invariants
- Use `rules/source-file-size.md` before editing implementation/source or test files and for whole-project reconcile audits
- Use `worktrees.md` for the canonical native Git worktree hierarchy, naming, shared-state model, environment ownership, and safety contract
- Use `kit rules add` to import or activate available registry rulesets from the Kit GitHub `main` branch
- Use `kit rules view <slug>` to preview a local or registry ruleset before importing it
- Use `kit init --refresh` to adopt existing registry rules into `.kit.yaml` registry state and pick up safe upstream ruleset updates
- Use `kit rules add --custom` for the interactive `$EDITOR` ruleset builder

- Use `rules/coding-agent-context-usage.md` for the capability, resolution, loading, and re-resolution sequence
- Store declarative coding-agent workflow contracts under `workflows/<slug>.md`

## Starter Files

- `testing.md` — repo-wide testing norms and evidence expectations
- `tooling.md` — local tooling and command references that are broader than one feature
- `external-systems.md` — durable notes about external systems, APIs, or integrations
- `rules/` — pointer-loaded durable rulesets such as frontend UI rules, testing rules, API conventions, security constraints, or domain rules
