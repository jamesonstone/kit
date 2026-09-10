package templates

import (
	"strconv"
	"strings"
)

// BuildAutoAssignWorkflow returns the Kit-managed GitHub Actions workflow that
// assigns newly opened or reopened issues and pull requests to the initiator
// and configured maintainers.
func BuildAutoAssignWorkflow(assignees []string) string {
	var builder strings.Builder
	builder.WriteString(`# Kit-managed auto-assignment workflow.
# Update github.default_assignees in .kit.yaml or ~/.config/kit/.kit.yaml, then run kit init --refresh.
# Configured maintainers are always included; the issue or pull request initiator is added automatically.
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
    name: Assign initiator and configured maintainers
    runs-on: ubuntu-latest

    steps:
      - name: Assign issue or pull request
        uses: actions/github-script@v7
        continue-on-error: true
        with:
          script: |
            const configured = [`)
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
            const author = context.payload?.issue?.user?.login ?? context.payload?.pull_request?.user?.login ?? null;
            const seen = new Set();
            const assignees = [];
            for (const login of [...configured, author]) {
              if (typeof login !== "string" || login.length === 0) {
                continue;
              }
              const key = login.toLowerCase();
              if (key.endsWith("[bot]")) {
                continue;
              }
              if (seen.has(key)) {
                continue;
              }
              seen.add(key);
              assignees.push(login);
            }
            if (assignees.length === 0) {
              core.info("No assignees resolved (no configured maintainers and no human initiator); skipping.");
              return;
            }
            if (author) {
              core.info(` + "`Assigning ${assignees.join(\", \")} (includes initiator ${author}).`" + `);
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
