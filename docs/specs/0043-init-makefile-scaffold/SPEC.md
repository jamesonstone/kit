---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: complete
feature:
  id: 0043
  slug: init-makefile-scaffold
  dir: 0043-init-makefile-scaffold
references:
  - id: core-init-contract
    name: Core project initialization contract
    type: doc
    target: docs/specs/0000_INIT_PROJECT.md
    relation: constrains
    read_policy: must
    used_for: canonical kit init and refresh behavior
    status: active
  - id: init-command
    name: Init command
    type: code
    target: pkg/cli/init.go
    selector_type: symbol
    selector: runInit
    relation: implements
    read_policy: must
    used_for: fresh initialization and generated initialization prompt
    status: active
  - id: init-scaffold
    name: Init scaffold files
    type: code
    target: pkg/cli/init_scaffold.go
    relation: implements
    read_policy: must
    used_for: create-if-missing project files
    status: active
  - id: init-refresh-files
    name: Init refresh file planning
    type: code
    target: pkg/cli/init_refresh_files.go
    selector_type: symbol
    selector: planRefreshInitScaffoldFiles
    relation: implements
    read_policy: must
    used_for: missing-file refresh and existing-file preservation
    status: active
  - id: project-templates
    name: Project scaffold templates
    type: code
    target: internal/templates/templates.go
    relation: implements
    read_policy: must
    used_for: safe starter Makefile content
    status: active
delivery_intent: issue_branch_pr_in_progress
---
# SPEC

## PURPOSE

Make `kit init` establish a safe, canonical Make entrypoint for each project and guide the initialization agent to wire targets such as `make dev` to the repository's real development commands.

## CONTEXT

- Fresh initialization already creates project configuration, local environment files, GitHub templates, repository instructions, and a Constitution before emitting a coding-agent prompt.
- The prompt currently asks only for `docs/CONSTITUTION.md`; it does not establish a consistent command entrypoint for development, build, test, lint, or validation workflows.
- The Kit, LabCore UI, and LabCore Makefiles share durable conventions: explicit `.PHONY` declarations, thin recipes over native toolchain commands, overridable tool variables where useful, and project-specific target sets rather than one universal recipe body.
- `kit init` runs before Kit can reliably infer every repository's package manager, runtime, service topology, or local development command. A static template must not pretend guessed commands are valid.
- Once the initialization agent maps the starter to real commands, the Makefile is project-owned. `kit init --refresh`, including targeted force refreshes, must not erase those commands.
- Issue `#60` is the delivery source, and branch `GH-60` was created from refreshed `origin/main` before implementation edits.

## REQUIREMENTS

- A fresh `kit init` creates `Makefile` when it is missing and preserves any existing Makefile byte-for-byte.
- `kit init --refresh` backfills a missing Makefile, while existing Makefiles remain preserved even with `--refresh --file=Makefile --force`.
- The starter Makefile provides a valid, useful basic structure without placeholder recipes or guessed project commands.
- The starter uses canonical Make conventions, including a default help entrypoint and explicit `.PHONY` declaration.
- The initialization prompt instructs the agent to inspect package scripts, toolchain configuration, development docs, and existing automation before adding recipes.
- The prompt requires canonical targets such as `dev`, `build`, `test`, `check`, `lint`, `fmt`, and `clean` only when a verified underlying repository command or workflow exists; `make dev` is the preferred local-development entrypoint when the project has a development/run workflow.
- Recipes remain thin wrappers around real project-native commands, composite targets reuse atomic targets, and no TODO, echo-only placeholder, guessed command, or duplicate build logic remains after initialization.
- Init help, next-step output, focused tests, the core initialization contract, and the project rollup describe the Makefile behavior accurately.
- Observable acceptance: fresh init and refresh tests create the starter, preservation tests protect customized content, prompt tests assert command-mapping guidance, full Go validation passes, and Kit document checks introduce no blocking finding.
- Non-goals: modifying sibling repositories, implementing toolchain auto-detection in Kit, forcing one target set on every project, or treating populated Makefiles as Kit-generated files that force refresh may overwrite.

## ACCEPTED PLAN

1. Add a minimal Makefile template with a default `help` target and no unverified project command recipes.
2. Add `Makefile` to fresh init and missing-file refresh scaffolding, with an explicit project-owned preservation rule that wins over targeted force refresh.
3. Expand the generated initialization prompt and visible next steps so an agent maps applicable canonical targets to verified repository-native commands and validates the resulting entrypoints.
4. Add focused template, init, refresh, preservation, and prompt regression tests.
5. Update the core initialization contract and project rollup, then run focused tests, formatting, full Go checks, lint, Kit document validation, and diff review.

