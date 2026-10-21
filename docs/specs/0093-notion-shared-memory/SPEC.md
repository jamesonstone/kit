---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "complete"
feature:
  id: "0093"
  slug: "notion-shared-memory"
  dir: "0093-notion-shared-memory"
---
# SPEC

## PURPOSE

Let a project opt into Notion shared memory through concise agent guidance.

## CONTEXT

GH-266. Init and reconcile already share an instruction planner; config set has
YAML-preserving updates and canonical reconciliation target handling.

## REQUIREMENTS

Default disabled. Configure destination and operating-guide HTTPS references.
Preserve custom instructions and MCP connections. Dry runs write nothing and
repeated reconciliation makes no further change. No private IDs in shipped files.

## ACCEPTED PLAN

Add repository-local shared_memory configuration and init/config-set flags.
Render one bounded optional block in supported agent entries through the existing
planner. Validate references before mutation. Add focused deterministic tests and
deliver a draft PR from a clean clone on GH-266, per the explicit user request.

## DECISIONS

Notion is the sole initial provider. References are user-supplied HTTPS URLs;
Kit does not fetch them. Keep repository memory authoritative for project facts.
Reference the operating guide without copying it. Do not change MCP settings,
authenticate, sync, schedule, manage credentials, run paid model tests, merge,
release, or manage local Mint binaries. Authentication is a separate setup task.
Use a separate managed block so the universal contract stays provider-neutral.
Reject malformed blocks instead of risking custom content.

## DISCOVERIES

Init already runs the shared refresh planner after scaffolding, so the optional
block uses the same path for init and reconcile. The instruction audit now uses
the configured planner too. Explicit forced config refresh retains the opt-in.

## VALIDATION

- `go test ./...`: PASS.
- Focused `TestSharedMemory`, `TestConfigSet`, and instruction tests: PASS.
- Init opt-in/default, dry-run, reconcile idempotency, custom guidance and MCP
  preservation, disable, invalid references, malformed markers, alias rejection,
  YAML comments, and forced config preservation are covered with local fixtures.
- `gofmt -l .`: no output; `git diff --check`: PASS.
- `make vet`, `make build`, `make lint`: PASS (lint reports 0 issues).
- `go run ./cmd/kit check --project`: PASS with the existing nonblocking advisory
  that edited `docs/references/README.md` is preserved.
- No live Notion calls, OAuth grants, credential changes, external-agent contact,
  paid model tests, merge, or release. Real host authentication is unverified.

## OUTCOME

Implemented repository-local configuration, init/config-set flags, and bounded
optional guidance in the existing instruction planner. Default behavior is
unchanged; no guide copy or MCP configuration is generated. Draft PR delivery
only. The primary checkout is unchanged. Live access depends on separate host
authentication and the configured guide; this integration does not verify it.

## REPOSITORY MEMORY

This spec preserves the integration boundary. Detailed usage belongs in the
command guide; README gets a short pointer with Maintainers last.
