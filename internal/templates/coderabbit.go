package templates

import "strings"

const CodeRabbitConfig = `reviews:
  path_filters:
    - "!docs/**"
    - "!AGENTS.md"
    - "!CLAUDE.md"
`

const IssueTemplate = `---
name: Issue
about: Describe the work and fill Verified when before an agent starts.
---

## Description

-

## Verified when

Fill this in before an agent starts.

- Done:
- Out of scope:
- Evidence: passing test, screenshot, log, or trace
- Hypothesis or expected answer (optional):
`

const PullRequestTemplate = `## Description

-

## How to Test

1.

## Evidence

- Commands: command and real output or excerpt
- Decisions and assumptions: defaults, scope calls, conflicting requirements reinterpreted
- Checks: for each check, the property it verified
- Not checked:
- Fresh-context review:

## Ticket

Closes #[ticket number]
`

// historicalPullRequestTemplates are exact Kit-owned bodies from earlier
// releases. Reconcile replaces only these, never a project edit.
var historicalPullRequestTemplates = []string{`## Description

-

## How to Test

1.

## Ticket

Closes #[ticket number]
`}

// BuildPullRequestTemplate keeps GitHub's existing closing syntax and selects
// an explicit full-URL placeholder for Linear repositories.
func BuildPullRequestTemplate(provider string) string {
	return renderPullRequestTemplate(PullRequestTemplate, provider)
}

// IsShippedPullRequestTemplate reports whether content is an exact Kit-owned
// pull request template, including earlier releases and the Linear ticket line.
func IsShippedPullRequestTemplate(content string) bool {
	if content == PullRequestTemplate || content == BuildPullRequestTemplate("linear") {
		return true
	}
	for _, historical := range historicalPullRequestTemplates {
		if content == historical || content == renderPullRequestTemplate(historical, "linear") {
			return true
		}
	}
	return false
}

// historicalIssueTemplates are exact Kit-owned issue templates from earlier
// releases. Reconcile replaces only these, never a project edit.
var historicalIssueTemplates []string

// IsShippedIssueTemplate reports whether content is an exact Kit-owned issue
// template, including earlier releases.
func IsShippedIssueTemplate(content string) bool {
	if content == IssueTemplate {
		return true
	}
	for _, historical := range historicalIssueTemplates {
		if content == historical {
			return true
		}
	}
	return false
}

func renderPullRequestTemplate(base, provider string) string {
	if provider == "linear" {
		return strings.Replace(base, "Closes #[ticket number]", "<!-- Full canonical Linear URL; preserve in the squash commit footer. Follow delivery for lifecycle. -->\n[Linear issue URL]", 1)
	}
	return base
}