## DECISIONS

- Use a safe `help`-only starter instead of guessed `dev`, build, or test recipes. The initialization prompt owns repository-specific command mapping because it can inspect live project context.
- Treat the populated Makefile as project-owned. Refresh may create it when absent but may never overwrite an existing file, including under targeted `--force`.
- Keep Make recipes as thin command aliases. The Makefile is a stable human interface, not a second implementation of package scripts, build pipelines, or service orchestration.
- Update the existing core init contract as canonical product documentation while retaining this feature spec for the rationale behind starter safety and refresh ownership.

## DISCOVERIES

- Plain `kit spec init-makefile-scaffold --output-only` still entered the deprecated V2 editor flow in this checkout, created a V2 skeleton, and then failed because configured editor `nvim` was unavailable. The generated artifact was semantically converted to this V3 spec before implementation.
- Targeted refresh validation uses a separate known-target registry from the scaffold planner. The first focused test run exposed the missing `Makefile` registration; adding it made targeted create and preserve behavior reachable.
- The documented `kit rollup` command is not available in the current CLI. V3 completion is the supported path that refreshes `PROJECT_PROGRESS_SUMMARY.md` after the living spec is complete.
- GitHub reports no explicit branch-protection resource for `main`; repository safety rules therefore require treating `main` as protected by assumption.
- The shell's default `PATH` omits `gh` and `go`; the installed `/opt/homebrew/bin/gh` and `/opt/homebrew/bin/go` binaries are the verified command paths for this task.

## VALIDATION

- Focused template, fresh-init, refresh, force-preservation, and prompt tests passed in `internal/templates` and `pkg/cli`.
- Full affected-package tests passed: `go test ./internal/templates ./pkg/cli -count=1`.
- A built Kit binary initialized a temporary project successfully; the generated starter ran `make help`, and the emitted prompt contained the absolute Makefile path, verified-command requirement, `make dev` guidance, and placeholder prohibition.
- `PATH="/opt/homebrew/bin:$PATH" make fmt` passed.
- `go test ./... -count=1` passed across all packages.
- `go vet ./...` and `go build ./cmd/kit` passed.
- `go test -race ./internal/templates ./pkg/cli -count=1` passed.
- `PATH="/opt/homebrew/bin:$PATH" golangci-lint run --new-from-rev=origin/main ./...` passed with `0 issues`.
- `go run ./cmd/kit check 0043-init-makefile-scaffold` passed.
- Initial `go run ./cmd/kit check --project` reported only the expected blocking absence of the new 0043 rollup row and heading, plus historical compatibility warnings and the existing Constitution-refresh advisory. V3 completion will refresh the rollup before the final project check.
- `go run ./cmd/kit complete 0043-init-makefile-scaffold` passed, set the V3 phase to `complete`, and refreshed `docs/PROJECT_PROGRESS_SUMMARY.md` with the 0043 row and summary.
- Final `go run ./cmd/kit check --project` exited successfully with 15 non-blocking historical V2 compatibility advisories and the existing project-refresh advisory; it reported no blocking findings attributable to this feature.
- `git diff --check` passed before completion curation; final staged and unstaged whitespace checks remain part of delivery validation.

## OUTCOME

- Fresh `kit init` now creates a safe Makefile starter with `.DEFAULT_GOAL := help` and `.PHONY: help` while preserving existing Makefiles byte-for-byte.
- Full and targeted refreshes backfill a missing Makefile, but existing project-owned content wins over targeted `--force` refresh.
- The initialization prompt now inspects repository-native command sources, exposes `make dev` when a verified development workflow exists, adds only applicable canonical targets, keeps recipes thin, rejects placeholders and guessed commands, and requires safe target validation.
- Init help, next steps, README, command guidance, the core initialization contract, focused tests, and this rationale record are aligned with the shipped behavior.
- No sibling repository was modified; its Makefile remained a read-only canonical example.

## REPOSITORY MEMORY

Decision: updated

Rationale: The code and tests can preserve scaffold mechanics, but they cannot fully preserve why the starter avoids guessed commands or why force refresh must yield to project ownership. This feature spec records that rationale, and the core init contract will remain the canonical shipped-behavior reference.

Artifacts:

- `docs/specs/0043-init-makefile-scaffold/SPEC.md`
- `docs/specs/0000_INIT_PROJECT.md`
