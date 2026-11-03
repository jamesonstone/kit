package cli

import (
	"strings"

	"github.com/jamesonstone/kit/v3/internal/templates"
)

const codeRabbitConfigPath = ".coderabbit.yaml"
const envPath = ".env"
const envrcPath = ".envrc"
const makefilePath = "Makefile"
const autoAssignWorkflowPath = ".github/workflows/auto-assign.yml"
const pullRequestTemplatePath = ".github/pull_request_template.md"
const issueTemplatePath = ".github/ISSUE_TEMPLATE/issue.md"
const gitignorePath = ".gitignore"

func missingGitignorePatterns(content string) []string {
	existing := make(map[string]struct{})
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		existing[line] = struct{}{}
	}

	var missing []string
	for _, pattern := range kitGitignorePatterns() {
		if _, ok := existing[pattern]; !ok {
			missing = append(missing, pattern)
		}
	}
	return missing
}

func kitGitignorePatterns() []string {
	var patterns []string
	for _, line := range strings.Split(templates.Gitignore, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns
}

func appendGitignorePatterns(content string, patterns []string) string {
	var builder strings.Builder
	builder.WriteString(strings.TrimRight(content, "\n"))
	if strings.TrimSpace(content) != "" {
		builder.WriteString("\n\n")
	}
	builder.WriteString("# Kit local generated environment, cache, and scratch artifacts\n")
	for _, pattern := range patterns {
		builder.WriteString(pattern)
		builder.WriteString("\n")
	}
	return builder.String()
}
