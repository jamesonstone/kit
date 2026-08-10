package bootstrap

import (
	"sort"
	"strconv"
	"strings"
)

const gitignoreHeader = "# Kit local environment, cache, and scratch artifacts"

var gitignorePatterns = []string{
	".env",
	".envrc",
	".kit/cache/",
	".kit/tmp/",
	".kit/*.tmp",
	".kit/*.lock",
	"tmp/",
}

const envrcStarter = `#!/bin/sh
set -eu

dotenv_if_exists
`

const makefileStarter = `.DEFAULT_GOAL := help

.PHONY: help

help:
	@printf '%s\n' 'Project developer workflow'
	@printf '%s\n' ''
	@printf '%s\n' 'Run the Kit repository-bootstrap workflow to add verified targets.'
`

const codeRabbitStarter = `reviews:
  path_filters:
    - "!docs/**"
    - "!AGENTS.md"
    - "!CLAUDE.md"
`

const pullRequestStarter = `## Description

-

## How to Test

1.

## Ticket

Closes #[ticket number]
`

const autoAssignStarter = `# Kit bootstrap auto-assignment workflow.
name: Auto assign

on:
  issues:
    types: [opened, reopened]
  pull_request_target:
    types: [opened, reopened, ready_for_review]

permissions:
  issues: write
  pull-requests: read

jobs:
  assign:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/github-script@v7
        continue-on-error: true
        with:
          script: |
            const assignees = [];
            if (assignees.length === 0) {
              core.info("No Kit auto-assignees configured; skipping.");
              return;
            }
            await github.rest.issues.addAssignees({
              owner: context.repo.owner,
              repo: context.repo.repo,
              issue_number: context.issue.number,
              assignees,
            });
`

func buildAutoAssign(values []string) string {
	seen := map[string]bool{}
	var assignees []string
	for _, value := range values {
		value = strings.TrimSpace(strings.TrimPrefix(value, "@"))
		if value != "" && !seen[value] {
			seen[value] = true
			assignees = append(assignees, value)
		}
	}
	sort.Strings(assignees)
	quoted := make([]string, 0, len(assignees))
	for _, assignee := range assignees {
		quoted = append(quoted, strconv.Quote(assignee))
	}
	return strings.Replace(autoAssignStarter, "const assignees = [];",
		"const assignees = ["+strings.Join(quoted, ", ")+"];", 1)
}

const readmeStarter = `# Project

Describe the repository purpose, boundaries, setup, and operating notes.
`

const readmeBadgesStart = "<!-- BEGIN KIT-MANAGED README BADGES -->"
const readmeBadgesEnd = "<!-- END KIT-MANAGED README BADGES -->"
const readmeBadgesBlock = readmeBadgesStart + `
<!-- The repository-bootstrap agent may add only badges supported by repository evidence. -->
` + readmeBadgesEnd

const readmeMaintainersStart = "<!-- BEGIN KIT-MANAGED README MAINTAINERS -->"
const readmeMaintainersEnd = "<!-- END KIT-MANAGED README MAINTAINERS -->"
const readmeMaintainersBlock = readmeMaintainersStart + `
## Maintainers

Document verified repository maintainers here.
` + readmeMaintainersEnd

const constitutionStart = "<!-- BEGIN KIT-MANAGED CONSTITUTION BASELINE -->"
const constitutionEnd = "<!-- END KIT-MANAGED CONSTITUTION BASELINE -->"
const constitutionBaseline = constitutionStart + `
## Kit-Managed Baseline

- Treat repository-local rules and workflows resolved by Kit as the coding-agent contract.
- Preserve project-owned work and validate changes before delivery.
- Promote only demonstrated durable project-wide truth into this Constitution.
- Keep feature rationale in current specifications and reusable practices in project references.
` + constitutionEnd

const constitutionStarter = `# CONSTITUTION

` + constitutionBaseline + `

## PRINCIPLES

<!-- Add only demonstrated durable project-wide principles. -->

## CONSTRAINTS

<!-- Add only demonstrated invariant constraints. -->

## NON-GOALS

<!-- Add explicit project-wide non-goals when repository evidence supports them. -->

## DEFINITIONS

<!-- Define stable project vocabulary when it emerges. -->
`

const progressSummaryStarter = `# PROJECT PROGRESS SUMMARY

## FEATURE PROGRESS TABLE

| ID | FEATURE | PATH | PHASE | SUMMARY |
| -- | ------- | ---- | ----- | ------- |

## PROJECT INTENT

No current specification evidence has been summarized yet.

## GLOBAL CONSTRAINTS

See docs/CONSTITUTION.md.

## FEATURE SUMMARIES

No current feature specifications were found during bootstrap.

## LAST UPDATED

Not yet populated from repository evidence.
`
