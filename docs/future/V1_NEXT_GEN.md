# V1 Next Gen

## Status

- Type: future architecture specification
- Binding status: non-binding until converted into normal `docs/specs/<feature>/` artifacts
- Intended audience: Kit maintainers and coding agents preparing implementation specs
- Current product impact: none

## Summary

Next-generation Kit should evolve from a static prompt and documentation harness into a local intent and alignment runtime for coding agents.

The current Kit harness improves agent work by forcing work through durable documents, prompt structure, readiness gates, and routing guidance. The next version should keep that durable Markdown foundation, but add a runtime layer that observes agent activity, compares it to a compact intent contract, and intervenes only when useful.

The primary product thesis is:

- better model understanding should reduce the need for a large upfront harness
- the harness should become thinner, more dynamic, and more precise
- Kit should supervise alignment at runtime instead of relying only on front-loaded instructions
- portability must remain grounded in Markdown, YAML, and open integration surfaces

The desired end state is:

```text
human request
  -> intent confidence loop
  -> compact intent contract
  -> context compiler
  -> coding agent or foundation model
  -> event stream
  -> local supervisor
  -> deterministic policy engine
  -> continue, inject context, nudge, pause, block, or resume
```

## Problem

Coding agents still fail in predictable ways:

- they over-read broad context instead of selecting the smallest useful source
- they implement before resolving important ambiguity
- they drift from the user's real goal after several tool calls
- they follow stale instructions or miss repo-local source-of-truth documents
- they waste model tokens restating context that was already known
- they continue after a tool result changes the risk profile
- they finish with plausible output even when the actual requested outcome is incomplete

Current Kit reduces these failures by generating strong prompts and keeping state in canonical documents. That is effective, but it is front-loaded. Once an external coding agent starts running, Kit has limited ability to observe and steer the run.

The next version should add a live control loop.

## Goals

- Define a local runtime that supervises coding-agent work while preserving Kit's Markdown-first portability.
- Convert weak human prompts into explicit, compact intent contracts.
- Ask clarifying questions until goal confidence is high enough for the selected risk level.
- Compile only the smallest relevant context into each agent or model interaction.
- Normalize live agent activity into a provider-agnostic event stream.
- Use deterministic rules for obvious policy decisions.
- Use a small local model only for ambiguous classification, summarization, drift detection, contract comparison, and routing decisions.
- Intervene through the narrowest available mechanism:
  - add context
  - nudge
  - pause and ask
  - block a tool call
  - stop and resume with corrected instructions
- Support multiple hosts through adapters:
  - Codex CLI and Codex Desktop
  - Claude Code and Claude Desktop
  - Warp terminal and Oz agents
  - future MCP-capable clients
- Keep all durable project truth inspectable in repository files.

## Non-Goals

- Do not train a new foundation model.
- Do not make the local supervisor model responsible for writing production code.
- Do not replace existing Kit workflow commands in the first implementation.
- Do not require one specific coding agent or model provider.
- Do not depend on hidden cloud state.
- Do not make raw terminal-text scraping the only integration path.
- Do not claim universal hard enforcement where a host only exposes weak hooks.
- Do not store secrets, API keys, or private model credentials in Kit documents.
- Do not require fine-tuning for V1.

## Core Concepts

### Intent Contract

An intent contract is the compact, machine-readable description of what the user wants.

It should be stored as YAML or JSON-compatible front matter plus a readable Markdown body. It is not a replacement for `SPEC.md`, `PLAN.md`, or `TASKS.md`. It is a runtime contract for the current work session.

Required fields:

```yaml
intent_contract_version: 1
goal: "one sentence outcome"
scope:
  allowed:
    - "paths, modules, docs, commands, or behavior explicitly in scope"
  disallowed:
    - "known out-of-scope work"
success_criteria:
  - "observable completion condition"
risk_level: "low|medium|high"
confidence:
  current: 0.0
  required: 0.95
open_questions:
  - id: "Q1"
    question: "question text"
    default: "recommended default"
    status: "open|answered|accepted_default"
authoritative_sources:
  - type: "file|command|url|memory|user"
    location: "source location"
    reason: "why this source matters"
policy:
  allow_edits: true
  allow_commands: true
  require_approval_for:
    - "destructive commands"
    - "scope expansion"
intervention_mode: "observe|advise|guard|enforce"
```

