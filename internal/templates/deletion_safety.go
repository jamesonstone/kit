package templates

const deletionSafetyGate = `## Deletion Safety Hard Gate

- Before designing deletion behavior or deleting persistent project, user, business, or external-system state, load ` + "`docs/references/rules/deletion-safety.md`" + `.
- An unqualified delete means soft delete: use a reversible lifecycle state with a supported, authorized, and tested restore path. Task-owned ephemeral scratch that never became authoritative state is outside this retained-state definition; ambiguity remains covered.
- Treat purge, destroy, force deletion, empty-trash operations, destructive replacement, history rewrite, retention expiry, backup or snapshot deletion, cryptographic erasure, and irreversible cascades as hard delete.
- Make the normal product and operational path soft-delete by default. Keep hard delete as a separate privileged, auditable, server-enforced action; a client prompt or ` + "`force`" + ` flag alone is insufficient.
- Before any hard delete, resolve and present the exact targets, or a bounded selector first resolved to the exact current target set with its current count and materialized target IDs or an immutable snapshot/version token, environment, cascades, why soft delete is insufficient, the loss of restore, backup state, retention or legal impact, and verification plan.
- After that outline, obtain a specific manual confirmation from the human for those exact current targets. Initial requests, general task or plan approval, automation, retention schedules, prior soft-delete approval, and broad cleanup language do not count.
- Bind confirmation to the actor, action, exact targets or immutable snapshot/version, environment, and consequences. Immediately before execution, compare the current target set or version with the confirmed snapshot; any difference requires a new outline and confirmation.
- Preserve stricter repository, legal, privacy, security, infrastructure, and provider controls. One post-outline confirmation may satisfy multiple deletion gates only when the combined outline contains every required field.

`
