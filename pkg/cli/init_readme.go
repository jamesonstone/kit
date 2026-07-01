package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

const (
	readmePath           = "README.md"
	readmeBadgeStart     = "<!-- BEGIN KIT-MANAGED README BADGES -->"
	readmeBadgeEnd       = "<!-- END KIT-MANAGED README BADGES -->"
	readmeStarterTagline = "Kit-managed project workspace"
)

func planRefreshReadmeFile(
	projectRoot string,
	cfg *config.Config,
	targets map[string]bool,
) (*initRefreshFileChange, error) {
	if !initRefreshTargetMatches(targets, readmePath) {
		return nil, nil
	}

	repository, err := readmeGitHubRepository(projectRoot, cfg)
	if err != nil {
		if len(targets) > 0 {
			return nil, err
		}
		return nil, nil
	}

	path := filepath.Join(projectRoot, readmePath)
	exists := document.Exists(path)
	before := ""
	if exists {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", readmePath, err)
		}
		before = string(data)
	}

	badgeBlock := managedReadmeBadgeBlock(repository, readmeCIWorkflow(projectRoot))
	after := upsertReadmeBadgeBlock(before, repository, badgeBlock)
	if !exists {
		return newInitRefreshFileChange(projectRoot, readmePath, before, after, instructionFileCreated), nil
	}
	if before == after {
		return newInitRefreshFileChange(projectRoot, readmePath, before, before, instructionFileSkipped), nil
	}
	result := instructionFileMerged
	if strings.Contains(before, readmeBadgeStart) {
		result = instructionFileUpdated
	}
	return newInitRefreshFileChange(projectRoot, readmePath, before, after, result), nil
}

func readmeGitHubRepository(projectRoot string, cfg *config.Config) (string, error) {
	if cfg != nil {
		if repository, ok := normalizeGitHubRepository(cfg.GitHub.Repository); ok {
			return repository, nil
		}
	}

	output, err := execCICommandRunner{}.Output(projectRoot, "git", "remote", "get-url", "origin")
	if err != nil {
		return "", fmt.Errorf("failed to resolve GitHub repository for README badges: %w", err)
	}
	owner, repo, err := parseGitHubRemoteURL(strings.TrimSpace(string(output)))
	if err != nil {
		return "", fmt.Errorf("failed to resolve GitHub repository for README badges: %w", err)
	}
	return owner + "/" + repo, nil
}

func normalizeGitHubRepository(raw string) (string, bool) {
	parts := strings.Split(strings.Trim(strings.TrimSpace(raw), "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", false
	}
	return parts[0] + "/" + strings.TrimSuffix(parts[1], ".git"), true
}

func managedReadmeBadgeBlock(repository, workflowPath string) string {
	badges := []string{
		readmeBadge("Last commit", "https://img.shields.io/github/last-commit/"+repository, "https://github.com/"+repository+"/commits"),
		readmeBadge("Open issues", "https://img.shields.io/github/issues/"+repository, "https://github.com/"+repository+"/issues"),
		readmeBadge("Pull requests", "https://img.shields.io/github/issues-pr/"+repository, "https://github.com/"+repository+"/pulls"),
	}
	if workflowPath != "" {
		workflowFile := filepath.Base(workflowPath)
		badges = append(badges, readmeBadge(
			"CI",
			"https://github.com/"+repository+"/actions/workflows/"+workflowFile+"/badge.svg",
			"https://github.com/"+repository+"/actions/workflows/"+workflowFile,
		))
	}
	badges = append(badges, readmeBadge(
		"Release",
		"https://img.shields.io/github/v/release/"+repository,
		"https://github.com/"+repository+"/releases",
	))

	return readmeBadgeStart + "\n" + strings.Join(badges, " ") + "\n" + readmeBadgeEnd + "\n"
}

func readmeBadge(label, imageURL, linkURL string) string {
	return fmt.Sprintf("[![%s](%s)](%s)", label, imageURL, linkURL)
}

func readmeCIWorkflow(projectRoot string) string {
	workflowDir := filepath.Join(projectRoot, ".github", "workflows")
	entries, err := os.ReadDir(workflowDir)
	if err != nil {
		return ""
	}

	var candidates []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := filepath.Ext(name)
		if ext != ".yml" && ext != ".yaml" {
			continue
		}
		base := strings.TrimSuffix(name, ext)
		if !readmeCIWorkflowName(base) {
			continue
		}
		candidates = append(candidates, filepath.ToSlash(filepath.Join(".github", "workflows", name)))
	}
	sort.Strings(candidates)
	if len(candidates) == 0 {
		return ""
	}
	return candidates[0]
}

