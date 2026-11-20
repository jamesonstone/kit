---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0102"
  slug: "work-ledger"
  dir: "0102-work-ledger"
relationships:
  - type: builds_on
    target: 0092-linear-issue-provider
    note: Reuses the offline, credential-free provider pattern. Native agents perform the API work.
---
# SPEC

## PURPOSE

Let a repository optionally keep a matching work-ledger item in step with Kit's ticket lifecycle (GH-300). The ledger tool can be swapped. Plane is the first adapter. An unconfigured repository does nothing.

## CONTEXT

Kit already selects GitHub Issues or Linear in repository-local `.kit.yaml` and tells native agents how to write tickets. Spec 0092 keeps Kit offline: no ticket SDK, no stored credentials, and tests that pin shipped instructions rather than live API calls.

The lifecycle those agents already follow is `docs/references/rules/delivery.md`: create or reuse the ticket, open the pull request, and record completion only after merge. This feature adds an optional ledger hook at those same moments.

GH-298 adds a separate Work ledger pointer paragraph for longer-running work. That change is unmerged on PR #299. This feature does not edit that paragraph. It configures the ticket-lifecycle hook.

## REQUIREMENTS

- Default off. An omitted `work_ledger`, an omitted `provider`, or `provider: none` is a no-op.
- One swappable adapter selected by `work_ledger.provider`. Implement `plane`. Reserve `linear`, `jira`, and `notion` without inventing their calls.
- Plane configuration is non-secret: workspace URL slug, project identifier, optional project name, and state names. Omitted state names mean `Active`, `In review`, and `Done`.
- Lifecycle when configured: at start, find or create the matching item, link the issue, set Active, and apply `agent:<x>` (Grok Bots use `grokbot:<bot_name>`). When the pull request opens, link it and set In review. When it merges, set Done.
- Kit stores no credentials, adds no ledger SDK, and does not call the ledger API. The native agent uses its already-authorized connector, such as the Plane MCP.
- A missing connector or a ledger error is reported and does not by itself stop the ticket. A failed search is not an empty project.
- Do not enable the ledger in this repository's `.kit.yaml`.

## ACCEPTED PLAN

Add `work_ledger` to project config and validate it on load. Add an adapter interface whose implementations only describe work: a no-op adapter and a Plane adapter. `Plan` returns ordered steps for start, review, and done. `Description` states the offline contract for the configured target.

Extend the existing delivery lifecycle section with the hook. Point at it from the universal contract's existing issues bullet. Leave GitHub issue and pull-request form templates unchanged, because the hook is default-off and is not a form field.

Do not add a CLI setter, a network client, or a credential field.

## DECISIONS

The adapter is a description the native agent executes. A Plane SDK or a Kit background sync would store or proxy credentials, which spec 0092 already rejected for tickets.

`work_ledger.plane.project` is the project identifier (`WL`), not the display title. Optional `name` carries the title (`Work Ledger`) when a repository wants it in the description. The example target workspace `reasoning-machines` is documentation and a test fixture, not a compiled-in default.

Owner labels stay a rule, not a secret and not a stored identity. The agent supplies `agent:<x>` or `grokbot:<bot_name>`. The adapter rejects any other shape and does not invent a name.

Reserved provider names fail validation so a later adapter can occupy the same key without a silent no-op. Unknown names fail closed.

No `kit config set` flag. Humans edit `.kit.yaml`. Validation runs on load. A setter would need a second YAML patcher and is not required to swap the provider.

The #298 Work ledger paragraph stays untouched so this branch can merge beside PR #299. The hook lives in the issue-provider section of `delivery.md` and in one existing contract sentence.

`kit reconcile` was not applied. A dry run from this worktree would clear the configured auto-assign login, retarget release workflows at a newer Mint pin, and rewrite registry hashes outside this feature.

## DISCOVERIES

Running `kit reconcile --dry-run --diff` on a current checkout plans unrelated updates to `.kit.yaml`, `README.md`, auto-assign, and the Mint release workflows. Those updates are not part of GH-300.

PR #299 also changes the delivery registry hash. This branch changes that same hash because it edits `delivery.md` in a different section. The paragraphs do not overlap.

A merge reads `github-pr-merge` and then `work-cleanup`. Cleanup already sends Linear state changes back to `delivery`. The ledger Done step now has the same pointer, so cleanup does not leave a configured item Active. An unmerged close still does not set Done. There is no canceled ledger state.

## VALIDATION

From this worktree, after the cleanup pointer:

- `gofmt -l .` printed nothing.
- `make vet` (`go vet ./...`) passed.
- `make test` (`go test -v ./...`) passed, including the no-op, Plane lifecycle, reserved providers, owner labels, config round trip, delivery hash, cleanup hash, and embedded hook contract.
- `make build` passed on the first head. The follow-up changed rules and tests only.
- `KIT_USAGE_DISABLED=1 ./bin/kit check --project` passed. It reported one existing compatibility advisory for `docs/references/README.md`. That advisory does not block validation and was not changed here.
- `make lint` (`golangci-lint` v2.11.0) reported 0 issues.

No live Plane or GitHub ledger call was made. Adapter tests scan planned actions for credential-like fields.

## OUTCOME

GH-300 adds optional `work_ledger` configuration and an offline adapter. Omitted configuration is a no-op. `plane` describes find-or-create, issue link, Active, and the owner label at start; find, pull-request link, and In review when the pull request opens; find and Done on merge. `linear`, `jira`, and `notion` are reserved. Kit stores no credentials and does not call a ledger API. The delivery rule and the universal contract tell native agents to perform those steps with their own connector. `work-cleanup` points a configured ledger at that Done step after a satisfying merge. This repository's ledger stays off.

A fresh-context review of `abe9277` found no acceptance miss. The cleanup pointer is the one valid gap it named. Empty plan URLs stay allowed because the adapter describes the step. The agent supplies the real links. An unmerged close still has no ledger state. No live Plane project was called.

## REPOSITORY MEMORY

This spec is the feature boundary. `delivery.md` and the universal contract carry the lifecycle agents follow. The Go adapter is the tested description of that lifecycle. Issue and pull-request templates stay as they are. This repository's ledger stays off.
