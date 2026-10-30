---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0094"
  slug: "local-runtime-lifecycle"
  dir: "0094-local-runtime-lifecycle"
---
# SPEC

## PURPOSE

Stop coding agents from filling the human's machine with duplicate local environments (GH-270). After an overnight agent run the human found over 20 instances of a local environment open in 20 Warp terminals, with the GPU pinned at 100%.

The human's words: "ensure warp/rungrid instances are cleaned up before starting new ones."

## CONTEXT

- The overnight run's post-merge step was `rungrid down` then `rungrid up`, repeated for each merged pull request.
- In rungrid's source, `up` calls `warp.Open` whenever opening is enabled, and the first tab URI carries `new_window=true`. This happens even when `up` reuses a running runtime. `down` stops services and does not touch Warp. Each restart therefore leaves one full window of tabs, and each tab holds a live terminal session.
- `rungrid up --no-open` and `--headless` open nothing.
- `work-cleanup` said to stop what the work started, but only when work ends. No rule covered a start, and agents read contextual rules only when a trigger applies.

## REQUIREMENTS

- Before any start, inspect what already runs for the project, then reuse it or stop it through its manager. Never run two runtimes of one project.
- Restart only the affected service; do not restart to check health.
- Agents open no terminal windows or tabs for a local environment unless the human asks.
- Verify teardown, and stop what the work started when it ends.
- Never stop an instance that another session or the human is using, never kill by a broad pattern, and never close a window the agent did not open.
- Non-goal: changing rungrid, or reconciling downstream repositories in this change.

## ACCEPTED PLAN

1. Ship a default (`downstream`, conditional) `local-runtime-lifecycle` ruleset. `kit init` installs it; `kit reconcile` adds it.
2. Add an always-loaded contract bullet under Engineering and a Contextual Rules route.
3. Regenerate this repository's three managed instruction files and list the rule in the references index.
4. Test init and reconcile installation across fixture generations, idempotence, and rule invariants.

## DECISIONS

- A new rule, not a section of `work-cleanup` or `testing-and-environment-validation`: the trigger is any start, which is neither work ending nor validation.
- The contract carries the core limits because the failure happened in routine steps, where an agent does not think of "local environment" as a trigger to load a rule.
- The rule names rungrid commands as examples of a general duty, because rungrid is the manager that caused this and the commands must be exact. Compose stacks and dev servers fall under the same duties.
- `--no-open` over `--headless` as the agent default: `--headless` rewrites the terminal mode in the effective manifest, while `--no-open` changes only whether a window opens.
- The rule tells agents to reuse or report a shared instance rather than always stop it first. The human's request was cleanup before a new start, but one session's stop would break another session's validation or the human's own terminals. A blocking stale instance is still stopped first.

## DISCOVERIES

- The root behavior is in rungrid: `up` opens a window for a runtime that is already running, and `down` leaves windows open. This rule limits agents only; a human's repeated `rungrid up` still stacks windows until rungrid changes.

## VALIDATION

`gofmt -l .`, `make vet`, `make lint`, `make test`, `make build`, and `kit check --project`. `TestLocalRuntimeLifecycleInstallsByDefaultAcrossGenerations` covers init, reconcile, and idempotence.

## OUTCOME

Shipped as a default rule. Downstream repositories get it on their next `kit reconcile` after a Kit release.

## REPOSITORY MEMORY

- Created: this spec, the `local-runtime-lifecycle` rule. Updated: universal contract, references index.
