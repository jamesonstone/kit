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
  - id: work-lane-gating
    name: Work lane gating ruleset
    type: rule
    target: docs/references/rules/work-lane-gating.md
    relation: constrains
    read_policy: must
    used_for: clean-preflight autonomy and in-progress continuation behavior
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
- The immutable `v1` release remains the historical policy. Issue `#68` and branch `GH-68` deliver a new `v2` current release so the historical artifact is not edited in place.
- Agents have been asking the new-lane question even when preflight proves that the default branch is clean, current, and has no active delivery lane. In that state the question has no material choice to resolve and should not block automatic issue and branch creation.
- Existing work remains different: when additional implementation scope is detected on an active branch or pull request, the user may still choose a new lane or explicitly continue the active one. Continuing must retain the active branch and pull request while giving the additional commits their own issue scope.

## REQUIREMENTS

- Add a visible root command named `kit instructions`.
- With no version flag, the command prints the complete current instruction payload as raw Markdown to stdout.
- Store each release as an immutable `vN` artifact embedded into the Kit binary, beginning with the supplied text as `v1`.
- Resolve default output through an explicit current-version pointer rather than relying on lexical file ordering.
- Accept explicit version identifiers in `vN` form, including `--version=v1`; reject empty, malformed, or unavailable identifiers with a clear error that identifies the requested value and available versions.
- Preserve the supplied Markdown content exactly, apart from ensuring the conventional single trailing newline in command output; do not include the request's surrounding `<instructions>` delimiter.
- Keep the command independent of a Kit project and free of network, file-write, Git, GitHub, editor, subprocess, configuration-remediation, and clipboard side effects.
- Expose the command in root help, shell completion through Cobra registration, `kit capabilities`, and `docs/commands.md`.
- Add focused registry and CLI tests for current resolution, explicit `v1` and `v2`, exact output, invalid/unavailable versions, help/metadata discoverability, and read-only preflight behavior.
- Observable acceptance: the built binary's default and `--version=v2` output are byte-identical to the embedded `v2` artifact, `--version=v1` remains byte-identical to immutable `v1`, unsupported versions fail non-zero with a useful message, targeted capability queries describe the command accurately, and full repository validation passes.
- Non-goals: provider-specific variants, user overrides, remote version fetching, version ranges or aliases such as `latest`, automatic prompt injection, changing scaffold instruction versions, or changing the prompt-library command.
- Preserve `v1` byte-for-byte and publish the revised policy as immutable `v2`, with `v2` becoming the explicit current version.
- In `v2`, when preflight proves all of the following, proceed to issue resolution and exact `GH-<issue-number>` branch creation without asking the new-lane question: the work is implementation work, the agent is on the clean default branch, the branch matches the refreshed remote default, and no issue, branch, or pull request already covers the requested work.
- Do not treat that automatic path as permission to skip identity, repository, base-refresh, issue-search, branch-naming, validation, staging, commit, push, or pull-request safeguards.
- Keep the lane-choice question when work is already in progress or the correct lane is otherwise material or ambiguous.
- When the user explicitly chooses to continue additional scope on work already in progress, create or reuse a separate issue for the new commits, keep the existing branch and pull request, scope the new commits to the additional issue, and update the pull request's issue references and complete validation description.
- Align the repo-local `work-lane-gating` ruleset with the same clean-preflight exception so repo-local rules cannot reintroduce the unnecessary question after `v2` says to proceed.

## ACCEPTED PLAN

1. Add a versioned embedded instruction payload under `internal/instructions`, plus a small registry with explicit current-version resolution, exact lookup, and deterministic available-version reporting.
2. Add `kit instructions` as a project-independent Cobra command that writes the selected payload directly to command stdout and accepts `--version=vN`.
3. Register the command in root help, automatic-config preflight exemptions, and complete `kit capabilities` metadata.
4. Add exact-output and error-path tests at the registry and CLI layers, then document the command and version contract.
5. Run focused tests, formatting, vet, full Go tests, build and binary smoke tests, Kit document checks, whitespace checks, and diff self-review before explicit staging and PR delivery.
6. Add immutable `v2` by copying the established policy and changing only the lane-allocation rules required by issue `#68`; move the explicit current-version pointer to `v2` while retaining the `v1` hash assertion.
7. Update `work-lane-gating` so a proven clean, current default branch with no active lane proceeds directly to `github-pr-delivery`, while an in-progress or ambiguous lane still asks and same-PR continuation retains its separate-issue traceability contract.
8. Extend registry, CLI, and ruleset tests for ordered `v1`/`v2` availability, exact current output, immutable hashes, clean-preflight autonomy, and in-progress continuation safeguards; update public version examples.
9. Validate both embedded versions byte-for-byte, run the complete repository checks, curate this living spec to the delivered behavior, and complete ready-for-review delivery on `GH-68`.

## DECISIONS

