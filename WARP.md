# WARP

Kit is a portable specification-driven development framework designed to avoid vendor lock-in.

## Kit is the source of truth

- Constitution: `docs/CONSTITUTION.md`
- Feature specs live under: `docs/specs/<feature>/`
  - `SPEC.md` (requirements)
  - `PLAN.md` (implementation plan)
  - `TASKS.md` (executable task list)
  - `ANALYSIS.md` (optional, analysis scratchpad)

## Workflow contract

- Specs drive code. Code serves specs.
- For any change:
  1. locate the relevant feature directory in `docs/specs/<feature>/`
  2. read `SPEC.md` → `PLAN.md` → `TASKS.md`
  3. implement tasks in order
  4. verify (tests / validation steps from plan)
  5. if reality diverges, update `SPEC.md` / `PLAN.md` / `TASKS.md` first, then code

## Multi-feature rule

- Never mix features in one `docs/specs/<feature>/` directory.
- If work spans features, update each feature's docs separately.

## Development Commands

```bash
make build              # build binary to bin/kit
make test               # run all tests
make vet                # static analysis
make fmt                # format code
make install            # install to $GOPATH/bin
```

## Architecture

```text
cmd/kit/                # entry point
internal/
  config/               # .kit.yaml loading
  document/             # markdown parsing/validation
  feature/              # feature numbering/directories
  git/                  # branch management
  rollup/               # PROJECT_PROGRESS_SUMMARY generation
  templates/            # embedded document templates
pkg/cli/                # cobra commands
```
