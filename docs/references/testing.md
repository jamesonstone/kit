# Testing Reference

## Purpose

- Record Kit's durable validation commands, automation, and evidence expectations
- Follow `rules/testing-and-environment-validation.md` for the cross-project testing contract
- Keep feature-specific testing details in the current feature's `SPEC.md` VALIDATION and OUTCOME sections

## Code-Level Validation

| Layer | Command | PR workflow or check | Required | Notes |
| --- | --- | --- | --- | --- |
| Formatting | `gofmt -l .` (must print nothing) | `CI / go` | yes | Fix with `make fmt` |
| Static analysis | `make vet` and `make lint` (`golangci-lint run ./...`) | `CI / go` | yes | No repository `.golangci` config; defaults apply |
| Unit and integration | `make test` (`go test -v ./...`) | `CI / go` | yes | Hermetic; tests use temp dirs, git fixtures, and stubbed registry/GitHub access |
| Build | `make build` | `CI / go`, local `.githooks/pre-commit` | yes | Hook enabled with `make install-git-hooks` |
| Release workflow contract | `go test ./internal/releaseworkflow` | included in `make test` | yes | Guards release scripts, workflows, and the v3 module path |
| Source size | `kit reconcile --all --dry-run` (source-file audit) | not in CI | yes before delivery | 300-line limit for handwritten Go source and tests |
| Improve smoke suites | `kit improve run --suite default --json` | `Kit Improve Validate` (path-filtered) | no | String-level CLI smoke checks; no model involved |

## High-Level Suites

| Suite | Type | Environment | Command | Automation | Evidence |
| --- | --- | --- | --- | --- | --- |
| None | not applicable | not applicable | not applicable | not applicable | Kit is a local CLI; end-to-end behavior is covered by Go integration tests |

## Environment Preflights

- Go version comes from `go.mod`
- No services, databases, browsers, or cloud targets are required; production environments are not applicable

## Credentials And Test Data

- Tests need no credentials; registry and GitHub calls are stubbed
- Avoid recording usage events from test or development binaries into `~/.config/kit/usage`; point `HOME` at a temporary directory when running built binaries manually
- Follow `rules/deletion-safety.md` for cleanup: default retained state to recoverable deletion and require exact post-outline manual confirmation before hard delete

## Evidence And Retention

- Keep `tmp/` ignored; CI results live in the GitHub Actions run for the pull request
- `tests/RUN_STATUS.md` is not applicable because Kit has no high-level suites

## Automation And Fallbacks

- `.github/workflows/ci.yml` runs formatting, vet, tests, build, and lint on every pull request
- Release workflows rerun `make vet` and `make test` before tagging

## Known Gaps

- The source-size audit is not yet enforced in CI
- `kit improve` smoke suites run only when their path filter matches
