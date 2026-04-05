# SPEC

## SUMMARY

- Add a canonical self-update command (`kit upgrade`) and keep `kit update` as
  a hidden deprecated compatibility entry point so users can move to the latest
  Kit release from GitHub Releases without manual install steps.
- Updates must be safe and predictable: never leave users with a broken binary, and always provide clear outcome and recovery guidance.

## PROBLEM

- Kit users currently need manual update flows (for example, reinstalling with Go tooling), which is slower and inconsistent.
- There is no built-in way to check installed version versus latest GitHub release and apply an in-place upgrade.
- Friction in the update process increases version drift and delays adoption of fixes and improvements.

## GOALS

- Provide a single built-in canonical command path for self-updating Kit from
  official GitHub Releases.
- Keep `kit update` callable with identical behavior while teaching
  `kit upgrade` as the canonical command.
- Detect current installed Kit version and latest stable release version using semantic version comparison.
- Upgrade in place when a newer compatible release is available and the install location is writable.
- Preserve a working Kit binary on all failure paths.
- Return explicit user-facing status for all outcomes: upgraded, already up to date, unsupported, or failed.

## NON-GOALS

- Managing updates for package-manager installs (for example Homebrew/Scoop) in this phase.
- Auto-updating in the background or on command startup.
- Supporting prerelease/nightly channel selection in this phase.
- Updating non-binary project artifacts or feature documents.
- Introducing release signing or notarization as part of this feature.

## USERS

- Developers and maintainers running a locally installed Kit binary.
- Contributors who want a low-friction path to stay on the latest stable Kit version.
- CI and automation contexts where explicit version updates are run as a command step.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## RELATIONSHIPS

none

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| release artifacts | code | `.github/workflows/`, `.goreleaser.yaml` | stable release downloads for self-update | active |
| version command | code | `pkg/cli/version.go` | local installed-version resolution | active |
| project docs | doc | `docs/PROJECT_PROGRESS_SUMMARY.md` | update-related user guidance and status references | active |
| GitHub Releases | external | `https://github.com/jamesonstone/kit/releases` | latest-release lookup and artifact download | active |

## REQUIREMENTS

- Expose a canonical `upgrade` command from the Kit root CLI.
- Keep `update` callable as a hidden deprecated compatibility entry point.
- The command must query the latest stable Kit release from `https://github.com/jamesonstone/kit/releases`.
- The command must compare current Kit version to latest stable release using semantic version rules.
- If already current, the command must make no filesystem changes and report both current and latest versions.
- If an update is available, the command must download the release artifact matching the running OS/ARCH.
- The command must verify artifact integrity against published release checksums before replacing the binary.
- The command must replace the Kit executable atomically or with equivalent safe semantics for the platform.
- On any failure, the existing Kit executable must remain runnable.
- If safe self-update is not possible (for example non-writable install path), the command must fail with actionable remediation steps.
- The command output must include current version, target version, and final status.

## ACCEPTANCE

- Running `kit upgrade` on an older installed release upgrades to latest stable release and subsequent `kit version` reports the new version.
- Running `kit upgrade` when already up to date exits successfully, changes nothing, and reports no upgrade needed.
- Running `kit update` produces the same results as `kit upgrade`.
- If checksum validation fails, upgrade aborts and the original binary remains functional.
- If release asset for current platform is missing, command fails with clear guidance and no binary replacement.
- If write permission is insufficient for the install path, command fails with actionable output and no partial update state.

## EDGE-CASES

- Current binary version is `dev` or otherwise non-semantic.
- GitHub API or asset download fails due to network errors, timeouts, or rate limits.
- Latest release is a prerelease while current behavior expects stable-only updates.
- Concurrent executions of `kit upgrade` target the same binary path.
- Running update from a symlinked binary path.
- Platform-specific replacement limitations (especially Windows executable replacement behavior).

## OPEN-QUESTIONS

- Should prereleases be ignored by default, with no opt-in in this phase?
- For `dev` builds, should the command refuse update or allow update to latest stable?
- Is Windows support required in the first release, or can v1 scope be macOS/Linux only?
- Is a dry-run mode (for example `--check`) required in v1?
- Should fallback guidance explicitly include `go install github.com/jamesonstone/kit/cmd/kit@latest` when in-place replacement is not possible?
- Should update behavior be blocked for package-manager detected installs, or attempted with warning?
- Is checksum verification against `checksums.txt` sufficient for v1 without signature verification?
- What exit code policy should be used for "already up to date" versus "update applied"?
- Should the command support pinning a target version (for example `--version vX.Y.Z`) in v1?
- What maximum acceptable runtime should be targeted for update checks on normal network conditions?
