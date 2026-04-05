# PLAN

## SUMMARY

Add a new canonical `resume` command, promote `status --all` into the explicit
project overview mode, and simplify the visible root command surface through
grouped help plus hidden deprecated compatibility wrappers.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02][SPEC-03][SPEC-04][SPEC-05][SPEC-06] Add a new
  `pkg/cli/resume.go` command that routes backlog items through the existing
  backlog pickup helper and routes non-backlog features through the existing
  catch-up prompt behavior
- [PLAN-02][SPEC-11][SPEC-12][SPEC-13][SPEC-14][SPEC-15][SPEC-16] Refactor
  `status` so the default mode stays active-feature focused while `--all`
  renders the fleet view in dedicated text and JSON output paths, with the
  human-readable path using a fixed-width lifecycle matrix instead of a
  Markdown-style table
- [PLAN-03][SPEC-17][SPEC-18] Rework root help rendering in `pkg/cli/root.go`
  and `pkg/cli/human_output.go` so only root help is grouped into product
  sections
- [PLAN-04][SPEC-19][SPEC-20][SPEC-21][SPEC-22][SPEC-23] Convert duplicate or
  maintenance commands into hidden deprecated compatibility surfaces while
  preserving invocation behavior
- [PLAN-05][SPEC-07][SPEC-08][SPEC-09][SPEC-10][SPEC-24] Update backlog,
  brainstorm, catchup, upgrade, skill, README, and canonical workflow docs so
  they teach the simplified command surface accurately
- [PLAN-06] Add or update focused tests for resume routing, status all-features
  output, grouped root help, and deprecated command visibility, then rerun the
  normal verification suite

## COMPONENTS

- `pkg/cli/resume.go`
  - new canonical resume command
  - mixed selector and routing logic
- `pkg/cli/backlog.go` and `pkg/cli/backlog_shared.go`
  - backlog guidance updates
  - reuse of shared pickup helpers
- `pkg/cli/catchup.go`
  - shared prompt path for non-backlog resume
  - deprecated hidden compatibility registration
- `pkg/cli/status.go` and `pkg/cli/status_output.go`
  - `--all` handling
  - dedicated all-features text and JSON rendering
  - fixed-width lifecycle matrix for terminal text output
- `pkg/cli/root.go` and `pkg/cli/human_output.go`
  - root command grouping
  - hidden/deprecated command visibility
- `pkg/cli/brainstorm.go`
  - deprecated hidden `--pickup` flag behavior
- `pkg/cli/upgrade.go` and `pkg/cli/skill.go`
  - canonical plus hidden-deprecated root command registration
- docs
  - `README.md`
  - `docs/specs/0000_INIT_PROJECT.md`
  - amended existing feature specs

## DATA

- no new persisted state
- reuse existing lifecycle data from:
  - `.kit.yaml`
  - feature phase derivation
  - backlog classification
- add a new `status --all --json` response shape without changing the existing
  default `status --json` payload

## INTERFACES

- new command:
  - `kit resume [feature]`
- new status mode:
  - `kit status --all`
- deprecated hidden compatibility surfaces:
  - `kit update`
  - `kit skills`
  - `kit catchup`
  - `kit scaffold`
  - `kit rollup`
  - `kit brainstorm --pickup`

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| backlog resume helper | code | `pkg/cli/backlog_shared.go` | canonical backlog resume path | active |
| catchup prompt behavior | code | `pkg/cli/catchup.go`, `pkg/cli/catchup_prompt.go` | non-backlog resume path | active |
| backlog classification | code | `internal/feature/backlog.go` | selector routing and status labeling | active |
| status model | code | `internal/feature/status.go` | per-feature overview data | active |
| Cobra hidden/deprecated support | library | `github.com/spf13/cobra` | compatibility command and flag migration | active |

## RISKS

- resume routing can drift from backlog classification if it reimplements the
  backlog rule instead of reusing the shared helper
- grouped root help can accidentally affect subcommand help if the rendering
  logic is not isolated to the root command
- changing `status` output can break expectations if the default JSON shape is
  modified instead of leaving it untouched
- hiding deprecated compatibility commands can confuse tests or docs if help
  assertions still expect the old visible surface
- the dirty worktree increases merge risk on CLI and doc files, so changes must
  stay tightly scoped and avoid overwriting unrelated edits

## TESTING

- add or update tests for:
  - resume routing and selector ordering
  - backlog guidance text
  - status active-only text output
  - `status --all` text output
  - `status --all --json` payload shape
  - root help grouping
  - hidden deprecated command visibility
  - deprecated `brainstorm --pickup`
- run:
  - `go test ./...`
  - `make build`
  - `go run ./cmd/kit/main.go --help`
  - `go run ./cmd/kit/main.go status`
  - `go run ./cmd/kit/main.go status --all`
  - `go run ./cmd/kit/main.go resume`
