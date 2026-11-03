---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0096"
  slug: "init-issue-template-prewrite"
  dir: "0096-init-issue-template-prewrite"
---
# SPEC

## PURPOSE

Keep `kit init` and `kit reconcile` from installing a GitHub issue template over an existing issue setup, and show every planned file write before it happens.

## CONTEXT

Issue: https://github.com/jamesonstone/kit/issues/276. GH-274 (PR #275) scaffolds `.github/ISSUE_TEMPLATE/issue.md` from `internal/templates/coderabbit.go`. A local template can override repository issue forms, `config.yml`, and templates inherited from an organization `.github` repository. Init previously wrote files without a single preview. Reconcile already previews with `--dry-run` and applies Kit-owned updates by default.

## REQUIREMENTS

- Do not create the issue template unless the project opts in with `kit init --issue-template`, `kit reconcile --issue-template`, or `.kit.yaml` `issues.issue_template: true`.
- Keep updating a file whose bytes exactly match a shipped Kit issue template, including when the setting is absent.
- Leave every other existing issue template unchanged.
- Before writing, `kit init` and `kit reconcile` list each file they would create, overwrite, or skip, and warn about existing files they would change and about existing issue templates, forms, and `config.yml`.
- Ask before an init overwrite. A non-interactive init skips that file unless `--force` is set.
- `kit init --dry-run` prints the plan and writes nothing. Tests compare a temp directory before and after.
- Warn when an organization owner can be identified. Otherwise document that inherited organization templates are not listed.
- Out of scope: a separate `kit precheck` command, and pull-request template write rules beyond the shared plan.

## ACCEPTED PLAN

Add `issues.issue_template` as an explicit boolean. `--issue-template` sets it and persists it. The scaffold planner creates a missing issue template only when that flag is set. Reconcile sets `updateShippedIssueTemplate` so an exact shipped copy still converges.

Build init's file plan with the existing refresh planner, seeded from the in-memory config, plus the default Mint pipeline init already adds. Print the plan, apply the overwrite gate, then write once. `--dry-run` returns before global config, directories, clipboard, and file writes.

The init gate confirms updates, merges, and removals, and confirms creating `issue.md` when any other issue surface exists. Non-interactive init skips those writes unless `--force`. Reconcile still applies other Kit-owned migration writes without a prompt. It uses the same confirm-or-skip rule only for creating the opt-in issue template beside existing issue surfaces, and it does not replace a custom issue template.

Scan `.github/ISSUE_TEMPLATE.md` and `.github/ISSUE_TEMPLATE/` for templates, YAML forms, and `config.yml`. Resolve the origin owner with git and, best effort, `gh api users/<owner>`. An organization owner produces a warning that a local template overrides inherited organization templates and that Kit does not download them. Any other failure produces the documented limitation only when an issue-template write is planned.

## DECISIONS

- Use both the flag and `.kit.yaml`. The flag matches `--mint`. The setting matches other repository issue choices and survives later reconcile runs.
- Default is off. An omitted `issues.issue_template` does not create the file.
- Exact shipped bytes are the ownership proof, same as the pull request template. Do not infer ownership from the path.
- Reconcile does not prompt before ordinary Kit-owned updates. The constitution requires migration to apply by default, and a non-interactive skip would stop agent reconcile. The new confirm-or-skip rule covers init overwrites and the risky issue-template create.
- `--force` on init authorizes those gated writes. It does not mean reconcile's `--force` (replace edited Kit sections). Reconcile `--force` also authorizes the gated issue-template create.
- Do not add a network requirement. `gh` failure does not fail the command.
- Kit's own repository already has the exact shipped template, so reconcile keeps updating it without setting `issues.issue_template`.

## DISCOVERIES

Init's file set is the refresh plan plus the default Mint pipeline. One plan, one apply, replaces the old scaffold-then-refresh sequence. `--output-only` keeps the prompt on stdout and sends the plan to stderr. An interactive run prints the plan and warnings before it asks. A non-interactive or dry run prints the decided plan. Surface warnings are included. A current reconcile stays quiet, except when a non-interactive run skipped the opt-in issue template beside an existing issue setup: that plan and its warnings are printed and nothing is written.

## VALIDATION

Passed `gofmt -l .` (empty), `make test` (`go test -v ./...`, exit 0, `pkg/cli` ok), `golangci-lint run ./...` v2.11.0 (`0 issues`), `make vet` (`go vet ./...`, exit 0), `make build` (`bin/kit` `v3.0.50`), and `KIT_USAGE_DISABLED=1 ./bin/kit check --project` (exit 0, one pre-existing non-blocking warning for edited `docs/references/README.md`). `./bin/kit init --dry-run` against a temp fixture with `ISSUE_TEMPLATE.md`, `config.yml`, and `bug.yml` exited 0, printed the plan and warnings, and left the fixture and the global config path unchanged.

## OUTCOME

Issue template creation is opt-in. Exact shipped copies still update on reconcile. Init and reconcile print the pre-write plan. Init asks before changing an existing file and skips that change when the run is not interactive unless `--force` is set. `--dry-run` writes nothing. Organization templates are warned when `gh` reports an organization owner, and the limitation is printed and documented otherwise. Reconcile still applies other Kit-owned updates without a prompt.

## REPOSITORY MEMORY

This spec keeps the opt-in, overwrite gate, and organization-template limitation. Help text, `docs/commands.md`, the README, and the Constitution init sentence record the operator contract. The checked-in issue template stays the scaffold source bytes.
