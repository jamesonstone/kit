---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: complete
feature:
  id: 0050
  slug: safe-worktree-workflow
  dir: 0050-safe-worktree-workflow
references:
  - id: safety-guardrails
    name: Safety guardrails
    type: ruleset
    target: docs/references/rules/safety-guardrails.md
    relation: constrains
    read_policy: must
    used_for: dirty-state and destructive-action boundaries
    status: active
  - id: github-pr-delivery
    name: GitHub PR delivery
    type: ruleset
    target: docs/references/rules/github-pr-delivery.md
    relation: constrains
    read_policy: must
    used_for: issue and pull-request lane semantics
    status: active
  - id: tooling
    name: Agent tooling
    type: documentation
    target: docs/agents/TOOLING.md
    relation: guides
    read_policy: must
    used_for: project-directory and worktree guidance
    status: active
skills:
  - name: github:github
    source: GitHub plugin
    path: github:github
    trigger: verify the existing issue and pull-request lanes in Kit and LabCore
    required: true
  - name: github:yeet
    source: GitHub plugin
    path: github:yeet
    trigger: publish the completed follow-up to the existing ready pull requests
    required: true
delivery_intent: issue_branch_pr_ready
---
# SPEC

## PURPOSE

Provide one safe, memorable `git wt` workflow for isolated Git issue and pull-request work while preserving in-flight primary checkouts and making the repository hierarchy predictable.

## CONTEXT

- The existing user-level `git wa` alias combines creation, pruning, forced removal, substring selection, and automatic `.env` symlinking in one opaque shell expression.
- Existing linked worktrees accumulate directly beneath `~/worktrees`, which hides repository ownership and makes similarly named lanes harder to scan.
- Kit-managed policy currently prohibits ordinary worktrees even when an unrelated dirty checkout makes a separate lane the safest way to preserve user work.
- A worktree has its own checkout, index, and `HEAD`, but shares refs, remotes, objects, configuration, and stash state with the other worktrees of the clone. Safe automation must respect both sides of that boundary.
- The accepted convention is `~/worktrees/<owner>/<repository>/<lane>`, with durable issue lanes named `GH-<number>` and temporary detached pull-request views named uppercase `PR-<number>`.
- The root checkout remains on `main`; each coding agent develops and tests in its assigned durable issue lane without checking that branch out in the root checkout.
- GitWT remains a thin wrapper around native Git worktree operations. Runtime services, databases, ports, Temporal state, and sibling repositories remain outside its scope.
- GitHub issue #100 and branch `GH-100` track the follow-up that annotates list
  rows with open pull request numbers without making GitHub availability a
  prerequisite for worktree navigation.
- GitHub issue #104 and branch `GH-104` track the follow-up that shares an
  ignored primary-checkout `.envrc` with writable lanes without requiring it to
  be committed or copied into every worktree.

## REQUIREMENTS

- Add a Kit-owned `git-wt` executable so standard Git external-command discovery makes it available as `git wt`.
- Make `make build` install or update `~/.local/bin/git-wt` from the same built artifact so source builds immediately refresh the Git subcommand on `PATH`.
- Default the hierarchy root to `~/worktrees`; allow a testable explicit override without requiring machine-specific paths in repository policy.
- Derive lowercase owner and repository segments from `origin`, preserve safe branch hierarchy below them, and reject absolute, empty, dot, or parent-traversal lane components.
- Provide a durable issue-lane command that creates or reuses exact `GH-<number>` from the freshly fetched remote default branch.
- Provide an existing-branch command that reuses a registered worktree, attaches a local branch, or creates a tracking branch from `origin`.
- Provide `PR-<number>` as a detached inspection lane fetched from the pull request head.
- Provide a repair command that resolves a same-repository pull request head and opens its durable branch worktree instead of editing the detached `PR-<number>` view.
- Provide read-only listing, exact safe removal, explicit pruning, root discovery, and dry-run-first migration of legacy flat linked worktrees.
- Make `git wt list` interactive by default when both input and output are terminals: render a colorized selector, support arrow keys and Tab for navigation, Enter for selection, and open a child shell in the selected worktree.
- Keep Git's primary worktree selectable and pinned at the top of `git wt list` by default, allow an explicit bottom pin, and make `h` open that home checkout immediately.
- Keep the primary worktree and every literal `main` branch row bright green even when selected, using one repository-independent identity color distinct from ordinary lane colors.
- Provide `git wt home` as read-only primary-worktree resolution followed by the same child-shell navigation used by `git wt cd`.
- Make terminal selection context cancellation interrupt idle input promptly, restore the original terminal mode and cursor state, and leave no background reader that can consume later input.
- Sanitize every dynamic terminal-table field before alignment and truncation so filesystem or Git metadata cannot inject terminal control sequences.
- Preserve script-safe listing when input or output is not a terminal and provide `--plain` to request the table explicitly from a terminal.
- Show `LAST UPDATED` as a human-readable calendar day plus `HH:MM` in the running user's local timezone, omit seconds and finer precision, and retain newest-first sorting by the full commit timestamp.
- Show `PR#` between `HEAD` and `LAST UPDATED` in both plain and interactive
  list output. Resolve same-repository open pull requests with one batched
  `gh` process invocation, match exact head branch names, and use `-` when the
  successful lookup finds no match. If one branch has multiple open pull
  requests, show their numbers in ascending comma-separated order.
