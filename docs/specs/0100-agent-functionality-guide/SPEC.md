---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0100"
  slug: "agent-functionality-guide"
  dir: "0100-agent-functionality-guide"
---
# SPEC

## PURPOSE

Make installed Kit functionality discoverable to coding agents through one local, read-only command and normal agent entry points.

## CONTEXT

Cobra command help is the authoritative command description. Execute prunes retired commands through internal/commandset. Config preflight can remediate configuration, so guide must bypass it. Usage recording can write local state, so guide is excluded. Independent GH-281 adds inspect; deriving commands from the live tree will include it after integration. GH-280 / PR #283 adds SOP guidance independently.

## REQUIREMENTS

Expose all visible installed commands, nested commands, usage, examples, flags and artifact instructions in deterministic text and JSON. Work outside projects without network, config remediation, or writes. Route discovery through the canonical contract, its three adapters, root help and init. Preserve the prohibition on duplicated capability catalogs. Describe SOPs as project-owned conventions, not a new creation command. Keep handwritten Go files below 300 lines and the contract at most 1475 words before SOP integration.

## ACCEPTED PLAN

1. Add a reusable command-tree snapshot and deterministic renderers.
2. Register kit guide --json, bypass config preflight, and exclude telemetry.
3. Add artifact guidance and discovery routes; regenerate checked-in contract adapters.
4. Test tree coverage, synthetic additions, deterministic JSON, flag metadata, artifact accuracy and read-only behavior. Run project validation.

## DECISIONS

Use live Cobra help rather than a parallel command catalog. Keep artifact guidance separate from command metadata because artifacts are concepts, not executable commands. Include project SOP convention with draft/active/retired states and no execution engine. Do not copy pending inspect or SOP implementations. Use a dedicated contract section to avoid pending SOP clause edits. Compact equivalent delivery prose to leave room under the contract budget.

## DISCOVERIES

Cobra lazily creates help and version flags, so the snapshot initializes both without executing commands. Root preflight can mutate config and released CLI usage can write local state; guide overrides the inherited preflight and has no telemetry path. Separate cleanup tests also pin contract phrases, so those markers remain intact. The contract is 1473 words, leaving room for the independent SOP integration. Rules list covers installed local rules only; view can read shipped rules, and the no-slug add selector can mutate.

## VALIDATION

Passed: focused guide, artifact, template/adapter, and cleanup-marker tests; make vet; make lint (0 issues); make build; kit check --project; go test -race ./internal/guide -count=1; gofmt -l . (empty); git diff --check. Project check reports the existing nonblocking preserved docs/references/README.md compatibility advisory. Full go test ./... passed (pkg/cli 54.925s). A final four-word contract clarification restored repository-local scope, target verification and useful-only delegation; the affected template, adapter, guide and cleanup tests passed again afterward. Two separate binary guide JSON invocations were byte-identical and included root --version. Initial full runs found compressed cleanup marker mismatches; restored the exact markers and verified their focused test. No production or network validation applies to this local read-only command.

## OUTCOME

Implemented kit guide and --json, derived from the installed live tree rather than a duplicate command catalog. Added artifact instructions, canonical contract/adapters, root help and init onboarding routes. Synthetic-command tests verify future registered commands appear automatically. No inspect or SOP implementation was copied and no execution engine was added. The independent lane is ready for review. A combined-tree check found only overlapping Constitution and root-help edits to resolve at merge time; the derived command tree automatically exposes separately registered commands.

## REPOSITORY MEMORY

Updated this living spec, README.md, docs/commands.md, and the Constitution command-surface invariant. Preserved the non-goal against a parallel or duplicated capability catalog while documenting derived help. Canonical contract and three generated entry adapters direct agents to the guide; no guarantee of agent compliance is claimed. Historical specs and the testing reference remain unchanged.
