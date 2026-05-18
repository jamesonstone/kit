package templates

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
