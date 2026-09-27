# References

## Purpose

- This directory holds durable repo-local references that are broader than one feature; keep long-lived background here instead of in agent entry files
- `rules/<slug>.md` holds contextual rules; the universal contract's Contextual Rules section says when to read each one, and `kit rules list` is the live inventory
- `workflows/<slug>.md` holds declarative workflow evidence lists used by `kit context resolve`
- `worktrees.md` describes the native Git worktree hierarchy, naming, shared-state model, and environment links
- Link these files from feature front matter references when they materially shape work

## Rules Management

- Use `kit rules add` to import or activate registry rulesets and `kit rules view <slug>` to preview one
- Use `kit init --refresh` to adopt registry rules into `.kit.yaml` state and pick up safe upstream updates
- Use `kit rules add --custom` for the interactive `$EDITOR` ruleset builder

## Starter Files

- `testing.md` — the project's validation commands, suites, and evidence expectations
- `tooling.md` — local tooling and command references that are broader than one feature
- `external-systems.md` — durable notes about external systems, APIs, or integrations

## Ruleset Index

Rulesets are loaded just in time when the universal contract's trigger applies. The managed downstream rules currently available here
are:

| Ruleset | Scope | Purpose |
| --- | --- | --- |
| `agent-completion-output` | coding-agent, task, completion, implementation, research, diagnosis, planning, validation, review, operations, coordination, handoff | No required response format; names only the facts a terminal completion or handoff must not leave out. |
| `agent-team-orchestration` | coding-agent, workflow, dispatch, subagent, verification | Safe, integrated, truthfully reported delegation using the host's native subagents. |
| `aws-agent-toolkit-guidance` | coding-agent, AWS, AWS CLI, AWS MCP, Agent Toolkit, infrastructure, documentation, secrets | Current AWS skills and official documentation, MCP or CLI execution, verified identity, material targets, infrastructure approval, and secret-safe handling. |
| `backend-service-architecture` | architecture, backend, API, service, repository, gateway | Responsibility boundaries for routes, controllers, services, repositories, and persistence adapters. |
| `coding-agent-context-usage` | coding-agent, workflow, rules, context, evidence | Capability lookup, deterministic context resolution, required evidence loading, and re-resolution. |
| `constitution-curation` | implementation, validation, repository-memory, constitution | Evidence-based promotion of durable rationale and project-wide invariants. |
| `cross-repository-program-coordination` | coding-agent, workflow, cross-repository, program, deployment, handoff, resume, dispatch | Coordinator-owned ledger, dependency frontier, exact evidence, checkpoints, reconciliation, and handoff for multi-repository programs. |
| `deadline-mode` | coding-agent, workflow, testing, implementation, prioritization | Explicit, user-signaled, invariant-preserving narrowing of testing scope and implementation complexity under a real deadline. |
| `deletion-safety` | implementation, data, persistence, filesystem, identity, API, UI, automation, cleanup, retention, migration, operations, cloud, infrastructure | Recoverable soft delete by default and exact post-outline manual confirmation before hard delete. |
| `frontend-application-architecture` | architecture, frontend, route, page, component, state | Responsibility and dependency boundaries for frontend routes, features, data adapters, state, and UI. |
| `github-pr-delivery` | git, GitHub, pull-request, documentation | Issue-to-PR delivery sequencing and post-PR verification. |
| `github-pr-merge` | git, GitHub, pull-request, merge, merge-queue, cross-repository | Explicit bounded standing authority, exact current merge readiness, pause/revocation precedence, and literal wave evidence. |
| `human-authorship` | git, GitHub, commit, pull-request, issue, attribution | Human-only displayed authorship on commits, pull requests, issues, comments, trailers, and other credits. |
| `infrastructure-change-approval` | cloud, infrastructure-as-code, AWS, GCP, Azure, Kubernetes, Terraform, Pulumi, CloudFormation | Separate standing standard deployment authority from infrastructure, security, data, destructive, and nonstandard effects that keep their own approval boundaries. |
| `kit-capabilities-usage` | Kit command discovery in downstream projects | Targeted, read-only capability lookup without maintaining Kit's internal catalog downstream. |
| `llms-txt` | web, website, API, documentation | `/llms.txt` contract for applicable public web and API surfaces. |
| `readme-header-tagline` | README and repository onboarding | Consistent top-level README identity and opening structure. |
| `safety-guardrails` | git, GitHub, safety | Recon, identity, worktree, secret, protected-branch, and failure-recovery boundaries. |
| `slack-read-only` | slack, messaging, communication, coding-agent, automation, collaboration | Slack is read-only by default; every send or other Slack mutation requires explicit, message-specific human approval. |
| `source-file-size` | implementation, testing, validation, refactor, reconcile, maintenance | Exact 300-line handwritten source/test limit, exclusions, semantic splits, and verification. |
| `testing-and-environment-validation` | implementation, testing, validation, CI, deployment, local, production, browser automation, browser testing | Code-level PR checks, high-level environment suites, browser lifecycle ownership, immutable evidence, status reporting, and safe production validation. |
| `work-lane-gating` | git, GitHub, workflow | Defaults every coding-agent repository mutation to a new pull-request worklane without asking. |

`command-capabilities` is a Kit-maintainer-only local ruleset. It requires
changes to `kit capabilities` metadata when Kit command behavior changes and is
not installed as a downstream project rule.
