# Kit Improve Evals

This directory contains committed benchmark definitions for `kit improve`.

- `suites/` selects held-in and held-out task groups.
- `tasks/` contains declarative benchmark tasks.
- `fixtures/repos/` contains small checked-in repositories copied into
  disposable `.kit/improve/runs/<id>/workspaces/` directories.
- `schemas/` documents the JSON/YAML contracts used by commands and artifacts.

Generated traces, candidates, scorecards, and reports live under
`.kit/improve/` and are ignored by git.

## Suite contracts

`default` is a capability smoke suite. Its tasks query `kit capabilities` and
verify small deterministic strings. A passing default run proves that those
catalog entries remain discoverable; it does not exercise generated prompts,
invoke a model, or measure coding-agent quality.

`prompt-system` renders representative current and legacy prompt surfaces with
the Kit executable selected by `--kit-binary`. It repeats every task three
times, treats required-output assertions as hard correctness gates, records
prompt/output size separately, and includes the generated loop model capability
contract. Use it for identical-definition before/after comparisons:

```bash
kit improve run --suite prompt-system --kit-binary /path/to/baseline/kit --json
kit improve run --suite prompt-system --kit-binary /path/to/candidate/kit --json
```

Compare runs only when `provenance.suite_definition_sha256` is identical. Also
record the runner and evaluated Kit binary hashes; the runner implements the
benchmark semantics, while the selected Kit binary produces the output under
test.

## What a run measures

- command success, exit code, error, timeout, and duration;
- configured assertion success and required-output completeness;
- changed files and allowed-surface violations;
- stdout line, word, byte, and estimated-token counts;
- SHA-256 output stability across repeated tasks after replacing the disposable
  workspace path with `{{workspace}}`;
- exact suite/fixture/task definition and binary provenance.

Estimated tokens are `ceil(stdout bytes / 4)`. This is a transparent size proxy,
not provider token accounting. Command duration is local process wall time, not
model latency unless a task actually invokes a model.

The deterministic suites do not observe provider cost, model latency,
conversation turns, live clarification/approval decisions, or actual tool and
subagent calls. Prompt assertions and fake-agent end-to-end tests can verify the
instructions governing those behaviors, but must not be reported as observed
model behavior.

Task execution is defined by `commands` and `assertions`. `input_prompt`,
`persona`, `expected_behavior`, `mutation_policy`, and known-failure metadata
describe intent but do not invoke a model or change scoring. Held-out selectors
exclude matching tasks from a run; V1 does not yet perform an automatic
baseline-versus-candidate held-out comparison. The fixed `seed` is provenance,
not evidence of randomized trials.

## Failure and score semantics

A nonzero or timed-out task command fails its trace even when a stdout assertion
happens to match. Any failed trace makes the run manifest fail, and `kit improve
run` exits nonzero after emitting the result. Failure signatures identify the
command, assertion, or allowed-surface cause rather than selecting a static
known failure mode. Mined clusters report the task surface encoded in that
signature and calculate flake rate from passing versus matching failed repeats
for the affected tasks.

`kit improve validate` currently validates candidate metadata only. Its score
is always `0` and its acceptance is `metadata-only` for a well-formed proposed
candidate. It is not a behavioral score. Use identical `run` suites plus normal
repository validation to judge a candidate.
