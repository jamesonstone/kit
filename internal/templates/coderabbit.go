package templates

import "strings"

const CodeRabbitConfig = `reviews:
  path_filters:
    - "!docs/**"
    - "!AGENTS.md"
    - "!CLAUDE.md"
`

const PullRequestTemplate = `## Description

-

## How to Test

1.

## Ticket

Closes #[ticket number]
`

// BuildPullRequestTemplate keeps GitHub's existing closing syntax and selects
// an explicit full-URL placeholder for Linear repositories.
func BuildPullRequestTemplate(provider string) string {
	if provider == "linear" {
		return strings.Replace(PullRequestTemplate, "Closes #[ticket number]", "<!-- Full canonical Linear URL; preserve in the squash commit footer. Follow delivery for lifecycle. -->\n[Linear issue URL]", 1)
	}
	return PullRequestTemplate
}