### Goal Confidence Loop

Before execution, Kit should estimate whether the goal is clear enough to proceed.

The confidence loop should evaluate:

- objective clarity
- scope boundaries
- success criteria
- source-of-truth availability
- risk and reversibility
- ambiguity that could change implementation direction
- whether the task is planning-only or implementation-ready

If confidence is below threshold, Kit should ask numbered clarifying questions with recommended defaults. If the user accepts defaults, the contract should record that acceptance.

Default thresholds:

| Risk Level | Default Required Confidence | Behavior |
| --- | ---: | --- |
| low | 0.80 | proceed with documented assumptions |
| medium | 0.90 | ask targeted questions before edits |
| high | 0.95 | ask until no unresolved assumptions remain |

### Event Stream

The event stream is a normalized record of what the agent is doing.

Kit should consume the strongest event source available:

1. native host hooks or SDK events
2. MCP tool calls through Kit-owned tools
3. shell and file events from a Kit wrapper
4. terminal output as a fallback

Events should use a provider-neutral schema:

```yaml
event_version: 1
event_id: "uuid or stable local id"
session_id: "session id"
turn_id: "turn id when available"
source: "codex|claude-code|claude-desktop|warp|pty|mcp|manual"
type: "prompt|model_delta|tool_call|tool_result|file_read|file_write|command|plan_update|final_output|error"
timestamp: "RFC3339"
payload: {}
derived:
  paths:
    - "path touched or referenced"
  risk_signals:
    - "scope_drift"
  summary: "short human-readable event summary"
```

### Local Supervisor

The local supervisor is a small model-backed judge used only after deterministic rules have handled obvious cases.

The supervisor should answer bounded questions:

- classify the current task phase
- summarize the last N events
- detect possible drift from the intent contract
- compare proposed actions against allowed scope
- choose the next control action
- identify missing context
- recommend one clarifying question when necessary

It must produce structured output:

```yaml
decision_version: 1
decision: "continue|inject_context|nudge|pause_and_ask|block|stop_and_resume"
confidence: 0.0
risk: "none|low|medium|high"
reason: "one or two sentences"
evidence:
  - "specific observed event or contract field"
recommended_message: "optional text to inject or show"
requires_user: false
```

The supervisor must not:

- edit files directly
- approve its own scope expansion
- invent new requirements
- override deterministic safety rules
- route around user approval requirements

### Policy Engine

The policy engine is deterministic code that decides what actions are allowed.

It should run before and after the local supervisor.

Pre-supervisor policy examples:

- block destructive commands unless explicitly allowed
- block edits outside allowed paths in enforce mode
- require user approval before changing public interfaces in high-risk mode
- require source-of-truth reads before implementation begins

Post-supervisor policy examples:

- reject malformed supervisor decisions
- downgrade unsupported actions for the current host
- cap intervention frequency
- prevent repeated nudges with the same message
- require explicit user confirmation for `block` when running in advisory mode

### Context Compiler

The context compiler selects the smallest useful context for the next model or agent interaction.

Inputs:

- intent contract
- repo-local docs
- feature artifacts
- recent event summaries
- files and commands referenced by the agent
- dependency and relationship metadata
- user corrections

Outputs:

- compact prompt fragment
- source list
- confidence score
- excluded-context notes when useful

The compiler should prefer:

- exact file sections over whole files
- current feature docs over project-wide docs
- explicit dependency links over broad search
- recent tool evidence over stale assumptions
- short corrective messages over restating full instructions

## Runtime Architecture

### Components

```text
kit CLI
  - existing commands remain
  - new commands manage contracts, runtime sessions, and adapters

kitd runtime
  - local process
  - owns event ingestion, policy checks, local supervisor calls, and intervention decisions

MCP server
  - portable tool and context surface for clients that support MCP

hook adapters
  - Codex hooks
  - Claude Code hooks
  - future host-specific hooks

PTY adapter
  - fallback wrapper for terminal tools without native hooks

rules emitter
  - keeps AGENTS.md, CLAUDE.md, WARP.md, and other rule files aligned with Kit docs

local model provider
  - pluggable local inference endpoint
  - optional remote fallback only when explicitly configured

event store
  - transparent JSONL files under repo-local Kit runtime state
  - no hidden database in V1
```

