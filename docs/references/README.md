# References

## Purpose

- This directory holds durable repo-local references that are broader than one feature
- `rules/<slug>.md` holds contextual rules shipped with the Kit binary; the universal contract's Contextual Rules section says when to read each one, and `kit rules list` is the live inventory
- `testing.md` records Kit's validation commands
- Link these files from feature front matter references when they materially shape work

## Ruleset Index

| Ruleset | Installed | Purpose |
| --- | --- | --- |
| `asd-ste100` | default | ASD-STE100 writing for agent explanations, with technical-content preservation and explicit style overrides. |
| `agent-team-orchestration` | default | Safe, integrated, truthfully reported delegation using the host's native subagents. |
| `aws-agent-toolkit-guidance` | default | Verified AWS identity, current AWS guidance and tools, and no secret exposure. |
| `backend-service-architecture` | default | Responsibility boundaries for routes, controllers, services, repositories, and persistence adapters. |
| `constitution-curation` | default | Evidence-based promotion of project-wide invariants. |
| `cross-repository-program-coordination` | default | One ledger and live-evidence reconciliation for multi-repository programs. |
| `deadline-mode` | default | Narrowed validation under an explicit user deadline without weakening safety. |
| `deletion-safety` | default | Recoverable soft delete by default and exact-target confirmation before hard delete. |
| `delivery` | default | Worktree lanes, human-owned pull requests, attribution, and commit conventions. |
| `frontend-application-architecture` | default | Responsibility boundaries for frontend routes, state, data adapters, and UI. |
| `github-pr-merge` | default | Explicit bounded merge authority and current `MERGE_READY` readiness. |
| `infrastructure-change-approval` | default | One outline and one approval per infrastructure batch; deletion always confirmed. |
| `slack-read-only` | default | Slack is read-only; every write needs explicit message-specific approval. |
| `testing-and-environment-validation` | default | Code-level tests primary, honest results, safe browser and production validation. |
| `llms-txt` | optional | `/llms.txt` contract for public web and API surfaces. |
| `readme-header-tagline` | optional | Consistent top-level README identity and opening structure. |

Optional rules install only through `kit rules add`.
