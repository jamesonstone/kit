# PLAN

## SUMMARY

- Add a dedicated `pkg/cli/version.go` command that prints the effective installed version to `stdout`.
- Register it on the root CLI, place it in command ordering, and document it in README.
- Correct the local Makefile build/install path so linker-injected versions default to the latest semantic Git tag.

## APPROACH

- Implement a no-arg Cobra command with `RunE` that writes the resolved version to `cmd.OutOrStdout()`.
- Keep formatting intentionally minimal so the command is script-friendly.
- Prefer existing release/build injection behavior and fall back to Go build info for `go install` builds.

## COMPONENTS

- `pkg/cli/version.go`: new command definition and runtime output.
- `pkg/cli/root.go`: help ordering so `version` appears in root command listings.
- `Makefile`: default `VERSION` derivation for local build/install flows.
- `README.md`: command table update.
- `pkg/cli/version_test.go`: regression coverage for output behavior.

## DATA

- Source of truth: `pkg/cli.Version`, with `runtime/debug.ReadBuildInfo()` as fallback when needed.
- No new config, files, or persistent state.

## INTERFACES

- New CLI interface: `kit version`.
- Existing interface retained: `kit --version`.

## RISKS

- Help ordering could hide the new command if not added to `commandOrder`.
- Tests could become flaky if they depend on global `Version` state without restoring it.

## TESTING

- Add tests for linker-version precedence, build-info fallback, and command stdout.
- Run `go test ./...`.
- Run `kit version` manually to confirm CLI behavior.
- Run `make build && ./bin/kit version` to confirm local tagged builds report the expected semantic version.
