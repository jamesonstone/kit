---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: complete
feature:
  id: 0048
  slug: native-plan-challenge
  dir: 0048-native-plan-challenge
relationships:
  - type: builds_on
    target: 0042-native-plan-repository-memory
  - type: related_to
    target: 0027-implement-readiness-gate
  - type: related_to
    target: 0025-v0-prompt-library
skills:
  - name: github:yeet
    source: GitHub plugin
    path: github:yeet
    trigger: publish the completed implementation through the issue branch and pull request
    required: true
references:
  - id: clipboard-transport
    name: Clipboard transport
    type: code
    target: pkg/cli/clipboard.go
    relation: implements
    read_policy: must
    used_for: explicit macOS clipboard reads and writes
    status: active
  - id: prompt-output
    name: Prompt output
    type: code
    target: pkg/cli/prompt_output.go
    relation: implements
    read_policy: must
    used_for: clipboard-first and raw-output behavior
    status: active
  - id: legacy-plan
    name: Legacy staged plan command
    type: code
    target: pkg/cli/plan.go
    relation: constrains
    read_policy: must
    used_for: preserving the deprecated staged plan generator under kit legacy plan
    status: active
  - id: root-help
    name: Root help routing
    type: code
    target: pkg/cli/root_help.go
    relation: implements
    read_policy: must
    used_for: native-plan utility discovery
    status: active
  - id: capability-catalog
    name: Capability catalog
    type: code
    target: pkg/cli/capabilities_catalog.go
    relation: implements
    read_policy: must
    used_for: command side-effect and usage discovery
    status: active
  - id: feature-notes
    name: Feature notes
    type: notes
    target: docs/notes/0048-native-plan-challenge
    relation: informs
    read_policy: conditional
    used_for: optional source material
    status: optional
delivery_intent: issue_branch_pr_in_progress
---
# SPEC

## PURPOSE

Make the proven cross-model plan-review workflow fast and explicit for Codex for Mac users: copy a plan produced by `/plan`, run `kit plan challenge`, and paste a supplemented adversarial-review prompt into a secondary model without Kit launching or calling any model.

## CONTEXT

- Codex for Mac produces a native plan through `/plan`, then offers only two return paths: implement the plan or tell Codex what to do differently.
- The current manual workflow copies that plan into Claude, asks for an adversarial review, then translates the response back into one of those Codex actions.
- Kit already treats native agent planning as the primary design surface and keeps the deprecated staged plan generator under `kit legacy plan`.
- Kit already copies prompt-producing command output with `pbcopy`, but it has no explicit clipboard-read helper.
- A top-level `kit plan` namespace is currently available and can represent utilities for native-agent plans without making Kit the plan author.
- Existing implementation-readiness guidance defines useful challenge dimensions, but the secondary reviewer must produce a response shaped for Codex for Mac rather than start implementation or rewrite repository memory.
- Delivery uses GitHub issue `#74`, branch `GH-74`, and a ready pull request based on refreshed `origin/main`.

## REQUIREMENTS

- Add a top-level `kit plan` command group described as utilities for plans produced by native coding agents.
- Add `kit plan challenge` with no positional arguments.
- Read the current macOS clipboard only when the user explicitly runs the challenge command.
- Reject clipboard read failures and empty or whitespace-only clipboard content with actionable errors.
- Treat the copied content as the complete candidate Codex plan and preserve its meaningful text in the generated challenge prompt.
- Ask the secondary model to identify only material plan defects: misunderstood goals, incomplete acceptance, contradictions, hidden assumptions, missing failure modes, unsafe sequencing, dependency or rollback gaps, weak validation, unnecessary complexity, scope creep, and risky actions without safeguards.
- Tell the reviewer not to implement, rewrite the entire plan, invent repository facts, or emit style-only feedback.
- Require the reviewer to map its complete response directly to Codex for Mac:
  - output exactly `IMPLEMENT THIS PLAN` when no material change is needed;
  - otherwise output `TELL CODEX WHAT TO DO DIFFERENT:` followed by concise numbered revision instructions suitable for the Codex input field.
- Default to copying the supplemented challenge prompt back to the clipboard and print only a concise acknowledgement.
- Support `--output-only` for raw prompt inspection without changing the clipboard and `--output-only --copy` for explicit dual output.
- Do not invoke a model, open or automate a desktop application, watch the clipboard, access chat history, call the network, or persist the copied plan.
- Keep `kit legacy plan` behavior unchanged.
- Add capability metadata, root-help discovery, command documentation, and focused tests for successful composition, exact reviewer output instructions, empty input, read failure, write failure, and raw-output behavior.
- Observable acceptance: a copied Codex plan becomes one complete paste-ready Claude review prompt; a Claude response following that prompt maps without manual translation to one of the two Codex for Mac plan controls; focused and full repository validation pass.
- Non-goals: model-provider configuration, model invocation, desktop Accessibility automation, background clipboard monitoring, review-result ingestion, canonical plan artifacts, generic clipboard-template expansion, file/stdin input modes, or automatic acceptance of secondary-model feedback.

