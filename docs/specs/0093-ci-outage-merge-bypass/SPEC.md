---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0093"
  slug: "ci-outage-merge-bypass"
  dir: "0093-ci-outage-merge-bypass"
relationships:
  - type: builds_on
    target: 0076-standing-merge-deploy-authority
    note: Reuses bounded standing merge authority and MERGE_READY readiness; adds one evidence substitute for hosted checks that cannot run.
---
# SPEC

## PURPOSE

Stop a hosted CI outage from stalling every Kit-managed repository (GH-268). The human's GitHub Actions minutes ran out, so jobs failed to start with a billing error. The managed contract forbade merging without passing required checks and outranks chat, so agents could not continue even with the human's consent.

The human's words: "If there's a CI outage or CI 'runs out of minutes', it should not be a blocker. We should inform the user but allow for manual bypass. The first approval of bypass is considered enough to continue building."

## CONTEXT

- The universal contract (`internal/templates/universal_contract.md`) renders the managed block in `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md`. It required `MERGE_READY` with required checks passed and forbade bypassing required checks.
- `github-pr-merge` said local tests never replace required hosted checks and forbade admin overrides.
- Standing authority (spec 0076) already lets one bounded grant cover later in-scope pull requests and new heads.
- `docs/references/testing.md` is each project's documented validation contract.

## REQUIREMENTS

- Applies only when hosted CI cannot start or run (outage, exhausted minutes, billing). Never to a check that ran and failed.
- The agent tells the human plainly and offers a manual bypass. Hosted checks are reported as not run, never as passed.
- The first bypass approval is a standing grant for in-scope merges while the outage lasts. It stays inside existing merge authority: scoped human authorization and a pull request that is otherwise ready.
- The full local validation in `docs/references/testing.md` must pass on the exact head being merged.
- Each bypass merge gets a pull-request comment with commands, results, and head SHA; the final report lists them.
- When the outage ends: re-run CI on every bypassed pull request, fix failures in follow-ups, stop using the bypass.
- Non-goal: reconciling downstream repositories in this change.

## ACCEPTED PLAN

1. Ship a default (`downstream`, conditional) `ci-outage` ruleset with an Applies When trigger and an explicit Precedence section. `kit init` installs it; `kit reconcile` adds it to existing projects.
2. Add an always-loaded contract bullet under Approval Boundaries and a Contextual Rules route, and qualify the two contract lines that forbade skipping required checks so they name `ci-outage` as the exception.
3. Qualify the Authority and Readiness lines of `github-pr-merge` the same way.
4. Regenerate this repository's three managed instruction files and list the rule in the references index.
5. Test init and reconcile installation across fixture generations, idempotence, the wording that resolves the gate conflict, and rule invariants.

## DECISIONS

- The bypass replaces only hosted required checks that could not run. Every other `MERGE_READY` gate stays, so the existing readiness vocabulary does not change. Not-run hosted checks use the literal state `NOT_RUN`.
- The bypass approval is a standing grant, but it is not merge authority. This keeps the merge boundary from spec 0076 intact: the human's bypass approval removes the per-PR prompt the human found blocking, and the scoped merge authorization still decides what may merge.
- Admin overrides stay forbidden. If branch protection or a ruleset still enforces the checks, the pull request is `BLOCKED` with its local evidence, and the human merges manually. This is the literal "manual bypass" and keeps the agent from gaining a new privileged action.
- Slow, queued, or policy-skipped checks do not qualify. Only a provider failure to start or run (outage, minutes, billing) does, so the bypass cannot hide a failing test.
- Superseded (after the first PR head): the procedure lived in a `github-pr-merge` section. The human asked for a default ruleset that `kit init` and `kit reconcile` add to every Kit project, so the procedure moved to `ci-outage`. A reviewer found the gates and the bypass read as contradictory, so `ci-outage` states that it is an explicit exception to "never bypass required checks" and "merge only `MERGE_READY`", taking precedence only for hosted checks that could not run, and the contract and `github-pr-merge` name it at both gates.
- The contract carries the trigger and core limits because agents read contextual rules only when a trigger applies, and the old contract text alone blocked the merge.
- Rejected: per-PR bypass approval (the human named the first approval as enough); treating local validation as passing hosted checks (it would make reports untruthful).

## DISCOVERIES

not required

## VALIDATION

- `go test ./internal/templates/ ./pkg/cli -run "Invariant|Contract|Standing" -count=1`: pass.
- Full validation: see the pull request body for `gofmt -l .`, `make vet`, `make lint`, `make test`, `make build`, and `kit check --project` results.
- Hosted CI on the pull request: expected `NOT_RUN` because of the billing outage that motivated this change.

## OUTCOME

- Shipped `docs/references/rules/ci-outage.md` as a default rule, the contract bullet and route, the `github-pr-merge` exception wording, regenerated instruction files, and tests in `pkg/cli/ci_outage_registry_test.go` plus invariant markers.
- Downstream repositories receive the rule at their next `kit reconcile` with the release that carries it.
- Risk: an agent can misread a failed check as an outage. The rule names the evidence (the provider's start or billing error) and excludes failed, slow, queued, and skipped checks.

## REPOSITORY MEMORY

- Created `docs/references/rules/ci-outage.md`. Updated `internal/templates/universal_contract.md`, `AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md`, `docs/references/rules/github-pr-merge.md`, `docs/references/README.md`.
- `docs/references/testing.md` unchanged: Kit's own validation commands did not change.
