# TASKS

## TASKS

- [x] Add feature spec, plan, and task docs for CI/CD release flow.
- [x] Add GoReleaser config for linux/darwin/windows on amd64/arm64.
- [x] Add `main` workflow to compute and push next patch semantic tag.
- [x] Add tag workflow to run vet/test, build artifacts, and publish GitHub release with generated notes.
- [x] Reflect on implementation and identify workflow trigger bug.
- [x] Fix workflow chaining by publishing release from `main` workflow after tag creation.
- [x] Add semantic-tag validation to manual tag release workflow.
- [x] Request CodeRabbit prompt-only findings from the user and address reflection findings with workflow comments.
- [x] Verify end-to-end workflow execution on next push to `main`.

## DEPENDENCIES

- Tag publication depends on successful semantic version calculation.
- Main release publication depends on successful tag creation and vet/test gates in the same workflow.
- Manual tag release publication depends on semantic tag format and successful vet/test gates.

## NOTES

- Semantic versions are sourced from Git tags, not a dedicated version file.
- Reflection finding: tag pushes made with `GITHUB_TOKEN` do not trigger downstream push workflows.
- Mitigation applied: main workflow now performs tag creation and release publication in one run.
- Policy clarified: automatic version bump is patch-only; major/minor remain manual semantic tags.
