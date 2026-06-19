package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/templates"
)

const codeRabbitConfigPath = ".coderabbit.yaml"
const envPath = ".env"
const envrcPath = ".envrc"
const pullRequestTemplatePath = ".github/pull_request_template.md"
const gitignorePath = ".gitignore"

func scaffoldGitignore(projectRoot string, outputOnly bool) error {
	path := filepath.Join(projectRoot, gitignorePath)
	if !document.Exists(path) {
		if err := document.Write(path, templates.Gitignore); err != nil {
			return fmt.Errorf("failed to create %s: %w", gitignorePath, err)
		}
		if !outputOnly {
			fmt.Printf("  ✓ Created %s\n", gitignorePath)
		}
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", gitignorePath, err)
	}
	missing := missingGitignorePatterns(string(data))
	if len(missing) == 0 {
		if !outputOnly {
			fmt.Printf("  ✓ %s exists, skipping\n", gitignorePath)
		}
		return nil
	}

	updated := appendGitignorePatterns(string(data), missing)
	if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to update %s: %w", gitignorePath, err)
	}
	if !outputOnly {
		fmt.Printf("  ✓ Updated %s\n", gitignorePath)
	}
	return nil
}

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

func scaffoldEnvFiles(projectRoot string, outputOnly bool) error {
	if err := scaffoldEnvFile(projectRoot, envPath, "", outputOnly); err != nil {
		return err
	}
	return scaffoldEnvFile(projectRoot, envrcPath, templates.Envrc, outputOnly)
}

func scaffoldEnvFile(projectRoot, relativePath, content string, outputOnly bool) error {
	path := filepath.Join(projectRoot, relativePath)
	if document.Exists(path) {
		if !outputOnly {
			fmt.Printf("  ✓ %s exists, skipping\n", relativePath)
		}
		return nil
	}

	if err := document.Write(path, content); err != nil {
		return fmt.Errorf("failed to create %s: %w", relativePath, err)
	}
	if !outputOnly {
		fmt.Printf("  ✓ Created %s\n", relativePath)
	}
	return nil
}

func scaffoldCodeRabbitConfig(projectRoot string, outputOnly bool) error {
	path := filepath.Join(projectRoot, codeRabbitConfigPath)
	if document.Exists(path) {
		if !outputOnly {
			fmt.Printf("  ✓ %s exists, skipping\n", codeRabbitConfigPath)
		}
		return nil
	}

	if err := document.Write(path, templates.CodeRabbitConfig); err != nil {
		return fmt.Errorf("failed to create %s: %w", codeRabbitConfigPath, err)
	}
	if !outputOnly {
		fmt.Printf("  ✓ Created %s\n", codeRabbitConfigPath)
	}
	return nil
}

func scaffoldPullRequestTemplate(projectRoot string, outputOnly bool) error {
	path := filepath.Join(projectRoot, pullRequestTemplatePath)
	if document.Exists(path) {
		if !outputOnly {
			fmt.Printf("  ✓ %s exists, skipping\n", pullRequestTemplatePath)
		}
		return nil
	}

	if err := document.Write(path, templates.PullRequestTemplate); err != nil {
		return fmt.Errorf("failed to create %s: %w", pullRequestTemplatePath, err)
	}
	if !outputOnly {
		fmt.Printf("  ✓ Created %s\n", pullRequestTemplatePath)
	}
	return nil
}