## ACCEPTED PLAN

1. Add a small clipboard-read helper using `pbpaste`, with an injectable function seam matching the existing clipboard-copy test seam.
2. Add a top-level native-plan command group and a `challenge` subcommand that reads the clipboard, validates the candidate, builds a fixed provider-neutral adversarial prompt, and routes output through the shared clipboard-first helper.
3. Keep the challenge prompt in a focused command file and delimit the copied plan as quoted review input while making the Codex for Mac return format exact.
4. Register `plan` and `plan challenge` in root help and capability metadata without changing the deprecated `kit legacy plan` command.
5. Add focused command, prompt, clipboard-error, output-mode, help, and capability tests; update README and command documentation.
6. Run formatting, vet, focused and full tests, build, changed-lines lint, V3 feature/project checks, prompt-system evaluation when affected, and diff/secret review before explicit staging and delivery.

## DECISIONS

- Accepted a prompt-only clipboard relay; rejected Kit-managed model invocation because the user selects and operates the Codex and Claude Mac applications.
- Accepted `kit plan challenge` as the public surface because it complements Codex-native `/plan` and leaves plan authorship outside Kit.
- Accepted one outbound transformation rather than a two-command round trip: the secondary-model prompt itself must produce Codex-ready feedback.
- Accepted a fixed built-in challenge contract for the first version; rejected generic clipboard templating and prompt-library placeholder expansion as unnecessary surface area.
- Accepted explicit clipboard access only; rejected background watching and automatic application switching.

## DISCOVERIES

- The checked-in `bin/kit` was stale relative to current source and initially scaffolded a V2 spec. The generated feature identity and notes were retained, while this spec was semantically replaced with the current V3 living-spec contract before implementation code changed.
- `pkg/cli/plan.go` is registered only beneath `legacyCmd`, so a distinct top-level `plan` command group does not collide with the existing public command tree.
- The shared prompt-output helper already implements the desired default-copy, `--output-only`, and explicit `--copy` behavior; the feature needs only clipboard input and command-specific composition.

## VALIDATION

- `make fmt` and `git diff --check` completed without formatting or whitespace errors.
- The focused `kit plan challenge`, root-help, and capability tests passed.
- `go test ./pkg/cli -count=1` passed.
- `go vet ./...` passed.
- `go test ./... -count=1` passed across every package.
- `go test -race ./pkg/cli -count=1` passed.
- `make build` produced `bin/kit` successfully.
- `golangci-lint run --new-from-rev=origin/main ./...` reported `0 issues`.
- The repository-wide `make lint` target still reports 45 pre-existing findings outside this change; changed-lines lint is clean.
- `./bin/kit capabilities plan challenge --json` returned the documented clipboard, model, network, persistence, and output-mode boundaries.
- `./bin/kit check native-plan-challenge` passed.
- `./bin/kit improve run --suite prompt-system --kit-binary ./bin/kit --json` run `20260723T171843.211414000Z-89cd0b` passed all 45 task runs and all 345 assertions with deterministic output across all 15 repeated tasks.
- Focused command tests exercised complete prompt composition, default clipboard replacement, raw output without replacement, explicit dual output, empty input, read failure, and write failure without accessing the user's real clipboard.

## OUTCOME

- Added a native-agent `kit plan` namespace without changing the deprecated staged `kit legacy plan`.
- Added `kit plan challenge` to turn a copied Codex for Mac plan into a complete adversarial-review prompt and copy it back for a human-selected secondary model.
- Constrained the secondary response to Codex's two actual plan controls so accepted plans and requested revisions require no manual translation.
- Kept Kit outside model execution, desktop automation, background clipboard monitoring, chat history, network access, and copied-plan persistence.
- Added command discovery, capability metadata, user documentation, and regression coverage for the complete clipboard handoff contract.

## REPOSITORY MEMORY

Decision: created

Rationale: The boundary between native plan authorship, human-selected secondary review, clipboard transport, and Codex-specific return semantics is a durable feature contract that code alone does not fully explain. The existing Constitution already establishes the applicable project-wide native-planning and explicit-execution invariants, so no Constitution change is warranted.

Artifacts:

- `docs/specs/0048-native-plan-challenge/SPEC.md`
