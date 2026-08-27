package templates

const workLaneMutationRoutingGate = `## Work Lane Mutation Hard Gate

- Before any coding-agent repository file or delivery mutation, including issue, branch, staging, commit, push, worktree, and pull-request mutations, load ` + "`docs/agents/GUARDRAILS.md`" + ` and ` + "`work-lane-gating`" + ` first and complete read-only safety recon.
- Default to a new worklane without asking for the accepted unit of work: create or reuse one human-assigned GitHub issue, exact ` + "`GH-<issue-number>`" + ` branch, canonical non-primary worktree, and ready pull-request plan. Reuse that recorded lane for subsequent in-scope mutations. A clean or dirty checkout, current feature branch, issue reference, or generic pull-request request does not change this default.
- Continue an existing lane only when the user explicitly directs that outcome for the same unit of work. Prove the non-primary owning worktree, branch, issue scope, protected base, and create-or-update pull-request target.
- Never offer or ask the user to choose between lanes.
- Record a Pull-Request Landing Plan covering the repository, issue, branch, canonical non-primary worktree, protected base, and create-or-update PR target. Verify that plan still matches before every mutation. Ask only when implementation intent or an explicitly named target is materially ambiguous and cannot be resolved from repository evidence.
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
3. When no complete lane is recorded for the accepted unit of work, apply the
   default. Default to a new worklane without asking. Create or reuse one
   human-assigned GitHub issue, exact ` + "`GH-<issue-number>`" + ` branch, canonical non-primary worktree,
   and ready pull-request plan. Reuse that recorded lane for subsequent
   in-scope mutations.
4. Continue an existing lane only when the user explicitly directs that
   outcome for the same unit of work. Prove its non-protected branch, exact
   owning linked worktree, issue scope, protected base, and create-or-update
   pull-request target. Never offer or ask the user to choose between lanes.
5. Record a Pull-Request Landing Plan with the repository, issue, branch,
   non-primary worktree, protected base, and create-or-update PR target, then
   verify that plan still matches before every mutation.

- This gate covers source, tests, documentation, specs, plans, generated files,
  configuration, and every other repository file. It also covers every delivery
  mutation, including issue, branch, staging, commit, push, worktree, and
  pull-request mutations, as well as merges. Read-only discovery and planning
  may precede it; write-capable commands such as ` + "`kit spec`" + `,
  ` + "`kit init`" + `, and ` + "`kit reconcile`" + ` may not.
- The new-worklane default applies on clean and dirty checkouts, protected and
  feature branches, issue references, and generic requests to produce a pull
  request. Do not infer permission to continue the current lane from any of
  those states.
- For the default new lane, create or reuse one human-assigned issue, exact
  ` + "`GH-<issue-number>`" + ` branch, canonical linked worktree at
  ` + "`~/worktrees/<owner>/<repository>/GH-<issue-number>`" + `, and one ready-PR plan
  before editing files.
- Continue existing work only after an explicit user direction and after proving the non-protected branch, its exact
  owning linked worktree, issue scope, protected base, and create-or-update PR
  target. Reuse an existing pull request; do not create a second delivery lane.
- Treat the clone's primary/root checkout as read-only for coding-agent work,
  regardless of branch or cleanliness. Never edit there with a plan to move the
  diff later.
- One lane allocation covers directly required tests, documentation,
  validation fixes, and delivery. Materially new or tangential scope defaults
  to another new lane once its implementation intent is accepted.
- Ask only to clarify implementation intent or a user-named target that is
  materially ambiguous and cannot be resolved from repository evidence. Never
  ask for a new-versus-existing lane preference.
- If an ungated or primary-checkout change is detected, stop and preserve it.
  Do not stage, commit, push, stash, reset, clean, discard, or silently transfer
  it; follow ` + "`work-lane-gating`" + ` recovery.

---

`
