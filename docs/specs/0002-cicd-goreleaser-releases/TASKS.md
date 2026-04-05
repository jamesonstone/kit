# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Add feature spec, plan, and task docs for CI/CD release flow. | done | agent | |
| T002 | Add GoReleaser config for linux/darwin/windows on amd64/arm64. | done | agent | T001 |
| T003 | Add `main` workflow to compute and push next patch semantic tag. | done | agent | T001, T002 |
| T004 | Add tag workflow to run vet/test, build artifacts, and publish GitHub release with generated notes. | done | agent | T002, T003 |
| T005 | Reflect on implementation and identify workflow trigger bug. | done | agent | T004 |
| T006 | Fix workflow chaining by publishing release from `main` workflow after tag creation. | done | agent | T005 |
| T007 | Add semantic-tag validation to manual tag release workflow. | done | agent | T004, T006 |
| T008 | Run reflection review and address findings with workflow comments. | done | agent | T006, T007 |
| T009 | Verify end-to-end workflow execution on next push to `main`. | done | agent | T006, T008 |

## TASK LIST

Use markdown checkboxes to track completion:

- [x] T001: Add feature spec, plan, and task docs for CI/CD release flow.
- [x] T002: Add GoReleaser config for linux/darwin/windows on amd64/arm64.
- [x] T003: Add `main` workflow to compute and push next patch semantic tag.
- [x] T004: Add tag workflow to run vet/test, build artifacts, and publish GitHub release with generated notes.
- [x] T005: Reflect on implementation and identify workflow trigger bug.
- [x] T006: Fix workflow chaining by publishing release from `main` workflow after tag creation.
- [x] T007: Add semantic-tag validation to manual tag release workflow.
- [x] T008: Run reflection review and address findings with workflow comments.
- [x] T009: Verify end-to-end workflow execution on next push to `main`.

## TASK DETAILS

### T001

- **GOAL**: add the docs that define the CI/CD release flow.
- **SCOPE**: create the spec, plan, and task docs for release automation.
- **ACCEPTANCE**: the feature docs exist and describe the release flow.

### T002

- **GOAL**: add the release artifact matrix configuration.
- **SCOPE**: configure GoReleaser for linux, darwin, and windows on amd64 and arm64.
- **ACCEPTANCE**: the release config covers the intended platforms and architectures.

### T003

- **GOAL**: compute and publish the next semantic tag from `main`.
- **SCOPE**: add the main workflow that derives and pushes the next patch release tag.
- **ACCEPTANCE**: a push to `main` can create the next release tag.

### T004

- **GOAL**: publish release artifacts after gates pass.
- **SCOPE**: add the release workflow that runs vet, test, build, and release publication.
- **ACCEPTANCE**: the workflow uploads GitHub release artifacts with generated notes.

### T005

- **GOAL**: capture the trigger-chain issue discovered during reflection.
- **SCOPE**: identify why tag pushes did not trigger the downstream release flow.
- **ACCEPTANCE**: the workflow gap is documented.

### T006

- **GOAL**: fix release chaining by publishing in the same workflow run.
- **SCOPE**: move release publication into the `main` workflow after tag creation.
- **ACCEPTANCE**: release publication no longer depends on downstream tag-trigger chaining.

### T007

- **GOAL**: validate manual semantic tag releases.
- **SCOPE**: add semantic-tag validation to the manual release workflow.
- **ACCEPTANCE**: non-semantic tags fail safely.

### T008

- **GOAL**: close the reflection loop on the release workflow.
- **SCOPE**: review the implementation and address findings in workflow comments.
- **ACCEPTANCE**: reflection findings are resolved or recorded.

### T009

- **GOAL**: verify the release pipeline end to end.
- **SCOPE**: confirm the main workflow creates the tag and publishes the release.
- **ACCEPTANCE**: the next push to `main` exercises the full flow successfully.

## DEPENDENCIES

- Tag publication depends on successful semantic version calculation.
- Main release publication depends on successful tag creation and vet/test gates in the same workflow.
- Manual tag release publication depends on semantic tag format and successful vet/test gates.

## NOTES

- Semantic versions are sourced from Git tags, not a dedicated version file.
- Reflection finding: tag pushes made with `GITHUB_TOKEN` do not trigger downstream push workflows.
- Mitigation applied: main workflow now performs tag creation and release publication in one run.
- Policy clarified: automatic version bump is patch-only; major/minor remain manual semantic tags.
