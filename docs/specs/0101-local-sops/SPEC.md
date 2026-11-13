---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "implement"
feature:
  id: "0101"
  slug: "local-sops"
  dir: "0101-local-sops"
---
# SPEC

## PURPOSE

Keep personal operating procedures available to agents on this machine without publishing them in Git.

## CONTEXT

Kit's shared SOPs live at `docs/sops/<slug>/SOP.md`. Git does not copy ignored files into linked worktrees. Kit's normal delivery work takes place in linked worktrees.

## REQUIREMENTS

- Init and reconcile add `/docs/sops/local/` to the root ignore file.
- Discover local SOPs from the primary checkout and shared SOPs from the current checkout. Use the existing format, applicability, and authorization rules.
- Preserve local SOP contents during reconcile and cleanup.

## ACCEPTED PLAN

Extend the existing ignore scaffold and SOP guidance. Document a Git-based search from a linked worktree and exercise it in a real fixture. Protect ignored local procedures from worktree removal.

## DECISIONS

Keep a single local copy per clone. Use Git's worktree list to locate the primary checkout. Keep discovery in agent guidance and use no executor, synchronization, or new CLI command.

## DISCOVERIES

Worktree cleanup distinguishes ignored local SOPs from regenerable build output. Primary sync refuses to overwrite ignored files because Git merges can overwrite them by default.

## VALIDATION

The real Git fixture verifies that the local file is ignored, absent from the
linked checkout, and found through the documented recipe. Cleanup fixtures
verify that a worktree with ignored SOP content is retained and that a primary
checkout fast-forward preserves a conflicting local file. Init and reconcile
tests cover the ignore entry and preservation of existing guides and SOPs.

## OUTCOME

GH-290 implements the Git ignore scaffold, linked-worktree discovery, and
cleanup guards. Local files remain with their clone.

## REPOSITORY MEMORY

The SOP guide and contextual rules carry the workflow. This spec records the design.
