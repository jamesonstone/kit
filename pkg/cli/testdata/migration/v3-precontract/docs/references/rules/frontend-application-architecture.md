---
kind: ruleset
slug: frontend-application-architecture
description: Defines framework-aware route, feature, domain, data-adapter, state, and component boundaries for frontend applications.
status: active
registry_scope: downstream
applies_to:
  - architecture
  - frontend
  - web
  - mobile
  - ui
  - route
  - page
  - feature
  - component
  - view-model
  - state-management
  - api-client
  - data-adapter
  - coding-agent
read_policy_default: must
---

# Ruleset: frontend-application-architecture

## Purpose

- Keep navigation, feature orchestration, business behavior, data access, state, and presentation responsibilities explicit in frontend applications.
- Prefer feature ownership and inward dependency direction over global folders organized only by technical type.
- Complement the frontend prompt profile's visual and interaction guidance without duplicating it.

## Applies When

- A Kit-managed project creates or changes a web, mobile, or desktop frontend route, page, screen, feature, component, state flow, API client, browser storage adapter, cache, or frontend domain service.
- A change moves behavior between routes, components, hooks, stores, domain logic, or data access.
- Frontend work introduces navigation, form orchestration, server state, optimistic updates, offline behavior, caching, or shared UI.

Load this rule before implementation for matching work even when the active feature does not link it explicitly.

This rule defines architectural roles, not mandatory directory, component, hook, or framework names. Preserve a stronger existing architecture and map framework-native constructs such as loaders, actions, server components, view models, presenters, composables, stores, or coordinators to the closest responsibility below.

## Rules

### Start From Repository Evidence

- Inspect the existing router, rendering model, feature boundaries, component library, state conventions, data-fetching layer, tests, and design system before adding structure.
- Apply these boundaries to new and materially changed code. Do not reorganize unrelated features to make the repository look uniform.
- Prefer the smallest structure that keeps real ownership clear. Do not manufacture layers or wrappers with no behavior.

### Routes And Pages

- Own URL or navigation mapping, route-level access gates, layout composition, route parameters, and feature entrypoint selection.
- Keep pages focused on composition and framework lifecycle integration.
- Treat loaders, actions, resolvers, and server components as transport or controller boundaries when the framework gives them that role; do not duplicate them with ceremonial wrappers.
- Do not place reusable business rules, ad hoc API clients, persistent state, or unrelated feature behavior in route modules.

### Features And Interaction Orchestration

- Co-locate a feature's screens, containers, interaction state, use-case orchestration, and feature-specific UI.
- Let feature controllers, view models, hooks, composables, or stores coordinate user intent with application services and data ports.
- Keep navigation effects, form submission flow, optimistic state, and user-visible error recovery owned by the feature that understands the interaction.
- Avoid cross-feature deep imports. Promote code to a shared boundary only after genuine reuse and stable ownership are demonstrated.

### Domain And Application Behavior

- Keep durable business calculations, policies, validation, and state transitions independent of the rendering framework when they are more than view-only behavior.
- Expose focused application operations that feature code can invoke without knowing protocol, cache, or storage details.
- Keep domain values and application services free of components, navigation APIs, browser globals, framework lifecycle objects, and transport DTOs.
- Do not extract trivial display formatting or one-screen interaction code into a domain layer solely for symmetry.

### Data Adapters

- Centralize API calls, transport serialization, authentication attachment, cache integration, browser or device storage, and external SDK access in focused clients, repositories, or adapters.
- Translate transport DTOs and storage records at the data boundary when their shape differs from stable application or domain models.
- Normalize error and cancellation behavior before returning it to feature orchestration.
- Keep data adapters free of rendering, navigation, toast, modal, and feature workflow decisions.

### Components

