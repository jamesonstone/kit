---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0086"
  slug: "deployment-recovery"
  dir: "0086-deployment-recovery"
---
# SPEC

## PURPOSE

Help coding agents recover a stalled authorized deployment through existing
pipelines before changing workflows or declaring delivery blocked.

## CONTEXT

Kit embeds `docs/references/rules/` in its binary. Reconcile installs missing
core rules and preserves locally customized rules. The universal contract
routes contextual work and currently says to diagnose before retrying.

The motivating incident involved missing publication, a rejected manual start,
and a premature blocker report. Those are distinct events; the original
missing-trigger cause remains unknown. The shared rule must avoid project,
cloud-provider, credential, and incident-specific prescriptions.

Issue #239, branch GH-239, and the canonical linked worktree own delivery.
The primary checkout and downstream projects remain untouched. The human later authorized review repair and merge of all open Kit pull
requests. Release and downstream rollout remain outside this repair task.

## REQUIREMENTS

- Ship a core, conditionally loaded deployment recovery rule.
- Prefer bounded safe recovery, preserve source identity and pipeline order,
  and distinguish missing runs, failed runs, deployment, and acceptance.
- Preserve approval, authentication, and production safety boundaries; a
  blocked alternative must not stop independent authorized work.
- Fresh init and every supported migration install the rule and routing.
  Reconcile remains idempotent and preserves project-owned rule edits.
- Avoid new runtime automation, mandatory agent lifecycle records, fixed
  provider commands, or blanket retry authority.

## ACCEPTED PLAN

Add the rule with `registry_scope: downstream`, add its universal routing
entry, clarify generic recovery wording, and route post-merge recovery from
the merge rule. Regenerate only affected Kit-owned surfaces and registry state.
Use native integration tests to verify distribution and preservation. Review
the rule against incident scenarios, run project validation, and open a ready PR.

## DECISIONS

- Keep the entry contract short and put task-specific guidance in a contextual
  rule. This follows [OpenAI's harness guidance](https://openai.com/index/harness-engineering/)
  on repository maps and invariants rather than oversized instruction manuals,
  and [Anthropic's context guidance](https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents)
  on concise instructions that leave room for agent judgment.
- Classify state and retry safety before a bounded recovery attempt; do not
  require complete historical root-cause analysis before a safe retry.
- Reuse the existing embedded-registry mechanism; no new CLI behavior or
  deployment automation is needed. Projects adopt the rule through a Kit
  binary containing this change, not merely by merging this source PR.

## DISCOVERIES

Baseline `kit status` reports an unrelated README refresh. Use targeted
reconciliation so this lane does not absorb that change.

Core rules install based on scope metadata, so no production Go change or
registry enumeration is needed. Reconcile reports a pre-existing non-blocking
advisory for project-owned `docs/references/README.md`; it stays preserved.

## VALIDATION

Review repair adopts the current main branch, including the default ASD-STE100
rule. Added explanatory test comments for the review warning and the missing
core-rule entry in the references index. The underlying recovery rule and
permissions are unchanged. All required local checks passed again on the integrated head. The same
nonblocking references-index advisory remains.

- PASS: `make test`, including fresh/current project installation, every
  released-generation migration fixture, idempotence, and preservation and
  reporting of project-owned deployment recovery content.
- PASS: `make vet`, `make lint` (zero issues), `make build`, `gofmt -l .`
  (no output), and `git diff --check`.
- PASS: `bin/kit check --project`, with the existing non-blocking reference
  README advisory. Targeted `reconcile --dry-run --diff` plans no further
  changes to the affected entry files or rules.
- The initial focused template check correctly detected unregenerated entry
  files. Targeted reconcile regenerated them; the final complete suite passes.

Manual scenario review of the rule (not a live model-behavior evaluation):

| Scenario | Expected agent decision | Review |
| --- | --- | --- |
| Failed matching run, safe to repeat | Bounded pipeline retry before workflow repair. | PASS |
| Missing run, no supported manual start | Inspect capability; identify that limitation without inventing a rerun or an AWS failure. | PASS |
| Expired authentication | Request configured reauthentication; keep identity and target. | PASS |
| One workaround needs approval | Hold that action and dependents; continue independent authorized recovery/work. | PASS |
| Repeated deterministic failure or uncertain persistent writes | Investigate and resolve safety before repeating. | PASS |
| Source branch advanced | Verify intended change and authorized scope before accepting the new run. | PASS |
| Pipeline succeeds, acceptance fails or is unrun | Report deployment separately; product correctness remains unvalidated. | PASS |

## OUTCOME

Implementation and local validation are complete on GH-239, ready for PR
review. The provider-neutral rule is conditional, reuses existing safety rules,
and adds no agent lifecycle or deployment automation. Merge, release, and
downstream rollout are not performed. Distribution requires building or
releasing the updated Kit and running reconcile in each managed repository;
locally customized rules remain preserved and reported.

## REPOSITORY MEMORY

The new contextual rule is durable shared guidance; this spec preserves the
scope and rationale. Updated the universal contract, generated agent entry
files, merge-rule routing, and self-host registry hashes. The Constitution and
historical specs need no rewrite.
