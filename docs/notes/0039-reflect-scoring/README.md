# Feature Notes: 0039-reflect-scoring

This directory stores optional source material for the feature. Notes can
include Slack excerpts, customer context, screenshots, research links, draft
responses, and other supporting inputs.

Notes are source material, not canonical truth. Promote durable decisions,
requirements, and implementation constraints into `SPEC.md` or another
canonical project document before relying on them for implementation.

## Directories

- `inbox/` - unsorted captured notes and conversation excerpts.
- `references/` - source material, links, research, examples, and assets.
- `responses/` - draft or sent responses related to the feature.
- `private/` - local-only sensitive conversations or tertiary context.

Agents should ignore `.gitkeep` files and read only the notes relevant to
the current task.