- Bound list-time GitHub lookup to two seconds and keep every failure
  non-blocking. Use stable, distinct markers: `NG` when `gh` is unavailable,
  `RL` for rate limiting, `TO` for timeout, and `??` for any other lookup or
  decode failure.
- Provide a read-only `path <lane>` command that prints only the exact registered worktree path so callers can navigate with `cd "$(git wt path <lane>)"` without fuzzy matching or filesystem mutation.
- Keep Kit-distributed rules and generated agent instructions portable: native `git worktree` and ordinary filesystem operations define the normative workflow, and no rule may require `git-wt`, `git wt`, `--no-link-env`, or another wrapper-specific command.
- Document `git wt` only as an optional convenience for manual users. The wrapper may mirror the portable contract but must not define, direct, or become an execution dependency of Kit-managed rules.
- Removal must never use force, must refuse the current checkout, dirty state, ignored material, and local branch commits that are not present on the configured upstream.
- Migration must preflight every candidate and destination before applying, use `git worktree move`, preserve dirty contents, skip already hierarchical directories, and stop rather than overwrite or force through a conflict.
- Link the clone's primary checkout repository-root `.env` and `.envrc` into
  writable `issue`, `add`, and `repair` lanes by default when each source
  exists, using exact symlinks only and allowing explicit `--no-link-env`
  isolation for both.
- Reusing an existing writable lane must ensure each expected environment
  symlink exists when linking is enabled.
- Missing source `.env` or `.envrc` files must not prevent lane creation;
  report each omitted link.
- Never overwrite a destination `.env`, and refuse symlinks for either name
  whose target is not the matching primary-checkout source. Preserve a regular
  destination `.envrc` supplied by Git or the user instead of replacing it.
- Treat `.envrc` as executable shell configuration: link only the exact
  primary-checkout source, document review of that source, and retain direnv's
  path-specific approval boundary.
- Detached `pr` inspection lanes must not create environment links, and
  `migrate` must preserve existing files and links without creating new ones.
- Safe removal may ignore only verified GitWT-managed `.env` and `.envrc`
  symlinks, must unlink only those symlinks before ordinary non-force worktree
  removal, and must restore all removed links if removal fails.
- Do not add substring-based targeting, implicit pruning during listing, forced `nuke`, stash, reset, clean, or branch deletion.
- Project validation must not require ignored local-only `.env` or `.envrc` scaffold files in a linked checkout.
- Update canonical Kit rules, generated instruction sources, active checked-in guidance, prompts, and tests so managed projects may use worktrees only beneath `~/worktrees` with one active branch per worktree and without nesting them inside repositories.
- Keep subagents from independently creating, switching, moving, or removing worktrees; a supervisor may assign an already prepared worktree explicitly.
- Document the mental model, command map, naming rules, lifecycle, shared-state caveats, and PR-review workflow.
- Observable acceptance: focused integration tests exercise issue, branch, PR, repair, remove, prune, and dirty migration behavior; full Kit validation passes; the installed command replaces `git wa`; every legacy worktree is relocated with branch and dirty-state parity.
- Non-goals: reconciling every managed project immediately, bypassing direnv
  approval, starting or stopping applications, multi-repository runtime
  orchestration, database reset or snapshot behavior, port allocation, Temporal
  namespace management, process supervision, automatic root-checkout branch
  switching, supporting fork pull-request repair automatically, deleting
  branches, force-removing worktrees, moving standalone clones, or merging
  delivery pull requests.

## ACCEPTED PLAN

1. Pass a default-enabled environment-link option through the shared writable-lane path used by `issue`, `add`, and `repair`, while keeping `pr`, `migrate`, and `GIT_WT_ROOT` behavior unchanged.
2. Add a narrow symlink manager for `.env` and `.envrc` that reports missing
   sources, refuses unexpected symlinks, preserves an existing regular
   `.envrc`, and validates an existing lane's expected links.
3. Extend conservative removal to recognize only verified expected `.env` and
   `.envrc` symlinks, preserve all other dirty, ignored, and unpublished-state
   protections, and restore every removed link if native removal fails.
