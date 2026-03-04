# PLAN

## SUMMARY

- Implement a primary `main` pipeline that creates next patch tag and publishes release in one workflow, plus manual semantic tag release support.

## APPROACH

- Add a `main` branch workflow that computes next `vMAJOR.MINOR.PATCH` tag from existing semantic tags and pushes it.
- In the same `main` workflow, run vet/test, build artifacts with GoReleaser, and publish GitHub release with generated notes for that tag.
- Keep a tag workflow (`v*`) for manual semantic tag release publication.
- Use Git tags as semantic version source of truth.

## COMPONENTS

- `.github/workflows/release-tag-main.yml`: semantic version tag creation and automated release publication on `main`.
- `.github/workflows/release-publish.yml`: gated build + release publication for manual semantic tag pushes.
- `.goreleaser.yaml`: build matrix, archive format, checksums, and linker version injection.

## DATA
- Semantic versions are derived from existing tags matching `vMAJOR.MINOR.PATCH`.
- If no semantic tags exist, initialize with `v1.0.0`.

## INTERFACES

- Trigger 1: GitHub push event to `main`.
- Trigger 2: GitHub push event for tags `v*` (manual tags).
- Release target: GitHub Releases API via `gh release create --generate-notes`.

## RISKS

- Concurrent `main` pushes could race on next tag computation.
- Mitigation: workflow concurrency group with queueing (`cancel-in-progress: false`).
- Existing tag re-run can collide.
- Mitigation: tag existence checks and idempotent release upload logic.
- Tag pushes created by `GITHUB_TOKEN` do not trigger downstream workflows.
- Mitigation: perform release publication inside the same `main` workflow after tag creation.

## TESTING

- Enforce `make vet` and `make test` before artifact publication.
- Validate YAML and GoReleaser config with `actionlint` and `goreleaser check`.
- Verify end-to-end by pushing to `main` and confirming tag + release creation.
