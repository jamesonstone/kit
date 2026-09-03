package templates

const infrastructureChangeApprovalGate = `## Infrastructure Change Approval Hard Gate

- Before mutating public-cloud resources, Kubernetes resources or cluster state, or infrastructure-as-code source, configuration, or state, load ` + "`docs/references/rules/infrastructure-change-approval.md`" + `.
- Classify planned effects as create, update, replace, delete, or remove. Proceed autonomously when the graph contains only additive or rollback-preserving effects, including additive IAM, network, or resource create-or-update and production activation.
- Routine application operations on already-provisioned workloads, including deployment image updates and ECS or equivalent service interactions that do not create, replace, or delete infrastructure, are not infrastructure-approval batches. Record classified non-destructive mutations; do not stop for a confirmation outline.
- Read-only discovery may precede mutation only when it does not alter cloud resources, Kubernetes objects, remote state, or repository-owned infrastructure source.
- Isolate delete, remove, destroy, purge, destructive replacement, state removal, history rewrite, data erasure, permission revocation, or loss of a supported recovery path. Present one exact-target outline and obtain one explicit user confirmation for that destructive batch.
- Deleting, destroying, or removing infrastructure always requires explicit confirmation after the consolidated outline, even when the initial request asked for it; one confirmation covers every deletion named in that batch. An accepted task, merge, image deployment, and routine ECS interactions never authorize deletion.
- During merge or release orchestration, do not execute infrastructure deletion, destruction, purge, destructive replacement, or state removal; isolate it as a separate task with its own exact post-outline authorization.
- After destructive confirmation, execute the exact approved batch and continue the rest of the task to completion in one pass without routine command-by-command approval.
- If additional destructive infrastructure changes become necessary, collect all then-known changes into one follow-up outline, obtain one confirmation, and execute that follow-up batch in one pass. Do not re-confirm actions already included in an approved batch.
- Unresolved destructive-effect classification fails closed. Compatible tools, commands, and retries inside an approved destructive boundary do not require another prompt. Explicit user holds such as "keep production default-off" prevail.

`
