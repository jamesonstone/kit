# Kit Workflows

Kit v2 centers feature work on one durable artifact: `SPEC.md`.

## V2 Single-SPEC Workflow

The v2 single-`SPEC.md` workflow is Kit's most structured operating engine. It
is the clearest path when a problem needs deliberate clarification, planning,
implementation, validation, reflection, documentation sync, delivery gating,
and evidence.

```text
Idea / input
  ↓
kit spec <feature>
  ↓
SPEC.md seeded with clarification.status=open
  ↓
clarify → ready → implement → validate → reflect → deliver/complete
```

`kit spec <feature>` remains prompt-producing by default. The generated
supervisor prompt instructs a coding agent to keep all durable workflow state in
`SPEC.md`, including the Agent Team Plan when implementation, validation, or
review benefits from specialist lanes. In `clarify`, it resolves repository
facts before asking only for material non-discoverable choices. A ready feature
continues with safe in-scope work without routine clarification or approval
pauses.

## Project Initialization

Run once, then refine as the project matures:

```bash
kit init
kit project refresh
kit reconcile
```

```text
┌──────────────┐
│ Constitution │  ← global constraints, principles, priors
└──────────────┘
```

Use `kit project refresh` when early feature work reveals durable
project-level rules, vocabulary, or constraints that should update
`CONSTITUTION.md`.

## Optional Research Material

```text
┌──────────────┐
│ Notes/Inputs │  ← reference materials, screenshots, research, constraints
└──────────────┘
```

Feature notes and design materials live under `docs/notes/<feature>/` when
needed. They are supporting inputs, not replacements for `SPEC.md`.

## V2 Artifact Details

1. 📜 **Constitution** - strategy, constraints, long-term project rules, and priors.
2. 📐 **SPEC.md** - front matter phase and clarification state plus thesis, context, clarifications, requirements, assumptions, acceptance criteria, implementation plan, task checklist, validation map, reflection notes, documentation updates, delivery decision, and evidence.
3. 🧠 **Legacy staged artifacts** - historical `BRAINSTORM.md`, `PLAN.md`, and `TASKS.md` files retained for upgraded projects or explicit legacy flows.

When a core workflow command runs without a feature argument, its selector only
shows features eligible for that command's next stage. Completed stages are
omitted from earlier-stage selectors.

If `kit spec` has no eligible existing feature candidates to list, it prompts
for a new feature name and starts the v2 single-SPEC intake.

## `kit spec` Intake

For a new `SPEC.md`, `kit spec <feature>` asks for:

1. one thesis/goal editor entry
2. one delivery-intent answer

Delivery intent options:

- `1` / Enter / `yes` - create a Kit-managed issue, branch, and pull request later
- `2` / `no` - capture the idea only; no issue/branch/PR intent yet
- `3` / `continue` - coding agent should continue on the current branch/current issue/current PR if one exists

Existing `SPEC.md` files are preserved by default. Use `--revise-thesis` to
append a dated thesis note and refresh delivery intent.

New specs default to `clarification.status: open`,
`clarification.confidence: 0`, and `clarification.unresolved_questions: 1`.
Normal adoption backfills missing clarification metadata without disturbing the
body; `--prompt-only` remains read-only and does not write adoption metadata.

`kit spec` does not create issues, branches, commits, pushes, or PRs during
intake. Delivery mutations remain behind the repo-local delivery hard gate.

## Typical Flow

```bash
kit spec my-feature
kit status
kit map
kit resume my-feature
```

```text
You / team idea
  ↓
kit spec my-feature
  ↓
SPEC.md + v2 supervisor prompt
  ↓
evidence-first clarification when needed → ready gate → same-thread implementation
  ↓
validate → reflect → deliver/complete
```

## Autonomous Loop

`kit loop workflow [feature]` is the execution wrapper for prompt-level
workflow automation. The durable state remains `SPEC.md`; direct execution
stays behind explicit loop/run behavior.

Each child run receives the durable feature context plus only its current phase
contract. Kit does not reinject the complete lifecycle supervisor prompt on
every fresh subprocess, and code still validates phase transitions.

