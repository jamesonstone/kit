---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: complete
feature:
  id: 0049
  slug: application-architecture-rules
  dir: 0049-application-architecture-rules
relationships:
  - type: builds_on
    target: 0021-project-validation-and-instruction-registry
  - type: related_to
    target: 0024-frontend-profile
  - type: builds_on
    target: 0045-constitution-curation
skills:
  - name: github:github
    source: GitHub plugin
    path: github:github
    trigger: create and verify the issue-first delivery lane
    required: true
  - name: github:yeet
    source: GitHub plugin
    path: github:yeet
    trigger: publish the completed change through the required pull request workflow
    required: true
references:
  - id: rules-registry
    name: Rules registry
    type: code
    target: pkg/cli/rules_registry.go
    relation: implements
    read_policy: must
    used_for: downstream ruleset visibility and refresh filtering
    status: active
  - id: backend-architecture-rule
    name: Backend service architecture rule
    type: rule
    target: docs/references/rules/backend-service-architecture.md
    relation: implements
    read_policy: must
    used_for: route controller service and repository boundaries
    status: active
  - id: frontend-architecture-rule
    name: Frontend application architecture rule
    type: rule
    target: docs/references/rules/frontend-application-architecture.md
    relation: implements
    read_policy: must
    used_for: route feature domain data and component boundaries
    status: active
  - id: rlm-routing
    name: RLM routing
    type: documentation
    target: docs/agents/RLM.md
    relation: guides
    read_policy: must
    used_for: just-in-time architecture rule loading
    status: active
  - id: instruction-templates
    name: Repository instruction templates
    type: code
    target: internal/templates/instruction_templates_v3.go
    relation: implements
    read_policy: must
    used_for: reliable downstream architecture routing
    status: active
delivery_intent: issue_branch_pr_in_progress
---
# SPEC

## PURPOSE

Provide reusable, framework-aware backend and frontend architecture rules that keep responsibilities and dependency direction explicit across Kit-managed application projects.

## CONTEXT

- The registry currently includes workflow, safety, documentation, and lifecycle rules, but no downstream rule governs backend or frontend application structure.
- API-adjacent `llms-txt` guidance covers service discoverability, not implementation architecture.
- Kit has a frontend prompt profile for visual quality, interaction states, and responsive validation. It does not define frontend code ownership or dependency boundaries.
- Registry refresh installs active downstream rules, but a registry file cannot guarantee its own just-in-time loading. Generated repository instructions need concise routing pointers for matching work.
- The requested backend convention names routes, controllers, services, and repositories. Those responsibilities should be explicit without forcing identical directory names on every language or framework.
- Issue `#76` and branch `GH-76` own this change.

## REQUIREMENTS

- Add an active downstream `backend-service-architecture` ruleset with `read_policy_default: must`.
- Define the default backend dependency direction as routes to controllers to services to repositories.
- Keep transport, request translation, business behavior, and persistence responsibilities separated.
- Prohibit direct route-to-repository access and business rules in transport or persistence layers.
- Keep transactions, authorization, validation, errors, observability, and tests owned at the narrowest correct boundary.
- Permit framework-native role combinations only when responsibilities remain explicit and the simpler structure is justified.
- Add an active downstream `frontend-application-architecture` ruleset with `read_policy_default: must`.
- Define frontend boundaries for routes or pages, feature orchestration, domain or application behavior, data adapters, and presentational components.
- Prefer feature ownership over global type-based dumping grounds, while extracting genuinely shared code only after reuse is demonstrated.
- Prohibit reusable UI components from owning feature business rules, network access, storage, or navigation side effects.
- Keep the existing frontend profile focused on visual and interaction quality instead of duplicating architecture guidance.
- Add concise generated routing so backend and frontend architecture work loads the applicable rule even when a feature has not yet linked it explicitly.
- Preserve Kit's progressive-disclosure model; do not inline either complete rule into always-loaded instructions.
- Verify both rules are valid, visible to downstream projects, adopted by refresh, and referenced by generated and checked-in repository instructions.
- Non-goals: automatically refactoring existing projects, enforcing literal directory names, inventing a universal framework abstraction, changing the frontend prompt profile, or changing `kit rules link` behavior.

