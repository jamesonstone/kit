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
- explicit dependency link over broad search
- repo-local docs before global model/vendor instructions

## Rules

- Keep map work file-scoped or narrowly bounded so synthesis stays deterministic
- Prefer repo-local docs before secondary global inputs
- For feature-scoped work, keep must-read inputs small: the current `TASKS.md` entry plus the linked `PLAN.md` and `SPEC.md` sections
- Use indices first: start with `kit map <feature>` and `docs/PROJECT_PROGRESS_SUMMARY.md` to shortlist candidate prior features under `docs/specs/`
- Treat prior feature docs, repo references, and secondary global inputs as conditional reads only
- Open a prior feature doc only when it affects a shared interface or contract, overlapping files or modules, migrations or data shape, acceptance criteria, or an explicit relationship or dependency link
- Inspect at most 5 prior feature directories before narrowing further or asking a clarifying question
- Extract only the concrete facts that change the current feature; do not paraphrase entire prior docs into chat or copy irrelevant history into the active artifact
- Treat RLM as discovery and context selection first; do not jump straight into parallel execution while the candidate set is still broad
- Always update affected documentation and ensure touched documents stay current and properly formatted before finishing the work
- Record the docs, skills, and references that materially shaped the feature in dependency tables
- Use `kit dispatch` only when the work moves from broad discovery into multi-lane execution planning
