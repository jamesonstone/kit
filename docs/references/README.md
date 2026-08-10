# References

## Purpose

- This directory holds durable repo-local references that are broader than one feature
- Keep long-lived background context here instead of in injected top-level instruction files
- Link these files from feature front matter references when they materially shape work
- Store durable rulesets under `rules/<slug>.md` and link them with `kit rules link` instead of copying rules into agent instruction files
- Use `rules/kit-capabilities-usage.md` in downstream projects for Kit command discovery guidance
- Use `rules/feature-notes.md` when deciding how to load, reference, promote, or ignore source material under `docs/notes/<feature>`
- Use `rules/constitution-curation.md` after implementation and validation to keep the Constitution aligned with demonstrated project-wide truth
- Use `rules/cross-repository-program-coordination.md` before implementing or resuming accepted plans that span multiple repositories with dependent deliverables, staged deployment or activation, or expected handoff
- Use `rules/infrastructure-change-approval.md` before mutating public-cloud resources, Kubernetes resources or cluster state, or infrastructure-as-code source, configuration, or state to require one plan-level confirmation per batch, one-pass execution, and explicit confirmation for deletion or removal
- Use `rules/testing-and-environment-validation.md` before implementation and validation, including browser automation and browser testing, to preserve code-level checks, browser lifecycle ownership, and environment evidence safely
- Use `rules/source-file-size.md` before editing implementation/source or test files and for whole-project reconcile audits
- Use `rules/codex-thread-initialization.md` to preserve Codex's ordered pre-response rename and pin gate during instruction refresh and reconciliation
- Use `worktrees.md` for the canonical native Git worktree hierarchy, naming, shared-state model, environment ownership, and safety contract
- Use `kit rules add` to import or activate available registry rulesets from the Kit GitHub `main` branch
- Use `kit rules view <slug>` to preview a local or registry ruleset before importing it
- Use `kit init --refresh` to adopt existing registry rules into `.kit.yaml` registry state and pick up safe upstream ruleset updates
- Use `kit rules add --custom` for the interactive `$EDITOR` ruleset builder
- `kit rule` is the singular alias for `kit rules`

## Starter Files

- `testing.md` — repo-wide testing norms and evidence expectations
- `tooling.md` — local tooling and command references that are broader than one feature
- `external-systems.md` — durable notes about external systems, APIs, or integrations
- `rules/` — pointer-loaded durable rulesets such as frontend UI rules, testing rules, API conventions, security constraints, or domain rules
- `../notes/<feature>/` — optional feature source material; not canonical truth and private contents remain ignored

## Ruleset Index

Rulesets are loaded just in time according to their `read_policy` and
`applies_to` metadata. The managed downstream rules currently available here
are:

| Ruleset | Scope | Purpose |
| --- | --- | --- |
| `agent-team-orchestration` | coding-agent, workflow, dispatch, subagent, verification | Accountable supervisor plans, bounded specialist lanes, overlap control, and read-only verification. |
| `backend-service-architecture` | architecture, backend, API, service, repository, gateway | Responsibility boundaries for routes, controllers, services, repositories, and persistence adapters. |
| `codex-thread-initialization` | codex, coding-agent, session, thread, session-management | Ordered pre-response thread renaming and pinning with verified or fail-visible status. |
| `constitution-curation` | implementation, validation, repository-memory, constitution, project-refresh | Evidence-based promotion of durable rationale and project-wide invariants. |
| `cross-repository-program-coordination` | coding-agent, workflow, cross-repository, program, deployment, handoff, resume, dispatch | Coordinator-owned ledger, dependency frontier, exact evidence, checkpoints, reconciliation, and handoff for multi-repository programs. |
| `feature-notes` | notes, source-material, documentation | Optional feature source material and promotion boundaries for `docs/notes/<feature>`. |
| `frontend-application-architecture` | architecture, frontend, route, page, component, state | Responsibility and dependency boundaries for frontend routes, features, data adapters, state, and UI. |
| `github-pr-delivery` | git, GitHub, pull-request, documentation | Issue-to-PR delivery sequencing and post-PR verification. |
| `infrastructure-change-approval` | cloud, infrastructure-as-code, AWS, GCP, Azure, Kubernetes, Terraform, Pulumi, CloudFormation | One plan-level confirmation and one-pass execution per infrastructure batch, with explicit confirmation for deletion or removal. |
| `kit-capabilities-usage` | Kit command discovery in downstream projects | Targeted, read-only capability lookup without maintaining Kit's internal catalog downstream. |
| `llms-txt` | web, website, API, documentation | `/llms.txt` contract for applicable public web and API surfaces. |
| `readme-header-tagline` | README and repository onboarding | Consistent top-level README identity and opening structure. |
| `safety-guardrails` | git, GitHub, safety | Recon, identity, worktree, secret, protected-branch, and failure-recovery boundaries. |
| `source-file-size` | implementation, testing, validation, refactor, reconcile, maintenance | Exact 300-line handwritten source/test limit, exclusions, semantic splits, and verification. |
| `testing-and-environment-validation` | implementation, testing, validation, CI, deployment, local, production, browser automation, browser testing | Code-level PR checks, high-level environment suites, browser lifecycle ownership, immutable evidence, status reporting, and safe production validation. |
| `work-lane-gating` | git, GitHub, workflow | Separates non-implementation documentation work from implementation delivery lanes. |

`command-capabilities` is a Kit-maintainer-only local ruleset. It requires
changes to `kit capabilities` metadata when Kit command behavior changes and is
not installed as a downstream project rule.
