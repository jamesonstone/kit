---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: complete
feature:
  id: 0044
  slug: versioned-agent-instructions
  dir: 0044-versioned-agent-instructions
references:
  - id: instruction-registry
    name: Repository instruction registry
    type: code
    target: internal/instructions/registry.go
    relation: informs
    read_policy: must
    used_for: package placement and existing instruction terminology
    status: active
  - id: command-capabilities
    name: Command capabilities ruleset
    type: rule
    target: docs/references/rules/command-capabilities.md
    relation: constrains
    read_policy: must
    used_for: complete read-only command metadata and discoverability
    status: active
  - id: command-guide
    name: Kit command guide
    type: doc
    target: docs/commands.md
    relation: guides
    read_policy: must
    used_for: public command behavior and version-selection examples
    status: active
delivery_intent: issue_branch_pr_in_progress
---
# SPEC

## PURPOSE

Provide one stable Kit command that emits the provider-neutral coding-agent instructions used directly by Codex, Claude, and GitHub Copilot, while preserving earlier instruction revisions for reproducible use.

## CONTEXT

- The supplied instruction body is the canonical first release, `v1`.
- Users need the latest instruction policy without remembering its version, but must be able to retrieve an older immutable version explicitly after later revisions ship.
- Kit already uses the term “repository instructions” for generated `AGENTS.md`, `CLAUDE.md`, and Copilot files. This command emits a separate provider-neutral payload and must not change scaffold-version detection or generated repository-file behavior.
- `kit prompt coding-agent instructions` is a prompt-library entry that asks an agent to produce task-specific instructions; it does not expose this static policy and is not a substitute for the new command.
- The command is a read-only, project-independent stdout surface. It must not perform configuration remediation, use the clipboard, access the network, or write files.
- Issue `#62` is the delivery source, and branch `GH-62` was created from refreshed `origin/main` before implementation edits.

## REQUIREMENTS

- Add a visible root command named `kit instructions`.
- With no version flag, the command prints the complete current instruction payload as raw Markdown to stdout.
- Store each release as an immutable `vN` artifact embedded into the Kit binary, beginning with the supplied text as `v1`.
- Resolve default output through an explicit current-version pointer rather than relying on lexical file ordering.
- Accept explicit version identifiers in `vN` form, including `--version=v1`; reject empty, malformed, or unavailable identifiers with a clear error that identifies the requested value and available versions.
- Preserve the supplied Markdown content exactly, apart from ensuring the conventional single trailing newline in command output; do not include the request's surrounding `<instructions>` delimiter.
- Keep the command independent of a Kit project and free of network, file-write, Git, GitHub, editor, subprocess, configuration-remediation, and clipboard side effects.
- Expose the command in root help, shell completion through Cobra registration, `kit capabilities`, and `docs/commands.md`.
- Add focused registry and CLI tests for current resolution, explicit `v1`, exact output, invalid/unavailable versions, help/metadata discoverability, and read-only preflight behavior.
- Observable acceptance: the built binary's default and `--version=v1` output are byte-identical to the embedded `v1` artifact, unsupported versions fail non-zero with a useful message, targeted capability queries describe the command accurately, and full repository validation passes.
- Non-goals: provider-specific variants, user overrides, remote version fetching, version ranges or aliases such as `latest`, automatic prompt injection, changing scaffold instruction versions, or changing the prompt-library command.

## ACCEPTED PLAN

1. Add a versioned embedded instruction payload under `internal/instructions`, plus a small registry with explicit current-version resolution, exact lookup, and deterministic available-version reporting.
2. Add `kit instructions` as a project-independent Cobra command that writes the selected payload directly to command stdout and accepts `--version=vN`.
3. Register the command in root help, automatic-config preflight exemptions, and complete `kit capabilities` metadata.
4. Add exact-output and error-path tests at the registry and CLI layers, then document the command and version contract.
5. Run focused tests, formatting, vet, full Go tests, build and binary smoke tests, Kit document checks, whitespace checks, and diff self-review before explicit staging and PR delivery.