### Proposed Packages

```text
internal/runtime/
  contract.go
  session.go
  events.go
  policy.go
  decisions.go
  store.go

internal/supervisor/
  provider.go
  prompt.go
  schema.go
  evaluator.go

internal/adapters/
  codex_hooks.go
  claude_hooks.go
  mcp_server.go
  pty.go
  warp_rules.go

pkg/cli/
  supervise.go
  contract.go
  runtime.go
```

### Durable Files

V1 should keep durable state visible and reviewable.

```text
.kit.yaml
docs/future/V1_NEXT_GEN.md
.kit/
  runtime/
    sessions/
      <session-id>/
        INTENT_CONTRACT.md
        events.jsonl
        decisions.jsonl
        context.jsonl
        summary.md
    adapters/
      codex-hooks.json
      claude-hooks.json
      warp-rules.md
```

If `.kit/` conflicts with current project assumptions, the first formal feature spec must decide whether runtime files belong under `.kit/`, `.codex/kit/`, or `docs/runtime/`.

## Command Surface

### `kit contract`

Create, inspect, or update the active intent contract.

Examples:

```bash
kit contract new "Fix the failing auth callback test"
kit contract status
kit contract questions
kit contract answer Q1 "Only touch AuthRouteGate"
```

Required behavior:

- infer initial contract from the user prompt
- ask clarifying questions when confidence is below threshold
- support `--risk low|medium|high`
- support `--required-confidence`
- write a readable contract artifact
- never overwrite an existing active contract without confirmation

### `kit supervise`

Run or attach supervision for a coding-agent session.

Examples:

```bash
kit supervise -- codex
kit supervise --adapter codex-hooks
kit supervise --adapter claude-hooks
kit supervise --pty -- claude
kit supervise status
```

Required behavior:

- require an active intent contract unless `--ad-hoc` is passed
- start event ingestion
- select the strongest available adapter
- report intervention mode
- write events and decisions to JSONL
- show concise user-facing notices for interventions

### `kit adapters`

Install or print adapter configuration.

Examples:

```bash
kit adapters list
kit adapters install codex
kit adapters install claude-code
kit adapters print warp
```

Required behavior:

- never silently modify global agent config
- show exact files that would change
- support `--output-only`
- support project-local configuration first

### `kit mcp`

Run Kit's MCP server.

Example:

```bash
kit mcp serve
```

Required tools:

- `kit_contract_read`
- `kit_context_search`
- `kit_context_compile`
- `kit_event_record`
- `kit_decision_request`
- `kit_policy_check`

The MCP server is primarily a portability surface. It should not be the only enforcement mechanism.

## Adapter Requirements

### Codex Adapter

The Codex adapter should use native hooks where available.

Minimum hook coverage:

- session start: inject current contract summary
- user prompt submit: add missing contract context or block unsafe prompts
- pre-tool use: block or annotate risky commands and edits
- post-tool use: summarize results and detect drift
- stop: prevent premature stop when completion criteria are unmet

V1 should treat Codex hooks as the strongest initial integration because they can participate in the agent loop directly.

### Claude Code Adapter

The Claude Code adapter should use hooks and MCP.

Minimum hook coverage:

- user prompt submit
- pre-tool use
- post-tool use
- stop or equivalent completion event

Claude Code should receive the same intent contract and policy decisions as Codex, but adapter behavior must account for Claude-specific hook semantics.

### Claude Desktop Adapter

Claude Desktop should be supported first through:

- MCP server tools
- repository instruction files
- contract documents

V1 should not assume Claude Desktop exposes strong runtime interception. If no hook-equivalent surface exists, Claude Desktop support is advisory rather than enforceable.

### Warp Adapter

Warp should be supported first through:

- `AGENTS.md` project rules
- optional `WARP.md` emitted from the same source contract
- MCP server configuration
- PTY or Oz CLI wrapper where practical

Warp support should distinguish:

- local desktop agents
- cloud agents through Oz
- terminal sessions that can only be monitored through visible output

