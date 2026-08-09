# Progressive Context Loading

## Start with the resolved contract

Use explicit hints to keep context narrow:

```bash
kit contract resolve \
  --workflow implementation-delivery \
  --path internal/registry/catalog.go \
  --applies-to registry
```

Read all artifacts marked mandatory and every selected workflow dependency.
Read conditional artifacts only when the resolver selects them or repository
evidence makes their applicability clear.

## Expand from repository evidence

1. Inspect the affected package, direct callers, and tests.
2. Read the current feature spec when material rationale exists.
3. Follow links to the Constitution or durable references only when the current
   decision depends on them.
4. Consult historical specs for rationale, not for current CLI behavior.
5. Re-resolve with new path or applicability hints when scope expands.

Do not load the entire documentation tree by default. The catalog and resolved
contract are the index; repository code and canonical docs remain the evidence.
