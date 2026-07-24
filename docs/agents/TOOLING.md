# Tooling

## Skills

- Repo-local canonical skills live under `.agents/skills/*/SKILL.md`
- For feature-scoped work, start with the current feature's canonical front matter `skills`, falling back to the legacy `SPEC.md` `## SKILLS` table only when front matter is absent
- Keep the selected skill set minimal and actionable

## Command Capability Discovery

- Use `kit capabilities` when choosing among Kit commands and the mutation, network, write, or git behavior is not already obvious.
- Use `kit capabilities <command> --json` for one command path, including nested paths such as `rules add` or `skill mine`.
- Use `kit capabilities --search <term> --json` for compact filtered discovery, and `kit capabilities --full --json` only when hidden or deprecated compatibility commands matter.
- Treat `kit capabilities` itself as read-only: it does not require a Kit project root and does not load project config, write files, call the network, run subprocesses, or mutate git.
- In downstream Kit-managed projects, load `docs/references/rules/kit-capabilities-usage.md` when command discovery affects the task.
- Downstream projects should use `kit capabilities` for command discovery; do not maintain Kit's internal command catalog from a downstream project.

## Dispatch

- Use `kit dispatch` after native planning when an accepted plan needs a safe multi-lane execution topology
- Load `docs/references/rules/agent-team-orchestration.md` when dispatch, direct subagent execution, or read-only verification topology affects the task
- Keep one accountable supervisor responsible for scope, integration, validation, evidence, delivery gating, and final reporting
- Use subagents when the work cleanly separates into low-overlap lanes after discovery
- Keep single-lane work in one supervisor lane when the task is trivial, tightly coupled, high-overlap, high-ambiguity, cannot spawn subagents, or the user requested single-agent execution
- Default to at most 3 concurrent lanes; never exceed 4
- Keep broad or noisy discovery in RLM first; use dispatch or direct subagent execution only after the relevant workstreams are narrow enough to predict overlap
- Predict overlap conservatively before parallelizing
- Use read-only verification subagents by default after implementation unless a recorded exception applies

## PR Review Feedback

- Use `kit pr fix` as the default PR review feedback entrypoint when current PR review feedback should become a copied dispatch prompt.
- With no `--pr`, `kit pr fix` lists open pull requests in the current repository and asks which one to repair.
- Use `kit pr fix --pr <target>` when the PR is known; accepted targets match dispatch PR intake: URL, Markdown link, `owner/repo#number`, or current-repo number.
- `kit pr fix` uses the prompt-producing `kit dispatch --pr` path and copies the resulting dispatch prompt directly for a coding agent.
- Pass `--edit` to review and change the task list in the default editor before copying; `--vim` and `--editor <cmd>` also opt into editing.
- The generated PR-fix prompt requires a post-push reflection cycle before review-thread resolution: the coding agent must review the pushed diff in context, confirm the PR head still matches the commit it pushed, and only then resolve verified addressed conversations.
- `kit pr fix` does not run the loop agent, edit files, write `.kit/loops` evidence, stage, commit, push, post PR comments, or resolve review threads.
- Use `kit loop review` when changed code should be locally reviewed and repaired by the configured loop agent until the final response reports at least 95% correctness and ends with `done`.
- Without `--pr`, `kit loop review` reviews current-branch changes relative to `origin/main`, falling back to local `main`, plus staged and unstaged changes.
- Use `kit loop review --pr <target>` when current unresolved CodeRabbit PR feedback should be opportunistically folded into the repair loop while local review starts immediately.
- Use `kit loop review --pr <target> --watch` or `--wait-for-coderabbit` only when finalization should block for CodeRabbit completion.
- Review prompts use one agent by default; pass `--subagents` to let the parent review agent pre-analyze the diff and choose subagents only when the lanes are clearly independent under `agent-team-orchestration` limits.
- Use `kit dispatch --loop --pr <target>` when current unresolved CodeRabbit PR review feedback should become a human-reviewed dispatch prompt instead of an agent repair loop.
- Use `kit dispatch --pr <target> --coderabbit` only when you need raw unresolved CodeRabbit review-thread intake without review-loop watch, classification, or summary behavior.
- Treat `kit loop review` as local repair only: it may edit files through the configured agent and write `.kit/loops` evidence, but it must not stage, commit, push, post PR comments, or resolve review threads.
- After fixes or no-op decisions are complete, validation has run, the repair is pushed, and reflection confirms no other code was pushed after the repair commit, resolve matching current unresolved review threads on the PR, including human reviewer and CodeRabbit feedback, with `kit dispatch --pr <target> --resolve --yes`.
- Resolve only feedback verified as fixed or intentionally no-op; do not resolve unfixed, uncertain, stale, or unrelated feedback.
- `kit dispatch --pr <target> --resolve --yes` is an explicit GitHub mutation and must not be run speculatively.

## Project Worktrees

- Work in the existing checkout when it already owns the requested lane
- For a separate lane, reuse or create `~/worktrees/<owner>/<repository>/<lane>`; never put a worktree inside a repository
- Use exact `GH-<number>` for durable issue lanes and uppercase detached `PR-<number>` only for temporary pull-request inspection
- Reuse the pull request head branch for writable repair; never edit the detached `PR-<number>` view
- Use native `git worktree` commands as the portable authority for creation, reuse, detached inspection, repair, removal, pruning, and migration; do not require `git-wt`, an alias, or another wrapper
- Optional wrappers are manual conveniences only and must preserve the same path and safety contract
- Keep the root checkout on the protected default branch and work directly in the assigned durable lane
- Do not stash, reset, clean, force-remove, or delete a branch to create or clear a worktree
- Link the primary checkout's `.env` into writable lanes by default when it exists, using only an exact verified symlink; omit the link when isolation is required
- Never copy `.env` contents or automatically share `.envrc`; worktree tooling does not manage runtime services, databases, ports, Temporal state, processes, or sibling repositories
- Remember that refs, remotes, objects, configuration, and stash state are shared across worktrees even though checkout, index, and `HEAD` are separate
- Load `docs/references/worktrees.md` when worktree creation, repair, migration, or removal affects the task

## Secondary Global Inputs

- `~/.claude/CLAUDE.md`
- `${CODEX_HOME}/AGENTS.md`
- `${CODEX_HOME}/instructions.md`
- `${CODEX_HOME}/skills/*/SKILL.md`

- Treat these as secondary context after repo-local docs
- Do not use `.claude/skills` as canonical discovery input