### PTY Adapter

The PTY adapter is the universal fallback.

Capabilities:

- launch a command under supervision
- read stdout and stderr
- detect common progress markers
- send interrupt signals when configured
- record command output summaries

Limitations:

- cannot reliably see hidden model state
- cannot cleanly intercept every host action
- may only intervene by interrupting the process or printing guidance

## Local Model Requirements

V1 does not require fine-tuning.

The local model should support:

- low-latency structured classification
- small context windows efficiently
- deterministic-enough output under constrained schemas
- local execution through a pluggable provider
- optional CPU-only mode for slower but portable behavior
- optional GPU acceleration when available

Initial implementation should use prompting and examples, not training.

Optimization order:

1. deterministic rules
2. input normalization
3. strict output schema
4. small few-shot prompt set
5. trace-based evaluation
6. model selection and quantization
7. fine-tuning only after enough labeled traces exist

Fine-tuning may be considered when:

- there are at least several hundred high-quality labeled traces
- prompt-only supervision has repeatable failure modes
- latency or cost matters enough to justify maintenance
- the output schema is stable

Full model training is out of scope.

## Intervention Semantics

Kit should prefer the smallest useful intervention.

| Decision | Meaning | Example |
| --- | --- | --- |
| `continue` | no action needed | agent is reading the correct source |
| `inject_context` | add missing facts without stopping | remind agent of allowed paths |
| `nudge` | visible steering, agent may continue | "Stay in planning mode" |
| `pause_and_ask` | stop for user clarification | confidence dropped below threshold |
| `block` | prevent a specific action | destructive command outside scope |
| `stop_and_resume` | interrupt and continue with corrected prompt | agent is solving the wrong task |

Mode behavior:

| Mode | Behavior |
| --- | --- |
| `observe` | record decisions only |
| `advise` | show non-blocking notices |
| `guard` | block high-confidence safety or scope violations |
| `enforce` | block any contract violation above configured threshold |

V1 should default to `advise` for new installations and require explicit opt-in for `guard` or `enforce`.

## Evaluation Plan

### Trace Dataset

Every supervised session should produce a replayable trace:

- intent contract
- event stream
- supervisor decisions
- policy decisions
- interventions
- final outcome
- user corrections

Trace files must be local and inspectable. Users must be able to delete them.

### Metrics

Quality metrics:

- intervention precision
- missed-drift rate
- false-block rate
- clarification usefulness
- final task completion rate
- user correction frequency

Efficiency metrics:

- input tokens avoided
- repeated-context reduction
- time to first useful action
- time lost to unnecessary intervention
- local supervisor latency

Safety metrics:

- destructive command blocks
- out-of-scope edit blocks
- skipped-source detections
- premature-stop detections

### Golden Trace Tests

Add a fixture format for trace replay:

```text
testdata/runtime/
  scope-drift/
    INTENT_CONTRACT.md
    events.jsonl
    expected_decisions.jsonl
  missing-source/
    INTENT_CONTRACT.md
    events.jsonl
    expected_decisions.jsonl
```

Replay tests should verify:

- deterministic policy decisions
- local supervisor prompt construction
- schema validation
- adapter downgrade behavior
- intervention mode behavior

## Security And Privacy

V1 must be conservative.

Rules:

- default to local-only runtime state
- never send event traces to a remote model unless explicitly configured
- redact known secret patterns before local model prompts and trace summaries
- do not store raw command output by default when it likely contains secrets
- expose a `kit supervise purge` command
- keep adapter installation explicit and reviewable
- treat MCP tools as untrusted boundaries
- record which adapter produced each event
- distinguish advisory decisions from enforced blocks

## Implementation Phases

### Phase 0: Contract And Trace Format

Deliverables:

- define intent contract schema
- define event schema
- define decision schema
- add JSONL read/write helpers
- add trace replay tests
- add `kit contract new/status`

Acceptance:

- a user can create and inspect a contract
- traces can be read back deterministically
- no coding-agent integration is required yet

### Phase 1: Deterministic Policy Engine

Deliverables:

- implement policy checks for path scope, command risk, source-of-truth requirements, and completion criteria
- add `observe`, `advise`, `guard`, and `enforce` modes
- add policy replay fixtures

