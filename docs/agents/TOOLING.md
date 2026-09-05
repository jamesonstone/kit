# Tooling

## Kit Evidence Sequence

- Use `kit capabilities <command> --json` when side effects are not already established.
- Resolve `kit context resolve --workflow <slug> --json` before coding-agent work and load the selected local evidence.
- Rerun resolution after material scope changes; never treat resolved JSON as a new source of truth.

## Skills

- Repo-local canonical skills live under `.agents/skills/*/SKILL.md`
- For feature-scoped work, start with the current feature's canonical front matter `skills`, falling back to the legacy `SPEC.md` `## SKILLS` table only when front matter is absent
- Keep the selected skill set minimal and actionable

## Command Capability Discovery

- Use `kit capabilities` when choosing among Kit commands and the mutation, network, write, or git behavior is not already obvious.
- Use `kit capabilities <command> --json` for one command path, including nested paths such as `rules add` or `context resolve`.
- Use `kit capabilities --search <term> --json` for compact filtered discovery, and `kit capabilities --full --json` only when full details are needed.
- Treat `kit capabilities` itself as read-only: it does not require a Kit project root and does not load project config, write files, call the network, run subprocesses, or mutate git.
- In downstream Kit-managed projects, load `docs/references/rules/kit-capabilities-usage.md` when command discovery affects the task.
- Downstream projects should use `kit capabilities` for command discovery; do not maintain Kit's internal command catalog from a downstream project.

## Dispatch

- Use `kit dispatch` after native planning when an accepted plan needs a safe multi-lane execution topology
- Load `docs/references/rules/agent-team-orchestration.md` when dispatch, direct subagent execution, or read-only verification topology affects the task
- When `cross-repository-program-coordination` applies, dispatch only the canonical program ledger's reconciled ready frontier and checkpoint program state after each material transition or handoff
- Keep one accountable supervisor responsible for scope, integration, validation, evidence, delivery gating, and final reporting
- Use subagents when the work cleanly separates into low-overlap lanes after discovery
- Keep single-lane work in one supervisor lane when the task is trivial, tightly coupled, high-overlap, high-ambiguity, cannot spawn subagents, or the user requested single-agent execution
- Let the host govern concurrency; never invent a static numeric cap
- Keep broad or noisy discovery in RLM first; use dispatch or direct subagent execution only after the relevant workstreams are narrow enough to predict overlap
- Predict overlap conservatively before parallelizing
- Use read-only verification subagents by default after implementation unless a recorded exception applies

## Capability-Aware Host Adapter

Use `docs/references/rules/agent-team-orchestration.md` for the canonical topology and lifecycle. This adapter translates semantic profiles into controls confirmed by the active host; host and model examples are illustrative, not fixed requirements.

| Profile | Capability target |
| --- | --- |
| `architect` | strongest justified architecture and systems reasoning |
| `orchestrator` | decomposition, dependency, risk, and overlap analysis |
| `mapper` | efficient read-heavy discovery and codebase mapping |
| `specialist` | bounded implementation with established contracts |
| `precision` | high-risk or ambiguous work requiring deeper judgment |
| `verifier` | fresh independent review with enough strength to challenge the implementation |

### Capability Negotiation

- Before delegation, inspect the host's confirmed child-launch, per-child model and effort, same-agent follow-up, wait or join, live roster or capacity, and fresh-verifier controls. Record confirmed absence as unavailable and unexposed controls as `unknown`.
- Let the host govern concurrency. When parallel capacity is unconfirmed, do not invent a numeric cap or launch a child to probe it; a confirmed child primitive may still be used serially.
- If the host exposes no child primitive, use one supervisor with logical lanes. Never report a role prompt, task list, handoff, or manually opened conversation as a child.
- Resolve profiles against the live roster. Fall back to an equal-or-stronger eligible configuration, then a narrower low-risk lane with stronger verification, then a runtime-selected configuration reported as unverified; otherwise report `BLOCKED`. An unavailable exact user model or effort pin remains blocked until the user changes it.
- If capacity changes or a spawn fails, keep unadmitted lanes pending, preserve accepted work, recompute the ready frontier from confirmed capacity, and report the failure and degradation. Do not retry indefinitely or let children create descendants.
- Report actual and logical lanes separately, requested and effective profiles plus model and effort when exposed, confirmed or unknown parallelism, continuity or replacement, verifier independence, and every fallback.

