# Kit agent workflow audit, November 17, 2026

Kit's provider-neutral design aligns with published GPT-6.2 and Claude coding guidance. Native planning and delegation, project-specific acceptance criteria, and validation against observable state are worth retaining. Two shipped behaviors need correction first. Cleanup can discard local work, and a validation command can return success after skipping a requested check. Both reproduced independently of model choice.

This audit covers Kit at [66d4591](https://github.com/jamesonstone/kit/tree/66d45912116bfb11d037df4709fd6810ad813b72). It examines the canonical contract, generated instructions, CLI entry points, local command-report schema, and project tests. The OpenAI and Anthropic sources were consulted on November 17, 2026. The findings apply to this revision. The suggested changes below are proposals.

Kit already demonstrates several of the intended usage patterns. Its own CI runs native Go tests and project-contract validation as separate steps. For an existing project, the documented maintenance path offers read-only diagnosis with `kit health` and a preview with `kit reconcile --dry-run --diff`. The merge rule distinguishes authority to merge from evidence that the current head passed. Preserve those distinctions when simplifying the instructions. [CI workflow](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/.github/workflows/ci.yml#L37) · [Maintenance commands](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/docs/commands.md#L57) · [Merge rule](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/docs/references/rules/github-pr-merge.md#L31)

## Findings

### 1. High — Cleanup can discard a file that no remote holds

**Owner** — Kit's shipped `work-cleanup` rule and its classifier tests.

In a disposable repository, a pushed worktree contained an ignored, handwritten local database. The shipped classifier returned `RECOVERABLE`. The prescribed worktree removal succeeded and the database disappeared. Checking for a remote copy of the commit establishes nothing about an ignored file. The classifier checks local SOPs and environment files explicitly but assumes the remaining ignored content can be regenerated. The existing classifier tests passed during the audit. [Cleanup classifier and definition](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/docs/references/rules/work-cleanup.md#L87) · [Removal rule](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/docs/references/rules/work-cleanup.md#L55) · [Classifier tests](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/pkg/cli/work_cleanup_classifier_test.go#L76)

**Change and verify.** Preserve the worktree whenever it contains unclassified ignored content. Limit automatic removal to positively identified, regenerable artifacts. Exercise the published workflow against both a disposable cache and a unique, ignored file. Assert that the unique file still exists after cleanup. Until the change lands, retain worktrees with unverified ignored contents.

### 2. High — Combined validation flags can hide a broken feature

**Owner** — The `check` command in `pkg/cli`. Its help also supplies the guide.

A disposable copy of the audited revision contained one additional feature directory with its required `SPEC.md` missing. The CLI built from source produced these results. [Flag handling](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/pkg/cli/check.go#L60)

| Invocation | Exit | What ran |
| --- | ---: | --- |
| `kit check --project` | 0 | Project contract |
| `kit check --all` | 1 | Feature checks detected the missing spec |
| `kit check --project --all` | 0 | Project contract only |

**Change and verify.** Reject the combined invocation with a usage error, or implement both checks and propagate either failure. Run the same fixture through the CLI and assert the exit status. Until then, run the two commands separately. Also correct the help text that describes a V2 spec as required. The current `kit spec` creates V3 specs, and `kit check 0100-agent-functionality-guide` passed on a V3 spec. [Check help](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/pkg/cli/check.go#L24) · [Spec contract](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/pkg/cli/spec.go#L26)

### 3. Medium — Some grants are broader than the task

**Owner** — The universal contract and the browser-validation and cleanup rules.

The contract gives repository rules precedence over chat, with a specific exception for a human stop. Browser validation permits use of any exposed connected profile. Low-disk cleanup can extend to machine-wide tool caches. A host may offer a signed-in browser or a shared cache for unrelated work. Tool availability alone gives no evidence that the current task needs it. This is a scope finding from the shipped text. [Precedence](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/internal/templates/universal_contract.md#L3) · [Browser selection](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/docs/references/rules/testing-and-environment-validation.md#L43) · [Cache cleanup](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/docs/references/rules/work-cleanup.md#L62)

**Change and verify.** Describe the order among Kit documents separately from host permissions and the user's current scope. Prefer a task-owned browser. Reuse a connected profile when the task calls for that account. Make machine-wide cache maintenance an explicit operator choice. Test that a local UI task with two available browser profiles uses the isolated one, and that a task explicitly directed to a connected account can proceed within its grant. [OpenAI computer-use controls](https://developers.openai.com/api/docs/guides/tools-computer-use-integration#restrict-the-environment) · [Anthropic approval guidance](https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/prompting-claude-fable-5-2#approvals)

### 4. Medium — The delivery rule has no path for a single-agent host

**Owner** — The `delivery` and `agent-team-orchestration` rules.

Delivery requires a second agent with no prior context before **any** pull request opens. The delegation rule says to disclose an unavailable verifier and continue. A host without a fresh reviewer can prepare and validate a change but cannot satisfy both instructions for delivery. [Delivery gate](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/docs/references/rules/delivery.md#L85) · [Capability fallback](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/docs/references/rules/agent-team-orchestration.md#L37)

**Change and verify.** Define one outcome for that host. Where project policy permits, record a self-review and route the PR to a human reviewer. Where an independent pre-PR review is mandatory, report that specific gate as blocked. Exercise both host configurations and verify the evidence accurately names who reviewed the work. Keep existing merge and branch-protection requirements.

### 5. Medium — A project convention can reach only one provider

**Owner** — The entry-file renderer and the project-additions guidance in `internal/guide`.

Kit writes the same managed block into each vendor's entry file and preserves additions to each file separately. A team that adds a shared test command only to `AGENTS.md` can get different behavior across hosts. Codex discovers `AGENTS.md`. By default, Claude Code prefers `CLAUDE.md` when both files are present. Configured Claude discovery modes can differ. [Project additions](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/internal/guide/artifacts.go#L12) · [Managed-block test](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/internal/templates/universal_contract_test.go#L119) · [Codex discovery](https://learn.chatgpt.com/docs/agent-configuration/agents-md) · [Claude Code discovery](https://code.claude.com/docs/en/memory#when-claude-code-reads-agentsmd)

**Change and verify.** Direct shared conventions to one project-owned source the managed block already references, such as `docs/references/testing.md` for project checks. Reserve each vendor entry's additions for that host. In disposable projects, give both hosts a shared test instruction and verify each runs the specified command. Preserve existing vendor-specific additions during reconciliation.

### 6. Medium — The contract requests full command discovery for every task

**Owner** — The universal contract's Kit Functionality section and the guide command.

The universal contract says to run `kit guide`. At this revision the text output is 588 lines, 4,100 words, and 30,206 bytes. This is additional to the contract's 1,483 words. Even an audit that already knows its commands receives maintenance flags and unrelated command descriptions. The writing rule adds a separate read before the first response. The sizes are measured. Their effect on latency, cost, or quality still needs a comparison. [Guide instruction](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/internal/templates/universal_contract.md#L38) · [CLI guide](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/pkg/cli/guide.go#L12) · [Writing-rule trigger](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/docs/references/rules/asd-ste100.md#L24)

**Change and verify.** Keep the full guide available for discovery. Trigger it when the agent needs to choose a Kit command and use the existing command-specific help for known commands. Compare both variants on a routine fix and an unfamiliar maintenance task. Measure correct tool choice, input volume, elapsed time, and completion. Evaluate the writing rule separately so the result identifies which change helped. [OpenAI on repository instructions](https://developers.openai.com/blog/rethinking-skills-and-prompts-for-gpt-6-astra) · [Anthropic on rechecking old prompts](https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/prompting-claude-fable-5-2)

## Model-specific implications

| Model | What to retain or measure |
| --- | --- |
| GPT-6.2 Sol and Astra | Retain clear project conventions, reversible decisions within scope, and final verification. Compare Sol on routine fixes and Astra on complex work. Kit never calls the model API, so prefix caching and reasoning controls belong to the host. [OpenAI GPT-6.2 guide](https://developers.openai.com/api/docs/guides/latest-model?model=gpt-6.2-astra) |
| Claude Opus 5.5 | Use ordinary coding and recovery workflows as a baseline. Anthropic recommends starting API evaluations at medium effort. Test whether the agent resumes after a recoverable tool error and reaches the requested outcome. [Opus 5.5 guide](https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/prompting-claude-opus-5-5) |
| Claude Fable 5.2 | Exercise the longest workflow. Anthropic starts API evaluations at high effort. Count premature stops, repeated approvals, and unnecessary delegation before adding instructions. [Fable 5.2 guide](https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/prompting-claude-fable-5-2) |
| Claude Mythos 5.2 | The name “Mythos” here means the contemporary 5.2 release. It shares Fable's underlying model and has separate access and safeguards. Only an actual authorized Mythos run can provide Mythos results. [Anthropic release and availability](https://www.anthropic.com/claude-fable-and-mythos-5-2) |

## Evidence and follow-up

In a synthetic project with a different marker in each entry file, Codex CLI 0.169.0 returned the `AGENTS.md` marker in two fresh sessions, one requested with GPT-6.2 Astra and one with GPT-6.2 Sol. A second read-only trial gave each model a disposable copy of Kit, a binary built from the audited commit, and one feature missing its spec. Asked to validate both scopes, both reported the project contract passed, the feature checks failed, and the exact failing feature. Each result comes from one trial per task and model. Claude Code 2.1.287 was available locally, but its authentication preflight reported no active login. The Claude findings use published host and provider guidance. No live Fable, Mythos, or Opus result is available from this audit.

Kit's Go tests verify generated instructions and deterministic CLI behavior. Its bounded local usage report records command exits, elapsed time, and a pseudonymous project ID. It omits `guide` calls, model identity, and task outcomes. A nonzero check can mean the validator caught a real problem. Use aggregates to choose test cases and keep local usage private. [Testing inventory](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/docs/references/testing.md#L15) · [Event fields](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/internal/usage/types.go#L12) · [Guide exclusion](https://github.com/jamesonstone/kit/blob/66d45912116bfb11d037df4709fd6810ad813b72/pkg/cli/guide_test.go#L40)

For the next evaluation, use three disposable scenarios.

1. Recover from a failing `kit check`, rerun it, and verify the corrected feature on disk.
2. Return an AWS identity mismatch from a fake service. Verify that the agent reports it and makes zero mutation calls.
3. Complete a small PR task against a fake remote, once with a fresh reviewer available and once without. Compare the resulting PR state and review evidence with project policy.

Repeat each scenario and compare the shipped contract to one change at a time. Start each variant in a fresh session so the host loads the intended instructions. Record the host version, model requested and returned when available, effort, final state, unauthorized action attempts, and time or token use. Score each model separately. Keep normal Go CI hermetic. [OpenAI evaluation practices](https://developers.openai.com/api/docs/guides/evaluation-best-practices) · [Anthropic agent evaluation practices](https://www.anthropic.com/research/demystifying-evals-for-ai-agents)