- Use `kit instructions` because it directly names the emitted artifact and avoids conflating retrieval with the task-specific `kit prompt` library.
- Use raw stdout with no clipboard default or human banner so the result can be redirected or injected into any supported coding agent without cleanup.
- Use `vN` strings at the public boundary and immutable embedded Markdown files internally. This keeps user-visible versions explicit while preserving the exact source text in reviewable files.
- Keep an explicit current-version constant and exact registry lookup. File ordering, semantic guessing, and a `latest` alias would make the default contract less reviewable.
- Extend the existing `internal/instructions` package but keep the payload registry separate from repository instruction-scaffold versioning so the two version domains cannot silently interact.
- Treat a proven clean default-branch preflight as implicit consent to allocate the normal issue-number lane because no competing work exists to preserve and the session-initialization contract already requires that lane.
- Keep explicit user choice for in-progress or ambiguous work because continuing can change traceability across an existing branch and pull request. A continue choice changes issue-to-commit mapping only; it does not allocate a replacement branch or pull request.
- Encode the same decision in both the provider-neutral `v2` payload and `work-lane-gating`. The embedded instructions establish the cross-project agent policy, while the repo-local ruleset is the higher-priority operational gate in Kit-managed repositories.

## DISCOVERIES

- Cobra scopes the new child-command `--version` flag independently from Kit's existing root `--version` flag. Built-binary smoke tests confirmed that `kit instructions --version=v1` selects the instruction payload rather than the installed CLI version path.
- A SHA-256 assertion over the embedded `v1` payload makes accidental edits to the immutable release visible in tests while keeping the Markdown file itself as the reviewable source.
- The recurring question is also encoded in the repo-local `work-lane-gating` ruleset, which outranks generic agent defaults in Kit-managed projects. Updating only the embedded payload would leave the reported behavior intact after agents load repository rules, so `v2` and the ruleset must agree.
- `github-pr-delivery` already contains the required additional-scope exception: a separate issue and issue-scoped commits remain on the existing pull-request head branch. Focused regression assertions now protect that behavior while the lane gate points continue choices to it.

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
- The `v2` update passed focused registry, CLI, root-help, capability, and ruleset tests: `go test ./internal/instructions ./pkg/cli -run 'TestAgentInstructions|TestAgentInstructionVersions|TestInstructions|TestGitHubPRDeliveryRuleset|TestWorkLaneGatingRuleset|TestRootHelpGroupsCanonicalCommands|TestCapabilityCatalogCoversVisibleRootCommands' -count=1`.
- `make fmt`, `make vet`, `go test ./... -count=1`, `make build`, and `golangci-lint run --new-from-rev=origin/main ./...` passed.
- `go test -race ./internal/instructions ./pkg/cli -count=1` passed.
- Built-binary comparisons proved default output and `--version=v2` are byte-identical to `versions/v2.md`, while `--version=v1` remains byte-identical to immutable `versions/v1.md`.
- Built-binary `--version=v3` validation returned non-zero with exactly `Error: unsupported instructions version "v3"; available versions: v1, v2`.
- `./bin/kit capabilities instructions --json` confirmed the command remains project-independent and read-only and lists current `v2` plus historical `v1` examples.
- Pre-curation `./bin/kit check 0044-versioned-agent-instructions` and `./bin/kit check --project` passed; the project check reported only the 15 existing legacy compatibility advisories and the existing Constitution-refresh cadence warning.
- Post-curation `./bin/kit complete 0044-versioned-agent-instructions` passed from the supported `deliver` phase, restored the feature to `complete`, and refreshed `docs/PROJECT_PROGRESS_SUMMARY.md`.
- The due semantic project-refresh review found no additional stale project-wide truth beyond the clean-preflight invariant curated in this change. `./bin/kit project refresh --now` recorded the completed review in `.kit.yaml` without rewriting the Constitution automatically.

## OUTCOME

- `kit instructions` now prints the current embedded provider-neutral coding-agent policy as raw Markdown with no banner, clipboard, project-config, file-write, network, subprocess, Git, or GitHub side effect.
- The immutable `v1` source is stored at `internal/instructions/versions/v1.md`, registered separately from repository scaffold versions, and protected by an exact SHA-256 regression assertion.
- The explicit `CurrentAgentVersion` pointer controls default output; `--version=v1` selects the release directly, while empty, malformed, and unavailable selectors fail with actionable available-version errors.
- Root help, Cobra completion registration, automatic-config preflight behavior, capabilities metadata, README guidance, command documentation, and the Constitution package map are aligned with the new command.
- No provider-specific variants, remote lookup, customization, prompt injection, scaffold-version change, or prompt-library behavior was added.
- Immutable `v2` is now the current provider-neutral policy. It proceeds automatically from a proven clean, current default branch with no active lane, while preserving every normal issue, identity, branch, validation, staging, commit, push, and pull-request safeguard.
- In-progress or ambiguous implementation work still requires the lane choice. Choosing to continue allocates separate issue-scoped commits on the existing branch and pull request rather than creating a replacement delivery lane.
- The higher-priority `work-lane-gating` ruleset now implements the same distinction, and focused tests pin both clean-preflight autonomy and the existing same-PR continuation exception.

## REPOSITORY MEMORY

Decision: updated

Rationale: The current-version pointer, immutable historical lookup, raw-output boundary, clean-versus-in-progress lane decision, and same-PR traceability exception are durable product and workflow decisions that code and tests alone would not explain fully.

Artifacts:

- `docs/specs/0044-versioned-agent-instructions/SPEC.md`
- `README.md`
- `docs/commands.md`
- `docs/CONSTITUTION.md`
- `docs/references/rules/work-lane-gating.md`
- `internal/instructions/versions/v2.md`
- `.kit.yaml`
- `docs/PROJECT_PROGRESS_SUMMARY.md`
