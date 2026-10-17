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

## AUTHORIZED CLI EXTENSION PLAN

Add `kit config set --issue-provider github|linear` with the init target flags,
noninteractive validation, `--dry-run` and JSON output. Patch only issue fields
in the YAML node document, preserving unrelated values/comments. Use the canonical
reconcile target resolver to protect primary checkouts; report the target and
explicit `kit reconcile --dry-run --diff`/`kit reconcile` next actions rather than
reimplement migration or silently applying unrelated repairs. Validate before
writes and use atomic replacement. A no-op stays a no-op.

Add `kit mint check` and explicit `kit mint update` using Mint's published official
Go module install mechanism (latest release inspected: v0.8.0, no binary assets).
Report locally resolved executable, repository action pins and Go dependency
separately; never imply updating one updates others. Updates are constrained to
verified user-owned Go-bin installations and staged/verified before replacement.
No production/labcore configuration or globally installed tools will be changed
in implementation tests. Linear support remains unreleased until a published tag
containing the companion change is verified; no invented minimum version.

## CLI EXTENSION OUTCOME AND VALIDATION

Implemented `kit config set --issue-provider github|linear`, `--dry-run`, `--json`
and shared init target flags. Settings are patched atomically with unrelated YAML
preserved; aliases/anchors/root merges that could alter unrelated semantics fail
before writes. The command reuses reconcile target safety and reports explicit
mandatory next_actions for canonical migration rather than duplicating it.

Implemented `kit mint check|update`, published `--version`, numeric local
`--require-version`, explicit `--require-linear`, JSON reports on operational
errors, and update dry-run using the same plan validation without installation.
Official published tags and Go proxy/checksum verification constrain staged Go
installs; only supported user-owned Go-bin targets update. Local executable and
repo module/action pins are separate; pins remain inventory/not_verified.
The verified Linear minimum stays unset until a real release proves support.
No Mint source/PR extension was needed for these Kit-owned commands.

Observed: complete Go suite, vet, build, formatting, lint (0 issues), project
check and targeted race tests passed. Project check retains its existing edited
retired README compatibility advisory. New tests cover provider transitions,
repeatability, optional-project clearing, config preservation, invalid/no partial
writes, YAML alias/merge rejection, canonical delivery target, JSON failures,
module/action inventory, missing/current/old/development/unreleased Mint versions,
unsupported installs, downgrade refusal, staging/build failures and stale-version
concurrency. A live read-only check reported Mint missing on this session's PATH,
latest v0.8.0 and five repository pins; updated=false and no pins changed.
No live installation/update, credentials changes, ticket writes or production
operations were performed; updater tests use isolated fixtures.

Independent review found and verified fixes for single-line Go dependencies,
YAML alias/root-merge preservation, early-error JSON output and concurrent update
no-downgrade guarantees. Final review confirmed the remaining findings resolved.
Hosted checks/review on the extension are recorded on PR #265 and remain the
source of current evidence. Both PRs must remain open and unmerged for review.

Automatic review on 5e6a422 identified two minor contract gaps: encoding changed
input indentation, and absent/error pin inventory serialized as null. The patcher
now infers block indentation from YAML source positions (four-space fallback),
with two/four-space regression tests. Report serialization normalizes missing pins
to an empty array for non-project and error paths. Focused tests and lint passed;
final hosted check/review evidence remains on the live PR.

## AUTHORIZED AUTOMATIC WORKFLOW INTEGRATION

Init plans Mint integration before writing any scaffolds. Empty workflow sets get
a complete release-state publisher pinned to an official published Mint commit;
no runtime, registry, cloud or deployment adapters are invented. Existing Mint
workflows retain their pins and publisher ownership. One clear release/deploy
workflow with one ordinary steps job gets a version/install step only; existing
release commands, permissions, triggers and deployment behavior stay byte-for-byte.
CI-only, multiple candidate or reusable workflows get a separate read-only Mint
validation workflow and a selection note. Explicit workflow/job selectors support
a reviewed target without guessing production. Reconcile uses the same planner
with --mint-workflows and its existing --dry-run --diff delivery boundary.

