# PLAN

## SUMMARY

- Add a real self-update command surface under canonical `kit upgrade` while
  preserving `kit update` as a hidden deprecated compatibility command.
- Reuse `currentVersion()` and existing CLI registration patterns while keeping
  backward compatibility without exposing both names in default help.
- Fetch the latest stable GitHub Release, verify checksums, extract the platform binary in memory, and replace the executable safely.

## APPROACH

- [PLAN-01][SPEC-01][SPEC-02][SPEC-03] Add a new `pkg/cli/upgrade.go` file
  that registers canonical `upgrade` and hidden deprecated `update` commands
  sharing the same behavior and `--yes` confirmation bypass flag.
- [PLAN-02][SPEC-04][SPEC-05][SPEC-06][SPEC-07] Resolve the current version with `currentVersion()`, fetch the latest stable release from the GitHub Releases API, skip prereleases, and compare versions with stdlib-only semver parsing.
- [PLAN-03][SPEC-08][SPEC-09][SPEC-10] Derive the expected release artifact from `.goreleaser.yaml` naming rules, download the artifact plus `checksums.txt`, and verify SHA-256 before any filesystem write.
- [PLAN-04][SPEC-11][SPEC-12][SPEC-13] Extract the executable from `tar.gz` or `zip` in memory and replace the installed binary using a same-directory temp file plus atomic rename semantics, with Windows-specific best-effort fallback.
- [PLAN-05][SPEC-14][SPEC-15][SPEC-16][SPEC-17] Keep user-facing output exact and actionable for already-current, confirmation, success, timeout, rate-limit, missing-asset, checksum, and permission-failure paths.
- [PLAN-06][SPEC-18][SPEC-19][SPEC-20] Add command ordering, README utility docs, and focused unit tests for asset naming, checksum parsing, version comparison, and executable path resolution.

## COMPONENTS

- `pkg/cli/upgrade.go`
  - command registration
  - release fetch and asset resolution
  - checksum validation
  - archive extraction
  - executable replacement
- `pkg/cli/root.go`
  - add help ordering for canonical `upgrade` and compatibility handling for
    `update`
- `pkg/cli/upgrade_test.go`
  - table-driven unit tests for helper logic
- `README.md`
  - add utility command docs
- `docs/specs/0003-inplace-upgrade-update/TASKS.md`
  - record execution order and completion state

## DATA

- External input comes only from the public GitHub Releases API and release asset downloads.
- Release metadata fields used:
  - `tag_name`
  - `prerelease`
  - `assets[].name`
  - `assets[].browser_download_url`
- No new persisted state is introduced.
- Temporary replacement files are created only in the installed binary directory and must be cleaned on failure.

## INTERFACES

- New root commands:
  - `kit upgrade`
  - hidden deprecated `kit update`
- Shared flag:
  - `--yes`, `-y`
- Output goes to `cmd.OutOrStdout()`
- Errors go to `cmd.ErrOrStderr()`

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| release workflow | code | `.github/workflows/` | stable-release source and artifact publication | active |
| GoReleaser config | code | `.goreleaser.yaml` | release asset naming and checksums | active |
| version command | code | `pkg/cli/version.go` | local version comparison and reporting | active |
| GitHub Releases | external | `https://github.com/jamesonstone/kit/releases` | update metadata and asset retrieval | active |

## RISKS

- Compatibility must preserve `kit update` invocation without keeping it
  visible in the default root help.
- `os.Executable()` may return a resolved target instead of a symlink path; replacement must intentionally target that resolved path.
- Windows cannot atomically rename over a running executable, so the fallback rename-to-`.old` flow must restore the original binary on failure.
- A partial or corrupted download must never reach the replacement step.
- API rate limiting and timeouts must fail fast with manual-install guidance.

## TESTING

- Add table-driven unit tests for:
  - `selectAssetName`
  - `parseChecksums`
  - `compareVersions`
  - `buildExecutablePath`
- Run:
  - `make vet`
  - `make test`
  - `make build`
  - `./bin/kit upgrade --help`
  - `./bin/kit update --help`
  - `./bin/kit --help`
