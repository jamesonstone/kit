---
kind: ruleset
slug: backend-service-architecture
description: Defines framework-aware route, controller, service, and repository boundaries for backend services and APIs.
status: active
registry_scope: downstream
applies_to:
  - architecture
  - backend
  - api
  - webservice
  - route
  - controller
  - handler
  - service
  - repository
  - persistence
  - data-access
  - gateway
  - coding-agent
read_policy_default: must
---

# Ruleset: backend-service-architecture

## Purpose

- Keep transport, request translation, business behavior, and persistence responsibilities explicit in backend services.
- Make the default runtime flow easy to trace: route → controller → service → repository.
- Preserve framework and language idioms without allowing business or persistence concerns to leak across boundaries.

## Applies When

- A Kit-managed project creates or changes an API, backend service, web service, HTTP or RPC route, controller or handler, application service, repository, persistence adapter, or external-service gateway.
- A change moves responsibilities between transport, application, domain, or data-access code.
- A backend feature introduces authorization, transactions, data mapping, caching, retries, or observable error behavior.

Load this rule before implementation for matching work even when the active feature does not link it explicitly.

This rule defines architectural roles, not mandatory directory or type names. Preserve a stronger existing architecture and map framework-native names such as handler, resolver, endpoint, use case, store, DAO, gateway, or adapter to the closest responsibility below.

## Rules

### Start From Repository Evidence

- Inspect the existing package structure, dependency direction, framework conventions, tests, and domain boundaries before adding a new layer or abstraction.
- Apply these boundaries to new and materially changed code. Do not perform an unrelated repository-wide rewrite.
- Prefer the smallest structure that keeps real responsibilities separate. Do not create empty pass-through layers solely to match a diagram.

### Route

- Own protocol registration: method and path, RPC or event binding, middleware ordering, versioning, and controller wiring.
- Keep route declarations declarative and free of business rules, persistence calls, and response construction beyond framework-required registration.
- Attach coarse transport concerns such as authentication middleware, rate limits, request size limits, tracing, and content negotiation at this boundary when supported.

### Controller

- Adapt transport input to an application request: decode, perform structural validation, resolve caller context, and invoke one application service or use case.
- Map application results and typed errors to protocol responses consistently.
- Keep controllers thin. Do not implement business policy, transactions, queries, persistence mapping, retry policy, or multi-step domain orchestration here.
- Pass authenticated identity and request context inward; enforce business authorization in the application or domain layer where the protected resource and action are understood.

### Service

- Own application use cases, business invariants, authorization decisions, orchestration, and transaction boundaries.
- Depend on explicit repository or gateway contracts rather than HTTP, RPC, ORM, or storage-framework request types.
- Keep services independently testable with repository and gateway substitutes.
- Split services by cohesive use case or domain capability instead of accumulating unrelated behavior in a generic manager or god service.

### Repository

- Own persistence and retrieval behavior: queries, storage mapping, optimistic locking, pagination mechanics, and storage-specific error translation.
- Expose domain- or application-oriented operations rather than leaking ORM models, SQL builders, driver errors, or transport response types.
- Do not decide business policy, authorization, workflow ordering, or protocol responses.
- Keep transaction primitives explicit. The service chooses the business transaction boundary; repositories participate in it without silently widening it.
- Model external systems as focused gateways or clients when repository terminology would obscure that the dependency is not persistent storage.

### Dependency And Data Boundaries

- Preserve the runtime flow route → controller → service → repository or gateway.
- Do not let routes or controllers call repositories directly when application behavior exists.
- Do not let repositories import controllers, routes, protocol response models, or application orchestration.
- Keep transport DTOs, domain values, and persistence records separate when their shapes or lifecycles differ; avoid mechanical duplicate models when one stable representation genuinely serves both boundaries.
- Centralize protocol error mapping and persistence error translation at their respective boundaries.
- Add interfaces at substitution, ownership, or test seams. Do not create one-method interfaces or generic base repositories without a concrete boundary need.

### Cross-Cutting Behavior

- Propagate cancellation, deadlines, correlation identifiers, and tracing context through established framework mechanisms.
- Keep logging and metrics useful at boundaries without logging the same failure at every layer.
- Put validation at the narrowest correct boundary: structural input validation in the controller, business invariants in the service or domain, and storage constraints in the repository.
- Make retry, idempotency, and concurrency behavior explicit in the service or gateway that owns the operation.

### Testing

- Test route registration and middleware wiring when configuration can regress.
- Test controllers for decoding, structural validation, service invocation, and response or error mapping.
- Test services for business behavior, authorization, orchestration, and transaction decisions without real transport or storage.
- Test repositories against the real storage boundary or a faithful integration fixture for queries, mapping, constraints, and concurrency semantics.
- Add end-to-end coverage for critical paths across all layers.

### Framework-Aware Exceptions

- A simple endpoint may combine route and controller roles when the framework does so naturally and no business logic leaks into registration code.
- A cohesive service may use a concrete repository directly when an interface adds no substitution or ownership value.
- A trivial read may collapse service and repository roles only when it has no business rule, authorization decision, orchestration, or transaction behavior; keep the exception local and easy to split when those concerns appear.
- Record material deviations in the active spec or canonical architecture documentation with the repository evidence that justifies them.

## Anti-Patterns

- Route declarations containing business decisions, SQL, ORM calls, or multi-step workflows.
- Controllers that authorize domain actions, manage transactions, or duplicate business rules.
- Services that accept framework request or response objects or return protocol status codes.
- Repositories that return HTTP responses, enforce workflow policy, or leak storage-driver errors to transport code.
- Direct route-to-repository or controller-to-database access used to bypass an existing service boundary.
- Generic base controllers, base services, or base repositories that erase domain language and make behavior harder to trace.
- Layers that only rename methods and add no ownership, translation, policy, or test seam.
- One service or repository accumulating unrelated domains because it is convenient to inject everywhere.

## Verification

- Trace at least one changed request from route registration through controller, service, repository or gateway, and back to the protocol response.
- Confirm imports and calls follow the intended direction and no transport or persistence type crosses an inappropriate boundary.
- Confirm business authorization and transaction ownership sit in the service or domain layer.
- Confirm structural, business, and storage validation each live at the correct boundary.
- Confirm tests exercise every changed responsibility at the narrowest useful level plus an integration path when practical.
- Confirm any collapsed roles or deviations are locally justified and do not hide business or persistence coupling.
- Run the repository's format, unit, integration, contract, and API-schema checks for the affected service.

## Examples

Standard request flow:

```text
POST /orders
  route: register method, path, middleware, and controller
  controller: decode CreateOrderRequest and map errors to the protocol
  service: authorize, enforce order rules, and own the transaction
  repository: persist the order and translate storage failures
```

Framework-native names:

```text
router → handler → use case → store
```

These names satisfy the rule when their responsibilities match route → controller → service → repository.