Mint #33 merged with human authorization as eac91490f772adbd2ab030cb6eba3910c1bca93c.
The normal release workflow published v0.9.0; its fetched annotated tag peels to
that commit and contains pkg/changelog/linear.go. Thus v0.9.0 is now the verified
Linear minimum, independent of local installation or existing repository pins.

## AUTOMATIC WORKFLOW VALIDATION

Added planning and YAML-preservation tests for empty repositories, init delivery
inventory, explicit workflow/job selection, single existing deployment, CI-only,
multiple candidates/jobs, reusable/dynamic workflows, existing Mint, repeat and
force reconciliation, dry-run/no writes, symlinks, malformed workflow and remote
default-branch selection. Workflow batches recheck snapshots before any writes,
so a concurrent edit fails without overwriting it or partially writing the batch.
Published capability tests cover local Linear minimum success and upgrade to
v0.9.0, and rejection below that verified release.

Independent read-only review found no actionable issues. The generated workflow
has not been executed in a consumer repository. Without remote default-branch
metadata its main starter may need a reviewed branch-filter adjustment; the
runtime default-branch guard prevents publication from the wrong branch. Existing
CI-only or ambiguous repositories gain read-only tooling, not inferred deployment.
Mint's normal post-merge Release and Documentation runs both succeeded for
merge eac91490f772adbd2ab030cb6eba3910c1bca93c; no manual release was started.

## AUTOMATIC REVIEW CORRECTIONS

CodeRabbit found that appending trusted Mint code to a deployment job would share
its ambient OIDC, environment and write credentials. Automatic selection now
requires explicit read-only permissions and no environment or ambient job/workflow
env; unknown inherited defaults and privileged jobs use the standalone read-only
check. Both workflow/job selectors provide the explicit opt-in for shared ambient
credentials. Regression cases cover writes, OIDC, production environments, ambient
secrets, unknown defaults and write-all.

The workflow snapshot check had also changed generic managed symlink handling.
Strict snapshot/path verification is now marked only on Mint-planned changes;
generic managed workflow behavior remains unchanged, with a regression test that
updates its linked target while strict Mint plans still reject that same link.

Independent follow-up review found prior credential-export steps could still
leave ambient credentials for the appended step. Automatic adoption additionally
rejects secret-context references, GITHUB_ENV exports and authentication/credential
actions in existing steps; these require explicit selection. Tests cover AWS
credential setup, bracket-form secret references and environment exports.

The follow-up automated review identified a validation-to-write interval inside
multi-file batches. Mint workflow writes now receive the original snapshot at
the individual write boundary; existing files reuse the repository's staged
conditional config replacement with byte and file-identity rechecks. New files
stage completely and publish via an exclusive hard link, so an appearing target
is never overwritten and failed writes/closes leave no partial destination.
Rollback also checks the expected just-written snapshot before touching a Mint
workflow. Deterministic tests inject edits/appearances during earlier batch writes,
write/close failures, a target appearing during staging, and edits before rollback.
The tests preserve human edits, roll back earlier files and retain file modes.

The next review correctly distinguished snapshot checks from a filesystem CAS.
All Kit Mint writes and rollbacks now share a per-target exclusive-create lock
held across validation, staging and replacement; existing locks fail closed and
are not reclaimed. Portable path replacement cannot conditionally compare bytes
against an arbitrary editor. These locks coordinate cooperating Kit operations;
external edits are checked at validation boundaries, with a residual final
check-to-replace window. This limit is documented rather than claiming absolute
CAS. Newly created Mint workflows are always retained and reported on rollback
instead of using unsafe check-then-unlink. Rollback is therefore best effort.
Tests cover lock contention for writes/rollback, release, existing-lock retention
and retained complete workflows with other batch files rolled back.

The existing optional controller scaffold now marks newly created controllers for
the same strict writer/lock protocol. Existing controllers still return skipped
changes and retain their original bytes and pins. This closes the review's scope
caveat without changing rendering or controller behavior.
