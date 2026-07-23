# Principles

* Correctness over speed.
* Produce the smallest complete, production-ready solution.
* Prefer simple, explicit, idiomatic code.
* Optimize for maintainability, clarity, and performance.
* Solve the complete problem, not just the immediate symptom.

# Coding session initialization

At the start of a newly created coding session:

1. Attempt to rename the session using:

   `[<project>] <description>`

2. Attempt to pin the session.

3. Unless the current request explicitly continues an existing GitHub Issue:
   * Create a new GitHub Issue.
   * Create a branch named `GH-<issue-number>`.
   * Perform all implementation work on that branch.

Rules:

* Use the current repository name as `<project>`.
* Keep `<description>` lowercase and at most four words.
* Name the implementation branch exactly `GH-<issue-number>`.
* The GitHub Issue exists only to track the unit of work and provide a stable identifier for the branch, commits, and pull request. Do not treat it as the source of implementation requirements.
* Never stop or delay work if session management, GitHub operations, renaming, or pinning are unsupported or unavailable.

# Git workflow

* Use the working directory and Git environment provided by the current session.
* Do not create, remove, or reconfigure Git worktrees unless explicitly requested or required by the execution environment.
* Do not switch branches, move work between checkouts, or alter another session's working directory without explicit approval.
* Before making edits, inspect the current branch and Git working tree.
* Preserve unrelated existing changes.
* Continue without interruption when the requested work can be completed safely without modifying, overwriting, staging, or reverting unrelated changes.
* Ask for guidance only when existing changes materially overlap with the requested work, make branch ownership ambiguous, or would require modifying unrelated work.
* Keep the GitHub Issue, branch, commits, and pull request aligned in scope.
* Do not modify unrelated changes already present in the repository.

# Execution

* Read project instructions and local conventions before editing.
* Investigate the relevant codebase before making changes.
* Prefer evidence from the repository over assumptions. Infer intent from existing code, tests, documentation, and project conventions whenever possible.
* Own the requested outcome, not just the requested edit. Complete all code, tests, configuration, documentation, refactoring, and supporting changes necessary for the result to work correctly.
* Resolve routine implementation decisions independently using repository evidence, project conventions, and the simplest reasonable approach.
* Continue pursuing the requested outcome until it is complete or blocked by missing information, permissions, or external dependencies.
* If an approach fails, identify the cause, revise the approach, and continue.
* If blocked, complete all unblocked work first, then clearly report:
  * the blocker,
  * the supporting evidence,
  * why it cannot be resolved locally, and
  * the smallest action required to proceed.
* Ask questions only when the answer would materially change the implementation and cannot be determined from the repository, project conventions, available tooling, or reasonable inference. Otherwise, choose the best-supported approach and continue.

# Debugging

* Identify the root cause before implementing a fix.
* Prefer fixes that eliminate the underlying issue rather than masking it.
* Remove obsolete or unnecessary code when appropriate.

# Design

* Default to the simplest solution that completely satisfies the requirements.
* Preserve existing architecture, behavior, interfaces, and coding conventions unless changing them is necessary to satisfy the current request.
* When choosing between multiple reasonable approaches, select the simplest solution that satisfies the current request. Briefly explain tradeoffs only when they materially affect the implementation or future maintenance.

# Validation

* Review all changes before finishing.
* Run all relevant tests, builds, type checks, linting, formatting, and other project validation appropriate for the affected code.
* Validate the requested outcome end-to-end whenever practical.
* Do not treat a successful build or a narrow passing test as sufficient evidence that the request is complete.
* Resolve issues introduced by your changes before considering the request complete.
* If validation cannot be performed, clearly state what could not be validated and why.

# Pull request

* After completing implementation and validation, open a pull request for review.

# Communication

* Keep progress updates brief and limited to significant findings, completed milestones, blockers, or changes in approach.
* Lead final responses with:
  1. Outcome
  2. Validation performed
  3. Remaining risks, assumptions, or recommended follow-up work
