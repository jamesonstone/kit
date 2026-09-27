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
| Project check | `kit check --project` (includes the 300-line source audit from `.kit.yaml`) | `CI / go` | yes | Also verifies the managed contract block and rule documents |

## High-Level Suites

| Suite | Type | Environment | Command | Automation | Evidence |
| --- | --- | --- | --- | --- | --- |
| None | not applicable | not applicable | not applicable | not applicable | Kit is a local CLI; end-to-end behavior is covered by Go integration tests |

## Environment Preflights

- Go version comes from `go.mod`
- No services, databases, browsers, or cloud targets are required; production environments are not applicable

## Credentials And Test Data

- Tests need no credentials; registry and GitHub calls are stubbed
- Test binaries and development builds never record usage; set `KIT_USAGE_DISABLED=1` when running a released binary against scratch projects
- Follow `rules/deletion-safety.md` for cleanup: default retained state to recoverable deletion and require exact post-outline manual confirmation before hard delete

## Evidence And Retention

- Keep `tmp/` ignored; CI results live in the GitHub Actions run for the pull request
- `tests/RUN_STATUS.md` is not applicable because Kit has no high-level suites

## Automation And Fallbacks

- `.github/workflows/ci.yml` runs formatting, vet, tests, build, lint, and `kit check --project` on every pull request
- Release workflows rerun `make vet` and `make test` before tagging

## Known Gaps

- None known
