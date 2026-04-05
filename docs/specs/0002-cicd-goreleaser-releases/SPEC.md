# SPEC

## SUMMARY

- Add an automated release pipeline that versions Kit from semantic Git tags and publishes cross-platform GitHub release artifacts from the main branch workflow.

## PROBLEM

- Kit lacks an automated release pipeline for cross-platform binary distribution.
- Releases are not consistently versioned and published from `main`.

## GOALS

- Trigger a release flow on every push to `main`.
- Enforce semantic versioning with patch increments, starting at `v1.0.0`.
- Build binaries for Linux, macOS, and Windows on `amd64` and `arm64`.
- Publish release artifacts to GitHub Releases.
- Generate release notes automatically.
- Keep release gating with `make vet` and `make test`.

## NON-GOALS

- Implement minor or major version auto-bumping.
- Add package manager publishing (Homebrew, Scoop, etc.).
- Add signing or notarization in this phase.

## USERS

- Maintainers publishing Kit binaries.
- Users downloading prebuilt binaries from GitHub Releases.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## RELATIONSHIPS

none

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| release workflow | code | `.github/workflows/` | release publication and tag automation | active |
| GoReleaser config | code | `.goreleaser.yaml` | cross-platform artifact matrix and archive settings | active |
| version command | code | `pkg/cli/version.go` | installed-version surface referenced by release docs | active |
| project progress summary | doc | `docs/PROJECT_PROGRESS_SUMMARY.md` | release visibility and project-state reporting | active |

## REQUIREMENTS

- On push to `main`, compute next semantic version patch tag.
- If no existing semantic tag, first release tag is `v1.0.0`.
- Create and push a Git tag for the computed version.
- On push to `main`, run vet/test gates and publish release artifacts for the computed tag.
- On semantic tag push (`v*`), support manual release publication.
- Use GoReleaser to generate archives and checksums for target OS/ARCH matrix.
- Create or update a GitHub Release with generated notes and upload artifacts.

## ACCEPTANCE

- A push to `main` creates a new semantic tag with patch increment.
- The same `main` workflow publishes a release containing Linux/macOS/Windows artifacts for amd64/arm64.
- Release contains auto-generated notes.
- Workflow fails if vet/test fail.

## EDGE-CASES

- Non-semver tags are ignored for version calculation.
- Existing tag collisions fail safely before release upload.
- Re-run of tag release updates artifacts on existing release via clobber upload.
- Release publication does not rely on tag-triggered workflow chaining from `GITHUB_TOKEN`.

## OPEN-QUESTIONS

- None.
