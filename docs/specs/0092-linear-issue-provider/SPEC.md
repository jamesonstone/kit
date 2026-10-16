---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "implement"
feature:
  id: "0092"
  slug: "linear-issue-provider"
  dir: "0092-linear-issue-provider"
---
# SPEC

## PURPOSE

Select GitHub Issues (default) or Linear per repository without changing GitHub pull request delivery.

## CONTEXT

Kit delegates ticket writes to the native agent. Canonical rules are embedded and reconciled; configuration is YAML. Mint separately renders commit history.

## REQUIREMENTS

Preserve GitHub defaults and history. Use Linear identifiers as scopes and full URLs as commit footers. Never query or publish private issue contents; no live test tickets.

## ACCEPTED PLAN

Add validated repository-only configuration and init selection; update canonical contract and rules; test persistence, switching and delivered policy.

## DECISIONS

Use existing agent capabilities for ticket writes. No new network client, credential store, or automatic private content enrichment. Two directly coupled PRs require no program ledger.

## DISCOVERIES

Kit has no ticket-creation command; issue payload and lifecycle are canonical agent policy. Mint has no provider API dependency.

## VALIDATION

Focused init/default selection, persisted configuration, invalid target/provider, switching, exact-template migration, custom-template preservation, badge and embedded policy tests pass. Full Go tests, vet, build, lint and project checks run before PR delivery. Ticket auth/errors/duplicates/payload/lifecycle are agent instructions verified by contract fixtures, not executable provider integration tests. No live Linear writes or credentials were used. Independent review found an ordinary-reconcile template switching gap; fixed exact shipped templates and added regression coverage. Project check has an existing retired edited README compatibility warning.

## OUTCOME

Implemented for review; no merge or release authorized. Merge Mint first, then Kit; adopt a Mint version containing Linear-link parsing in consumer workflows before activating Linear. There is no Go API dependency or required dependency version bump between these PRs.

## REPOSITORY MEMORY

Feature spec and public configuration/changelog documentation. Historical specs remain unchanged.

## DELIVERY EVIDENCE

Issue: [GH-264](https://github.com/jamesonstone/kit/issues/264). Branch: `GH-264`.
Ready PR: [Kit #265](https://github.com/jamesonstone/kit/pull/265).
Companion: [Mint #33](https://github.com/jamesonstone/mint/pull/33).
Local full tests, vet, build, formatting, lint and project checks passed.
Independent review finding resolved and re-reviewed. Hosted checks and review
were pending when this record was written; the live PR is authoritative.
No merge, release, credential configuration or Linear writes performed.

## PROVIDER PARITY AUDIT

At base 7472eda, delivery.md instructs native agents to find/create human-assigned
GitHub issues and populate original ask, scope, acceptance criteria and verification;
work-cleanup.md instructs agent closure after observed completion. No ticket creation
API or kit issue command exists in cmd/internal/pkg. rules.go embeds the canonical
corpus, rules_registry_embedded.go loads it, and init/reconcile install the rules.
Linear uses this same complete mechanism, with repository target configuration and
provider-specific instructions for Markdown description, human assignment, duplicate
search/readback, failures, branch/reference/commit formatting and verified states.

Configuration and template behavior have executable Go tests. Ticket creation,
population, auth/error recovery, duplicate prevention and lifecycle writes depend
on the host's authorized capability and the agent following shipped instructions.
Their tests are contract fixtures, not API execution tests. Kit does not install
credentials, create server-side integrations or automatically map workflow states.
No background ticket transition service or automatic closure-on-merge was added.
The README now makes these operational limits explicit. No live Linear mutation
was attempted; such an execution test needs an authorized precise destination.

Independent continuation review confirmed these mechanism and scope claims without
new defects. CodeRabbit completed successfully on `1b54c0f` at 2026-10-16
22:01 UTC, reporting no actionable comments. CI passed on that source head.
