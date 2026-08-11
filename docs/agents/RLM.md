# RLM

## Purpose

- RLM is Kit's just-in-time context-routing pattern
- Use it for any task where loading full context would be noisy or wasteful
- The goal is progressive disclosure: load only the smallest relevant subset of repo knowledge needed for the immediate decision

## Trigger Signals

- codebase-wide analysis
- scan repository
- audit all integrations
- many files or services
- high uncertainty about where the relevant logic lives
- feature work with many possible prior docs or references
- any request where broad upfront reading would slow correctness

## Coding Agent Contract

1. Run `kit context resolve --workflow <slug> --json` with relevant feature and path hints.
2. Load every required selected artifact before acting.
3. Treat blocked resolution as a hard evidence gap.
4. Rerun resolution after material scope changes.

## Runtime Loop

1. identify the immediate decision
2. load the smallest relevant artifact
3. extract only required facts
4. act if context is sufficient
5. recurse only when uncertainty remains
6. stop loading once the decision is supported

## Context Budget Rules

- specific section over full file
- current feature over all features
- explicit reference link over broad search
- repo-local docs before global model/vendor instructions

## Rules

- Keep map work file-scoped or narrowly bounded so synthesis stays deterministic
- Prefer repo-local docs before secondary global inputs
- For living-spec feature work, keep must-read inputs small: the current `SPEC.md` section or decision, plus directly linked references, relationships, rules, evidence, or historical staged artifacts only when they affect that decision
- Treat rulesets under `docs/references/rules/` as just-in-time context; load only the linked ruleset sections whose `read_policy` and `applies_to` match the current decision
- Load `docs/references/rules/backend-service-architecture.md` before implementing API or backend routes, controllers or handlers, application services, repositories, persistence adapters, or gateways
- Load `docs/references/rules/frontend-application-architecture.md` before implementing frontend routes or pages, feature orchestration, state flows, data adapters, or reusable components
- Load `docs/references/rules/testing-and-environment-validation.md` and `docs/references/testing.md` before implementation or validation, including browser automation and browser testing
- Load `docs/references/rules/infrastructure-change-approval.md` before planning or performing mutations to public-cloud resources, Kubernetes resources or cluster state, or infrastructure-as-code source, configuration, or state
- Load `docs/references/rules/github-pr-merge.md` and resolve `pull-request-merge` before any merge or merge-queue mutation
- Load `docs/references/rules/cross-repository-program-coordination.md` before implementing or resuming an accepted plan that spans multiple repositories with dependent deliverables, staged deployment or activation, or expected agent or session handoff
- Load `docs/references/rules/agent-team-orchestration.md` only when the immediate decision includes execution topology, subagent lanes, or read-only verification; do not load it for trivial single-lane tasks
- Use indices first: start with `docs/PROJECT_PROGRESS_SUMMARY.md` and explicit SPEC relationships to shortlist candidate prior features under `docs/specs/`
- Treat prior feature docs, repo references, and secondary global inputs as conditional reads only
- Do not load every ruleset by default; feature front matter references determine when a ruleset is must-read, conditional, evidence, or skipped
- Open a prior feature doc only when it affects a shared interface or contract, overlapping files or modules, migrations or data shape, acceptance criteria, or an explicit relationship or reference link
- Inspect at most 5 prior feature directories before narrowing further or asking a clarifying question
- Extract only the concrete facts that change the current feature; do not paraphrase entire prior docs into chat or copy irrelevant history into the active artifact
- Treat RLM as discovery and context selection first; do not jump straight into parallel execution while the candidate set is still broad
- Always update affected documentation and ensure touched documents stay current and properly formatted before finishing the work
- Record the docs, skills, and references that materially shaped the feature in canonical front matter references
- Use `kit dispatch` only after native planning has established a narrow implementation topology
