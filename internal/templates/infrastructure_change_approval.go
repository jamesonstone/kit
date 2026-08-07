package templates

const infrastructureChangeApprovalGate = `## Infrastructure Change Approval Hard Gate

- Before mutating public-cloud resources, Kubernetes resources or cluster state, or infrastructure-as-code source, configuration, or state, load ` + "`docs/references/rules/infrastructure-change-approval.md`" + `.
- Read-only discovery may precede confirmation only when it does not alter cloud resources, Kubernetes objects, remote state, or repository-owned infrastructure source.
- Before the first covered mutation, present one consolidated outline of the target context, resource actions, execution boundary, material impact and risk, rollback or recovery, and validation evidence; obtain explicit user confirmation.
- A sufficiently detailed initial request may satisfy the gate only when it contains that complete outline and clearly authorizes the exact bounded batch.
- After confirmation, execute the exact approved batch to completion without routine command-by-command approval.
- If the provider identity, environment, region or cluster, resource set, action type, material impact, or rollback plan changes, stop, revise the outline, and obtain renewed confirmation.

`
