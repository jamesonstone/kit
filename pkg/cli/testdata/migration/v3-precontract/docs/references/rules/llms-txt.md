---
kind: ruleset
slug: llms-txt
description: Requires Kit-managed web services, websites, and APIs to expose an LLM-friendly /llms.txt endpoint.
status: active
registry_scope: downstream
applies_to:
  - web
  - website
  - webservice
  - api
  - documentation
  - llms-txt
  - coding-agent
read_policy_default: conditional
---

# Ruleset: llms-txt

## Purpose

- Make Kit-managed web services, websites, and APIs discoverable and usable by LLMs at inference time.
- Provide a stable, concise, Markdown entrypoint for agents that need to understand the service, public API, docs, SDKs, and integration workflows.
- Keep LLM-facing service context current as routes, public documentation, API contracts, and product behavior change.

## Applies When

- The project is a website, web application, web service, API service, or exposes HTTP routes intended for users, integrators, agents, or machines.
- A change adds, removes, renames, or materially changes a public route, API endpoint, OpenAPI contract, user-facing workflow, SDK, public docs page, integration guide, or service capability.
- A Kit-managed project is initialized or refreshed and the repository has a web or API surface.

This rule does not apply to repositories with no web, HTTP, API, or hosted documentation surface.

## Rules

- Every applicable service must expose `/llms.txt` at the service or site root.
- Serve `/llms.txt` without authentication unless the whole service is private and the deployment intentionally has no public surface.
- Return stable Markdown content, not an HTML page or marketing redirect.
- Prefer `text/markdown; charset=utf-8` when the framework supports it; `text/plain; charset=utf-8` is acceptable when that is the platform convention.
- Follow the `llms.txt` structure:
  - H1 with the service, product, or site name.
  - Blockquote summary with the most important context for understanding the service.
  - Optional concise details explaining how to interpret the linked resources.
  - H2 sections containing Markdown link lists.
  - Each list item uses a Markdown link and, when useful, a short description after `:`.
  - Use an `Optional` section only for secondary resources that can be skipped when context must stay short.
- For API services, include links to the most relevant machine-readable or agent-useful resources, such as:
  - OpenAPI or API reference documentation.
  - Authentication and authorization documentation.
  - Core public endpoints or workflow guides.
  - SDKs, client examples, schemas, changelog, and status or support documentation when available.
- For websites and web applications, include links to the canonical pages, docs, feeds, product guides, support pages, and other pages most useful for understanding the site.
- Keep `/llms.txt` concise. It should orient and link; it should not duplicate the full documentation set.
- Consider adding `/llms-full.txt` or linked Markdown documentation for expanded context when the service has substantial documentation.
- Update `/llms.txt` in the same change when public routes, APIs, docs, SDKs, product capabilities, or integration workflows change.
- Do not include secrets, private keys, internal-only credentials, environment-specific tokens, non-public customer data, privileged admin paths, or sensitive internal runbooks.
- Do not use `/llms.txt` as a replacement for authorization controls, `robots.txt`, `sitemap.xml`, OpenAPI, or human documentation. It should complement those artifacts.

## Anti-Patterns

- Do not leave web/API projects without a root `/llms.txt` endpoint.
- Do not serve a stale static file that omits newly added public APIs, routes, docs, or workflows.
- Do not link every page indiscriminately. Curate the resources an LLM actually needs.
- Do not expose private operational details, machine-local URLs, development-only ports, secrets, or customer-specific data.
- Do not make `/llms.txt` depend on JavaScript rendering.
- Do not return a 404, auth challenge, HTML shell, SPA fallback, or generic homepage for `/llms.txt` unless the entire deployment intentionally has no accessible web surface.

## Verification

Before completing web or API work in an applicable project, verify:

- `/llms.txt` is reachable at the deployed or local service root.
- The response status is `200 OK`.
- The response body is Markdown and starts with an H1 naming the service, product, or site.
- The file includes a concise blockquote summary and relevant H2 link-list sections.
- API services link to current API reference or OpenAPI documentation when available.
- The content reflects the changed route, API, documentation, SDK, or workflow.
- The content contains no secrets, credentials, local-only tokens, customer data, or privileged internal-only operational details.
- Focused route, handler, static asset, or integration tests cover `/llms.txt` when the project has a testable web stack.

Recommended checks:

```bash
curl -fsS "$BASE_URL/llms.txt"
curl -fsSI "$BASE_URL/llms.txt"
```

## Examples

Minimal service file:

```markdown
# Example API

> Example API provides account and billing endpoints for partner integrations.

Use the OpenAPI reference for request and response shapes. Use the auth guide before calling protected endpoints.

## Docs

- [OpenAPI reference](https://example.com/openapi.json): Machine-readable API contract.
- [Authentication](https://example.com/docs/auth): OAuth and API key setup.
- [Errors](https://example.com/docs/errors): Error envelope and retry guidance.

## Workflows

- [Create account](https://example.com/docs/accounts#create): Required fields and lifecycle notes.
- [Billing webhook](https://example.com/docs/webhooks#billing): Event payloads and idempotency guidance.

## Optional

- [Changelog](https://example.com/changelog): Recent API changes.
```

Testing locally:

```bash
BASE_URL=http://localhost:3000
curl -fsS "$BASE_URL/llms.txt"
```