## ACCEPTED PLAN

1. Add canonical backend and frontend architecture rules with explicit responsibilities, dependency direction, anti-patterns, framework-aware exceptions, and verification.
2. Add short architecture routing pointers to the shared generated repository instructions and RLM guidance, then align checked-in instruction artifacts.
3. Add focused tests for ruleset validity, downstream refresh adoption, and generated instruction routing.
4. Update the project rollup and keep this living spec current with implementation discoveries, validation, outcome, and repository-memory curation.
5. Run focused tests, full Go tests, Kit document checks, formatting, changed-lines lint, and diff review before explicit staging and ready-PR delivery.

## DECISIONS

- Treat routes, controllers, services, and repositories as responsibility boundaries rather than mandatory folder names.
- Keep backend and frontend guidance in separate rules because their ownership models, framework exceptions, and anti-patterns differ.
- Keep frontend visual-quality guidance in the existing prompt profile and frontend code-structure guidance in the new registry rule.
- Use concise always-loaded routing pointers because active registry installation alone does not guarantee just-in-time rule loading.
- Make both rules mandatory when their architecture triggers match, while allowing a project to preserve a stronger existing architecture.

## DISCOVERIES

- `kit init --refresh` iterates over every downstream-visible registry rule and creates missing rule files, so adding active downstream files is sufficient for distribution.
- RLM currently prioritizes feature-linked rules and has explicit exceptions for a small number of globally triggered rules; application architecture needs the same concise routing treatment.
- The rules registry is sourced from `docs/references/rules` on the Kit `main` branch, so no separate catalog entry is required.
- New source rules appear as `untracked` registry state inside the Kit source repository until they merge to `main`; this is expected bootstrap behavior, while focused refresh tests prove downstream installation and managed-state recording.
- Generated V3 provider instructions and generated RLM guidance share existing template sources, so focused equality tests can prevent checked-in routing files from drifting.

## VALIDATION

- Focused ruleset and downstream refresh tests passed in `pkg/cli`.
- Focused generated-instruction routing and checked-in RLM equality tests passed in `internal/templates`.
- `make fmt` passed.
- `go vet ./...` passed.
- `go test ./... -count=1` passed across all packages.
- Focused race tests passed for `internal/templates` and `pkg/cli`.
- `make build` passed and produced `bin/kit` at version `v1.0.96`.
- `go run ./cmd/kit check --all` passed all 46 features, including this feature.
- `go run ./cmd/kit check --project` passed with a coherent project contract and reported project refresh not due.
- `golangci-lint run --new-from-rev=origin/main ./...` passed with `0 issues`.
- `go run ./cmd/kit rules list` showed both rules as active with their expected applicability tags; their Kit-source registry state is `untracked` until the files merge to the registry source branch.
- `git diff --check` passed.

## OUTCOME

- Added active downstream `backend-service-architecture` and `frontend-application-architecture` rules with mandatory read defaults.
- Backend guidance defines route, controller, service, repository, gateway, transaction, validation, dependency, testing, and framework-exception boundaries.
- Frontend guidance defines route or page, feature orchestration, domain or application, data-adapter, component, state, dependency, testing, and framework-exception boundaries.
- Both rules define responsibilities rather than mandatory directory names and preserve stronger repository-native architecture.
- Generated and checked-in Codex, Claude, Copilot, and RLM instructions now route matching work to the applicable rule without inlining the full rule.
- Focused tests validate ruleset content, downstream refresh adoption, and generated instruction alignment.
- The existing frontend prompt profile remains focused on visual and interaction quality.

## REPOSITORY MEMORY

Decision: created

Rationale: The layer responsibilities, dependency directions, framework exceptions, and separation between frontend visual guidance and application architecture are durable cross-project decisions that code and tests alone cannot preserve. The existing Constitution already defines downstream registry and repository-memory behavior, so no Constitution change is warranted.

Artifacts:

- `docs/specs/0049-application-architecture-rules/SPEC.md`
- `docs/references/rules/backend-service-architecture.md`
- `docs/references/rules/frontend-application-architecture.md`
- `docs/agents/RLM.md`