- Keep reusable components presentational and controlled where practical: receive data and callbacks, render states, and emit user intent.
- Do not let shared components fetch feature data, write persistent storage, navigate as a hidden side effect, or enforce feature business policy.
- Allow feature-specific components to own local view behavior, but keep network, cross-screen workflow, and durable business decisions in feature orchestration or application services.
- Reuse the established component library, tokens, accessibility primitives, and composition patterns before creating new shared primitives.

### State Ownership

- Keep state as local as its lifetime and sharing needs allow.
- Distinguish server state from transient view state, form state, navigation state, and durable client state.
- Use the established query, cache, or store mechanism instead of maintaining duplicate copies of server data.
- Promote state to a broader store only when multiple owners or lifecycle requirements justify it.
- Make synchronization, optimistic updates, rollback, persistence, and invalidation behavior explicit at the feature or data boundary that owns it.

### Dependency Direction

- Preserve the user-flow direction route or page → feature orchestration → application or domain operation → data port.
- Keep compile-time dependencies pointing inward: UI and data adapters may depend on application or domain contracts; application and domain code must not depend on UI frameworks, network clients, or storage implementations.
- Wire concrete adapters at the application composition boundary instead of importing them into domain behavior.
- Keep shared UI independent of feature modules, and keep features independent of other features' internal files.

### Testing

- Test pure domain and application behavior without rendering or network dependencies.
- Test feature orchestration for state transitions, user intent, async behavior, cancellation, optimistic updates, and recovery.
- Test components for rendered states, accessibility semantics, and emitted events.
- Test data adapters for serialization, authentication, error normalization, caching, and cancellation.
- Add route or end-to-end coverage for critical user journeys across navigation, feature, and data boundaries.
- Apply the frontend prompt profile separately when visual quality, responsive behavior, and interaction-state inspection matter.

### Framework-Aware Exceptions

- A simple page may combine route, feature, and presentation roles when it has no reusable business policy, shared state, or independent data workflow.
- A framework loader, action, or server component may call an application service directly when it already serves as the route's controller boundary.
- A component may fetch its own data only when it is the explicit feature boundary rather than a reusable presentation component.
- Record material deviations in the active spec or canonical architecture documentation with the repository evidence that justifies them.

## Anti-Patterns

- Reusable components that fetch feature data, write storage, navigate, or enforce business policy.
- Route or page files that accumulate domain rules, API calls, caching, and unrelated feature state.
- Global `components`, `hooks`, `stores`, `services`, or `utils` folders used as dumping grounds without stable ownership.
- Domain or application code importing UI frameworks, browser globals, network clients, or storage implementations.
- Multiple caches or stores holding conflicting copies of the same server state.
- Cross-feature imports into another feature's internal components, state, or adapters.
- Generic wrapper hooks, repositories, or service classes that only rename framework APIs.
- Moving every small formatter or view interaction into a domain layer despite having no reusable business meaning.

## Verification

- Trace at least one changed user interaction from route or page through feature orchestration and application behavior to the data adapter and back to rendered state.
- Confirm compile-time dependencies point inward and domain or application code remains framework- and transport-independent where required.
- Confirm reusable components do not own hidden network, storage, navigation, or business-policy side effects.
- Confirm state is owned at the narrowest correct lifetime and duplicate server-state sources were not introduced.
- Confirm feature internals are not imported across feature boundaries and shared code has demonstrated reuse.
- Confirm framework-native role combinations are explicit and do not hide business or data-access coupling.
- Run the repository's format, type, unit, component, integration, accessibility, build, and end-to-end checks appropriate to the affected frontend.

## Examples

Feature-first request flow:

```text
/orders/:id route
  page: compose the order feature and read route parameters
  feature controller: coordinate loading, edits, optimistic state, and recovery
  application service: enforce order-edit behavior
  orders API adapter: serialize requests and normalize transport failures
  components: render controlled order states and emit user intent
```

Framework-native server boundary:

```text
route loader → application service → data port implemented by an API adapter
```

No separate controller wrapper is required when the loader already owns request adaptation and response mapping.
