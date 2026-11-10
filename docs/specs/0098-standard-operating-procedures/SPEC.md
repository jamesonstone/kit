---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0098"
  slug: "standard-operating-procedures"
  dir: "0098-standard-operating-procedures"
---
# SPEC

## PURPOSE

Introduce maintained standard operating procedures that coding agents can execute for authorized recurring work. Keep operating instructions separate from feature history and general policy.

## CONTEXT

The user clarified that the LLM/coding agent is the execution engine, with workbook sync in LabCore as a concrete example. Existing contextual rules already cover operating constraints, but projects need a stable SOP location, scaffold, and instructions for agent use. Issue GH-280 delivers these foundations; GH-281 delivers read-only project discovery; GH-282 separately adds an agent functionality guide.

## REQUIREMENTS

- Store project-owned procedures at `docs/sops/<slug>/SOP.md`.
- Init and reconcile scaffold `docs/sops/README.md` with a reusable format and preserve existing content.
- Define draft, active, and retired procedures with applicability, ownership, prerequisites, ordered steps, verification, and recovery.
- Ship contextual guidance for creating, reading, suggesting, executing, and maintaining SOPs.
- Preserve existing permission boundaries. An SOP does not authorize actions.
- Keep discovery/listing commands in a separate PR. No process execution engine or production operations.

## ACCEPTED PLAN

1. Add an embedded SOP guide/template and scaffold it using the existing instruction-artifact pipeline.
2. Add the core standard-operating-procedures ruleset and route relevant tasks to it from the canonical contract.
3. Update current user documentation and generated contract copies.
4. Test init/reconcile creation, preservation, dry-run behavior, and rule distribution; run full validation and fresh-context review.
5. Summarize before opening a ready-for-review PR. Implement project discovery independently in the follow-up.

## DECISIONS

- Use SOP as the canonical name. A separate workflow format or alias would duplicate the same content.
- Agents create and maintain procedure Markdown from the scaffold; no creation command is needed for this first PR.
- The guide is create-once project content, like the testing reference. The shipped contextual rule owns reusable agent behavior.
- The original proposed workflow CLI, formal workflow metadata validator, and automatic lifecycle are superseded by the user’s narrower first-PR scope.
- Public examples use generic workbook synchronization, not private endpoints, datasets, or production commands.
- The follow-up discovery command can read conventional Markdown paths without a code dependency, allowing separate PRs against the default branch.
- Keep spec number 0098. The slug is `standard-operating-procedures`, not the superseded `active-workflows` name. 0097 remains the shell-environment spec on the earlier branch.

## DISCOVERIES

The universal contract is at its 1500-word limit. The SOP route stays in that one contract. Room came from redundant wording, not from paraphrasing policy: the contextual-rules intro repeated the opening "only when triggered" rule; "ready pull request" was already required by the delivery sentence; "in that order" repeats "the first existing"; the optional-rules sentence keeps the same obligation. Repair kinds, including ordered merges, stay in the contract. The `ENG-123` example stays. "State `repository memory: not required` once" no longer also says "in the final response." Headroom remains for the shell-environment route and the guide route. A reconcile preview showed unrelated registry, README badge, and Mint pin updates; those are not part of this change.

## VALIDATION

Before implementation, baseline `go test ./... -count=1`, `go vet ./...`, `golangci-lint run ./...`, and formatting passed. The original change passed `gofmt -l .`, `make vet`, `make lint` (0 issues), `make test`, `make build`, and `KIT_USAGE_DISABLED=1 ./bin/kit check --project`, with the existing nonblocking `docs/references/README.md` advisory. Focused SOP tests cover migration generations, idempotency, preservation including force, targeted reconcile, dry-run, symlink protection, missing-guide diagnostics, and template sections.

Review repair, 2026-11-10: `gofmt -l .` printed nothing; `go vet ./...` passed; `go test ./... -count=1` passed (`pkg/cli` 34.603s), including `TestUniversalContractStaysSmall` (1485 words); `golangci-lint run ./...` reported 0 issues; `make build` passed; `KIT_USAGE_DISABLED=1 ./bin/kit check --project` exited 0 with the same nonblocking references advisory. Hosted CI was not yet rerun for this repair commit. No live workbook sync, production validation, or guarantee of agent compliance was tested.

## OUTCOME

Implemented the project-owned SOP guide/template, create-once init/reconcile integration, missing-guide diagnostics, core contextual rule, and shared contract routing. Generated entry files match the canonical contract. The contract keeps the SOP route under 1500 words by deleting redundant wording, not by paraphrasing policy. The spec directory is `0098-standard-operating-procedures`. The implementation adds no execution engine or command. LabCore inspection remained read-only; only a synthetic workbook outline appears in the public template.

## REPOSITORY MEMORY

This spec preserves the scope change and rejected alternatives. Update the current overview, command documentation, Constitution memory definitions, and SOP guide. Preserve historical specs and existing procedures.