4. Add focused real-Git integration coverage plus command-help assertions for creation, reuse, dual opt-out, collisions, checked-in `.envrc` compatibility, safe removal, and no-copy semantics.
5. Align the canonical guide, README, command docs, Constitution, active registry rules, generated V3 support files, and current V3 instruction payload while preserving immutable V1 and V2 payloads.
6. Validate Kit fully, then update only LabCore's dedicated worktree-policy PR if its managed rules and V3 guidance require the new contract.
7. Add exact registered-lane path lookup for shell command substitution, document the parent-shell limitation, and register the command in Kit capabilities.
8. Refactor Kit-managed rules, generated instructions, and downstream LabCore guidance to express lane creation, reuse, environment linking, exact path validation, and removal with native Git and generic filesystem semantics; retain GitWT only in manual command documentation.
9. Resolve environment ownership from the clone's primary worktree rather than the invoking linked checkout, then prove that lanes created from another lane remain valid after the invoking lane is removed.
10. Promote the portable worktree guide into the V3 instruction support-document registry so `kit reconcile` creates and refreshes the workflow without changing immutable V1 or V2 payloads.
11. Move list behavior into a focused component, keep non-terminal output deterministic, and add a raw-terminal selector with color, arrow/Tab navigation, explicit cancellation, and child-shell entry for the chosen worktree.
12. Expose the primary worktree as `git wt home`, pin it independently of list sorting, add the `h` selector shortcut, and reserve bright green for primary/`main` identity.
13. Add one bounded open-pull-request resolver for list presentation, project
    exact branch matches into a `PR#` field, and preserve successful listing
    through explicit failure markers.
14. Cover successful, empty, missing-CLI, rate-limit, timeout, malformed, and
    generic failure paths without allowing tests to depend on live GitHub.
15. Align help, capabilities, command documentation, the canonical worktree
    guide and template, and the worktree-sync non-mutation boundary.

## DECISIONS

- Use an external `git-wt` binary instead of a Git alias so the workflow is testable, documented, and maintainable in Kit.
- Keep `PR-<number>` detached and inspection-only. Writable repair always targets the pull request's durable same-repository head branch.
- Make migration preview-only unless `--apply` is explicit.
- Make safe removal conservative. Manual intervention is preferable to losing ignored files, untracked work, or unpushed commits.
- Preserve arbitrary safe branch path components beneath owner and repository while reserving uppercase `GH-<number>` and `PR-<number>` identities for standard lanes.
- Make exact `.env` and `.envrc` symlinking the bounded writable-lane
  convenience. It shares primary-checkout configuration without copying
  credentials and remains explicitly disableable with `--no-link-env`.
- Preserve the earlier decision not to share `.envrc` as superseded by the
  user's explicit workflow requirement on issue #104. The replacement keeps
  the executable-config boundary narrow through exact-source validation,
  destination preservation, and direnv's separate path approval.
- Recognize removable environment links by exact destination name, symlink
  type, and target match only; do not add broad environment-file deletion or
  dirty-state exceptions.
- Keep application processes, databases, ports, Temporal state, and sibling repository coordination outside GitWT so the command remains a thin Git worktree wrapper.
- Preserve exact `path` output as the only way to change the invoking shell with `cd "$(git wt path GH-101)"`. The later `cd` command and interactive list selector intentionally open a child shell in the chosen worktree and must not claim to change their parent shell.
- Build `git-wt` once into `bin/` and install that artifact from the shared Make target so `make build`, `make install-git-wt`, and `make install` cannot diverge.
- Make native `git worktree` the policy authority. GitWT is a manually invoked convenience implementation of that policy, never a prerequisite for agents, reconciliation, or teammates using other tooling.
- Treat the primary worktree as the stable owner of shared `.env` and `.envrc`
  sources. An invoking linked lane is an ephemeral consumer and must never
  become another lane's environment source.
- Distribute the native worktree guide as a V3 support document rather than a ruleset: it is required operational reference material for current scaffolds, while the active safety and delivery rules remain the normative policy.
- Resolve “home” from the first entry in Git's porcelain worktree list. That entry is available offline and identifies the primary checkout without guessing from branch names or paths.
- Pin home after the requested list sort and reversal so it stays predictably selectable; `--root-position bottom` is the only option that moves the pinned row.
- Reserve bright green for both the primary checkout and literal `main` rows, even while selected. The pointer still communicates selection without replacing stable home/default-branch identity.
- Convert commit timestamps into the running user's local timezone only for `git wt list` presentation. Keep the full parsed instant for sorting and retain the existing sync report representation so list localization does not make sync JSON environment-dependent.
- Keep the optional pull-request annotation read-only and fail-soft rather
  than adding an opt-in flag: one batched lookup measured 0.37 seconds in Kit,
  and a two-second timeout bounds the new cost below the accepted three-second
  threshold.
