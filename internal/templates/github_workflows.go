package templates

import (
	"strconv"
	"strings"
)

// BuildAutoAssignWorkflow returns the Kit-managed GitHub Actions workflow that
// assigns newly opened or reopened issues and pull requests to configured users.
func BuildAutoAssignWorkflow(assignees []string) string {
	var builder strings.Builder
	builder.WriteString(`# Kit-managed auto-assignment workflow.
# Update github.default_assignees in .kit.yaml or ~/.config/kit/.kit.yaml, then run kit init --refresh.
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
    name: Assign configured maintainers
    runs-on: ubuntu-latest

    steps:
      - name: Assign issue or pull request
        uses: actions/github-script@v7
        continue-on-error: true
        with:
          script: |
            const assignees = [`)
	for i, assignee := range assignees {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString("\n              ")
		builder.WriteString(strconv.Quote(assignee))
	}
	if len(assignees) > 0 {
		builder.WriteString("\n            ")
	}
	builder.WriteString(`];
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
`)
	return builder.String()
}
