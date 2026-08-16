package templates

const workLaneMutationRoutingGate = `## Work Lane Mutation Hard Gate

- Before any coding-agent repository file or delivery mutation, including issue, branch, staging, commit, push, worktree, and pull-request mutations, load ` + "`docs/agents/GUARDRAILS.md`" + ` and ` + "`work-lane-gating`" + ` first, complete read-only safety recon, then ask exactly: "Before I make any repository changes, should I create a new GitHub issue, GH-<issue-number> branch, canonical worktree, and pull request for this work, or continue in the existing branch/worktree and land it through that branch's pull request?"
- Interpret a leading standalone response token case-insensitively: ` + "`c`" + ` means continue existing; ` + "`n`" + ` or ` + "`y`" + ` means new lane. In a longer response, shorthand is the primary lane choice and the remaining text is supplemental lane instructions. Full-form choices remain valid; ambiguous or contradictory responses fail closed.
- Wait for the explicit choice and record a Pull-Request Landing Plan covering the repository, issue, branch, canonical non-primary worktree, protected base, and create-or-update PR target. Verify that plan still matches before every mutation. Never infer the choice from clean state or a generic PR request.
- Treat the primary/root checkout as read-only. If an ungated or root change exists, preserve it: Do not stage, commit, push, stash, reset, clean, discard, or silently transfer it.

`

const workLaneMutationHardGate = `## Work Lane Mutation Hard Gate

Before a coding agent performs any repository file or delivery mutation, it
must:

1. Load ` + "`docs/agents/GUARDRAILS.md`" + ` and
   ` + "`docs/references/rules/work-lane-gating.md`" + `.
2. Complete read-only safety recon, including the current branch, dirty state,
   remote, active pull requests, registered worktrees, and exact primary
   checkout.
3. Ask exactly:

   > Before I make any repository changes, should I create a new GitHub issue, ` + "`GH-<issue-number>`" + ` branch, canonical worktree, and pull request for this work, or continue in the existing branch/worktree and land it through that branch's pull request?

   Interpret the response's first standalone token after trimming surrounding
   whitespace, case-insensitively: ` + "`c`" + ` means continue existing, while
   ` + "`n`" + ` or ` + "`y`" + ` means new lane. When shorthand leads a longer response,
   shorthand is the primary lane choice and the remaining text is supplemental
   lane instructions. Continue accepting explicit full-form choices; ambiguous
   or contradictory responses fail closed.
4. Wait for the user's explicit choice unless that exact choice is already
   recorded for the same unit of work.
5. Record a Pull-Request Landing Plan with the repository, issue, branch,
   non-primary worktree, protected base, and create-or-update PR target, then
   verify that plan still matches before every mutation.

- This gate covers source, tests, documentation, specs, plans, generated files,
  configuration, and every other repository file. It also covers every delivery
  mutation, including issue, branch, staging, commit, push, worktree, and
  pull-request mutations, as well as merges. Read-only discovery and planning
  may precede it; write-capable commands such as ` + "`kit spec`" + `,
  ` + "`kit init`" + `, and ` + "`kit reconcile`" + ` may not.
- Never infer the choice from a clean default branch, an issue reference, or a
  generic request to produce a pull request.
- For a new lane, create or reuse one human-assigned issue, exact
  ` + "`GH-<issue-number>`" + ` branch, canonical linked worktree at
  ` + "`~/worktrees/<owner>/<repository>/GH-<issue-number>`" + `, and one ready-PR plan
  before editing files.
- Continue existing work only after proving the non-protected branch, its exact
  owning linked worktree, issue scope, protected base, and create-or-update PR
  target. Reuse an existing pull request; do not create a second delivery lane.
- Treat the clone's primary/root checkout as read-only for coding-agent work,
  regardless of branch or cleanliness. Never edit there with a plan to move the
  diff later.
- One choice covers directly required tests, documentation, validation fixes,
  and delivery. Ask again for materially new or tangential scope.
- If an ungated or primary-checkout change is detected, stop and preserve it.
  Do not stage, commit, push, stash, reset, clean, discard, or silently transfer
  it; follow ` + "`work-lane-gating`" + ` recovery.

---

`