- Use short ASCII markers instead of font-dependent glyphs so redirected
  tables, narrow terminals, logs, and tests retain stable width and meaning.

## DISCOVERIES

- The existing flat worktree inventory contains both clean and dirty linked worktrees; most branches have no upstream, so migration must not depend on clean or published state.
- Kit already resolves feature allocation through the shared Git common directory, confirming that worktrees share clone-level Git state even though their checkout state is isolated.
- Active no-worktree language exists in registry rules, generated tooling guidance, legacy instruction versions, dispatch/improvement prompts, and the Constitution.
- Published `kit instructions` versions are immutable. The historical `v1` and `v2` payloads retain their hashes and former no-worktree contract; new current `v3` carries the project-oriented policy.
- Git reserves `git <command> --help` for manual-page discovery before invoking an external subcommand. Inline discovery therefore uses `git wt help` or direct `git-wt --help`.
- A real dry run found 29 legacy linked worktrees with collision-free destinations: 25 dirty and 4 clean.
- The existing writable-lane implementation already converged issue, add, repair, and reuse behavior through `addBranch`; passing one default-enabled option through that path keeps the environment policy consistent without touching detached PR or migration flows.
- Git reports an ignored or untracked managed link as an exact `.env` porcelain entry. Filtering only that entry after verifying the symlink target preserves every existing dirty, ignored, and unpublished removal protection.
- GitWT must restore the original symlink text rather than recreate a normalized target when native worktree removal fails, so relative and absolute matching links retain their original representation.
- LabCore's live `GH-78` / PR `#79` is unrelated order-to-hold product work. Its dedicated worktree-policy lane is issue `#80`, branch `GH-80`, and PR `#81`; only that lane may receive downstream policy updates.
- Navigation must be implemented as path resolution because a child process cannot change its invoking shell's current directory.
- Kit registry rules and generated instruction sources had made the optional wrapper authoritative by naming `git wt`, GitWT, and `--no-link-env`. Expressing the same behavior with `git worktree list --porcelain`, `git worktree add`, `git worktree move`, non-force `git worktree remove`, and exact `ln -s` validation keeps reconciliation portable without weakening the contract.
- GitWT previously selected the invoking worktree as the `.env` source. A lane created from another linked lane therefore depended on that intermediate lane's lifetime and could be left with a broken link after conservative removal of the intermediate lane.
- LabCore's primary checkout had regular ignored `.env` and `.envrc` files,
  while writable lane `GH-90` had neither. `direnv exec .` in the primary
  checkout exported the three required service tokens, but the same command in
  `GH-90` exported none, proving the failure was missing lane links rather than
  a requirement to commit `.envrc`.
- V3 guidance routed agents to `docs/references/worktrees.md` only “when present,” but the file was absent from `instructions.SupportDocs`; `kit reconcile` therefore could not create or refresh the canonical workflow.
- Interactive navigation must be gated on both terminal input and terminal output. Redirected and piped invocations need stable plain text, while terminal users can safely receive raw input handling and ANSI color.
- Darwin pseudo-terminals do not support `os.File` read deadlines, and a goroutine-only cancellation wrapper can leave the underlying read blocked long enough to consume later terminal input. Context-aware readiness reads with exact descriptor-flag restoration avoid both failure modes.
- Git's porcelain worktree output lists the primary worktree first, so one parsed `primary` marker can drive `git wt home`, list pinning, the `h` shortcut, and identity coloring without network access.
- RFC 3339 commit dates retain the commit's recorded offset when parsed. An explicit `time.Local` conversion is required before formatting when `LAST UPDATED` must represent the user running the command rather than the commit author.
- A live `gh pr list --state open --limit 1000 --json
  number,headRefName` benchmark completed in 0.37 seconds while the existing
  local list completed in 0.36 seconds. Sequential execution therefore remains
  comfortably below three seconds in the current repository, and the timeout
  prevents slow GitHub access from becoming an unbounded navigation delay.

## VALIDATION