## DECISIONS

- Use `kit instructions` because it directly names the emitted artifact and avoids conflating retrieval with the task-specific `kit prompt` library.
- Use raw stdout with no clipboard default or human banner so the result can be redirected or injected into any supported coding agent without cleanup.
- Use `vN` strings at the public boundary and immutable embedded Markdown files internally. This keeps user-visible versions explicit while preserving the exact source text in reviewable files.
- Keep an explicit current-version constant and exact registry lookup. File ordering, semantic guessing, and a `latest` alias would make the default contract less reviewable.
- Extend the existing `internal/instructions` package but keep the payload registry separate from repository instruction-scaffold versioning so the two version domains cannot silently interact.

## DISCOVERIES

- Cobra scopes the new child-command `--version` flag independently from Kit's existing root `--version` flag. Built-binary smoke tests confirmed that `kit instructions --version=v1` selects the instruction payload rather than the installed CLI version path.
- A SHA-256 assertion over the embedded `v1` payload makes accidental edits to the immutable release visible in tests while keeping the Markdown file itself as the reviewable source.

## VALIDATION

- Focused registry, CLI, root-help, preflight, and capability tests passed: `go test ./internal/instructions ./pkg/cli -run 'TestAgentInstructions|TestAgentInstructionVersions|TestInstructions|TestRootHelpGroupsCanonicalCommands|TestCapabilityCatalogCoversVisibleRootCommands' -count=1`.
- `make fmt` passed and left the Go sources formatted.
- `make vet` passed (`go vet ./...`).
- `go test ./... -count=1` passed across all packages.
- `make build` passed and produced `bin/kit` at version `v1.0.91`.
- `cmp <(./bin/kit instructions) internal/instructions/versions/v1.md` and the equivalent `--version=v1` comparison both passed, proving byte-identical raw output.
- Built-binary `--version=v2` validation returned non-zero with exactly `Error: unsupported instructions version "v2"; available versions: v1`.
- Targeted `kit capabilities instructions --json` assertions confirmed mutation level, network, file-write, and Git behavior are all `none`, and the only important flag is `--version`.
- `kit capabilities --search 'agent instructions' --json` discovered the new command.
- `golangci-lint run --new-from-rev=origin/main ./...` passed with no findings.
- `go test -race ./internal/instructions ./pkg/cli -count=1` passed.
- `./bin/kit check 0044-versioned-agent-instructions` passed before completion curation.
- `./bin/kit complete 0044-versioned-agent-instructions` passed, set the V3 phase to `complete`, and refreshed `docs/PROJECT_PROGRESS_SUMMARY.md`.
- Final `./bin/kit check 0044-versioned-agent-instructions` passed.
- Final `./bin/kit check --project` exited successfully with 15 historical V2 compatibility advisories and the existing project-refresh cadence warning; no blocking finding was introduced by this feature.

## OUTCOME

- `kit instructions` now prints the current embedded provider-neutral coding-agent policy as raw Markdown with no banner, clipboard, project-config, file-write, network, subprocess, Git, or GitHub side effect.
- The immutable `v1` source is stored at `internal/instructions/versions/v1.md`, registered separately from repository scaffold versions, and protected by an exact SHA-256 regression assertion.
- The explicit `CurrentAgentVersion` pointer controls default output; `--version=v1` selects the release directly, while empty, malformed, and unavailable selectors fail with actionable available-version errors.
- Root help, Cobra completion registration, automatic-config preflight behavior, capabilities metadata, README guidance, command documentation, and the Constitution package map are aligned with the new command.
- No provider-specific variants, remote lookup, customization, prompt injection, scaffold-version change, or prompt-library behavior was added.

## REPOSITORY MEMORY

Decision: created

Rationale: The current-version pointer, immutable historical lookup, raw-output boundary, and separation from repository scaffold versions are durable product decisions that code alone would not explain fully.

Artifacts:

- `docs/specs/0044-versioned-agent-instructions/SPEC.md`
- `README.md`
- `docs/commands.md`
- `docs/CONSTITUTION.md`
