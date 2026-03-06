# SPEC

## SUMMARY

- Add an explicit `kit version` subcommand that prints the installed Kit version from the same build metadata already used by `--version`.
- The command must be stable, script-friendly, and visible in CLI help so users can inspect their installed release version directly.

## PROBLEM

- Kit currently exposes version information only through the root `--version` flag.
- Users and scripts do not have a first-class `kit version` command path.
- Existing update-related documentation already refers to `kit version`, creating a product gap.

## GOALS

- Expose `kit version` as a root subcommand.
- Print the installed Kit version to stdout with a stable format.
- Reuse the existing linker-injected version source of truth.
- Reuse the existing linker-injected version source of truth with a fallback for module-installed binaries.
- Ensure local `make build` and `make install` flows do not bake a stale hard-coded semantic version into the binary.
- Keep the command lightweight with no filesystem or network requirements.
- Surface the command in CLI help and README command listings.

## NON-GOALS

- Checking GitHub for the latest available release.
- Printing extended build metadata such as commit SHA or build date.
- Changing the behavior of the existing `--version` flag.
- Enforcing semantic-version formatting for non-release `dev` builds in this phase.

## USERS

- Developers verifying which Kit version is installed locally.
- Automation or scripts that need a stable command form for reading the installed version.
- Future self-update flows that need a canonical local version command.

## REQUIREMENTS

- Add a new root command: `kit version`.
- The command must accept no positional arguments.
- The command must print the currently installed Kit version followed by a newline.
- The command must read from the same `pkg/cli.Version` value used by existing build and release flows.
- The command must prefer the existing `pkg/cli.Version` value used by build and release flows.
- If the linker version is empty or `dev`, the command must fall back to Go build info when a module version is available.
- Local Makefile-driven builds must derive their default semantic version from the latest matching Git tag rather than a fixed literal.
- The command must exit successfully for both release builds and non-release builds such as `dev`.
- The command must appear in root help output and command ordering.
- Public documentation must mention the new command.

## ACCEPTANCE

- Running `kit version` on an installed release prints the installed version string and exits with status `0`.
- Running `kit version` on a binary installed via `go install ...@latest` prints the module version when available.
- Running a local `make build` or `make install` from a tagged repository produces a binary whose `kit version` matches the latest semantic tag by default.
- Running `kit version` on a local development build prints the current build version identifier and exits with status `0`.
- Root help output includes `version` in the available commands list.
- README command documentation includes `kit version`.

## EDGE-CASES

- The build version is `dev` rather than a release tag.
- The build version includes a leading `v` prefix from Git tag injection.
- The command is used from scripts that expect stdout-only output.

## OPEN-QUESTIONS

- None.