- `go test ./internal/worktree -count=1` passed integration coverage for issue lanes, existing branches, detached PR views, same-repository repair branches, safe removal, explicit pruning, and dirty migration.
- `make fmt` and `make vet` passed.
- `go test ./... -count=1` passed across every package.
- `go test -race ./internal/worktree ./internal/templates ./pkg/cli -count=1` passed.
- `golangci-lint run --new-from-rev=origin/main ./...` reported `0 issues`.
- `goreleaser check` validated the two-binary release configuration.
- `make build` produced both `bin/kit` and `bin/git-wt`; `make build-windows` produced both Windows executables.
- `./bin/kit improve run --suite prompt-system --kit-binary ./bin/kit --json` final run `20260723T212544.763147000Z-d1a63f` passed all 45 task runs and all 345 assertions with deterministic output across 15 repeated tasks.
- `./bin/kit check safe-worktree-workflow` passed.
- Focused reconcile tests proved primary checkouts still report missing local environment scaffold while linked checkouts do not require ignored `.env` or `.envrc` files.
- Pre-completion `./bin/kit check --project` first exposed the linked-checkout environment-file mismatch; after the validator fix and progress-summary refresh, the final project check passed coherently.
- `make install-git-wt`, `git wt help`, and `git wt root` passed; the installed binary and build artifact had identical SHA-256 values.
- Post-PR review regression coverage passed for dot-segment project identity rejection, offline local-lane reuse, linked-worktree versus submodule detection, narrow local-only scaffold filtering, `.envrc` policy, and V3 repository-memory completion guidance.
- `git wt migrate` previewed 29 collision-free moves. `git wt migrate --apply` moved all 29 with `git worktree move`; post-move verification matched every branch, `HEAD`, status count, and complete status hash, preserving all 25 dirty and 4 clean worktrees. A second preview reported no legacy flat linked worktrees.
- LabCore downstream validation ran `make check` successfully and passed explicit worktree-policy assertions. Its `kit check --project` remains blocked by pre-existing invalid reference relations in feature `0012` and `0013` plus pre-existing scaffold/progress warnings; no unrelated feature artifacts were changed.
- `git diff --check` passed in both repositories.
- Follow-up focused integration tests passed for default linking through issue, add, and repair; opt-out; missing sources; destination collisions; existing-lane reuse; detached PR isolation; migration preservation; matching-link removal; regular and unexpected destination refusal; restoration on Git removal failure; and help output.
- Follow-up `make fmt`, `go vet ./...`, and `go test ./... -count=1` passed across all packages.
- Follow-up race tests passed for `internal/worktree`, `internal/templates`, `internal/instructions`, and `pkg/cli`.
- Follow-up `golangci-lint run --new-from-rev=origin/main ./...` reported `0 issues`.
- Follow-up `make build`, `make build-windows`, and `goreleaser check` passed for both `kit` and `git-wt`.
- `./bin/git-wt help` showed `--no-link-env` for issue, add, and repair while leaving detached PR usage unchanged.
- Immutable V1 and V2 instruction payloads retained SHA-256 values `50cbfd80732e7b1912dc65f160cbf8555d2da95cb79079f33d7131cd51a86be5` and `811842c5c87a1b8c7f82831c7c76739071921583c44b0ab9c5dc62cbc08b27fc`.
- Follow-up `./bin/kit check safe-worktree-workflow`, `./bin/kit check --all`, and `./bin/kit check --project` passed; all 47 features and the project contract were coherent.
- Follow-up prompt-system run `20260724T143503.212928000Z-6e1a13` passed all 45 task runs and all 345 assertions with deterministic output across 15 repeated tasks.
- Exact-path integration coverage passed for durable issue lanes, nested branch lanes, unknown-lane refusal, invalid usage, and path-only output across filesystem-equivalent macOS path aliases.
- `make fmt`, `go vet ./...`, `go test ./... -count=1`, `go test -race ./internal/worktree ./pkg/cli -count=1`, and `golangci-lint run --new-from-rev=origin/main ./...` passed; lint reported `0 issues`.
- `make build`, `make build-windows`, and `goreleaser check` passed for both binaries and target platforms.
- The built command completed a real shell navigation check with `resolved_lane="$(./bin/git-wt path GH-78)"`, matched the active worktree, and confirmed branch `GH-78` after changing into the result.
- `make build` rebuilt both executables, installed `~/.local/bin/git-wt`, and produced identical SHA-256 values for `bin/git-wt` and the installed command.
- `./bin/kit capabilities --json git wt path`, `./bin/kit check safe-worktree-workflow`, and `./bin/kit check --project` passed.
- Prompt-system run `20260724T145815.411364000Z-d99cf1` passed all 45 task runs, all 345 assertions, and all 15 repeated-task determinism checks.
- Native-first follow-up tests passed for `internal/instructions`, `internal/templates`, and focused registry/capability assertions in `pkg/cli`; they require native `git worktree` authority and reject wrapper-specific policy in generated V3 guidance and active rules.
- `make fmt`, `make vet`, `go test ./... -count=1`, and `go test -race ./internal/instructions ./internal/templates ./pkg/cli -count=1` passed after the native-first policy refactor.
- `golangci-lint run --new-from-rev=origin/main ./...` reported `0 issues`, and `make build` rebuilt Kit and installed the optional `git-wt` manual wrapper from the validated artifact.
- LabCore branch `GH-80` passed `make check` after its guardrails, delivery rule, agent guidance, references index, and worktree guide were aligned to the native Git contract.
- Exact policy scans found no `GitWT`, `git wt`, or `--no-link-env` dependency in Kit or LabCore agent guidance, active registry rules, or reference indexes; those names remain only in manual wrapper documentation and Kit's command capability surface.
- `./bin/kit check safe-worktree-workflow`, `./bin/kit check --all`, and `./bin/kit check --project` passed; all 47 features and the project contract remained coherent after Constitution curation.
- Prompt-system run `20260724T152424.240242000Z-9863de` passed all 45 task runs, all 345 assertions, and all 15 repeated-task determinism checks.
- Primary-source regression coverage created a writable lane from another linked lane, removed the intermediate lane conservatively, and proved the second lane still read the primary checkout's `.env`.
- V3 registry, template, and init-refresh tests proved `docs/references/worktrees.md` is a required managed support document, while V2 support files remain unchanged.
- `make fmt`, `make vet`, and `go test ./... -count=1` passed after the stable-source and reconcile fixes.
- `go test -race ./internal/worktree ./internal/instructions ./internal/templates ./pkg/cli -count=1` passed.
- `golangci-lint run --new-from-rev=origin/main ./...` reported `0 issues`; `go build ./...` and `goreleaser check` passed.
- V1 and V2 instruction payloads retained SHA-256 values `50cbfd80732e7b1912dc65f160cbf8555d2da95cb79079f33d7131cd51a86be5` and `811842c5c87a1b8c7f82831c7c76739071921583c44b0ab9c5dc62cbc08b27fc`; current V3 is `a75fb2b02d37a7fbdc5926b9c71130210c6e929366b09707b410ab2f5b90792f`.
- LabCore branch `GH-80` passed `make check` after its rules and guide adopted the primary-checkout environment source.
- Final `git diff --check` passed in Kit and LabCore.
- Final `go run ./cmd/kit check safe-worktree-workflow`, `go run ./cmd/kit check --all`, and `go run ./cmd/kit check --project` passed; all 47 features and the project contract remained coherent.
- Interactive-list regression coverage passed for TTY selection and child-shell entry, `--plain` fallback, arrow and Tab key decoding, ANSI state/selection colors, supported sort flags, and human-readable day formatting.
- A real pseudo-terminal run moved from `GH-86` to `main` with the down arrow, opened a child shell, confirmed the selected worktree path with `pwd`, and returned cleanly after `exit`.
- `make fmt`, `make vet`, `go test ./... -count=1`, `go test -race ./internal/worktree ./pkg/cli -count=1`, and `golangci-lint run --new-from-rev=origin/main ./...` passed after the selector follow-up; lint reported `0 issues`.
- `go run ./cmd/kit capabilities git wt list --json`, `go run ./cmd/kit check safe-worktree-workflow`, `go run ./cmd/kit check --project`, `make build`, and `git diff --check` passed; the built and installed `git-wt` binaries had identical SHA-256 values.
- Review-repair coverage passed three consecutive idle-PTY cancellation runs, proving prompt return, raw-mode and cursor restoration, descriptor-flag restoration, and preservation of a post-cancellation sentinel; dynamic-field coverage rejected ESC, C1 CSI, CR, and LF injection before layout.
- Darwin package tests, race tests, and vet passed for `internal/worktree` and `pkg/cli`; the Windows worktree test binary compiled successfully, and the capability JSON now distinguishes full-timestamp default sorting from day-precision display.
- Issue `#95` follow-up coverage passed for `home` resolution from a linked lane, default top and explicit bottom pinning after alternate sort/reverse choices, `h`/`H` key decoding, stable primary/`main` bright-green selection color, capabilities, help, and exact guide/template parity.
- `make fmt`, `make vet`, `go test ./... -count=1`, `go test -race ./internal/worktree ./pkg/cli -count=1`, `make build-git-wt`, `git diff --check`, `kit check safe-worktree-workflow`, and `kit check --project` passed for the home-navigation follow-up.
- Built-binary smoke checks from linked worktree `GH-93` proved `git wt list --plain --sort path` pins the exact Git primary path first, `--root-position bottom` pins it last, and capability JSON advertises read-only home resolution, the `h` shortcut, and repository-independent bright-green identity.
- Local-time follow-up coverage proved date rollover from a commit offset into a different user zone and exact `Jan 02, 2006 15:04` rendering without seconds. Full tests, vet, race tests, feature/project checks, capabilities, build, and diff checks passed.
- Built-binary smoke output for the same `GH-93` commit was `Jul 26, 2026 17:44` under `America/New_York` and `Jul 26, 2026 21:44` under `UTC`, proving user-zone conversion at minute precision.
- Issue #100 focused coverage passed exact branch matching, ascending
  multi-PR rendering, successful no-match display, missing-`gh`, rate-limit,
  timeout, malformed-response, and generic-failure paths. All failure paths
  retained successful plain list output.
