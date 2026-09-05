package templates

const infrastructureChangeApprovalGate = `## Infrastructure Change Approval Hard Gate

- Before mutating public-cloud resources, Kubernetes resources or cluster state, or infrastructure-as-code source, configuration, or state, load ` + "`docs/references/rules/infrastructure-change-approval.md`" + `.
- Standing merge/deploy authority covers only a named existing standard deployment workflow for an authorized environment and exact merged artifact on already-provisioned application resources, followed by deployed-identity, health, runtime, and rollback verification. Generic task acceptance does not authorize deployment.
- IAM, network topology, KMS, secrets, persistent data-store or database-schema change, data loss, cluster control-plane change, infrastructure creation, replacement, or deletion, new targets, workflow mutation, and nonstandard deployment effects are outside standing authority and require their own applicable approval boundary.
- Routine application operations on already-provisioned workloads are not infrastructure-approval batches when they stay inside the recorded standard deployment boundary. Record the target, workflow, environment, actor, and artifact; do not stop for another prompt solely because a later in-scope PR or head was unknown when authority was granted.
- Read-only discovery may precede confirmation only when it does not alter cloud resources, Kubernetes objects, remote state, or repository-owned infrastructure source.
- Put one consolidated outline of the target context, resource actions, execution boundary, material impact and risk, rollback or recovery, and validation evidence into the task plan before the first covered infrastructure mutation. Obtain one explicit user confirmation for that complete bounded batch.
- Approval of a task plan containing the complete outline counts as confirmation. A standing merge/deploy grant never substitutes for that infrastructure outline.
- Deleting, destroying, or removing infrastructure always requires explicit confirmation after the consolidated outline, even when the initial request asked for it; one confirmation covers every deletion named in that batch. Standing authority, merge, image deployment, and routine operations never authorize deletion.
- During merge or release orchestration, do not execute infrastructure deletion, destruction, purge, destructive replacement, or state removal; isolate it as a separate task with its own exact post-outline authorization.
- After infrastructure confirmation, execute the exact approved batch and continue in one pass without routine command-by-command approval. Additional or materially different covered changes require one follow-up outline and confirmation.
- The most recent direct human instruction wins. Pause, hold, or revocation stops affected actions and dependents until explicit human resume. Unresolved effect classification fails closed.

`
