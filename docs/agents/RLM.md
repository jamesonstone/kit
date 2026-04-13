# RLM

## Purpose

- RLM is Kit's repository-scale context-routing pattern: discover broadly, narrow to the smallest relevant context, then synthesize from sourced reads
- Use this pattern for repository-scale analysis, broad audits, or tasks that span many files or services
- Use RLM when the task is broad enough that loading the whole repo context would be noisy or wasteful
- The goal is progressive disclosure: load only the relevant subset of repo knowledge instead of flooding context

## Trigger Signals

- codebase-wide analysis
- scan repository
- audit all integrations
- many files or services
- high uncertainty about where the relevant logic lives

## Execution Pattern

1. index candidate docs, files, skills, and references
2. filter to the smallest set likely to matter
3. map bounded reads or file-scoped workers across the filtered set
4. reduce those results into a sourced synthesis

## Rules

- Keep map work file-scoped or narrowly bounded so synthesis stays deterministic
- Prefer repo-local docs before secondary global inputs
- Treat RLM as discovery and context selection first; do not jump straight into parallel execution while the candidate set is still broad
- Always update affected documentation and ensure touched documents stay current and properly formatted before finishing the work
- Record the docs, skills, and references that materially shaped the feature in dependency tables
- Use `kit dispatch` only when the work moves from broad discovery into multi-lane execution planning
