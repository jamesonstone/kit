# Security

## Purpose and trust boundaries

Treat security as a system property. Identify trust boundaries, minimize attack surface, apply least privilege and defense in depth, prefer secure defaults, and fail closed when authorization or integrity cannot be established. Use trusted primitives rather than adding unnecessary mechanisms.

Use [OWASP ASVS](https://asvs.owasp.org/) for applicable verification requirements, [OWASP Top 10](https://owasp.org/projects/top-ten) for risk awareness, [NIST SSDF](https://csrc.nist.gov/projects/ssdf) for development practices, and [CWE](https://cwe.mitre.org/) for weakness analysis. Apply them according to actual risk; this document is engineering guidance, not a compliance certification.

## Authentication, authorization, and isolation

- Distinguish authentication from authorization. Enforce authorization at trusted server-side boundaries; client-side checks are insufficient. Verify resource ownership and tenant isolation consistently with least privilege.
- Use established libraries and framework security primitives. Do not introduce custom cryptography or authentication protocols unnecessarily.
- Bound API resource use and apply appropriate rate limits. Protect sensitive operations against replay and unintended duplication; use idempotency when repeated requests could cause unsafe effects. Keep internal details out of external errors.

## Untrusted input and sensitive data

- Treat external input as untrusted and validate it at trust boundaries. Use parameterized queries and destination-appropriate output encoding. Prevent injection; avoid unsafe deserialization and command execution.
- Never hardcode credentials, tokens, or secrets. Use established secret-management mechanisms. Keep secrets out of logs, exceptions, and generated artifacts.
- Minimize sensitive-data collection, retention, and exposure. Test fixtures must contain neither real secrets nor sensitive production data.

## Dependencies, implementation, and review

- Prefer maintained, minimal dependencies, reproducible builds, and dependency locking where supported. Evaluate vulnerabilities by exploitability and impact, not severity labels alone. Do not add infrastructure solely for generic recommendations.
- Review changes at trust boundaries. Prefer proven framework controls over custom security abstractions. Test meaningful security failure modes. Never weaken security controls merely to make tests pass.
- Record security-relevant actions with appropriate attribution and trace consequential actions where required. Preserve diagnostic context without credentials or sensitive request payloads. Distinguish operational telemetry from security audit records.

## Vulnerability reporting

Report suspected exploitable vulnerabilities privately. Coordinate remediation and disclosure with the maintainers before public disclosure; do not include exploitable details in public issues or pull requests.

- **Supported versions (project-owned):** TODO: maintainers must state the supported-version policy when known. No support range is implied by this template.
- **Private reporting channel (project-owned):** TODO: maintainers must publish a verified private channel. Prefer GitHub Private Vulnerability Reporting only when enabled and appropriate; otherwise document the project's actual private channel. No contact or reporting service is implied by this template.
- **Response and disclosure expectations (project-owned):** TODO: maintainers must state actual expectations. Do not assume a response SLA or disclosure deadline.

## Project guidance

Keep project-specific security constraints and reporting instructions here or in the reporting section. Kit updates unchanged defaults and preserves customized documents during reconciliation. This guidance does not enable security settings or enforce runtime controls.