func readmeCIWorkflowName(name string) bool {
	switch strings.ToLower(name) {
	case "ci", "test", "tests", "build":
		return true
	default:
		return false
	}
}

func upsertReadmeBadgeBlock(content, repository, badgeBlock string) string {
	if strings.TrimSpace(content) == "" {
		return newReadmeStarter(repository, badgeBlock)
	}

	if start := strings.Index(content, readmeBadgeStart); start >= 0 {
		end := strings.Index(content[start:], readmeBadgeEnd)
		if end >= 0 {
			end = start + end + len(readmeBadgeEnd)
			return joinReadmeParts(content[:start], badgeBlock, content[end:])
		}
	}

	insertAt := readmeBadgeInsertIndex(content)
	return joinReadmeParts(content[:insertAt], badgeBlock, content[insertAt:])
}

func readmeBadgeInsertIndex(content string) int {
	lines := strings.SplitAfter(content, "\n")
	line := firstReadmeContentLine(lines)
	if line >= len(lines) {
		return len(content)
	}

	trimmed := strings.TrimSpace(lines[line])
	if trimmed == "```text" {
		if closing := readmeFenceClosingLine(lines, line+1); closing >= 0 {
			return readmeAfterOpeningParagraph(lines, closing+1)
		}
	}

	if strings.HasPrefix(trimmed, "#") {
		return readmeAfterOpeningParagraph(lines, line+1)
	}
	return 0
}

func firstReadmeContentLine(lines []string) int {
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			return i
		}
	}
	return len(lines)
}

func readmeFenceClosingLine(lines []string, start int) int {
	for i := start; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "```" {
			return i
		}
	}
	return -1
}

func readmeAfterOpeningParagraph(lines []string, start int) int {
	i := start
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	if i >= len(lines) || strings.HasPrefix(strings.TrimSpace(lines[i]), "#") {
		return readmeOffset(lines, start)
	}
	for i < len(lines) && strings.TrimSpace(lines[i]) != "" {
		i++
	}
	return readmeOffset(lines, i)
}

func readmeOffset(lines []string, count int) int {
	offset := 0
	for i := 0; i < count && i < len(lines); i++ {
		offset += len(lines[i])
	}
	return offset
}

func joinReadmeParts(before, middle, after string) string {
	before = strings.TrimRight(before, "\n")
	middle = strings.TrimRight(middle, "\n")
	after = strings.TrimLeft(after, "\n")

	switch {
	case before == "" && after == "":
		return middle + "\n"
	case before == "":
		return middle + "\n\n" + after
	case after == "":
		return before + "\n\n" + middle + "\n"
	default:
		return before + "\n\n" + middle + "\n\n" + after
	}
}

func newReadmeStarter(repository, badgeBlock string) string {
	name := strings.TrimSpace(strings.TrimPrefix(repository, strings.Split(repository, "/")[0]+"/"))
	title := readmeTitle(name)
	wordmark := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(name, "-", " "), "_", " "))
	paragraph := fmt.Sprintf("%s is a Kit-managed project. Update this README with the repository purpose, boundaries, setup, and operating notes.", title)
	return fmt.Sprintf("```text\n%s\n\n                         %s\n```\n\n%s\n\n%s", wordmark, readmeStarterTagline, paragraph, badgeBlock)
}

func readmeTitle(name string) string {
	parts := strings.Fields(strings.ReplaceAll(strings.ReplaceAll(name, "-", " "), "_", " "))
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	if len(parts) == 0 {
		return "This project"
	}
	return strings.Join(parts, " ")
}