During `clarify`, loop automation researches repository facts and updates
`SPEC.md` without guessing user intent. If a material non-discoverable choice
remains, it stops with numbered `Open Questions`, recommended defaults, and the
impact of each answer. It advances only when
`clarification.status` is `ready`, confidence meets the configured threshold,
and unresolved questions are `0`.

```yaml
loop:
  min_confidence: 95
  max_iterations: 20
  agent:
    command: codex
    args: ["--ask-for-approval", "never", "exec", "--model", "gpt-5.6", "--sandbox", "workspace-write", "--ignore-user-config", "--color", "never", "-"]
```

```bash
kit loop workflow my-feature --dry-run
kit loop workflow my-feature
kit loop workflow my-feature --until validate
kit loop review
kit loop review --pr 14
```

Loop evidence is written under `.kit/loops/<run-id>/`.

## Prompt-System Benchmarks

`kit improve run --suite default` is a deterministic command-capability smoke
test. It does not render prompts or measure model quality.

`kit improve run --suite prompt-system --kit-binary <path>` renders the
representative prompt surfaces three times and records command/assertion
success, normalized output hashes, output size, local duration, and exact
runner/binary/suite provenance. A failed command or assertion makes the CLI
exit nonzero. Compare binaries only when the suite-definition SHA-256 matches.
Provider cost, model latency, conversational turns, and actual tool/subagent
routing remain unobservable unless a benchmark explicitly invokes and
instruments a model.

## V2 Readiness And Completion

The v2 supervisor prompt performs readiness gates inside `SPEC.md` before
implementation begins. It requires:

- clarified assumptions
- `clarification.status: ready`
- confidence at or above the configured goal
- `clarification.unresolved_questions: 0`
- binary-verifiable acceptance criteria
- a task checklist
- a validation map
- an Agent Team Plan or recorded single-lane exception
- documentation sync
- reflection notes
- evidence before delivery

## Legacy V1 Foundations

Kit v2 was built from the original staged workflow:

```text
brainstorm → specification → plan → tasks → implement → reflection
```

```text
┌─────────────┐   ┌───────────────┐   ┌─────────────┐   ┌─────────────┐   ┌──────────────┐
│ BRAINSTORM  │ → │ SPECIFICATION │ → │    PLAN     │ → │    TASKS    │ → │  REFLECTION  │
└─────────────┘   └───────────────┘   └─────────────┘   └─────────────┘   └──────────────┘
      idea             clarified           approach          checklist          review
```

That foundation still matters: v1 made ambiguity explicit, forced planning
before execution, kept task progress durable, and closed the loop with review.
Kit v2 keeps those semantics but removes the user-facing command sequence.

Historical `BRAINSTORM.md`, `PLAN.md`, and `TASKS.md` files remain readable and
non-disruptive in upgraded projects. Their commands live under `kit legacy` for
finishing old staged work.

## Legacy Staged Commands

Use `kit legacy <command>` only when finishing existing v1 staged work or
capturing backlog research that intentionally lives outside the active v2 lane.

```bash
kit legacy --help
kit legacy brainstorm my-feature --prepare
kit legacy brainstorm --backlog shared-refactor
```

## Project Structure

```text
.kit.yaml                    # configuration and local prompt overrides
docs/
  CONSTITUTION.md            # project-wide constraints
  PROJECT_PROGRESS_SUMMARY.md
  notes/
    0001-my-feature/
      .gitkeep
      design/                # frontend materials when --profile=frontend is used
        .gitkeep
        screenshots/
          .gitkeep
        references/
          .gitkeep
  specs/
    0001-my-feature/
      SPEC.md                # v2 durable feature workflow artifact
      BRAINSTORM.md          # optional legacy staged research artifact
      PLAN.md                # optional legacy staged plan artifact
      TASKS.md               # optional legacy staged task artifact
      ANALYSIS.md            # optional
  references/
    rules/
      frontend-ui.md         # optional durable pointer-loaded rulesets
```

New v2 `SPEC.md` files include front matter with `workflow_version: 2`, a
workflow `phase`, and structured clarification state. Legacy staged artifacts
still include front matter when created, and Kit commands read front matter
first before falling back to legacy
body metadata.