- The first complete test run exposed one new capability wording mismatch
  (`none` versus the asserted `no local mutation`); the capability text was
  corrected and the final `go test ./... -count=1` passed, including
  `internal/worktree` in 40.200 seconds and `pkg/cli` in 9.548 seconds.
- `make fmt`, `make vet`, `go test -race ./internal/worktree ./pkg/cli
  -count=1`, and `golangci-lint run --new-from-rev=origin/main ./...` passed;
  lint reported `0 issues` and the race packages passed in 43.037 and 12.302
  seconds.
- `make build-windows`, `goreleaser check`, and `make build` passed. The built
  and installed `git-wt` binaries shared SHA-256
  `a7bda569bb5e5dbb4a160e3488f41531abddb754b8405f65b6b70a411ddbc6a0`.
- A built-binary live lookup rendered the five-column
  `STATE HEAD PR# LAST UPDATED PATH` table and completed in 0.75 seconds. The
  same command under a PATH with Git but without `gh` still listed every row
  and rendered `NG`.
- `kit capabilities git wt list --json`, `kit check
  safe-worktree-workflow`, `kit check worktree-sync`, and `kit check
  --project` passed; the project contract remained coherent.
- After ready pull request #103 opened, the built command rendered
  `clean GH-100 103` in the live `PR#` column and completed in 0.92 seconds.
