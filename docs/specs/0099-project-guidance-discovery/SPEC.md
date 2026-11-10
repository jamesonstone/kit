---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: deliver
feature:
  id: "0099"
  slug: project-guidance-discovery
  dir: 0099-project-guidance-discovery
---
# SPEC

## PURPOSE

Expose current repository-local guidance through `kit inspect`, for people and agents.

## CONTEXT

Rules list describes installed local rules; rules view can read a local or shipped artifact. Project discovery needs current local evidence instead. The existing config parser supports read-only inspection, but cwd preflight can repair configuration. SOP scaffolding is separate work; this command must work without it.

## REQUIREMENTS

- Resolve cwd or `--path` to the nearest containing Kit project without changing cwd.
- Distinguish non-project, nonexistent, inaccessible, and invalid paths/configuration.
- Report local rulesets, SOPs, agent entry files, Constitution, design/security and testing/SOP guides deterministically, with stable JSON.
- Preserve valid findings alongside invalid entries. Invalid metadata must never imply active status.
- Read only bounded guidance locations. Reject path escapes and symlinks; do not run instructions, use the network, repair configuration, or write project files.
- Keep handwritten Go files at or below 300 lines.

## ACCEPTED PLAN

1. Add an internal discovery package for safe resolution, local inventory, metadata validation, and report types.
2. Add thin `kit inspect` CLI with `--path` and `--json`. Exclude it from repair and telemetry.
3. Update command protection, root help, current user docs and Constitution.
4. Test path resolution, errors, metadata, stable output, no writes, and command registration. Run repository validation.

## DECISIONS

- Use `kit inspect` rather than overload shipped registry commands.
- Reject all guidance symlinks, including links inside the project, to keep the read boundary explicit and simple.
- Omit inspect telemetry rather than attribute an explicit target to the caller's cwd.
- Report metadata verbatim only after successful YAML decoding; report unknown status as invalid and inactive.
- Empty config findings encode as `[]`. `ParseWithInspection` validates `shell.command`, so file and byte parsers share that check.

## DISCOVERIES

- Ruleset lifecycle is active/optional/stale; SOP lifecycle is draft/active/retired. Ruleset descriptions are optional. Each artifact must match its file or directory slug before it can be active.
- Config parsing is reused through a byte-based inspection helper. An os.Root handle confines config and guidance reads to the resolved project.
- Configured missing paths remain visible. Optional default paths are omitted when absent. Inventory categories normalize registered project guidance to avoid duplicate entries.
- Scan targets must be directories before opening them; regular files and FIFOs are reported rather than opened as directories.
- Registry scope follows the executable canonical validator: omitted, downstream, or optional. A stale diagnostic mentioned kit-maintainer; corrected it and added embedded-corpus coverage plus 27 parity cases.

## VALIDATION

- Final validation passed: `gofmt -l .` (no output), `make vet`, `make lint` (0 issues), `make test`, `make build`, and `KIT_USAGE_DISABLED=1 ./bin/kit check --project`.
- `go test -race ./internal/guidance ./pkg/cli -run 'Test(Inspect|Metadata|Resolve)' -count=1` passed, including directory-type/FIFO and missing-reference regressions.
- Final built-binary self-inspection and inspection of the independent SOP lane both passed with zero errors, including all three optional-scope shipped rules. Permission tests passed on the selected Mac.
- Project check reports one nonblocking compatibility advisory for the pre-existing `docs/references/README.md` content.
- Review repair after #279 merged: `kit inspect --json` emits `"findings": []` when empty; shell validation stays inside `ParseWithInspection`; inspect docs sit beside the Command Surface table; root help no longer lists retired command names. `gofmt -l .` printed nothing; `go vet ./...` passed; `go test ./... -count=1` passed (`pkg/cli` 34.946s); `golangci-lint run ./...` reported 0 issues; `make build` passed; `KIT_USAGE_DISABLED=1 ./bin/kit check --project` exited 0 with the existing nonblocking references advisory. Hosted CI for this repair was not yet rerun.

## OUTCOME

Implemented `kit inspect [--path <directory>] [--json]` with deterministic local reports, error-preserving metadata validation, safe project resolution, and no repair, network, instruction execution, or usage recording. The command does not depend on SOP scaffolding. A fresh-context review confirmed the FIFO fix and metadata parity regressions. The independent lane is ready for review; no production operations or execution engine were added.

## REPOSITORY MEMORY

This spec preserves the discovery contract. README, docs/commands.md, and Constitution record the supported command surface and read-only behavior. Historical specs remain unchanged.