Acceptance:

- obvious out-of-scope edits are detected from event fixtures
- destructive commands are flagged or blocked according to mode
- policy output is schema-valid and explainable

### Phase 2: Local Supervisor

Deliverables:

- provider interface for local model calls
- supervisor prompt builder
- structured decision parser
- fallback behavior when no local model is available
- eval fixture runner

Acceptance:

- ambiguous traces produce structured supervisor decisions
- invalid model output fails closed to deterministic behavior
- local model is optional in observe mode

### Phase 3: MCP Server

Deliverables:

- `kit mcp serve`
- contract read tool
- context compile tool
- policy check tool
- event record tool
- decision request tool

Acceptance:

- MCP-capable clients can read the active contract
- MCP tool calls are logged as events
- clients can ask Kit for compact context instead of loading all docs

### Phase 4: Codex Adapter

Deliverables:

- project-local Codex hook config generator
- hook handlers for session start, user prompt, pre-tool, post-tool, and stop
- event ingestion from hook payloads
- intervention output mapping for Codex hook semantics

Acceptance:

- Codex receives contract context on session start
- risky tool calls can be blocked in guard mode
- post-tool drift can inject corrective context
- stop hook can continue a turn when completion criteria are unmet

### Phase 5: Claude Code Adapter

Deliverables:

- project-local Claude Code hook config generator
- hook handlers equivalent to Codex where supported
- MCP setup guidance
- host-specific downgrade behavior

Acceptance:

- Claude Code can consume the active contract
- supported tool calls can be intercepted
- unsupported actions degrade to advisory mode with clear notices

### Phase 6: Warp And PTY Adapters

Deliverables:

- `AGENTS.md` and optional `WARP.md` rule output
- MCP setup guidance for Warp
- PTY wrapper for terminal agents
- Oz CLI integration notes if stable

Acceptance:

- Warp can use Kit contract/rules as context
- PTY wrapper can record and summarize visible output
- unsupported enforcement is clearly labeled as advisory

### Phase 7: Product Hardening

Deliverables:

- docs
- examples
- privacy controls
- trace purge
- config reference
- cross-adapter compatibility matrix

Acceptance:

- users can understand what is observed, stored, and enforced
- all adapters have documented capabilities and limitations
- V1 can be enabled per project without changing global behavior

## Configuration

Proposed `.kit.yaml` additions:

```yaml
runtime:
  enabled: false
  default_mode: advise
  event_store: .kit/runtime
  redact_secrets: true
  local_model:
    provider: "ollama"
    model: "local-supervisor"
    timeout_ms: 2000
  confidence:
    low_risk: 0.80
    medium_risk: 0.90
    high_risk: 0.95
  adapters:
    codex:
      enabled: false
      install_hooks: project
    claude_code:
      enabled: false
      install_hooks: project
    warp:
      enabled: false
      emit_rules: true
```

## Open Questions

- Should runtime state live under `.kit/runtime`, `docs/runtime`, or another visible location?
- Should `kitd` be a long-running daemon or should hook handlers start short-lived processes?
- Which local model provider should be the default recommendation?
- Should Kit ship a default local-model prompt pack before any model-specific tuning?
- What is the minimum viable Codex hook set for a useful first release?
- Should `WARP.md` become a first-class scaffold target or remain generated guidance only?
- How much raw terminal output should be retained by default?
- Should V1 support remote supervisor models, or keep that explicitly out of scope until V2?

## Definition Of Done For V1

V1 is complete when:

- Kit can create a compact intent contract from a user request.
- Kit can ask clarifying questions until required confidence is reached.
- Kit can record provider-neutral runtime events.
- Kit can replay traces through deterministic policy checks.
- Kit can optionally call a local supervisor model for ambiguous decisions.
- Kit can expose contract and context tools through MCP.
- Kit can install project-local Codex and Claude Code adapter configuration.
- Kit can supervise at least one coding agent with real pre-tool and post-tool interventions.
- Kit can support weaker hosts through Markdown, MCP, and advisory PTY monitoring.
- All durable state is local, inspectable, and removable.
- Existing Kit prompt and document workflows continue to work when runtime supervision is disabled.