- Issue #104 focused integration coverage passed for default `.env` and
  `.envrc` linking through issue, add, and repair; dual opt-out; missing
  sources; existing-lane repair; checked-in `.envrc` preservation; unexpected
  and broken-link refusal; detached isolation; exact dual-link removal; and
  multi-link restoration after simulated Git removal failure.
- `make fmt`, `make vet`, `go test ./... -count=1`, and
  `go test -race ./internal/worktree ./internal/templates
  ./internal/instructions ./pkg/cli -count=1` passed. `golangci-lint run
  --new-from-rev=origin/main ./...` reported `0 issues`.
- `make build`, `make build-windows`, and `goreleaser check` passed. The built
  and installed `git-wt` binaries shared SHA-256
  `7a617bc547f52288a9dac4244c3fb57cd843ac394696cf22c67b0f10669edb48`.
- `kit check safe-worktree-workflow`, `kit check worktree-sync`, `kit check
  --all`, and `kit check --project` passed; all 50 features and the project
  contract were coherent.
- Built-binary reuse linked Kit lane `GH-104` to both primary-checkout
  environment sources. The same binary reused LabCore `GH-90`, verified both
  exact links, and `direnv exec .` reported `CUSTOMER_API_TOKEN`,
  `LABCORE_INTERNAL_TOKEN`, and `FLOWCORE_LABCORE_CALLER_TOKEN` as set while
  LabCore tracked status remained clean.
- `gitleaks git --redact --no-banner`, `git diff --check`, capability JSON
  generation, and exact worktree guide/template comparison passed.

## OUTCOME

- Added the Kit-owned `git-wt` executable with durable `GH-<number>` issue lanes, existing-branch reuse, detached `PR-<number>` views, writable PR-head repair, read-only listing, conservative exact removal, explicit pruning, canonical root discovery, and dry-run-first migration.
- Installed `git-wt` at `~/.local/bin/git-wt`, removed only the obsolete global `alias.wa`, and intentionally removed forced cleanup, substring targeting, and implicit list-time pruning from the workflow.
- Writable issue, add, repair, and existing-lane reuse now symlink the primary
  checkout's `.env` and `.envrc` by default when each exists, with explicit
  `--no-link-env` isolation for both; missing sources remain successful,
  existing regular `.envrc` files are preserved, and detached PR or migration
  flows create no links.
- Safe removal recognizes only exact expected `.env` and `.envrc` symlinks,
  preserves regular or unexpected destinations, retains every other dirty,
  ignored, and unpublished-state refusal, and restores all removed links if
  native non-force removal fails.
- GitWT remains limited to lane paths, branches, native worktree operations,
  and the bounded environment-link convenience; it does not orchestrate
  applications, databases, ports, Temporal state, processes, or sibling
  repositories.
