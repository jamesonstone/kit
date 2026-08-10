Act as the one accountable supervisor for this pull-request feedback repair.

## Resolve the repository contract

Run this local-only, read-only command before planning or editing:

```bash
kit contract resolve --workflow pr-feedback-repair --json
```

Read the returned workflow, all dependencies, and repository-local RLM routing. Kit generated this prompt and prepared the lane; it did not launch agents or perform repairs.

## Pinned repair lane

| Field | Pinned value |
| --- | --- |
| Pull request | `https://github.com/acme/app/pull/9` |
| Repository | `acme/app` |
| Head branch | `GH-9` |
| Expected remote head | `abc123` |
| Local head | `abc123` |
| Repair worktree | `/tmp/acme/app/GH-9` |
| Push target | `origin/GH-9` |
| Dirty ownership | `exclude` |

- Run every filesystem and Git operation in the recorded repair worktree.
- Before editing, recheck the PR head, local HEAD, branch, worktree registration, push target, and dirty status. Stop if any pinned identity changed.
- Never repair from a detached `PR-<number>` lane or create a second repair branch or pull request.
- Existing dirty changes were explicitly marked `exclude` for this repair.
  - `local.txt`
- Preserve excluded paths without editing or staging them; stop if repair scope overlaps them.

## Current active feedback

### Finding 1

- Kind: review-thread
- Review thread: PRRT_1
- Source: internal/app.go:12
- Author: reviewer
- URL: https://example.com/thread
- Fingerprint: sha256:one

```text
Fix the valid issue.
```

## Agent Team Plan

- Publish an Agent Team Plan before spawning. Use at most 3 independent low-overlap concurrent lanes; the hard ceiling is 4.
- One supervisor owns scope, finding disposition, lane assignment, integration, validation, delivery, reflection, and thread resolution.
- Serialize shared or ambiguous files and queue excess work. Subagents may not create, switch, move, or remove worktrees or mutate Git/GitHub delivery state.
- After nontrivial repair, use a separate read-only verification lane. If no agent is actually spawned, report `single supervisor lane; no specialist or verification agents spawned`.

## Repair and delivery contract

1. Verify every finding against current `HEAD`, its current path and line, and the integrated implementation. Fix only still-valid findings.
2. Record an evidence-based disposition for every item: fixed, stale, false-positive, out-of-scope, or human-needed. Do not silently drop feedback.
3. Keep changes minimal, integrate every repair in the supervisor lane, and validate the complete combined diff under repository rules.
4. Explicitly stage intended files and push one coherent batch only to the recorded existing PR branch after the delivery gate allows it.
5. Verify the exact pushed commit equals the remote PR head. Re-read the findings, review the full pushed diff, reflect, and rerun required validation.
6. Resolve only current unresolved, non-outdated review threads that are verified addressed by that pushed head. Leave stale, partial, human-needed, and non-thread feedback visible.
7. Stop after two head epochs or two repair passes and report remaining feedback instead of creating an infinite loop.

## Explicit thread resolution

After exact pushed-head verification, select only verified addressed IDs from this list:
- `PRRT_1`

Remove every unaddressed ID, replace `PUSHED_HEAD_SHA`, then run explicitly:

```bash
kit pr fix --pr "https://github.com/acme/app/pull/9" --resolve --head PUSHED_HEAD_SHA --yes --thread "PRRT_1"
```

## Completion report

Report the Agent Team Plan and actual agents used, every finding disposition, files changed or preserved, validation and full-diff review, pushed commit and exact-head proof, reflection, explicitly resolved thread IDs, remaining feedback, and any blocked or human-needed action.