### Illustrative Current Mappings

| Host | Example profile mapping |
| --- | --- |
| Codex | strongest justified live configuration for `architect` and `precision`; balanced read-heavy configuration for `orchestrator` and `mapper`; fast configuration for bounded `specialist` work; fresh strong configuration for `verifier` |
| Claude Code | Opus-class for architecture, precision, and strong verification; Sonnet-class for orchestration, mapping, and general implementation; Haiku-class only for narrow low-risk bounded work |
| GitHub Copilot | request semantic profiles conservatively and record the effective model or agent controls the host confirms; do not infer child continuation, parallelism, or per-child selection |
| Warp/Oz | use native parent-child orchestration, continuation, and per-child model controls only when exposed; parallelism and admission remain host-owned |
| Unknown or single-agent host | serialize logical lanes in the supervisor and perform a distinct self-review, reporting that no child or independent verifier was confirmed |

Current provider references: [Codex subagents](https://learn.chatgpt.com/docs/agent-configuration/subagents), [Claude Code subagents](https://code.claude.com/docs/en/sub-agents), [GitHub Copilot custom-agent invocation](https://docs.github.com/en/copilot/how-tos/copilot-cli/use-copilot-cli/invoke-custom-agents), [Warp orchestration](https://docs.warp.dev/platform/orchestration/), and [Warp project rules](https://docs.warp.dev/agents/capabilities/rules/). Treat these as evidence sources to re-check, not as pinned capability promises.

## PR Review Feedback

- Use `kit pr fix` as the default PR review feedback entrypoint when current PR review feedback should become a copied dispatch prompt.
- With no `--pr`, `kit pr fix` lists open pull requests in the current repository and asks which one to repair.
- Use `kit pr fix --pr <target>` when the PR is known; accepted targets match dispatch PR intake: URL, Markdown link, `owner/repo#number`, or current-repo number.
- `kit pr fix` uses the prompt-producing `kit dispatch --pr` path and copies the resulting dispatch prompt directly for a coding agent.
- Before producing a PR repair prompt, Kit resolves the same-repository PR head and reuses or creates its exact writable worktree; the user does not need to navigate to that lane first.
- If the resolved lane is dirty, Kit asks whether those changes belong in the repair and records `include` or `exclude`, the porcelain status, remote/local head SHAs, and the exact push target in the prompt.
- Pass `--edit` to review and change the task list in the default editor before copying; `--vim` and `--editor <cmd>` also opt into editing.
- The generated PR-fix prompt requires a post-push reflection cycle before review-thread resolution: the coding agent must review the pushed diff in context, confirm the PR head still matches the commit it pushed, and only then resolve verified addressed conversations.
- `kit pr fix` remains prompt-producing: except for preparing the writable worktree and its exact `.env` and `.envrc` links when needed, it does not launch an agent, edit source files, stage, commit, push, post PR comments, or resolve review threads.
- Use `kit dispatch --loop --pr <target> --watch` only for bounded expected CodeRabbit intake; waiting is deterministic and model-free.
- Use `kit dispatch --pr <target> --coderabbit` only when you need raw unresolved CodeRabbit review-thread intake without review-loop watch, classification, or summary behavior.
- After fixes or no-op decisions are complete, validation has run, the repair is pushed, and reflection confirms no other code was pushed after the repair commit, resolve matching current unresolved review threads on the PR, including human reviewer and CodeRabbit feedback, with `kit dispatch --pr <target> --resolve --yes`.
- Resolve only feedback verified as fixed or intentionally no-op; do not resolve unfixed, uncertain, stale, or unrelated feedback.
- `kit dispatch --pr <target> --resolve --yes` is an explicit GitHub mutation and must not be run speculatively.

## PR Merge

- Standing merge authority exists only when a human explicitly authorizes a bounded task, goal, or program to merge its resulting work. Generic task acceptance does not create it.
- Standing authority may bind later-created in-scope PRs and refreshed heads. Resolve the exact current PR and head before mutation; do not ask again solely because its number or final OID was unknown when authority was granted.
- A commit SHA or head OID identifies readiness evidence only; it is never an authorization identity. Never request exact-head reauthorization.
- Before any merge or merge-queue mutation, resolve `pull-request-merge` and load `docs/references/rules/github-pr-merge.md`.
- Merge only exact current `MERGE_READY` nodes from the standing-authority frontier; revalidate actor, head/base, repository policy, checks, dependencies, deployment workflow, environment, and material effects before every wave.
- A changed in-scope head invalidates readiness, not standing authority. Scope, repository, base, environment, actor, identity, method, workflow, or material-effect expansion requires explicit updated authority.
- After final-head evidence restores `MERGE_READY`, continue any already-authorized standard deployment and browser retry without another permission prompt.
- IAM, network, KMS, secrets, database-schema or data-loss changes, infrastructure creation/replacement/deletion, destructive deletion, nonstandard deployment effects, and unresolved risk classifications are outside standing merge/deploy authority.
- The most recent direct human instruction wins. Pause, hold, or revocation stops affected actions and dependents until explicit human resume or replacement authority.
- Report merge, hosted workflow, deployment/runtime, and production evidence separately.

## PR Release Orchestration

- Use `kit pr orchestrate` to turn explicit repository scope into a deterministic dependency-aware release prompt; Kit does not enumerate PRs, merge, deploy, mutate infrastructure, or launch an agent.
- Use repeatable `--repos` for exact repositories or `--root` for the root and immediate child repositories. Noninteractive use requires one of those scope flags.
- Use `--dry-run` to inspect resolved configuration provenance and the full prompt without clipboard access.
- Release agents resolve `release-orchestration` and, before any authorized merge, `pull-request-merge`. Preserve `MERGE_READY`, `BLOCKED`, and `UNKNOWN` literally.

## Project Worktrees

- Default to a new worklane without asking before any coding-agent repository
  file or delivery mutation; continue an existing lane only when the user
  explicitly directs it for the same unit of work
- Treat exact existing-PR review repair, CI repair, base refresh, and ordered
  merge coordination as continuation; reuse every targeted head and never
  create a coordinator or corrective pull request for scope-preserving work
- Record the repository, issue, branch, non-primary worktree, protected base,
  and create-or-update pull-request target before editing files
- Work in the existing checkout only when it is the exact non-primary linked
  worktree that owns the explicitly continued lane
- For a separate lane, reuse or create `~/worktrees/<owner>/<repository>/<lane>`; never put a worktree inside a repository
- Use exact `GH-<number>` for durable issue lanes and uppercase detached `PR-<number>` only for temporary pull-request inspection
- Reuse the pull request head branch for writable repair; never edit the detached `PR-<number>` view
- For target-bearing Kit repair commands, let Kit resolve the exact PR or branch lane from command context; do not require the user to navigate to it first
- When Kit reports a dirty target lane, ask whether its existing changes belong in the repair and carry that explicit include-or-exclude decision into the generated agent instructions
- Use native `git worktree` commands as the portable authority for creation, reuse, detached inspection, repair, removal, pruning, and migration; do not require `git-wt`, an alias, or another wrapper
- Optional wrappers are manual conveniences only and must preserve the same path and safety contract
- Treat the primary/root checkout as read-only for coding-agent work regardless
  of branch or cleanliness; never edit there with a plan to move the diff later
- Work directly in the user-selected durable lane and re-open the choice before
  materially new or tangential scope
- Do not stash, reset, clean, force-remove, or delete a branch to create or clear a worktree
- Link the primary checkout's `.env` and `.envrc` into writable lanes by default when each exists, using only exact verified symlinks; omit both links when isolation is required
- Never copy environment contents or overwrite destination environment material; preserve a repository- or user-supplied `.envrc`, and remember that direnv approval remains path-specific; worktree tooling does not manage runtime services, databases, ports, Temporal state, processes, or sibling repositories
- Remember that refs, remotes, objects, configuration, and stash state are shared across worktrees even though checkout, index, and `HEAD` are separate
- Load `docs/references/worktrees.md` when worktree creation, repair, migration, or removal affects the task

## Secondary Global Inputs

- `~/.claude/CLAUDE.md`
- `${CODEX_HOME}/AGENTS.md`
- `${CODEX_HOME}/instructions.md`
- `${CODEX_HOME}/skills/*/SKILL.md`

- Treat these as secondary context after repo-local docs
- Do not use `.claude/skills` as canonical discovery input