- Migrated the live worktree root to lowercase owner/repository hierarchy while preserving each branch and dirty checkout exactly.
- Added immutable current agent instructions `v3`, generated/legacy template alignment, prompt boundaries, active registry policy, release/build/install support, and a practical worktree reference guide.
- Updated project validation to recognize Git-file linked checkouts and avoid pressuring them to recreate or share ignored environment files.
- Hardened project identity containment, preserved offline reuse of existing local lanes, and distinguished linked-worktree metadata from submodule `.git` files after review.
- Updated LabCore's active rules and guidance on issue `#80` and branch `GH-80` without changing its existing `GH-78` lane or reconciling any other managed project.
- Added read-only `git wt path <lane>` lookup for exact registered lanes, enabling portable navigation with `cd "$(git wt path GH-101)"` while rejecting unknown lanes, fuzzy matches, and traversal.
- Registered `git wt path` in Kit capabilities and updated help, command docs, the canonical worktree guide, and delivery guidance with path-based navigation.
- `git wt list` now opens a colorized terminal selector by default, supports arrow keys and Tab with Enter-to-open child-shell behavior, retains deterministic table output for pipelines and `--plain`, and displays last updates as local calendar days through `HH:MM` while sorting by the full commit timestamp.
- Idle selector reads are context-cancellable without leaving a competing terminal reader, and all dynamic selector fields are control-sanitized before alignment and truncation.
- Registered `git wt list` in Kit capabilities and documented its terminal, child-shell, plain-output, sorting, and local-minute-precision boundaries.
- Added `git wt home`; default top pinning with explicit `--root-position bottom`; immediate `h` navigation; and constant bright-green identity for primary and `main` rows.
- Tracked the home-navigation follow-up as issue `#95` and kept its scoped commit on existing branch `GH-93` / pull request `#94`.
- `make build` now installs or updates `~/.local/bin/git-wt` from the same `bin/git-wt` artifact used for validation.
- Kit rules, generated V3 instructions, and LabCore policy now use native `git worktree` plus ordinary filesystem operations as the portable authority. Reconciled guidance does not require the optional wrapper, its command names, or its flags.
- The canonical guides map creation, reuse, detached inspection, repair, exact path validation, environment linking, migration, pruning, and conservative removal to native Git. Kit's command docs and LabCore's optional manual section retain `git wt` as a convenience cheat sheet only.
- Writable lanes now resolve `.env` ownership from Git's primary worktree, so a lane created from another linked lane never depends on that intermediate lane's lifetime.
- V3 reconciliation now creates and refreshes the portable worktree guide as a managed support document; V1 and V2 payloads and V2 support-document membership remain unchanged.
- `git wt list` now shows `PR#` between `HEAD` and `LAST UPDATED`, resolves
  exact same-repository open pull requests with one two-second batched `gh`
  invocation, renders multiple matches in ascending comma-separated order, and
  uses `-` for a successful no-match result.
- Missing `gh`, rate limiting, timeout, malformed output, and other GitHub
  failures no longer affect list success; they render `NG`, `RL`, `TO`, and
  `??` respectively in both plain and interactive output.
- The pull-request annotation remains read-only, performs no fetch or Git
  mutation, and leaves explicit worktree synchronization solely with `git wt
  sync`.
- The follow-up is tracked by issue `#100`, branch `GH-100`, commit
  `2040f73a164c92f18a4a84e9bbbd50b790086d16`, and ready pull request
  <https://github.com/jamesonstone/kit/pull/103>.
- Issue `#104` and branch `GH-104` extend the same primary-checkout ownership
  model to ignored `.envrc` files, preserve existing destination material, and
  retain direnv's per-path approval without requiring `.envrc` to be committed.
- LabCore lane `GH-90` now has ignored exact `.env` and `.envrc` links to its
  primary checkout and a lane-specific direnv approval; no LabCore tracked file
  changed.

## REPOSITORY MEMORY

Decision: updated

Rationale: Exact `.envrc` source ownership, executable-config trust, collision
preservation, direnv approval, and multi-link removal restoration are durable
workflow and security decisions that code and tests alone do not explain
completely. The spec and canonical worktree guide preserve those decisions.

Constitution curation result: updated the project-wide worktree invariant so
writable lanes may share exact primary-checkout `.envrc` sources without
copying or requiring tracked files, while preserving destination material and
direnv's approval boundary.

Artifacts:

- `internal/worktree`
- `pkg/cli/capabilities_catalog_*.go`
- `internal/instructions/versions/v3.md`
- `internal/templates`
- `docs/CONSTITUTION.md`
- `docs/agents/GUARDRAILS.md`
- `docs/agents/TOOLING.md`
- `docs/references/rules/safety-guardrails.md`
- `docs/references/rules/github-pr-delivery.md`
- `docs/specs/0050-safe-worktree-workflow/SPEC.md`
- `docs/specs/0052-worktree-sync/SPEC.md`
- `docs/references/worktrees.md`
- `internal/templates/worktrees_reference.md`
- `docs/commands.md`
- `docs/man/git-wt.1`
- `docs/PROJECT_PROGRESS_SUMMARY.md`
