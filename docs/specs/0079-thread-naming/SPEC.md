---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0079"
  slug: "thread-naming"
  dir: "0079-thread-naming"
---
# SPEC

## PURPOSE

Make long-lived and forked coding conversations searchable by the work they currently own, while preserving their identity, history, context/cache and tool state.

## CONTEXT

Kit installs versioned repository instructions and support documents from internal/templates, and registry-backed rules from docs/references/rules. Codex and Cursor read AGENTS.md; Claude Code reads CLAUDE.md. Ordinary Claude Chat/Cowork is not Claude Code and repository instruction loading is unverified. Existing Codex initialization fixes the initial title and forbids later changes; that conflicts with mutable ownership.

## REQUIREMENTS

- One canonical policy: [scope] domain / objective, normally repository scope, stable conceptual domain, current outcome. Aim for 60 characters without truncating meaning.
- Rename at semantic checkpoints only; suppress cosmetic differences and execution-stage churn. Fork children own their new objective without ancestry suffixes.
- Apply only verified native metadata operations to the exact current/child session. Preserve context and never create, restart, compact, duplicate or fork for naming.
- Install through existing scaffolds; preserve idempotency and user-owned instructions. Unknown capability degrades to one suggested title at a meaningful transition.
- No private-state writes, UI automation, polling, additional model calls, conversation store or lineage subsystem.

## ACCEPTED PLAN

1. Research host mechanisms and publish capability evidence before implementation.
2. Add a shared installed naming reference and small always-loaded pointer; extend Codex initialization to canonical titles and semantic reevaluation. Cursor reuses AGENTS.md.
3. Add an optional read-only title formatter/decision command under kit instructions, taking agent-resolved semantic fields; semantic inference remains with the active agent, never a keyword classifier or separate model.
4. Document native Codex and Claude SDK application with conservative identity/capability checks and assisted fallbacks; no hooks without preexisting semantic inputs.
5. Cover deterministic behavior, real scaffold installation/refresh and semantic scenario evaluation; run full Go validation, independent read-only review, then deliver one ready PR for GH-209.

Topology: root owns all integration and policy writes; read-only host researcher; bounded helper implementation may be delegated with disjoint files; fresh read-only verifier. Host confirms four total slots, stable references and follow-up. Agents inherit runtime model/effort (exact effective profile unverified); no descendants. Exit after verified implementation and ready PR; diagnose failures before retries, and stop only for a genuine external blocker. No merge authority.

## DECISIONS

Use existing repository scope and program context, not a new workspace model. Semantic interpretation is agent-owned; deterministic formatting/comparison consumes resolved fields. Optional helper calls are not mandatory at every checkpoint. No SDK dependency, daemon or hook installation is necessary: use already-exposed native tools/SDK only when safely bound, otherwise assist. Startup hooks cannot infer a future prompt or changed objective.

## DISCOVERIES

Codex native set_thread_title succeeded in this task; official app-server documents thread/name/set for loaded and persisted sessions. Claude SDK documents renameSession and hooks support SessionStart.sessionTitle, but host identity and SDK availability must be proven. Cursor documents manual rename but no verified existing-chat agent-callable setter. Host capability details and sources will live in the installed reference.

## VALIDATION

- PASS: `go test ./...`; affected-package race tests with `go test -race ./internal/threadtitle ./internal/instructions ./internal/templates ./pkg/cli`.
- PASS: `go vet ./...`, `golangci-lint run ./...` (0 issues), build, gofmt, `git diff --check`, affected Go source-size audit (all <=300 lines), changed-file gitleaks scan (no leaks).
- PASS: real refresh installs the shared policy, preserves bytes on repeat, and preserves unrecognized custom sections with the existing append-only conflict. Migration test retains old initialization text, adds an explicit naming override, satisfies reconciliation and remains idempotent.
- PASS: feature check `kit check 0079-thread-naming` and context resolution.
- FAIL, pre-existing: `kit check --project` reports duplicate feature number 0077 and that feature's missing progress entries/reference warning. Those unrelated historical records are preserved. Naming introduces no remaining project-check finding.
- Observed native Codex title update returned the same task ID. Claude/Cursor live title mutation and Desktop sidebar reflection remain UNOBSERVED; documented capability limits are not live acceptance claims.
- Fresh read-only verifier closed wording and append-only audit findings. Independent agent semantic evaluation passed initial webhook403, known JWT routing, remaining delivery verification, stage-only no-op, EUID/UI child ownership, known cross-repository program scope and one unsupported-host suggestion. This evaluates policy reasoning, not provider execution. Without evidence of JWT, use authorization routing rather than infer an authentication format.

## OUTCOME

Implementation and local validation complete. GH-209 delivers one shared embedded policy and installed projection, checkpoint routing across hosts, conservative supported metadata adapters, optional read-only naming/title commands and frozen global instructions v16. Existing v1-v15 remain immutable. No host daemon, hook, dependency, conversation store or private-state manipulation was introduced. Delivery is GH-209 to main; issue #209 tracks the current pull request. Merge/release/global installation are outside this delivery.

## REPOSITORY MEMORY

This spec retains design rationale and acceptance boundaries. 0081 retired the shared naming policy, always-loaded Conversation Naming pointers, `kit instructions naming`/`title`, and `internal/threadtitle`. v16 is the last global instruction version that required naming. Historical instruction versions remain immutable.
